package auth

import (
	"context"
	"modelcraft/pkg/bizutils"
	"modelcraft/pkg/logfacade"
	"time"

	domainauth "modelcraft/internal/domain/auth"
)

// CleanupService 负责定期清理过期的 token 和 key 记录
type CleanupService struct {
	refreshTokenRepo domainauth.RefreshTokenRepository
}

// NewCleanupService 创建新的 CleanupService
func NewCleanupService(
	refreshTokenRepo domainauth.RefreshTokenRepository,
) *CleanupService {
	return &CleanupService{
		refreshTokenRepo: refreshTokenRepo,
	}
}

// Start 启动后台清理 goroutine，每 24 小时执行一次
func (s *CleanupService) Start(ctx context.Context) {
	bizutils.GoWithCtx(ctx, func(ctx context.Context) {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.runCleanup(ctx)
			case <-ctx.Done():
				return
			}
		}
	})
}

func (s *CleanupService) runCleanup(ctx context.Context) {
	logger := logfacade.GetLogger(ctx)
	if err := s.refreshTokenRepo.DeleteExpired(ctx); err != nil {
		logger.Errorf(ctx, err, "cleanup refresh tokens failed")
	}
}
