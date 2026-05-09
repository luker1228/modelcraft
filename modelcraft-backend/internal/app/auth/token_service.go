package auth

import (
	"context"
	"modelcraft/internal/app/organization"
	"modelcraft/internal/domain/membership"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/repository"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"
	"modelcraft/pkg/logfacade"
	"time"

	"github.com/golang-jwt/jwt/v5"

	domainauth "modelcraft/internal/domain/auth"

	domainProfile "modelcraft/internal/domain/profile"

	domainUser "modelcraft/internal/domain/user"
)

// MembershipOrgProvider 是 TokenService 用于获取用户主 Org 信息的最小接口。
// 仅在签发 Token 时需要，不依赖完整的 MembershipRepository。
type MembershipOrgProvider interface {
	ListByUserWithDetails(ctx context.Context, userID string, limit int) ([]*membership.MembershipWithDetails, error)
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

// TokenService 处理认证令牌操作：注册、登录、刷新、登出。
// 使用有状态的 DB 存储 Refresh Token（opaque token），支持轮换和盗用检测。
type TokenService struct {
	refreshTokenRepo domainauth.RefreshTokenRepository
	userRepo         domainUser.UserRepository
	profileRepo      domainProfile.Repository
	membershipRepo   MembershipOrgProvider
	auditLogRepo     domainauth.SecurityAuditLogRepository
	passwordHasher   domainauth.PasswordHasher
	refreshTTL       time.Duration
	createOrgService OrgCreationService
	txManager        repository.TxManager
	jwtSigner        *domainauth.JWTSigner
}

// NewTokenService 创建新的 TokenService。
func NewTokenService(
	refreshTokenRepo domainauth.RefreshTokenRepository,
	userRepo domainUser.UserRepository,
	profileRepo domainProfile.Repository,
	auditLogRepo domainauth.SecurityAuditLogRepository,
	passwordHasher domainauth.PasswordHasher,
	refreshTTL time.Duration,
	createOrgService OrgCreationService,
	membershipRepo MembershipOrgProvider,
	txManager repository.TxManager,
	jwtSigner *domainauth.JWTSigner,
) *TokenService {
	if refreshTTL == 0 {
		refreshTTL = 7 * 24 * time.Hour
	}
	return &TokenService{
		refreshTokenRepo: refreshTokenRepo,
		userRepo:         userRepo,
		profileRepo:      profileRepo,
		membershipRepo:   membershipRepo,
		auditLogRepo:     auditLogRepo,
		passwordHasher:   passwordHasher,
		refreshTTL:       refreshTTL,
		createOrgService: createOrgService,
		txManager:        txManager,
		jwtSigner:        jwtSigner,
	}
}

// Register 手机号+密码注册新用户。
// 注册成功后会同事务初始化 profile，并创建用户的个人组织。
func (s *TokenService) Register(ctx context.Context, cmd RegisterCommand) (*RegisterResult, error) {
	logger := logfacade.GetLogger(ctx)

	// 1. 校验 userName（格式、保留字）
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

	// 4. 检查 userName 是否已被占用
	nameExists, err := s.userRepo.ExistsByName(ctx, cmd.UserName)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if nameExists {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.UserAlreadyExists, cmd.UserName)
	}

	// 5. 检查手机号是否已注册
	phoneExists, err := s.userRepo.ExistsByPhone(ctx, cmd.Phone)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if phoneExists {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.UserAlreadyExists, phone.Masked())
	}

	// 6. 哈希密码
	hashedPassword, err := s.passwordHasher.Hash(ctx, cmd.Password)
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "hash password")
	}

	// 7. 生成用户 ID
	id, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate user id")
	}

	// 8. 创建用户实体（使用用户提供的 userName）
	u, err := domainUser.NewUser(id, cmd.UserName, phone, hashedPassword)
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "create user entity")
	}

	// 9. 创建初始 profile
	profileID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate profile id")
	}
	defaultAvatarURL := "mock://avatar/default-1.png"
	initialProfile, err := domainProfile.NewInitialProfile(profileID, u.ID, "", &defaultAvatarURL, nil)
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "create profile entity")
	}

	// 10. 同事务持久化 user + profile
	if err := s.createUserAndProfile(ctx, u, initialProfile, phone.Masked()); err != nil {
		return nil, err
	}

	logger.Infof(ctx, "User registered with profile: id=%s, userName=%s, phone=%s, profileID=%s",
		u.ID, cmd.UserName, phone.Masked(), initialProfile.ID)

	result := &RegisterResult{
		UserID: u.ID,
		Profile: RegisterProfileSnapshot{
			ID:        initialProfile.ID,
			UserID:    initialProfile.UserID,
			Nickname:  initialProfile.Nickname,
			AvatarURL: initialProfile.AvatarURL,
			Bio:       initialProfile.Bio,
		},
	}

	// 11. 创建个人组织并返回 orgName
	if s.createOrgService == nil {
		// 兜底：测试环境可不注入组织服务，仍返回稳定 slug
		result.OrgName = bizutils.GenerateSlugWithLength(cmd.UserName, 6, 24)
		return result, nil
	}

	orgOutput, err := s.createOrgService.Execute(ctx, &organization.CreateOrganizationInput{
		DisplayName:          cmd.UserName,
		OrganizationName:     "", // 由组织服务根据 displayName 生成 slug
		OwnerUserID:          u.ID,
		EndUserAdminPassword: cmd.Password, // builtin admin 密码与注册密码保持一致
	})
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "create personal organization")
	}

	result.OrgName = orgOutput.OrganizationName
	logger.Infof(ctx, "Personal organization created: orgName=%s for user=%s", result.OrgName, u.ID)

	return result, nil
}

