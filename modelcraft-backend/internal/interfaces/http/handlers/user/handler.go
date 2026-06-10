package user

import (
	"context"
	"fmt"
	"modelcraft/internal/interfaces/http/generated"
	"modelcraft/pkg/logfacade"

	domainUser "modelcraft/internal/domain/user"
)

// Handler handles user-related HTTP requests.
type Handler struct {
	userRepo domainUser.UserRepository
	logger   logfacade.Logger
}

// NewHandler creates a new user handler.
func NewHandler(
	userRepo domainUser.UserRepository,
	logger logfacade.Logger,
) *Handler {
	return &Handler{
		userRepo: userRepo,
		logger:   logger,
	}
}

// GetUserMemberships returns the current user's organization memberships.
// GET /api/user/memberships
func (h *Handler) GetUserMemberships(
	ctx context.Context,
	userID string,
) (*generated.GetMembershipsResponse, error) {
	h.logger.Infof(ctx, "Getting memberships for user: %s", userID)

	u, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		h.logger.Errorf(ctx, "Failed to query user: user_id=%s, error=%v", userID, err)
		return nil, fmt.Errorf("failed to query user: %w", err)
	}
	if u == nil {
		return &generated.GetMembershipsResponse{Memberships: []generated.MembershipInfo{}}, nil
	}

	membershipInfos := []generated.MembershipInfo{
		{
			OrgId:       u.OrgName,
			OrgName:     u.OrgName,
			DisplayName: "",
			Role:        "",
			JoinedAt:    u.CreatedAt.UnixMilli(),
		},
	}

	h.logger.Infof(ctx, "Found membership for user: %s, org: %s", userID, u.OrgName)

	return &generated.GetMembershipsResponse{
		Memberships: membershipInfos,
	}, nil
}
