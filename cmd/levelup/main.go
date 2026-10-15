package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	core_config "github.com/cdlinkin/LevelUp/internal/core/config"
	core_logger "github.com/cdlinkin/LevelUp/internal/core/logger"
	core_pgx_pool "github.com/cdlinkin/LevelUp/internal/core/repository/postgres/pool/pgx"
	core_http_middleware "github.com/cdlinkin/LevelUp/internal/core/transport/http/middleware"
	core_http_server "github.com/cdlinkin/LevelUp/internal/core/transport/http/server"
	statistics_postgres_repository "github.com/cdlinkin/LevelUp/internal/features/statistics/repository/postgres"
	statistics_service "github.com/cdlinkin/LevelUp/internal/features/statistics/service"
	statistics_transport_http "github.com/cdlinkin/LevelUp/internal/features/statistics/transport/http"
	tasks_postgres_repository "github.com/cdlinkin/LevelUp/internal/features/tasks/repository/postgres"
	tasks_service "github.com/cdlinkin/LevelUp/internal/features/tasks/service"
	tasks_transport_http "github.com/cdlinkin/LevelUp/internal/features/tasks/transport/http"
	users_postgres_repository "github.com/cdlinkin/LevelUp/internal/features/users/repository/postgres"
	users_service "github.com/cdlinkin/LevelUp/internal/features/users/service"
	users_transport_http "github.com/cdlinkin/LevelUp/internal/features/users/transport/http"
	"go.uber.org/zap"

	_ "github.com/cdlinkin/LevelUp/docs"
)

// @title 			LevelUp API
// @version 		1.0
// @description 	Todo Application REST-API scheme
// @host 			127.0.0.1:5050
// @BasePath 		/api/v1
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

	logger.Debug("initializing HTTP server")
	httpServer := core_http_server.NewHTTPServer(
		core_http_server.NewConfigMust(),
		logger,
		core_http_middleware.CORS(),
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(logger),
		core_http_middleware.Trace(),
		core_http_middleware.Panic(),
	)
	apiVersionRouterV1 := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion1, nil)
	apiVersionRouterV1.RegisterRoutes(userTransportHTTP.Routes()...)
	apiVersionRouterV1.RegisterRoutes(tasksTransportHTTP.Routes()...)
	apiVersionRouterV1.RegisterRoutes(statisticsTransportHTTP.Routes()...)

	httpServer.RegisterAPIRouters(apiVersionRouterV1)
	httpServer.RegisterSwagger()

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("HTTP server run error", zap.Error(err))
	}
}
