package auth

import (
	"context"
	"database/sql"
	"fmt"
	"modelcraft/internal/app/organization"
	"modelcraft/internal/domain/enduser"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/repository"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"
	"modelcraft/pkg/logfacade"
	"os"
	"time"

	domainauth "modelcraft/internal/domain/auth"
	domainOrg "modelcraft/internal/domain/organization"
	domainPermission "modelcraft/internal/domain/permission"
	domainProfile "modelcraft/internal/domain/profile"
	domainUser "modelcraft/internal/domain/user"
)

// EndUserRepositoryFactory creates end-user repositories from a DB connection.
// Abstracted as an interface to avoid a direct import cycle between auth and enduser packages.
type EndUserRepositoryFactory interface {
	// NewEndUserRepository creates an EndUserRepository scoped to an org.
	NewEndUserRepository(db SQLDBTX, orgName string) enduser.EndUserRepository
}

// SQLDBTX is the minimal database interface accepted by repository factories.
// Both *sql.DB and *sql.Tx satisfy this interface.
type SQLDBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// OrgCreationService 是 TokenService 依赖的组织创建最小接口。
// 注册时用于自动为新用户创建个人组织（含 builtin admin EndUser）。
// 抽象为接口以便单测注入 spy。
type OrgCreationService interface {
	Execute(
		ctx context.Context,
		input *organization.CreateOrganizationInput,
	) (*organization.CreateOrganizationOutput, error)
}

type TxOrgCreationService interface {
	ExecuteWithQuerier(
		ctx context.Context,
		q dbgen.Querier,
		input *organization.CreateOrganizationInput,
	) (*organization.CreateOrganizationOutput, error)
}

// TokenService 处理认证令牌操作：注册、登录、刷新、登出。
// 使用有状态的 DB 存储 Refresh Token（opaque token），支持轮换和盗用检测。
// 同时处理 EndUser（终端用户）的 login/refresh/logout/me，统一两套认证路径。
type TokenService struct {
	refreshTokenRepo   domainauth.RefreshTokenRepository
	userRepo           domainUser.UserRepository
	orgRepo            domainOrg.OrganizationRepository
	profileRepo        domainProfile.Repository
	auditLogRepo       domainauth.SecurityAuditLogRepository
	passwordHasher     domainauth.PasswordHasher
	refreshTTL         time.Duration
	createOrgService   OrgCreationService
	txManager          repository.TxManager
	jwtSigner          *domainauth.JWTSigner
	endUserRepoFactory EndUserRepositoryFactory
	systemDB           *sql.DB
}

// NewTokenService 创建新的 TokenService。
func NewTokenService(
	refreshTokenRepo domainauth.RefreshTokenRepository,
	userRepo domainUser.UserRepository,
	orgRepo domainOrg.OrganizationRepository,
	profileRepo domainProfile.Repository,
	auditLogRepo domainauth.SecurityAuditLogRepository,
	passwordHasher domainauth.PasswordHasher,
	refreshTTL time.Duration,
	createOrgService OrgCreationService,
	txManager repository.TxManager,
	jwtSigner *domainauth.JWTSigner,
) *TokenService {
	if refreshTTL == 0 {
		refreshTTL = 7 * 24 * time.Hour
	}
	return &TokenService{
		refreshTokenRepo: refreshTokenRepo,
		userRepo:         userRepo,
		orgRepo:          orgRepo,
		profileRepo:      profileRepo,
		auditLogRepo:     auditLogRepo,
		passwordHasher:   passwordHasher,
		refreshTTL:       refreshTTL,
		createOrgService: createOrgService,
		txManager:        txManager,
		jwtSigner:        jwtSigner,
	}
}

// WithEndUserSupport attaches the end-user repository factory and system DB so that
// LoginEndUser / RefreshEndUserToken / GetEndUserMe can be called on this service.
func (s *TokenService) WithEndUserSupport(factory EndUserRepositoryFactory, systemDB *sql.DB) *TokenService {
	s.endUserRepoFactory = factory
	s.systemDB = systemDB
	return s
}

