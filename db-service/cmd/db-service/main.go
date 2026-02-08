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

	_ "github.com/lib/pq"
	"google.golang.org/grpc"

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
	grpcAddr := getEnv("GRPC_ADDR", ":50051")
	dbDSN := getEnv(
		"POSTGRES_DSN",
		"postgres://postgres:postgres@postgres:5432/tasks?sslmode=disable",
	)

	// ========================
	// Logger
	// ========================
	logg := logger.New("db-service", "main-logs").WithComponent("main")

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

	defer func() {
		if err := db.Close(); err != nil {
			logg.Error("failed to close db", err)
		}
	}()

	// ========================
	// Repository
	// ========================
	taskRepo := repository.NewTaskRepository(db, logg)

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

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
