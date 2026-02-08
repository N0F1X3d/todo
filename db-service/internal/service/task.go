package service

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/N0F1X3d/todo/db-service/internal/models"
	"github.com/N0F1X3d/todo/db-service/internal/repository"
	"github.com/N0F1X3d/todo/pkg/logger"
)

//go:generate mockery --name=TaskServiceInterface --filename=task_service_interface.go --output=../../mocks --case=underscore
type TaskServiceInterface interface {
	CreateTask(req models.CreateTaskRequest) (*models.Task, error)
	GetTaskByID(id int) (*models.Task, error)
	GetAllTasks() ([]models.Task, error)
	CompleteTask(id int) (*models.Task, error)
	DeleteTask(id int) error
}

// TaskService предоставляет бизнес-логику для работы с задачами.
type TaskService struct {
	repo repository.TaskRepositoryInterface
	log  *logger.Logger
}

// NewTaskService создает новый экземпляр TaskService
func NewTaskService(repo repository.TaskRepositoryInterface, log *logger.Logger) *TaskService {
	return &TaskService{
		repo: repo,
		log:  log.WithComponent("service").WithFunction("TaskService"),
	}
}

// CreateTask создает новую задачу с применением бизнес-логики и валидации
func (t *TaskService) CreateTask(req models.CreateTaskRequest) (*models.Task, error) {
	const op = "CreateTask"

	t.log.LogRequest(op, req)

	if strings.TrimSpace(req.Title) == "" {
		err := errors.New("title can not be empty")
		t.log.ErrorWithContext("validation failed", err, op, "request", req)
		return nil, err
	}
	if len(req.Title) > 255 {
		err := errors.New("title too long, maximum 255 characters")
		t.log.ErrorWithContext("validation failed", err, op, "request", req)
		return nil, err
	}
	task, err := t.repo.CreateTask(req)
	if err != nil {
		t.log.ErrorWithContext("failed to create task in repository", err, op, "request", req)
		return nil, err
	}

	t.log.LogResponse(op, task)
	return task, nil
}

// GetTaskByID возвращает задачу по ее ID
func (t *TaskService) GetTaskByID(id int) (*models.Task, error) {
	const op = "GetTaskByID"

	t.log.LogRequest(op, map[string]interface{}{"id": id})

	if id <= 0 {
		err := errors.New("invalid task id")
		t.log.ErrorWithContext("validation failed", err, op, "task_id", id)
		return nil, err
	}
	task, err := t.repo.GetTaskByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			t.log.Warn("task not found", "function", op, "task_id", id)
			return nil, errors.New("task not found")
		}
		t.log.ErrorWithContext("database error", err, op, "task_id", id)
		return nil, errors.New("internal server error")
	}

	t.log.LogResponse(op, task)

	return task, nil
}

// GetAllTasks возвращает слайс всех задач
func (t *TaskService) GetAllTasks() ([]models.Task, error) {
	const op = "GetAllTasks"
	t.log.LogRequest(op, nil)

	tasks, err := t.repo.GetAllTasks()
	if err != nil {
		t.log.ErrorWithContext("database error", err, op)
		return nil, errors.New("internal server error")
	}

	t.log.LogResponse(op, tasks)

	return tasks, nil
}

// CompleteTask помечает задачу как выполненную
func (t *TaskService) CompleteTask(id int) (*models.Task, error) {
	const op = "CompleteTask"

	t.log.LogRequest(op, map[string]interface{}{"id": id})

	if id <= 0 {
		err := errors.New("invalid task id")
		t.log.ErrorWithContext("validation error", err, op, "task_id", id)
		return nil, err
	}

	task, err := t.repo.GetTaskByID(id)
	if err != nil {
		t.log.ErrorWithContext("task not found", err, op, "task_id", id)
		return nil, err
	}

	if task.Completed {
		err := errors.New("task already completed")
		t.log.ErrorWithContext("failed to complete task", err, op, "task_id", id, "current_status", task.Completed)
		return nil, err
	}

	taskCompleted, err := t.repo.CompleteTask(id)
	if err != nil {
		t.log.ErrorWithContext("failed to complete task", err, op, "task_id", id)
		return nil, err
	}

	t.log.LogResponse(op, taskCompleted)

	return taskCompleted, nil
}

// DeleteTask удаляет задачу по id
func (t *TaskService) DeleteTask(id int) error {
	const op = "DeleteTask"

	t.log.LogRequest(op, map[string]interface{}{"id": id})

	if id <= 0 {
		err := errors.New("invalid id")
		t.log.ErrorWithContext("validation error", err, op, "task_id", id)
		return err
	}
	task, err := t.repo.GetTaskByID(id)
	if err != nil {
		err := errors.New("failed to find task")
		t.log.ErrorWithContext("task not found", err, op, "task_id", id)
		return err
	}

	err = t.repo.DeleteTask(id)
	if err != nil {
		t.log.ErrorWithContext("failed to delete task", err, op, "task_id", id)
		return err
	}

	t.log.LogResponse(op, map[string]interface{}{"deleted": true, "task_id": id, "task_title": task.Title})

	return nil
}