// Register 手机号+密码注册新用户。
// 注册成功后会同事务初始化 profile，并创建用户的个人组织。
// 流程：先建 Org（含 phone 全局唯一校验）→ 再建 User（绑定 orgName）→ 建 Profile。
func (s *TokenService) Register(ctx context.Context, cmd RegisterCommand) (*RegisterResult, error) {
	if err := domainUser.ValidateUserName(cmd.UserName); err != nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.AuthParamInvalid, err.Error())
	}

	// 2. 校验手机号格式
	phone, err := domainUser.NewPhoneNumber(cmd.Phone)
	if err != nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.AuthParamInvalid, err.Error())
	}

	// 3. 校验密码强度
	if err := domainauth.ValidatePasswordStrength(cmd.Password); err != nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.AuthParamInvalid, err.Error())
	}

	// 4. 检查手机号是否已注册 Org（全局唯一）
	if s.orgRepo != nil {
		phoneExists, err := s.orgRepo.ExistsByPhone(ctx, cmd.Phone)
		if err != nil {
			return nil, bizerrors.ConvertRepositoryError(ctx, err)
		}
		if phoneExists {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.PhoneAlreadyExists, phone.Masked())
		}
	}

	// 5. 哈希密码
	hashedPassword, err := s.passwordHasher.Hash(ctx, cmd.Password)
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "hash password")
	}

	// 6. 生成用户 ID
	userID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate user id")
	}

	// 7. 生成 profile ID
	profileID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate profile id")
	}

	orgDisplayName := cmd.OrgDisplayName
	if orgDisplayName == "" {
		orgDisplayName = cmd.UserName // fallback
	}

	result := &RegisterResult{UserID: userID}

	orgInput := &organization.CreateOrganizationInput{
		DisplayName:      orgDisplayName,
		OrganizationName: cmd.OrganizationName,
		OwnerUserID:      userID,
		Phone:            cmd.Phone,
		IsNewUser:        true, // user not yet persisted at this point; skip DB lookup in ExecuteWithQuerier
	}

	if s.txManager != nil {
		if txOrgService, ok := s.createOrgService.(TxOrgCreationService); ok {
			return s.registerWithTxOrgService( //nolint:wrapcheck
				ctx, cmd, userID, profileID, hashedPassword, phone, result, orgInput, txOrgService,
			)
		}
	}

	return s.registerFallback(ctx, cmd, userID, profileID, hashedPassword, phone, result, orgInput)
}

// registerFallback handles registration without a tx-aware org service. //nolint:cyclop
func (s *TokenService) registerFallback(
	ctx context.Context,
	cmd RegisterCommand,
	userID, profileID, hashedPassword string,
	phone domainUser.PhoneNumber,
	result *RegisterResult,
	orgInput *organization.CreateOrganizationInput,
) (*RegisterResult, error) {
	logger := logfacade.GetLogger(ctx)

	// Fallback path (no tx support): create org first, then user
	orgName, err := s.resolveOrgNameForFallback(ctx, cmd, userID, orgInput)
	if err != nil {
		return nil, err
	}

	u, err := domainUser.NewUser(userID, cmd.UserName, phone, hashedPassword, orgName)
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "create user entity")
	}

	defaultAvatarURL := "mock://avatar/default-1.png"
	initialProfile, err := domainProfile.NewInitialProfile(profileID, u.ID, "", &defaultAvatarURL, nil)
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "create profile entity")
	}

	if err := s.createUserAndProfile(ctx, u, initialProfile, phone.Masked()); err != nil {
		return nil, err
	}
	logger.Infof(ctx, "User registered: id=%s, userName=%s, phone=%s, orgName=%s",
		u.ID, cmd.UserName, phone.Masked(), orgName)

	result.OrgName = orgName
	result.Profile = RegisterProfileSnapshot{
		ID:        initialProfile.ID,
		UserID:    initialProfile.UserID,
		Nickname:  initialProfile.Nickname,
		AvatarURL: initialProfile.AvatarURL,
		Bio:       initialProfile.Bio,
	}
	return result, nil
}

