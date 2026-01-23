package server_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/N0F1X3d/todo/db-service/internal/models"
	"github.com/N0F1X3d/todo/db-service/internal/server"
	"github.com/N0F1X3d/todo/db-service/mocks"
	"github.com/N0F1X3d/todo/db-service/pkg/logger"
	"github.com/N0F1X3d/todo/db-service/pkg/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestTaskServer_CreateTask_Success(t *testing.T) {
	mockService := mocks.NewTaskServiceInterface(t)
	testLogger := logger.New("test-logs")

	createdTime := time.Now()
	mockService.On("CreateTask", models.CreateTaskRequest{
		Title:       "test task",
		Description: "test desc",
	}).Return(&models.Task{
		ID:          1,
		Title:       "test task",
		Description: "test desc",
		Completed:   false,
		CreatedAt:   createdTime,
		UpdatedAt:   createdTime,
	}, nil)

	server := server.NewTaskServer(mockService, testLogger)
	req := &proto.CreateTaskRequest{
		Title:       "test task",
		Description: "test desc",
	}

	resp, err := server.CreateTask(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(1), resp.Id)
	assert.Equal(t, "test task", resp.Title)
	assert.Equal(t, "test desc", resp.Description)
	assert.False(t, resp.Completed)
	assert.Equal(t, createdTime.Format(time.RFC3339), resp.CreatedAt)
}

func TestTaskServer_CreateTask_EmptyTitle(t *testing.T) {
	mockService := mocks.NewTaskServiceInterface(t)
	testLogger := logger.New("test-logs")

	mockService.On("CreateTask", models.CreateTaskRequest{
		Title:       "",
		Description: "test",
	}).Return(nil, errors.New("title can not be empty"))

	server := server.NewTaskServer(mockService, testLogger)

	req := &proto.CreateTaskRequest{
		Title:       "",
		Description: "test",
	}

	resp, err := server.CreateTask(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)

	grpcStatus, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, grpcStatus.Code())
	assert.Equal(t, "title can not be empty", grpcStatus.Message())
}

func TestTaskServer_CreateTask_TitleTooLong(t *testing.T) {
	mockService := mocks.NewTaskServiceInterface(t)
	testLogger := logger.New("test-logs")

	mockService.On("CreateTask", models.CreateTaskRequest{
		Title:       string(make([]byte, 256)),
		Description: "test",
	}).Return(nil, errors.New("title too long, maximum 255 characters"))

	server := server.NewTaskServer(mockService, testLogger)

	req := &proto.CreateTaskRequest{
		Title:       string(make([]byte, 256)),
		Description: "test",
	}

	resp, err := server.CreateTask(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)

	grpcStatus, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, grpcStatus.Code())
	assert.Equal(t, "title too long, maximum 255 characters", grpcStatus.Message())
}

func TestTaskServer_CreateTask_InternalError(t *testing.T) {
	mockService := mocks.NewTaskServiceInterface(t)
	testLogger := logger.New("test-logs")
	mockService.On("CreateTask", mock.Anything).Return(nil, errors.New("error"))

	server := server.NewTaskServer(mockService, testLogger)

	req := &proto.CreateTaskRequest{
		Title:       "test",
		Description: "test desc",
	}

	resp, err := server.CreateTask(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)

	grpcStatus, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, grpcStatus.Code())
	assert.Equal(t, "internal server error", grpcStatus.Message())
}

func TestTaskServer_GetTaskByID_Success(t *testing.T) {
	// Arrange
	mockService := mocks.NewTaskServiceInterface(t)
	testLogger := logger.New("test-logs")

	createdTime := time.Now()
	updatedTime := createdTime.Add(time.Hour)

	mockService.On("GetTaskByID", 1).Return(&models.Task{
		ID:          1,
		Title:       "Test Task",
		Description: "Test Description",
		Completed:   true,
		CreatedAt:   createdTime,
		UpdatedAt:   updatedTime,
	}, nil)

	server := server.NewTaskServer(mockService, testLogger)

	req := &proto.GetTaskByIDRequest{
		Id: 1,
	}

	// Act
	resp, err := server.GetTaskByID(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(1), resp.Id)
	assert.Equal(t, "Test Task", resp.Title)
	assert.True(t, resp.Completed)
	assert.Equal(t, createdTime.Format(time.RFC3339), resp.CreatedAt)
	assert.Equal(t, updatedTime.Format(time.RFC3339), resp.UpdatedAt)
}

