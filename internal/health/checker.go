package health

import (
	"context"
	"time"

	"mcp-agent/internal/repository"
	"mcp-agent/internal/service"
	"mcp-agent/pkg/logger"

	"go.uber.org/zap"
)

type Checker struct {
	healthSvc *service.HealthService
	statsRepo *repository.StatsRepository
	interval  time.Duration
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewChecker(healthSvc *service.HealthService, statsRepo *repository.StatsRepository, interval time.Duration) *Checker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Checker{
		healthSvc: healthSvc,
		statsRepo: statsRepo,
		interval:  interval,
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (c *Checker) Start() {
	logger.Info("health checker started", zap.Duration("interval", c.interval))

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	// 立即执行一次
	c.check()

	for {
		select {
		case <-ticker.C:
			c.check()
		case <-c.ctx.Done():
			logger.Info("health checker stopped")
			return
		}
	}
}

func (c *Checker) Stop() {
	c.cancel()
}

func (c *Checker) check() {
	results := c.healthSvc.CheckAll()
	if results == nil {
		return
	}

	for _, result := range results {
		if err := c.statsRepo.UpdateHealthStatus(result.Name, result.Status); err != nil {
			logger.Error("failed to update health status",
				zap.String("tool", result.Name),
				zap.Error(err))
		}

		logger.Debug("health check result",
			zap.String("tool", result.Name),
			zap.String("status", result.Status),
			zap.Int64("latency_ms", result.Latency))
	}
}
