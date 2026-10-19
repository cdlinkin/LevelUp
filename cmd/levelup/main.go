package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	core_auth "github.com/cdlinkin/LevelUp/internal/core/auth"
	core_config "github.com/cdlinkin/LevelUp/internal/core/config"
	core_logger "github.com/cdlinkin/LevelUp/internal/core/logger"
	core_pgx_pool "github.com/cdlinkin/LevelUp/internal/core/repository/postgres/pool/pgx"
	core_http_middleware "github.com/cdlinkin/LevelUp/internal/core/transport/http/middleware"
	core_http_server "github.com/cdlinkin/LevelUp/internal/core/transport/http/server"
	auth_service "github.com/cdlinkin/LevelUp/internal/features/auth/service"
	auth_transport_http "github.com/cdlinkin/LevelUp/internal/features/auth/transport/http"
	statistics_postgres_repository "github.com/cdlinkin/LevelUp/internal/features/statistics/repository/postgres"
	statistics_service "github.com/cdlinkin/LevelUp/internal/features/statistics/service"
	statistics_transport_http "github.com/cdlinkin/LevelUp/internal/features/statistics/transport/http"
	tasks_postgres_repository "github.com/cdlinkin/LevelUp/internal/features/tasks/repository/postgres"
	tasks_service "github.com/cdlinkin/LevelUp/internal/features/tasks/service"
	tasks_transport_http "github.com/cdlinkin/LevelUp/internal/features/tasks/transport/http"
	users_postgres_repository "github.com/cdlinkin/LevelUp/internal/features/users/repository/postgres"
	users_service "github.com/cdlinkin/LevelUp/internal/features/users/service"
	users_transport_http "github.com/cdlinkin/LevelUp/internal/features/users/transport/http"
	web_fs_repository "github.com/cdlinkin/LevelUp/internal/features/web/repository/file_system"
	web_service "github.com/cdlinkin/LevelUp/internal/features/web/service"
	web_transport_http "github.com/cdlinkin/LevelUp/internal/features/web/transport/http"
	"go.uber.org/zap"

	_ "github.com/cdlinkin/LevelUp/docs"
)

// @title 			LevelUp API
// @version 		1.0
// @description 	Todo Application REST-API scheme
// @host 			127.0.0.1:5050
// @BasePath 		/api/v1
//
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description					JWT access token из POST /auth/login. Формат: `Bearer <access_token>`
func main() {
	cfg := core_config.NewConfigMust()
	time.Local = cfg.TimeZone

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	logger, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Println("failed to init logger:", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Debug("application time zone", zap.Any("zone", time.Local))

	logger.Debug("initiaziling postgres connection pool")
	pool, err := core_pgx_pool.NewPool(
		ctx,
		core_pgx_pool.NewConfigMust(),
	)
	if err != nil {
		logger.Fatal("failed to init postgres connection pool", zap.Error(err))
	}
	defer pool.Close()

	logger.Debug("initializing feature", zap.String("feature", "users"))
	usersRepository := users_postgres_repository.NewUsersRepository(pool)
	usersService := users_service.NewUsersService(usersRepository)
	userTransportHTTP := users_transport_http.NewUsersHTTPHandler(usersService)

	logger.Debug("initializing feature", zap.String("feature", "tasks"))
	tasksRepository := tasks_postgres_repository.NewTasksRepository(pool)
	tasksService := tasks_service.NewTasksService(tasksRepository)
	tasksTransportHTTP := tasks_transport_http.NewTasksHTTPHandler(tasksService)

	logger.Debug("initializing feature", zap.String("feature", "statistics"))
	statisticsRepository := statistics_postgres_repository.NewStatisticsRepository(pool)
	statisticsService := statistics_service.NewStatisticsService(statisticsRepository)
	statisticsTransportHTTP := statistics_transport_http.NewStatisticsHTTPHandler(statisticsService)

	logger.Debug("initializing feature", zap.String("feature", "web"))
	webRepository := web_fs_repository.NewWebRepository()
	webService := web_service.NewWebService(webRepository)
	webTransportHTTP := web_transport_http.NewWebHTTPHandler(webService)

	logger.Debug("initializing feature", zap.String("feature", "auth"))
	tokenManager := core_auth.NewTokenManager(core_auth.NewConfigMust())
	authMiddleware := core_http_middleware.Auth(tokenManager)
	authService := auth_service.NewAuthService(usersRepository, tokenManager)
	authTransportHTTP := auth_transport_http.NewAuthHTTPHandler(authService, authMiddleware)

	logger.Debug("initializing HTTP server")
	httpConfig := core_http_server.NewConfigMust()
	httpServer := core_http_server.NewHTTPServer(
		httpConfig,
		logger,
		core_http_middleware.CORS(httpConfig.AllowedOrigins),
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(logger),
		core_http_middleware.Trace(),
		core_http_middleware.Panic(),
	)
	apiVersionRouterV1 := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion1, nil)
	apiVersionRouterV1.RegisterRoutes(
		core_http_server.WithMiddleware(
			userTransportHTTP.Routes(),
			authMiddleware,
		)...,
	)
	apiVersionRouterV1.RegisterRoutes(
		core_http_server.WithMiddleware(
			tasksTransportHTTP.Routes(),
			authMiddleware,
		)...,
	)
	apiVersionRouterV1.RegisterRoutes(
		core_http_server.WithMiddleware(
			statisticsTransportHTTP.Routes(),
			authMiddleware,
		)...,
	)
	apiVersionRouterV1.RegisterRoutes(
		authTransportHTTP.Routes()...,
	)

	httpServer.RegisterAPIRouters(apiVersionRouterV1)
	httpServer.RegisterRoutes(webTransportHTTP.Routes()...)
	httpServer.RegisterSwagger()

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("HTTP server run error", zap.Error(err))
	}
}
