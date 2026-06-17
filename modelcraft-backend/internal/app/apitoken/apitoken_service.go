package apitoken

import (
	"context"
	"fmt"
	"modelcraft/internal/app/auth"
	"modelcraft/internal/domain/shared"
	"modelcraft/pkg/bizutils"
	"time"

	domainenduser "modelcraft/internal/domain/enduser"
)

const maxTokensPerUser = 20

// APITokenService handles CRUD and validation for EndUser Personal Access Tokens.
type APITokenService struct {
	repo domainenduser.APITokenRepository
}

// NewAPITokenService creates a new APITokenService.
func NewAPITokenService(repo domainenduser.APITokenRepository) *APITokenService {
	return &APITokenService{repo: repo}
}

// CreateAPITokenCommand holds the input for token creation.
type CreateAPITokenCommand struct {
	OrgName   string
	EndUserID string
	Name      string
	ExpiresAt *time.Time
}

// CreateAPITokenResult holds the created token and the one-time plaintext.
type CreateAPITokenResult struct {
	Token     *domainenduser.APIToken
	Plaintext string
}

// CreateAPIToken mints a new PAT and stores its hash.
// The plaintext is only available in the returned result and must be shown to the user once.
func (s *APITokenService) CreateAPIToken(
	ctx context.Context, cmd CreateAPITokenCommand,
) (*CreateAPITokenResult, error) {
	existing, err := s.repo.ListByUser(ctx, cmd.OrgName, cmd.EndUserID)
	if err != nil {
		return nil, fmt.Errorf("check token count: %w", err)
	}
	if len(existing) >= maxTokensPerUser {
		return nil, fmt.Errorf("token limit reached: max %d tokens per user", maxTokensPerUser)
	}

	plaintext, hash, err := auth.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}
	fullPlaintext := "mc_pat_" + plaintext

	id, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, fmt.Errorf("generate token id: %w", err)
	}

	token := &domainenduser.APIToken{
		ID:          id,
		OrgName:     cmd.OrgName,
		EndUserID:   cmd.EndUserID,
		Name:        cmd.Name,
		TokenHash:   hash,
		ExpiresAt:   cmd.ExpiresAt,
		CreatedAt:   time.Now(),
		DeletedAt:   0,
		DeleteToken: 0,
	}
	if err := s.repo.Save(ctx, token); err != nil {
		return nil, fmt.Errorf("save token: %w", err)
	}

	return &CreateAPITokenResult{Token: token, Plaintext: fullPlaintext}, nil
}

// ListAPITokens returns all active tokens for the given end-user.
func (s *APITokenService) ListAPITokens(
	ctx context.Context, orgName, endUserID string,
) ([]*domainenduser.APIToken, error) {
	return s.repo.ListByUser(ctx, orgName, endUserID)
}

// RevokeAPIToken soft-deletes a token by ID, scoped to the user.
func (s *APITokenService) RevokeAPIToken(ctx context.Context, id, orgName, endUserID string) error {
	return s.repo.SoftDelete(ctx, id, orgName, endUserID)
}

// ValidateToken authenticates a Bearer mc_pat_xxx token and returns the domain entity.
func (s *APITokenService) ValidateToken(
	ctx context.Context, plaintext string,
) (*domainenduser.APIToken, error) {
	raw := plaintext[len("mc_pat_"):]
	hash := auth.HashToken(raw)
	token, err := s.repo.FindByHash(ctx, hash)
	if err != nil {
		if shared.IsNotFoundError(err) {
			return nil, fmt.Errorf("token not found")
		}
		return nil, fmt.Errorf("find token: %w", err)
	}
	if !token.IsValid() {
		return nil, fmt.Errorf("token expired or revoked")
	}
	return token, nil
}

// UpdateLastUsedAt records the most recent use timestamp for a token.
func (s *APITokenService) UpdateLastUsedAt(ctx context.Context, id string, at time.Time) error {
	return s.repo.UpdateLastUsed(ctx, id, at)
}
