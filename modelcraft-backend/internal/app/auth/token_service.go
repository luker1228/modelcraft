package auth

import (
	"context"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"
	"modelcraft/pkg/logfacade"
	"time"

	domainauth "modelcraft/internal/domain/auth"
	domainUser "modelcraft/internal/domain/user"
)

// TokenService 处理认证令牌操作：登录、刷新、登出。
// 使用有状态的 DB 存储 Refresh Token（opaque token），支持轮换和盗用检测。
type TokenService struct {
	refreshTokenRepo domainauth.RefreshTokenRepository
	userRepo         domainUser.UserRepository
	auditLogRepo     domainauth.SecurityAuditLogRepository
	refreshTTL       time.Duration
}

// NewTokenService 创建新的 TokenService。
func NewTokenService(
	refreshTokenRepo domainauth.RefreshTokenRepository,
	userRepo domainUser.UserRepository,
	auditLogRepo domainauth.SecurityAuditLogRepository,
	refreshTTL time.Duration,
) *TokenService {
	if refreshTTL == 0 {
		refreshTTL = 7 * 24 * time.Hour
	}
	return &TokenService{
		refreshTokenRepo: refreshTokenRepo,
		userRepo:         userRepo,
		auditLogRepo:     auditLogRepo,
		refreshTTL:       refreshTTL,
	}
}

// Login 查找或创建用户，生成 Refresh Token 存入 DB，返回明文给 BFF。
func (s *TokenService) Login(ctx context.Context, cmd LoginCommand) (*LoginResult, error) {
	logger := logfacade.GetLogger(ctx)

	// 查找用户，不存在则创建
	u, err := s.userRepo.GetByExternalID(ctx, cmd.ExternalID)
	if err != nil {
		// 非 NotFound 的系统错误
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if u == nil {
		// 创建新用户
		id, err := bizutils.GenerateUUIDV7()
		if err != nil {
			return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate user id")
		}
		u, err = domainUser.NewUser(id, cmd.ExternalID, cmd.Name, "")
		if err != nil {
			return nil, bizerrors.WrapError(err, bizerrors.SystemError, "create user entity")
		}
		if err := s.userRepo.Create(ctx, u); err != nil {
			return nil, bizerrors.ConvertRepositoryError(ctx, err)
		}
		logger.Infof(ctx, "Created new user: id=%s, external_id=%s", u.ID, cmd.ExternalID)
	}

	// 生成 opaque refresh token
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

	logger.Infof(ctx, "Login success: user_id=%s", u.ID)
	return &LoginResult{
		UserID:       u.ID,
		RefreshToken: plaintext,
		ExpiresAt:    expiresAt,
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
	return &RefreshResult{
		UserID:       token.UserID,
		RefreshToken: plaintext,
		ExpiresAt:    expiresAt,
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
