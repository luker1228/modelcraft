package user

import (
	"context"
	"fmt"
	"modelcraft/internal/domain/membership"
	"modelcraft/internal/interfaces/http/generated"
	"modelcraft/pkg/logfacade"
)

// Handler handles user-related HTTP requests
type Handler struct {
	membershipRepo membership.MembershipRepository
	logger         logfacade.Logger
}

// NewHandler creates a new user handler
func NewHandler(
	membershipRepo membership.MembershipRepository,
	logger logfacade.Logger,
) *Handler {
	return &Handler{
		membershipRepo: membershipRepo,
		logger:         logger,
	}
}

// GetUserMemberships returns the current user's organization memberships
// GET /api/user/memberships
func (h *Handler) GetUserMemberships(
	ctx context.Context,
	userID string,
) (*generated.GetMembershipsResponse, error) {
	h.logger.Infof(ctx, "Getting memberships for user: %s", userID)

	// Query user memberships with details (organization and role info)
	memberships, err := h.membershipRepo.ListByUserWithDetails(ctx, userID, 100)
	if err != nil {
		h.logger.Errorf(ctx, "Failed to query user memberships: user_id=%s, error=%v", userID, err)
		return nil, fmt.Errorf("failed to query memberships: %w", err)
	}

	// Convert to API response format
	membershipInfos := make([]generated.MembershipInfo, len(memberships))
	for i, m := range memberships {
		membershipInfos[i] = generated.MembershipInfo{
			OrgId:       m.OrgName,
			OrgName:     m.OrgName,
			DisplayName: m.DisplayName,
			Role:        m.RoleName,
			JoinedAt:    m.JoinedAt.UnixMilli(), // Convert to Unix milliseconds
		}
	}

	h.logger.Infof(ctx, "Found %d memberships for user: %s", len(memberships), userID)

	return &generated.GetMembershipsResponse{
		Memberships: membershipInfos,
	}, nil
}