func (s *TokenService) registerWithTxOrgService(
	ctx context.Context,
	cmd RegisterCommand,
	userID, profileID, hashedPassword string,
	phone domainUser.PhoneNumber,
	result *RegisterResult,
	orgInput *organization.CreateOrganizationInput,
	txOrgService TxOrgCreationService,
) (*RegisterResult, error) {
	logger := logfacade.GetLogger(ctx)
	err := s.txManager.WithTx(ctx, func(ctx context.Context, q dbgen.Querier) error {
		// Step 1: Create Org first to get orgName
		var orgName string
			var ownerRoleID int
		if s.createOrgService != nil {
			orgOutput, txErr := txOrgService.ExecuteWithQuerier(ctx, q, orgInput)
			if txErr != nil {
				return bizerrors.WrapError(
					txErr, bizerrors.SystemError,
					fmt.Sprintf("create personal organization: %v", txErr),
				)
			}
			orgName = orgOutput.OrganizationName
				ownerRoleID = int(orgOutput.RoleID)
		} else {
			orgName = bizutils.GenerateSlugWithLength(cmd.UserName, 6, 24)
		}
		result.OrgName = orgName

		// Step 2: Create User entity (now orgName is known)
		u, txErr := domainUser.NewUser(userID, cmd.UserName, phone, hashedPassword, orgName)
		if txErr != nil {
			return bizerrors.WrapError(txErr, bizerrors.SystemError, "create user entity")
		}

		// Step 3: Create Profile entity
		defaultAvatarURL := "mock://avatar/default-1.png"
		initialProfile, txErr := domainProfile.NewInitialProfile(profileID, u.ID, "", &defaultAvatarURL, nil)
		if txErr != nil {
			return bizerrors.WrapError(txErr, bizerrors.SystemError, "create profile entity")
		}
		result.Profile = RegisterProfileSnapshot{
			ID:        initialProfile.ID,
			UserID:    initialProfile.UserID,
			Nickname:  initialProfile.Nickname,
			AvatarURL: initialProfile.AvatarURL,
			Bio:       initialProfile.Bio,
		}

		// Step 4: Persist user + profile
		userRepo := repository.NewSqlUserRepository(q)
		profileRepo := repository.NewSqlProfileRepository(q)
			if err := s.persistUserAndProfile(ctx, userRepo, profileRepo, u, initialProfile, phone.Masked()); err != nil {
				return err
			}

			// Step 5: Assign owner role (user must exist first for fk_user_roles_user FK)
			if ownerRoleID != 0 {
				userRoleRepo := repository.NewSqlCasbinUserRoleRepository(q)
				userRole := &domainPermission.UserRole{
					UserID: userID, RoleID: ownerRoleID, OrgName: orgName,
				}
				if err := userRoleRepo.AssignRole(ctx, userRole); err != nil {
					return bizerrors.WrapError(err, bizerrors.SystemError, "assign owner role")
				}
			}
			return nil
	})
	if err != nil {
		if _, ok := err.(*bizerrors.BusinessError); ok {
			return nil, err
		}
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "register transaction")
	}
	logger.Infof(ctx, "User registered: id=%s, userName=%s, phone=%s, orgName=%s",
		userID, cmd.UserName, phone.Masked(), result.OrgName)
	return result, nil
}

func (s *TokenService) createUserAndProfile(
	ctx context.Context,
	u *domainUser.User,
	p *domainProfile.Profile,
	maskedPhone string,
) error {
	persist := func(
		ctx context.Context, userRepo domainUser.UserRepository, profileRepo domainProfile.Repository,
	) error {
		return s.persistUserAndProfile(ctx, userRepo, profileRepo, u, p, maskedPhone)
	}

	if s.txManager == nil {
		if s.profileRepo == nil {
			return bizerrors.NewErrorFromContext(ctx, bizerrors.SystemError, "profile repository not configured")
		}
		return persist(ctx, s.userRepo, s.profileRepo)
	}

	err := s.txManager.WithTx(ctx, func(ctx context.Context, q dbgen.Querier) error {
		userRepo := repository.NewSqlUserRepository(q)
		profileRepo := repository.NewSqlProfileRepository(q)
		return persist(ctx, userRepo, profileRepo)
	})
	if err != nil {
		if _, ok := err.(*bizerrors.BusinessError); ok {
			return err
		}
		return bizerrors.WrapError(err, bizerrors.SystemError, "register transaction")
	}

	return nil
}

func (s *TokenService) persistUserAndProfile(
	ctx context.Context,
	userRepo domainUser.UserRepository,
	profileRepo domainProfile.Repository,
	u *domainUser.User,
	p *domainProfile.Profile,
	maskedPhone string,
) error {
	if err := userRepo.Create(ctx, u); err != nil {
		return s.classifyCreateUserError(ctx, userRepo, u, maskedPhone, err)
	}

	if err := profileRepo.CreateInitialProfile(ctx, p); err != nil {
		return bizerrors.ConvertRepositoryError(ctx, err)
	}
	return nil
}

