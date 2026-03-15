package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"mcp-agent/internal/config"
	"mcp-agent/internal/handler"
	"mcp-agent/internal/health"
	"mcp-agent/internal/middleware"
	"mcp-agent/internal/permission"
	"mcp-agent/internal/repository"
	"mcp-agent/internal/service"
	"mcp-agent/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	logger.Init(cfg.Log.Level)
	defer logger.Sync()

	// 确保数据目录存在
	if cfg.Database.Driver == "sqlite" {
		dir := filepath.Dir(cfg.Database.DSN)
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.Fatal("failed to create data directory", zap.Error(err))
		}
	}

	// 初始化数据库
	db, err := repository.NewDB(cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect database", zap.Error(err))
	}

	// Seed 默认管理员用户
	if err := repository.Seed(db); err != nil {
		logger.Error("failed to seed database", zap.Error(err))
	}

	// 初始化 Repositories
	userRepo := repository.NewUserRepository(db)
	toolRepo := repository.NewToolRepository(db)
	logRepo := repository.NewLogRepository(db)
	statsRepo := repository.NewStatsRepository(db)

	// 初始化 Services
	authSvc := service.NewAuthService(userRepo, cfg.JWT)
	toolSvc := service.NewToolService(toolRepo, logRepo, statsRepo)
	healthSvc := service.NewHealthService(toolRepo, cfg.MCPServers, time.Duration(cfg.HealthCheck.TimeoutSeconds)*time.Second)
	logSvc := service.NewLogService(logRepo)
	statsSvc := service.NewStatsService(statsRepo)

	// 初始化权限管理器
	permManager := permission.NewManager("configs/permissions.yaml")

	// 启动健康检查定时任务
	healthChecker := health.NewChecker(healthSvc, statsRepo, time.Duration(cfg.HealthCheck.IntervalSeconds)*time.Second)
	go healthChecker.Start()
	defer healthChecker.Stop()

	// 初始化 Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	toolHandler := handler.NewToolHandler(toolSvc)
	healthHandler := handler.NewHealthHandler(healthSvc)
	logHandler := handler.NewLogHandler(logSvc)
	statsHandler := handler.NewStatsHandler(statsSvc)

	// 配置 Gin
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	// 路由注册
	api := r.Group("/api/v1")
	{
		// 公开接口
		api.GET("/health", healthHandler.CheckAll)
		api.POST("/auth/login", authHandler.Login)

		// 需要认证的接口
		auth := api.Group("")
		auth.Use(middleware.Auth(cfg.JWT.Secret))
		{
			// Token 刷新
			auth.POST("/auth/refresh", authHandler.Refresh)
			auth.GET("/auth/profile", authHandler.Profile)

			// 工具列表（所有登录用户可查看）
			auth.GET("/tools", toolHandler.List)

			// 工具调用（需要权限检查）
			auth.POST("/tools/:name/call", middleware.PermissionCheck(permManager), toolHandler.Call)

			// 工具健康检查
			auth.GET("/tools/:name/health", healthHandler.CheckTool)

			// 工具统计信息
			auth.GET("/tools/:name/stats", statsHandler.GetToolStats)
			auth.GET("/stats", statsHandler.ListAllStats)

			// 管理员接口
			admin := auth.Group("")
			admin.Use(middleware.AdminOnly())
			{
				admin.POST("/tools", toolHandler.Create)
				admin.PUT("/tools/:name", toolHandler.Update)
				admin.DELETE("/tools/:name", toolHandler.Delete)
				admin.GET("/logs", logHandler.Query)
				admin.POST("/auth/users", authHandler.CreateUser)
			}
		}
	}

	// 启动服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("server starting", zap.String("addr", addr), zap.String("mode", cfg.Server.Mode))

	if err := r.Run(addr); err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}
}
