package auth

import (
	"context"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"
	"modelcraft/pkg/logfacade"
	"time"

	domainauth "modelcraft/internal/domain/auth"
)

// CreateAPIKeyCommand 创建 API Key 的命令
type CreateAPIKeyCommand struct {
	UserID    string
	Name      string
	RoleIDs   []int
	ExpiresAt *time.Time
}

// CreateAPIKeyResult 创建 API Key 的结果（包含明文 key，只返回一次）
type CreateAPIKeyResult struct {
	Key      *domainauth.APIKey
	PlainKey string // 明文，只在创建时返回一次
}

// UpdateAPIKeyCommand 更新 API Key 的命令
type UpdateAPIKeyCommand struct {
	ID        string
	UserID    string
	Name      string
	RoleIDs   []int
	ExpiresAt *time.Time
}

// APIKeyService 处理 API Key CRUD 用例
type APIKeyService struct {
	apiKeyRepo domainauth.APIKeyRepository
}

// NewAPIKeyService 创建新的 APIKeyService
func NewAPIKeyService(apiKeyRepo domainauth.APIKeyRepository) *APIKeyService {
	return &APIKeyService{
		apiKeyRepo: apiKeyRepo,
	}
}

// ListAPIKeys 返回当前用户所有有效 key
func (s *APIKeyService) ListAPIKeys(ctx context.Context, userID string) ([]*domainauth.APIKey, error) {
	keys, err := s.apiKeyRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	return keys, nil
}

// CreateAPIKey 创建新 key，返回包含明文 key 的结果（明文只返回一次）
func (s *APIKeyService) CreateAPIKey(ctx context.Context, cmd CreateAPIKeyCommand) (*CreateAPIKeyResult, error) {
	// 检查限额（≤20）
	count, err := s.apiKeyRepo.CountActiveByUserID(ctx, cmd.UserID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if count >= domainauth.APIKeyMaxPerUser {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.APIKeyLimitExceeded, domainauth.APIKeyMaxPerUser)
	}

	// 生成 key
	plaintext, hash, err := GenerateAPIKey()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate api key")
	}

	keyID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate key id")
	}

	key := &domainauth.APIKey{
		ID:        keyID,
		UserID:    cmd.UserID,
		Name:      cmd.Name,
		KeyHash:   hash,
		KeyPrefix: plaintext[:10], // 前10位，如 mc_a1b2c3
		RoleIDs:   cmd.RoleIDs,
		ExpiresAt: cmd.ExpiresAt,
		CreatedAt: time.Now(),
	}

	if err := s.apiKeyRepo.Save(ctx, key); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	logfacade.GetLogger(ctx).Infof(ctx, "API key created: user_id=%s, key_id=%s", cmd.UserID, key.ID)
	return &CreateAPIKeyResult{
		Key:      key,
		PlainKey: plaintext,
	}, nil
}

// RevokeAPIKey 吊销指定 key（幂等）
func (s *APIKeyService) RevokeAPIKey(ctx context.Context, id, userID string) (*domainauth.APIKey, error) {
	key, err := s.apiKeyRepo.FindByID(ctx, id)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if key == nil || key.UserID != userID {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.APIKeyNotFound, id)
	}

	if err := s.apiKeyRepo.Revoke(ctx, id, userID); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	// Fetch updated key
	updated, err := s.apiKeyRepo.FindByID(ctx, id)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	return updated, nil
}

// UpdateAPIKey 更新 key 名称或过期时间
func (s *APIKeyService) UpdateAPIKey(ctx context.Context, cmd UpdateAPIKeyCommand) (*domainauth.APIKey, error) {
	key, err := s.apiKeyRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if key == nil || key.UserID != cmd.UserID {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.APIKeyNotFound, cmd.ID)
	}

	if err := s.apiKeyRepo.Update(ctx, cmd.ID, cmd.UserID, cmd.Name, cmd.RoleIDs, cmd.ExpiresAt); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	updated, err := s.apiKeyRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	return updated, nil
}

// VerifyAPIKey 验证 API Key 并返回 userID（实现 middleware.APIKeyVerifier 接口）
func (s *APIKeyService) VerifyAPIKey(ctx context.Context, rawKey string) (string, error) {
	hash := HashToken(rawKey)
	key, err := s.apiKeyRepo.FindByHash(ctx, hash)
	if err != nil {
		return "", bizerrors.ConvertRepositoryError(ctx, err)
	}
	if key == nil || !key.IsValid() {
		return "", nil
	}

	// 异步防抖更新 last_used_at（不阻塞请求）
	keyID := key.ID
	lastUsedAt := key.LastUsedAt
	bizutils.GoWithCtx(ctx, func(ctx context.Context) {
		if lastUsedAt == nil || time.Since(*lastUsedAt) > time.Minute {
			_ = s.apiKeyRepo.UpdateLastUsed(ctx, keyID)
		}
	})

	return key.UserID, nil
}
