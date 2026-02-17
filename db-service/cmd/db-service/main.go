package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	"github.com/redis/go-redis/v9"

	appconfig "github.com/N0F1X3d/todo/db-service/internal/config"
	"github.com/N0F1X3d/todo/db-service/internal/repository"
	"github.com/N0F1X3d/todo/db-service/internal/server"
	"github.com/N0F1X3d/todo/db-service/internal/service"
	"github.com/N0F1X3d/todo/pkg/logger"

	pb "github.com/N0F1X3d/todo/pkg/proto"
)

func main() {
	// ========================
	// Config
	// ========================
	cfg, err := appconfig.Load(appconfig.GetConfigPath())
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	grpcAddr := cfg.GRPC.Address()
	dbDSN := cfg.DB.DSNWithTimeout()

	// ========================
	// Logger
	// ========================
	logg := logger.New(cfg.App.Name, "main-logs").WithComponent("main")

	// ========================
	// Database
	// ========================
	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	logg.Info("connected to postgres")

	// ========================
	// Redis (кеш задач)
	// ========================
	var redisClient *redis.Client
	if cfg.Redis.Enabled {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Address(),
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})

		redisCtx, redisCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer redisCancel()

		if err := redisClient.Ping(redisCtx).Err(); err != nil {
			logg.Warn("failed to connect to redis, caching disabled", "error", err)
			redisClient = nil
		} else {
			logg.Info("connected to redis",
				"addr", cfg.Redis.Address(),
				"db", cfg.Redis.DB,
				"ttl", cfg.Redis.TTL.String(),
			)
		}
	}

	// ========================
	// Migrations
	// ========================
	runMigrations(db)

	defer func() {
		if err := db.Close(); err != nil {
			logg.Error("failed to close db", err)
		}
	}()

	// ========================
	// Repository
	// ========================
	taskRepo := repository.NewTaskRepository(db, logg, redisClient, cfg.Redis.TTL)

	// ========================
	// Service
	// ========================
	taskService := service.NewTaskService(taskRepo, logg)

	// ========================
	// gRPC Server
	// ========================
	grpcServer := grpc.NewServer()
	taskServer := server.NewTaskServer(taskService, logg)

	pb.RegisterTaskServiceServer(grpcServer, taskServer)

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// ========================
	// Graceful shutdown
	// ========================
	shutdownCtx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	go func() {
		logg.Info("gRPC server started", "addr", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("grpc serve error: %v", err)
		}
	}()

	<-shutdownCtx.Done()
	logg.Info("shutdown signal received")

	grpcServer.GracefulStop()
	logg.Info("server stopped gracefully")
}

// getEnv оставлен для обратной совместимости, но не используется
// в текущей версии, где конфигурация загружается через cleanenv.
func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// runMigrations применяет миграции через уже открытое соединение *sql.DB.
// Это решает проблему "no scheme", потому что migrate.New(...) ожидает URL вида postgres://...
func runMigrations(db *sql.DB) {
	driver, err := migratepg.WithInstance(db, &migratepg.Config{})
	if err != nil {
		log.Fatalf("migration driver init error: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file:///app/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("migration init error: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migration up error: %v", err)
	}

	log.Println("migrations applied successfully")
}