func (s *TokenService) createUserAndProfile(
	ctx context.Context,
	u *domainUser.User,
	p *domainProfile.Profile,
	maskedPhone string,
) error {
	persist := func(
		ctx context.Context,
		userRepo domainUser.UserRepository,
		profileRepo domainProfile.Repository,
	) error {
		if err := userRepo.Create(ctx, u); err != nil {
			if shared.IsDuplicateKeyError(err) {
				return bizerrors.NewErrorFromContext(ctx, bizerrors.UserAlreadyExists, maskedPhone)
			}
			return bizerrors.ConvertRepositoryError(ctx, err)
		}

		if err := profileRepo.CreateInitialProfile(ctx, p); err != nil {
			return bizerrors.ConvertRepositoryError(ctx, err)
		}
		return nil
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

// Login 登录（支持手机号或用户名），生成 Refresh Token 存入 DB，返回明文给 BFF。
func (s *TokenService) Login(ctx context.Context, cmd LoginCommand) (*LoginResult, error) {
	logger := logfacade.GetLogger(ctx)

	// Resolve the effective identifier and type (backward compat: use Phone if Identifier is empty)
	identifier := cmd.Identifier
	idType := cmd.IdentifierType
	if identifier == "" && cmd.Phone != "" {
		identifier = cmd.Phone
		idType = IdentifierTypePhone
	}
	// Default to PHONE if not specified
	if idType == "" {
		idType = IdentifierTypePhone
	}

	var u *domainUser.User
	var err error

	switch idType {
	case IdentifierTypePhone:
		// Validate phone format
		if _, validateErr := domainUser.NewPhoneNumber(identifier); validateErr != nil {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.AuthParamInvalid, validateErr.Error())
		}
		// Find user by phone
		u, err = s.userRepo.GetByPhone(ctx, identifier)
		if err != nil {
			if shared.IsNotFoundError(err) {
				return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.AuthenticationFailed, "phone number not found")
			}
			return nil, bizerrors.ConvertRepositoryError(ctx, err)
		}

	case IdentifierTypeUsername:
		// Find user by username (name field)
		u, err = s.userRepo.GetByName(ctx, identifier)
		if err != nil {
			if shared.IsNotFoundError(err) {
				return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.AuthenticationFailed, "username not found")
			}
			return nil, bizerrors.ConvertRepositoryError(ctx, err)
		}

	default:
		return nil, bizerrors.NewErrorFromContext(
			ctx,
			bizerrors.AuthParamInvalid,
			"invalid identifier type: "+string(idType),
		)
	}

	// Verify password
	if err := s.passwordHasher.Verify(ctx, cmd.Password, u.PasswordHash); err != nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.AuthenticationFailed, "incorrect password")
	}

	// Generate opaque refresh token
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

	// Fetch user's primary organization (first one)
	var orgName string
	if s.membershipRepo != nil {
		memberships, listErr := s.membershipRepo.ListByUserWithDetails(ctx, u.ID, 1)
		if listErr == nil && len(memberships) > 0 {
			orgName = memberships[0].OrgName
		}
	}

	logger.Infof(ctx, "Login success: user_id=%s, identifier_type=%s", u.ID, idType)

	accessToken, err := s.jwtSigner.IssueAccessToken(u.ID, orgName, jwt.ClaimStrings{domainauth.AudienceTenant}, nil)
	if err != nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.SystemError, "failed to issue access token")
	}

	return &LoginResult{
		UserID:       u.ID,
		UserName:     u.Name,
		OrgName:      orgName,
		AccessToken:  accessToken,
		RefreshToken: plaintext,
		ExpiresIn:    s.jwtSigner.TTLSeconds(),
	}, nil
}

// OAuthLogin 通过外部认证提供者（OAuth）登录。
// Deprecated: 保留向后兼容，新流程使用 Login。
func (s *TokenService) OAuthLogin(ctx context.Context, cmd OAuthLoginCommand) (*LoginResult, error) {
	logger := logfacade.GetLogger(ctx)

	// 查找用户，不存在则创建
	u, err := s.userRepo.GetByExternalID(ctx, cmd.ExternalID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if u == nil {
		id, err := bizutils.GenerateUUIDV7()
		if err != nil {
			return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate user id")
		}
		u, err = domainUser.NewOAuthUser(id, cmd.ExternalID, cmd.Name, "")
		if err != nil {
			return nil, bizerrors.WrapError(err, bizerrors.SystemError, "create user entity")
		}
		if err := s.userRepo.Create(ctx, u); err != nil {
			return nil, bizerrors.ConvertRepositoryError(ctx, err)
		}
		logger.Infof(ctx, "Created new OAuth user: id=%s, external_id=%s", u.ID, cmd.ExternalID)
	}

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

	logger.Infof(ctx, "OAuth login success: user_id=%s", u.ID)
	return &LoginResult{
		UserID:       u.ID,
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

	// 查询用户的主 org（与 Login 路径一致，取第一个 membership 的 OrgName）
	var refreshOrgName string
	if s.membershipRepo != nil {
		memberships, listErr := s.membershipRepo.ListByUserWithDetails(ctx, token.UserID, 1)
		if listErr == nil && len(memberships) > 0 {
			refreshOrgName = memberships[0].OrgName
		}
	}

	accessToken, err := s.jwtSigner.IssueAccessToken(
		token.UserID, refreshOrgName, jwt.ClaimStrings{domainauth.AudienceTenant}, nil,
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

// mustGenerateID 用于审计日志等非关键路径的 ID 生成，忽略错误。
func mustGenerateID() string {
	id, _ := bizutils.GenerateUUIDV7()
	return id
}