// classifyCreateUserError maps a duplicate-key error on user creation to a specific business error.
func (s *TokenService) classifyCreateUserError(
	ctx context.Context,
	userRepo domainUser.UserRepository,
	u *domainUser.User,
	maskedPhone string,
	createErr error,
) error {
	if !shared.IsDuplicateKeyError(createErr) {
		return bizerrors.ConvertRepositoryError(ctx, createErr)
	}
	phoneExists, phoneErr := userRepo.ExistsByPhone(ctx, u.OrgName, u.Phone.String())
	if phoneErr == nil && phoneExists {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.PhoneAlreadyExists, maskedPhone)
	}
	nameExists, nameErr := userRepo.ExistsByName(ctx, u.OrgName, u.Name)
	if nameErr == nil && nameExists {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.UserNameAlreadyExists, u.Name)
	}
	return bizerrors.NewErrorFromContext(ctx, bizerrors.Conflict, "duplicate user record")
}

// Login 管理员用户名登录：全局按 userName 查找用户。
func (s *TokenService) Login(ctx context.Context, cmd LoginCommand) (*LoginResult, error) {
	logger := logfacade.GetLogger(ctx)

	if cmd.UserName == "" {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.AuthParamInvalid, "userName is required")
	}

	// Step 1: 全局按用户名查找用户
	u, err := s.userRepo.GetByNameGlobal(ctx, cmd.UserName)
	if err != nil {
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.AuthenticationFailed, "user not found")
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	// Step 2: 验证密码
	if u.PasswordHash == "" {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.AuthenticationFailed, "incorrect password")
	}
	if err := s.passwordHasher.Verify(ctx, cmd.Password, u.PasswordHash); err != nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.AuthenticationFailed, "incorrect password")
	}

	// Step 4: 签发 refresh token
	plaintext, hash, err := GenerateRefreshToken()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate refresh token")
	}
	tokenID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate token id")
	}
	expiresAt := time.Now().Add(s.refreshTTL)
	token := &domainauth.RefreshToken{
		ID:        tokenID,
		UserID:    u.ID,
		TokenHash: hash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	if err := s.refreshTokenRepo.Save(ctx, token); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	// Step 5: 签发 JWT（is_admin=true for Admin login）
	accessToken, err := s.jwtSigner.IssueAccessToken(u.ID, u.OrgName, true)
	if err != nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.SystemError, "failed to issue access token")
	}

	logger.Infof(ctx, "Admin login success: user_id=%s, org_name=%s", u.ID, u.OrgName)

	return &LoginResult{
		UserID:       u.ID,
		UserName:     u.Name,
		OrgName:      u.OrgName,
		AccessToken:  accessToken,
		RefreshToken: plaintext,
		ExpiresIn:    s.jwtSigner.TTLSeconds(),
	}, nil
}