func TestTaskServer_GetTaskByID_InvalidID(t *testing.T) {
	// Arrange
	mockService := mocks.NewTaskServiceInterface(t)
	testLogger := logger.New("test-logs")

	mockService.On("GetTaskByID", 0).Return(nil, errors.New("invalid task id"))

	server := server.NewTaskServer(mockService, testLogger)

	req := &proto.GetTaskByIDRequest{
		Id: 0,
	}

	// Act
	resp, err := server.GetTaskByID(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)

	grpcStatus, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, grpcStatus.Code())
	assert.Equal(t, "invalid task id", grpcStatus.Message())
}

func TestTaskServer_GetTaskByID_NotFound(t *testing.T) {
	// Arrange
	mockService := mocks.NewTaskServiceInterface(t)
	testLogger := logger.New("test-logs")

	mockService.On("GetTaskByID", 999).Return(nil, errors.New("task not found"))

	server := server.NewTaskServer(mockService, testLogger)

	req := &proto.GetTaskByIDRequest{
		Id: 999,
	}

	// Act
	resp, err := server.GetTaskByID(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)

	grpcStatus, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, grpcStatus.Code())
	assert.Equal(t, "task not found", grpcStatus.Message())
}

func TestTaskServer_GetTaskByID_InternalError(t *testing.T) {
	// Arrange
	mockService := mocks.NewTaskServiceInterface(t)
	testLogger := logger.New("test-logs")

	mockService.On("GetTaskByID", mock.Anything).Return(nil, errors.New("database error"))

	server := server.NewTaskServer(mockService, testLogger)

	req := &proto.GetTaskByIDRequest{
		Id: 1,
	}

	// Act
	resp, err := server.GetTaskByID(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)

	grpcStatus, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, grpcStatus.Code())
	assert.Equal(t, "internal server error", grpcStatus.Message())
}
func TestTaskServer_GetAllTasks_Success(t *testing.T) {
	// Arrange
	mockService := mocks.NewTaskServiceInterface(t)
	testLogger := logger.New("test-logs")

	createdTime := time.Now()
	tasks := []models.Task{
		{
			ID:          1,
			Title:       "Task 1",
			Description: "Description 1",
			Completed:   false,
			CreatedAt:   createdTime,
			UpdatedAt:   createdTime,
		},
		{
			ID:          2,
			Title:       "Task 2",
			Description: "Description 2",
			Completed:   true,
			CreatedAt:   createdTime,
			UpdatedAt:   createdTime.Add(time.Hour),
		},
	}

	mockService.On("GetAllTasks").Return(tasks, nil)

	server := server.NewTaskServer(mockService, testLogger)

	req := &proto.GetAllTasksRequest{}

	// Act
	resp, err := server.GetAllTasks(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Tasks, 2)
	assert.Equal(t, int32(1), resp.Tasks[0].Id)
	assert.Equal(t, "Task 1", resp.Tasks[0].Title)
	assert.Equal(t, int32(2), resp.Tasks[1].Id)
	assert.Equal(t, "Task 2", resp.Tasks[1].Title)
	assert.True(t, resp.Tasks[1].Completed)
}

func TestTaskServer_GetAllTasks_Empty(t *testing.T) {
	// Arrange
	mockService := mocks.NewTaskServiceInterface(t)
	testLogger := logger.New("test-logs")

	mockService.On("GetAllTasks").Return([]models.Task{}, nil)

	server := server.NewTaskServer(mockService, testLogger)

	req := &proto.GetAllTasksRequest{}

	// Act
	resp, err := server.GetAllTasks(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Empty(t, resp.Tasks)
}

func TestTaskServer_GetAllTasks_InternalError(t *testing.T) {
	// Arrange
	mockService := mocks.NewTaskServiceInterface(t)
	testLogger := logger.New("test-logs")

	mockService.On("GetAllTasks").Return(nil, errors.New("database error"))

	server := server.NewTaskServer(mockService, testLogger)

	req := &proto.GetAllTasksRequest{}

	// Act
	resp, err := server.GetAllTasks(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)

	grpcStatus, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, grpcStatus.Code())
	assert.Equal(t, "internal server error", grpcStatus.Message())
}

func TestTaskServer_CompleteTask_Success(t *testing.T) {
	// Arrange
	mockService := mocks.NewTaskServiceInterface(t)
	testLogger := logger.New("test-logs")

	createdTime := time.Now()
	completedTime := createdTime.Add(time.Hour)

	mockService.On("CompleteTask", 1).Return(&models.Task{
		ID:          1,
		Title:       "Test Task",
		Description: "Test Description",
		Completed:   true,
		CreatedAt:   createdTime,
		UpdatedAt:   completedTime,
	}, nil)

	server := server.NewTaskServer(mockService, testLogger)

	req := &proto.CompleteTaskRequest{
		Id: 1,
	}

	// Act
	resp, err := server.CompleteTask(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(1), resp.Id)
	assert.True(t, resp.Completed)
	assert.Equal(t, completedTime.Format(time.RFC3339), resp.UpdatedAt)
}

