package server

import (
	"context"

	"github.com/N0F1X3d/todo/db-service/internal/service"
	"github.com/N0F1X3d/todo/db-service/pkg/logger"
	"github.com/N0F1X3d/todo/db-service/pkg/proto"
)

// TaskServer реализует gRPC сервер для работы с базой данных
type TaskServer struct {
	proto.UnimplementedTaskServiceServer
	service *service.TaskService
	log     *logger.Logger
}

// NewTaskServer
func NewTaskServer(service *service.TaskService, log *logger.Logger) *TaskServer {
	return &TaskServer{
		service: service,
		log:     log.WithComponent("Server").WithFunction("NewTaskServer"),
	}
}

// CreateTask обрабатывает gRPC запрос на создание задачи
func (s *TaskServer) CreateTask(ctx context.Context, req *proto.CompleteTaskRequest) (*proto.TaskResponse, error) {
	return nil, nil
}

// GetTaskByID обрабатывает gRPC запрос на поиск задачи по ID
func (s *TaskServer) GetTaskByID(ctx context.Context, req *proto.GetTaskByIDRequest) (*proto.TaskResponse, error) {
	return nil, nil
}

// GetAllTasks обрабатывает gRPC запрос на поиск всех задач
func (s *TaskServer) GetAllTasks(ctx context.Context, req *proto.GetAllTasksRequest) (*proto.GetAllTasksResponse, error) {
	return nil, nil
}

// CompleteTask обрабатывает gRPC запрос на завершение задачи по ID
func (s *TaskServer) CompleteTask(ctx context.Context, req *proto.CompleteTaskRequest) (*proto.TaskResponse, error) {
	return nil, nil
}

// DeleteTask обрабатывает gRPC запрос на удаление задачи по ID
func (s *TaskServer) DeleteTask(ctx context.Context, req *proto.DeleteTaskRequest) (*proto.DeleteTaskResponse, error) {
	return nil, nil
}