// Refresh 验证旧 token → 盗用检测 → 轮换生成新 token。
func (s *TokenService) Refresh(ctx context.Context, cmd RefreshCommand) (*RefreshResult, error) {
	logger := logfacade.GetLogger(ctx)

	// 1. 计算 hash，查 DB
	hash := HashToken(cmd.RefreshToken)
	token, err := s.refreshTokenRepo.FindByHash(ctx, hash)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	// 2. token 不存在 → 401
	if token == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.AuthUnauthorized, "refresh token not found")
	}

	// 3. 已 revoked → 盗用检测
	if token.IsRevoked() {
		_ = s.refreshTokenRepo.RevokeAllByUserID(ctx, token.UserID)
		_ = s.auditLogRepo.Insert(ctx, &domainauth.SecurityAuditLog{
			ID:     mustGenerateID(),
			UserID: token.UserID,
			Event:  domainauth.EventReuseDetected,
			Detail: map[string]any{"token_id": token.ID},
		})
		logger.Warnf(ctx, "Token reuse detected: user_id=%s, token_id=%s", token.UserID, token.ID)
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.AuthUnauthorized, "token reuse detected")
	}

	// 4. 已过期 → 401
	if !token.IsValid() {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.AuthUnauthorized, "refresh token expired")
	}

	// 5. 正常轮换：revoke 旧 token，生成新 token
	if err := s.refreshTokenRepo.Revoke(ctx, token.ID); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	plaintext, newHash, err := GenerateRefreshToken()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate new refresh token")
	}

	tokenID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate token id")
	}

	expiresAt := time.Now().Add(s.refreshTTL)
	newToken := &domainauth.RefreshToken{
		ID:        tokenID,
		UserID:    token.UserID,
		TokenHash: newHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	if err := s.refreshTokenRepo.Save(ctx, newToken); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	logger.Infof(ctx, "Token refreshed: user_id=%s", token.UserID)

	// orgName/isAdmin 从 user_orgs 直接取（与 Login 路径一致）
	var refreshOrgName string
	var refreshIsAdmin bool
	if s.endUserRepoFactory != nil {
		userRepo := s.endUserRepoFactory.NewEndUserRepository(s.systemDB, "")
		if u, uErr := userRepo.GetByIDGlobal(ctx, token.UserID); uErr == nil && u != nil {
			refreshOrgName = u.OrgName
			refreshIsAdmin = u.IsAdmin
		}
	}
	// Fallback: get orgName from userRepo (used in tests and non-enduser flows)
	if refreshOrgName == "" {
		if u, uErr := s.userRepo.GetByID(ctx, token.UserID); uErr == nil && u != nil {
			refreshOrgName = u.OrgName
		}
	}

	accessToken, err := s.jwtSigner.IssueAccessToken(
		token.UserID, refreshOrgName, refreshIsAdmin,
	)
	if err != nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.SystemError, "failed to issue access token")
	}

	return &RefreshResult{
		AccessToken:  accessToken,
		RefreshToken: plaintext,
		ExpiresIn:    s.jwtSigner.TTLSeconds(),
	}, nil
}

// Logout 吊销指定的 refresh token。
func (s *TokenService) Logout(ctx context.Context, cmd LogoutCommand) error {
	hash := HashToken(cmd.RefreshToken)
	token, err := s.refreshTokenRepo.FindByHash(ctx, hash)
	if err != nil {
		return bizerrors.ConvertRepositoryError(ctx, err)
	}
	if token == nil {
		return nil // 已不存在的 token 无需吊销
	}
	return s.refreshTokenRepo.Revoke(ctx, token.ID)
}

// resolveOrgNameForFallback resolves or creates the org name when no tx-aware org service is available.
func (s *TokenService) resolveOrgNameForFallback(
	ctx context.Context,
	cmd RegisterCommand,
	userID string,
	orgInput *organization.CreateOrganizationInput,
) (string, error) {
	if s.createOrgService != nil {
		orgOutput, err := s.createOrgService.Execute(ctx, orgInput)
		if err != nil {
			return "", bizerrors.WrapError(err, bizerrors.SystemError, "create personal organization")
		}
		return orgOutput.OrganizationName, nil
	}
	orgName := bizutils.GenerateSlugWithLength(cmd.UserName, 6, 24)
	// When no org creation service (e.g. tests), create a minimal org in orgRepo
	// so the phone-based login lookup works.
	if s.orgRepo != nil {
		if minOrg, orgErr := domainOrg.NewOrganization(orgName, cmd.UserName, userID, cmd.Phone); orgErr == nil {
			_ = s.orgRepo.Create(ctx, minOrg)
		}
	}
	return orgName, nil
}

// mustGenerateID 用于审计日志等非关键路径的 ID 生成，忽略错误。
func mustGenerateID() string {
	id, _ := bizutils.GenerateUUIDV7()
	return id
}

// DemoLogin issues a guest JWT for anonymous demo access. No authentication required.
// Returns NotFound if DEMO_ENABLED is not set to "true".
func (s *TokenService) DemoLogin(ctx context.Context) (*LoginResult, error) {
	if os.Getenv("DEMO_ENABLED") != "true" {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.NotFound, "demo mode not enabled")
	}

	accessToken, err := s.jwtSigner.IssueAccessToken("guest", "demo", false)
	if err != nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.SystemError, "failed to issue guest token")
	}

	return &LoginResult{
		UserID:      "guest",
		UserName:    "guest",
		OrgName:     "demo",
		AccessToken: accessToken,
		ExpiresIn:   s.jwtSigner.TTLSeconds(),
	}, nil
}
