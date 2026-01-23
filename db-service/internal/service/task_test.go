package service_test

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/N0F1X3d/todo/db-service/internal/models"
	"github.com/N0F1X3d/todo/db-service/internal/service"
	"github.com/N0F1X3d/todo/db-service/mocks"
	"github.com/N0F1X3d/todo/db-service/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTaskService_CreateTask_Success(t *testing.T) {
	mockRepo := mocks.NewTaskRepositoryInterface(t)
	mockRepo.On("CreateTask", mock.AnythingOfType("models.CreateTaskRequest")).Return(&models.Task{
		ID:          1,
		Title:       "test task",
		Description: "test description",
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil)

	testLogger := logger.New("test-logs")
	taskService := service.NewTaskService(mockRepo, testLogger)

	req := models.CreateTaskRequest{
		Title:       "test task",
		Description: "test description",
	}
	task, err := taskService.CreateTask(req)

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, 1, task.ID)
	assert.Equal(t, "test task", task.Title)
	assert.False(t, task.Completed)
}

func TestTaskService_CreateTask_EmptyTitle(t *testing.T) {
	mockRepo := mocks.NewTaskRepositoryInterface(t)
	testLogger := logger.New("test-logs")
	taskService := service.NewTaskService(mockRepo, testLogger)

	req := models.CreateTaskRequest{Title: " "}
	task, err := taskService.CreateTask(req)

	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Equal(t, "title can not be empty", err.Error())
}

func TestTaskService_CreateTask_TitleTooLong(t *testing.T) {
	mockRepo := mocks.NewTaskRepositoryInterface(t)
	testLogger := logger.New("test-logs")
	taskService := service.NewTaskService(mockRepo, testLogger)

	longTitle := string(make([]byte, 256))
	req := models.CreateTaskRequest{Title: longTitle}
	task, err := taskService.CreateTask(req)

	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Equal(t, "title too long, maximum 255 characters", err.Error())
}

func TestTaskService_GetTaskByID_Success(t *testing.T) {
	mockRepo := mocks.NewTaskRepositoryInterface(t)
	mockRepo.On("GetTaskByID", 1).Return(&models.Task{
		ID:          1,
		Title:       "test task",
		Description: "test description",
		Completed:   false,
	}, nil)

	testLogger := logger.New("test-logs")
	taskService := service.NewTaskService(mockRepo, testLogger)

	task, err := taskService.GetTaskByID(1)

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, 1, task.ID)
	assert.Equal(t, "test task", task.Title)
}

func TestTaskService_GetTaskByID_InvalidTaskID(t *testing.T) {
	mockRepo := mocks.NewTaskRepositoryInterface(t)
	testLogger := logger.New("test-logs")
	taskService := service.NewTaskService(mockRepo, testLogger)

	task, err := taskService.GetTaskByID(0)

	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Equal(t, "invalid task id", err.Error())
}

func TestTaskService_GetTaskByID_TaskNotFound(t *testing.T) {
	mockRepo := mocks.NewTaskRepositoryInterface(t)
	mockRepo.On("GetTaskByID", 999).Return(nil, sql.ErrNoRows)
	testLogger := logger.New("test-logs")
	taskService := service.NewTaskService(mockRepo, testLogger)

	task, err := taskService.GetTaskByID(999)

	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Equal(t, "task not found", err.Error())
}

func TestTaskService_GetAllTasks_Success(t *testing.T) {
	mockRepo := mocks.NewTaskRepositoryInterface(t)
	tasks := []models.Task{
		{ID: 1, Title: "test task 1", Completed: false, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: 2, Title: "test task 2", Completed: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	mockRepo.On("GetAllTasks").Return(tasks, nil)
	testLogger := logger.New("test-logs")
	taskService := service.NewTaskService(mockRepo, testLogger)

	tasks, err := taskService.GetAllTasks()

	assert.NoError(t, err)
	assert.NotNil(t, tasks)
	assert.Len(t, tasks, 2)
	assert.Equal(t, "test task 1", tasks[0].Title)
	assert.Equal(t, "test task 2", tasks[1].Title)
	assert.False(t, tasks[0].Completed)
	assert.True(t, tasks[1].Completed)
}

func TestTaskService_GetAllTasks_DatabaseError(t *testing.T) {
	mockRepo := mocks.NewTaskRepositoryInterface(t)
	mockRepo.On("GetAllTasks").Return(nil, errors.New("any error"))

	testLogger := logger.New("test-logs")
	taskService := service.NewTaskService(mockRepo, testLogger)

	tasks, err := taskService.GetAllTasks()

	assert.Error(t, err)
	assert.Nil(t, tasks)
	assert.Equal(t, "internal server error", err.Error())
}

func TestTaskService_GetAllTasks_EmptyList(t *testing.T) {
	// Arrange
	mockRepo := mocks.NewTaskRepositoryInterface(t)

	mockRepo.On("GetAllTasks").Return([]models.Task{}, nil)

	testLogger := logger.New("test-logs")
	taskService := service.NewTaskService(mockRepo, testLogger)

	// Act
	tasks, err := taskService.GetAllTasks()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, tasks)
	assert.Empty(t, tasks)
}

func TestTaskService_CompleteTask_Success(t *testing.T) {
	mockRepo := mocks.NewTaskRepositoryInterface(t)
	mockRepo.On("GetTaskByID", 1).Return(&models.Task{
		ID:        1,
		Title:     "test task",
		Completed: false,
	}, nil)
	completedTime := time.Now()
	mockRepo.On("CompleteTask", 1).Return(&models.Task{
		ID:        1,
		Title:     "test task",
		Completed: true,
		UpdatedAt: completedTime,
	}, nil)

	testLogger := logger.New("test-logs")
	taskService := service.NewTaskService(mockRepo, testLogger)

	task, err := taskService.CompleteTask(1)

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, 1, task.ID)
	assert.True(t, task.Completed)
	assert.Equal(t, completedTime, task.UpdatedAt)
}

