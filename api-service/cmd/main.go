package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"github.com/N0F1X3d/todo/api-service/internal/clients/grpcclient"
	"github.com/N0F1X3d/todo/api-service/internal/config"
	"github.com/N0F1X3d/todo/api-service/internal/http-server/handlers"
	"github.com/N0F1X3d/todo/api-service/internal/http-server/middleware"
	pkgKafka "github.com/N0F1X3d/todo/pkg/kafka"
	"github.com/N0F1X3d/todo/pkg/logger"
)

func main() {
	// ===== Config =====
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{
			ServiceName: "todo-api-service",
			HTTPHost:    "0.0.0.0",
			HTTPPort:    8080,
			GRPCHost:    "localhost",
			GRPCPort:    50051,
		}
	}

	// ===== Logger =====
	appLogger := logger.New("api-service", "api-logs").WithComponent("main")

	appLogger.Info("Starting API Service",
		"http_addr", cfg.HTTPAddress(),
		"grpc_addr", cfg.GRPCAddress(),
	)

	// ===== gRPC client =====
	grpcClient, err := grpcclient.NewTaskClient(cfg.GRPCAddress(), appLogger)
	if err != nil {
		appLogger.Fatal("Failed to connect to db-service", err)
	}
	defer grpcClient.Close()

	// ===== Kafka producer =====
	producer := pkgKafka.NewProducer([]string{"kafka:9092"}, "task-events")
	defer producer.Close()

	// ===== Handlers =====
	taskHandler := handlers.NewTaskHandler(grpcClient, producer, appLogger)

	// ===== Router =====
	router := mux.NewRouter().StrictSlash(true)

	// === API routes (по ТЗ) ===
	router.HandleFunc("/create", taskHandler.CreateTask).Methods(http.MethodPost)
	router.HandleFunc("/list", taskHandler.ListTasks).Methods(http.MethodGet)
	router.HandleFunc("/delete", taskHandler.DeleteTask).Methods(http.MethodDelete)
	router.HandleFunc("/done", taskHandler.CompleteTask).Methods(http.MethodPut)

	// ===== Health check =====
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"service": cfg.ServiceName,
		})
	}).Methods(http.MethodGet)

	// ===== Middleware =====
	handler := middleware.Chain(
		router,
		middleware.CORSMiddleware,
		middleware.SecurityHeadersMiddleware,
		func(next http.Handler) http.Handler {
			return middleware.LoggingMiddleware(next, appLogger)
		},
		middleware.JSONContentTypeMiddleware,
	)

	// ===== HTTP Server =====
	server := &http.Server{
		Addr:         cfg.HTTPAddress(),
		Handler:      handler,
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
	}

	// ===== Graceful shutdown =====
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		appLogger.Info("API Service started", "address", cfg.HTTPAddress())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Failed to start server", err)
		}
	}()

	<-stop
	appLogger.Info("Shutting down API Service...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		appLogger.Error("Server forced to shutdown", err)
	}

	appLogger.Info("API Service stopped")
}
