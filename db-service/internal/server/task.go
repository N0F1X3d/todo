package server

import (
	"context"
	"time"

	"github.com/N0F1X3d/todo/db-service/internal/models"
	"github.com/N0F1X3d/todo/db-service/internal/service"
	"github.com/N0F1X3d/todo/db-service/pkg/logger"
	"github.com/N0F1X3d/todo/db-service/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:generate mockery --name=TaskServerInterface --filename=task_server_interface.go --output=../../mocks --case=underscore

// TaskServerInterface определяет контракт для gRPC сервера задач
type TaskServerInterface interface {
	CreateTask(ctx context.Context, req *proto.CreateTaskRequest) (*proto.TaskResponse, error)
	GetTaskByID(ctx context.Context, req *proto.GetTaskByIDRequest) (*proto.TaskResponse, error)
	GetAllTasks(ctx context.Context, req *proto.GetAllTasksRequest) (*proto.GetAllTasksResponse, error)
	CompleteTask(ctx context.Context, req *proto.CompleteTaskRequest) (*proto.TaskResponse, error)
	DeleteTask(ctx context.Context, req *proto.DeleteTaskRequest) (*proto.DeleteTaskResponse, error)
	// Наследуем методы от встроенного интерфейса
	proto.TaskServiceServer
}

// TaskServer реализует gRPC сервер для работы с базой данных
type TaskServer struct {
	proto.UnimplementedTaskServiceServer
	service service.TaskServiceInterface
	log     *logger.Logger
}

// NewTaskServer
func NewTaskServer(service service.TaskServiceInterface, log *logger.Logger) *TaskServer {
	return &TaskServer{
		service: service,
		log:     log.WithComponent("Server").WithFunction("NewTaskServer"),
	}
}

// CreateTask обрабатывает gRPC запрос на создание задачи
func (s *TaskServer) CreateTask(ctx context.Context, req *proto.CreateTaskRequest) (*proto.TaskResponse, error) {
	const op = "CreateTask"

	s.log.LogRequest(op, map[string]interface{}{
		"title":       req.GetTitle(),
		"description": req.GetDescription(),
	})
	start := time.Now()

	createReq := models.CreateTaskRequest{
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
	}

	task, err := s.service.CreateTask(createReq)

	if err != nil {
		s.log.ErrorWithContext("failed to create task", err, op, "title", req.GetTitle(), "description", req.GetDescription())
		// Конвертируем ошибки в статусы
		switch err.Error() {
		case "title can not be empty":
			return nil, status.Error(codes.InvalidArgument, "title can not be empty")
		case "title too long, maximum 255 characters":
			return nil, status.Error(codes.InvalidArgument, "title to long, maximum 255 characters")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	response := &proto.TaskResponse{
		Id:          int32(task.ID),
		Title:       task.Title,
		Description: task.Description,
		Completed:   task.Completed,
		CreatedAt:   task.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   task.UpdatedAt.Format(time.RFC3339),
	}

	duration := time.Since(start).Microseconds()

	s.log.LogResponse(op, response)
	s.log.LogQueryResult(op, duration, 1)
	return response, nil
}

// GetTaskByID обрабатывает gRPC запрос на поиск задачи по ID
func (s *TaskServer) GetTaskByID(ctx context.Context, req *proto.GetTaskByIDRequest) (*proto.TaskResponse, error) {
	const op = "GetTaskByID"

	s.log.LogRequest(op, map[string]interface{}{"id": req.GetId()})

	start := time.Now()

	task, err := s.service.GetTaskByID(int(req.GetId()))
	if err != nil {
		s.log.ErrorWithContext("failed to get task", err, op, "id", req.GetId())
		switch err.Error() {
		case "invalid task id":
			return nil, status.Error(codes.InvalidArgument, "invalid task id")
		case "task not found":
			return nil, status.Error(codes.NotFound, "task not found")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	response := &proto.TaskResponse{
		Id:          int32(task.ID),
		Title:       task.Title,
		Description: task.Description,
		Completed:   task.Completed,
		CreatedAt:   task.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   task.UpdatedAt.Format(time.RFC3339),
	}

	duration := time.Since(start).Milliseconds()

	s.log.LogResponse(op, response)
	s.log.LogQueryResult(op, duration, 1)

	return response, nil
}

// GetAllTasks обрабатывает gRPC запрос на поиск всех задач
func (s *TaskServer) GetAllTasks(ctx context.Context, req *proto.GetAllTasksRequest) (*proto.GetAllTasksResponse, error) {
	const op = "GetAllTasks"

	s.log.LogRequest(op, nil)

	start := time.Now()

	tasks, err := s.service.GetAllTasks()
	if err != nil {
		s.log.ErrorWithContext("failed to get all tasks", err, op)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	response := &proto.GetAllTasksResponse{
		Tasks: make([]*proto.TaskResponse, 0, len(tasks)),
	}

	for _, task := range tasks {
		response.Tasks = append(response.Tasks, &proto.TaskResponse{
			Id:          int32(task.ID),
			Title:       task.Title,
			Description: task.Description,
			Completed:   task.Completed,
			CreatedAt:   task.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   task.UpdatedAt.Format(time.RFC3339),
		})
	}

	duration := time.Since(start).Milliseconds()

	s.log.LogResponse(op, map[string]interface{}{"tasks_count": len(tasks)})
	s.log.LogQueryResult(op, duration, int64(len(tasks)))

	return response, nil
}

// CompleteTask обрабатывает gRPC запрос на завершение задачи по ID
func (s *TaskServer) CompleteTask(ctx context.Context, req *proto.CompleteTaskRequest) (*proto.TaskResponse, error) {
	const op = "CompleteTask"

	s.log.LogRequest(op, map[string]interface{}{"id": req.GetId()})

	start := time.Now()

	task, err := s.service.CompleteTask(int(req.GetId()))
	if err != nil {
		s.log.ErrorWithContext("failed to complete task", err, op, "task_id", req.GetId())
		switch err.Error() {
		case "invalid task id":
			return nil, status.Error(codes.InvalidArgument, "invalid task id")
		case "task not found":
			return nil, status.Error(codes.NotFound, "task not found")
		case "task already completed":
			return nil, status.Error(codes.FailedPrecondition, "task already completed")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	response := &proto.TaskResponse{
		Id:          int32(task.ID),
		Title:       task.Title,
		Description: task.Description,
		Completed:   task.Completed,
		CreatedAt:   task.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   task.UpdatedAt.Format(time.RFC3339),
	}

	duration := time.Since(start).Milliseconds()

	s.log.LogResponse(op, response)
	s.log.LogQueryResult(op, duration, 1)
	return response, nil
}

// DeleteTask обрабатывает gRPC запрос на удаление задачи по ID
func (s *TaskServer) DeleteTask(ctx context.Context, req *proto.DeleteTaskRequest) (*proto.DeleteTaskResponse, error) {
	const op = "DeleteTask"

	s.log.LogRequest(op, map[string]interface{}{"id": req.GetId()})

	start := time.Now()

	err := s.service.DeleteTask(int(req.GetId()))
	if err != nil {
		s.log.ErrorWithContext("failed to delete task", err, op, "task_id", req.GetId())
		switch err.Error() {
		case "invalid id":
			return nil, status.Error(codes.InvalidArgument, "invalid id")
		case "failed to find task":
			return nil, status.Error(codes.NotFound, "task not found")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	duration := time.Since(start).Milliseconds()

	s.log.LogResponse(op, map[string]interface{}{"deleted": true, "task_id": req.GetId()})
	s.log.LogQueryResult(op, duration, 1)
	return &proto.DeleteTaskResponse{Success: true}, nil
}