func TestTaskService_CompleteTask_InvalidID(t *testing.T) {
	mockRepo := mocks.NewTaskRepositoryInterface(t)
	testLogger := logger.New("test-logs")
	taskService := service.NewTaskService(mockRepo, testLogger)

	_, err := taskService.CompleteTask(0)

	assert.Error(t, err)
	assert.Equal(t, "invalid task id", err.Error())
}

func TestTaskService_CompleteTask_TaskNotFound(t *testing.T) {
	mockRepo := mocks.NewTaskRepositoryInterface(t)
	mockRepo.On("GetTaskByID", 99).Return(nil, sql.ErrNoRows)
	testLogger := logger.New("test-logs")
	taskService := service.NewTaskService(mockRepo, testLogger)

	_, err := taskService.GetTaskByID(99)

	assert.Error(t, err)
}

func TestTaskService_CompleteTask_AlreadyCompleted(t *testing.T) {
	mockRepo := mocks.NewTaskRepositoryInterface(t)
	mockRepo.On("GetTaskByID", 1).Return(&models.Task{
		ID:        1,
		Title:     "test",
		Completed: true,
	}, nil)
	testLogger := logger.New("test-logs")
	taskService := service.NewTaskService(mockRepo, testLogger)

	_, err := taskService.CompleteTask(1)

	assert.Error(t, err)
	assert.Equal(t, "task already completed", err.Error())
}

func TestTaskService_CompleteTask_CompleteError(t *testing.T) {
	// Arrange
	mockRepo := mocks.NewTaskRepositoryInterface(t)

	mockRepo.On("GetTaskByID", 1).
		Return(&models.Task{
			ID:        1,
			Title:     "Test Task",
			Completed: false,
		}, nil)

	mockRepo.On("CompleteTask", 1).
		Return(nil, errors.New("update failed"))

	testLogger := logger.New("test-logs")
	taskService := service.NewTaskService(mockRepo, testLogger)

	// Act
	task, err := taskService.CompleteTask(1)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Equal(t, "update failed", err.Error())
}

func TestTaskService_DeleteTask_Success(t *testing.T) {
	mockRepo := mocks.NewTaskRepositoryInterface(t)
	mockRepo.On("GetTaskByID", 1).Return(&models.Task{
		ID:    1,
		Title: "test",
	}, nil)
	mockRepo.On("DeleteTask", 1).Return(nil)
	testLogger := logger.New("test-logs")
	taskService := service.NewTaskService(mockRepo, testLogger)

	err := taskService.DeleteTask(1)

	assert.NoError(t, err)
}

func TestTaskService_DeleteTask_InvalidID(t *testing.T) {
	mockRepo := mocks.NewTaskRepositoryInterface(t)
	testLogger := logger.New("test-logs")
	taskService := service.NewTaskService(mockRepo, testLogger)

	err := taskService.DeleteTask(0)

	assert.Error(t, err)
	assert.Equal(t, "invalid id", err.Error())
}

func TestTaskService_DeleteTask_TaskNotFound(t *testing.T) {
	mockRepo := mocks.NewTaskRepositoryInterface(t)
	mockRepo.On("GetTaskByID", 99).Return(nil, sql.ErrNoRows)
	testLogger := logger.New("test-logs")
	taskService := service.NewTaskService(mockRepo, testLogger)

	err := taskService.DeleteTask(99)

	assert.Error(t, err)
	assert.Equal(t, "failed to find task", err.Error())
}

func TestTaskService_DeleteTask_DeleteError(t *testing.T) {
	mockRepo := mocks.NewTaskRepositoryInterface(t)
	mockRepo.On("GetTaskByID", 1).Return(&models.Task{
		ID:    1,
		Title: "test",
	}, nil)
	mockRepo.On("DeleteTask", 1).Return(errors.New("failed to delete task"))
	testLogger := logger.New("test-logs")
	taskService := service.NewTaskService(mockRepo, testLogger)

	err := taskService.DeleteTask(1)

	assert.Error(t, err)
	assert.Equal(t, "failed to delete task", err.Error())
}

func TestTaskService_DeleteTask_GetTaskError(t *testing.T) {
	mockRepo := mocks.NewTaskRepositoryInterface(t)
	mockRepo.On("GetTaskByID", 1).Return(nil, errors.New("connection error"))
	testLogger := logger.New("test-logs")
	taskService := service.NewTaskService(mockRepo, testLogger)

	err := taskService.DeleteTask(1)

	assert.Error(t, err)
	assert.Equal(t, "failed to find task", err.Error())
}