func TestTaskServer_CompleteTask_InvalidID(t *testing.T) {
	// Arrange
	mockService := mocks.NewTaskServiceInterface(t)
	testLogger := logger.New("test-logs")

	mockService.On("CompleteTask", 0).Return(nil, errors.New("invalid task id"))

	server := server.NewTaskServer(mockService, testLogger)

	req := &proto.CompleteTaskRequest{
		Id: 0,
	}

	// Act
	resp, err := server.CompleteTask(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)

	grpcStatus, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, grpcStatus.Code())
	assert.Equal(t, "invalid task id", grpcStatus.Message())
}

func TestTaskServer_CompleteTask_NotFound(t *testing.T) {
	// Arrange
	mockService := mocks.NewTaskServiceInterface(t)
	testLogger := logger.New("test-logs")

	mockService.On("CompleteTask", 999).Return(nil, errors.New("task not found"))

	server := server.NewTaskServer(mockService, testLogger)

	req := &proto.CompleteTaskRequest{
		Id: 999,
	}

	// Act
	resp, err := server.CompleteTask(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)

	grpcStatus, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, grpcStatus.Code())
	assert.Equal(t, "task not found", grpcStatus.Message())
}

func TestTaskServer_CompleteTask_AlreadyCompleted(t *testing.T) {
	// Arrange
	mockService := mocks.NewTaskServiceInterface(t)
	testLogger := logger.New("test-logs")

	mockService.On("CompleteTask", 1).Return(nil, errors.New("task already completed"))

	server := server.NewTaskServer(mockService, testLogger)

	req := &proto.CompleteTaskRequest{
		Id: 1,
	}

	// Act
	resp, err := server.CompleteTask(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)

	grpcStatus, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.FailedPrecondition, grpcStatus.Code())
	assert.Equal(t, "task already completed", grpcStatus.Message())
}

func TestTaskServer_CompleteTask_InternalError(t *testing.T) {
	// Arrange
	mockService := mocks.NewTaskServiceInterface(t)
	testLogger := logger.New("test-logs")

	mockService.On("CompleteTask", mock.Anything).Return(nil, errors.New("database error"))

	server := server.NewTaskServer(mockService, testLogger)

	req := &proto.CompleteTaskRequest{
		Id: 1,
	}

	// Act
	resp, err := server.CompleteTask(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)

	grpcStatus, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, grpcStatus.Code())
	assert.Equal(t, "internal server error", grpcStatus.Message())
}

func TestTaskServer_DeleteTask_Success(t *testing.T) {
	// Arrange
	mockService := mocks.NewTaskServiceInterface(t)
	testLogger := logger.New("test-logs")

	mockService.On("DeleteTask", 1).Return(nil)

	server := server.NewTaskServer(mockService, testLogger)

	req := &proto.DeleteTaskRequest{
		Id: 1,
	}

	// Act
	resp, err := server.DeleteTask(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
}

func TestTaskServer_DeleteTask_InvalidID(t *testing.T) {
	// Arrange
	mockService := mocks.NewTaskServiceInterface(t)
	testLogger := logger.New("test-logs")

	mockService.On("DeleteTask", 0).Return(errors.New("invalid id"))

	server := server.NewTaskServer(mockService, testLogger)

	req := &proto.DeleteTaskRequest{
		Id: 0,
	}

	// Act
	resp, err := server.DeleteTask(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)

	grpcStatus, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, grpcStatus.Code())
	assert.Equal(t, "invalid id", grpcStatus.Message())
}

func TestTaskServer_DeleteTask_NotFound(t *testing.T) {
	// Arrange
	mockService := mocks.NewTaskServiceInterface(t)
	testLogger := logger.New("test-logs")

	mockService.On("DeleteTask", 999).Return(errors.New("failed to find task"))

	server := server.NewTaskServer(mockService, testLogger)

	req := &proto.DeleteTaskRequest{
		Id: 999,
	}

	// Act
	resp, err := server.DeleteTask(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)

	grpcStatus, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, grpcStatus.Code())
	assert.Equal(t, "task not found", grpcStatus.Message())
}

func TestTaskServer_DeleteTask_InternalError(t *testing.T) {
	// Arrange
	mockService := mocks.NewTaskServiceInterface(t)
	testLogger := logger.New("test-logs")

	mockService.On("DeleteTask", mock.Anything).Return(errors.New("database error"))

	server := server.NewTaskServer(mockService, testLogger)

	req := &proto.DeleteTaskRequest{
		Id: 1,
	}

	// Act
	resp, err := server.DeleteTask(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)

	grpcStatus, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, grpcStatus.Code())
	assert.Equal(t, "internal server error", grpcStatus.Message())
}
