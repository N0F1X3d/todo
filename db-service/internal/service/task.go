package service

import (
	"errors"
	"strings"
	"time"

	"github.com/N0F1X3d/todo/db-service/internal/models"
	"github.com/N0F1X3d/todo/db-service/internal/repository"
	"github.com/N0F1X3d/todo/db-service/pkg/logger"
)

// TaskService предоставляет бизнес-логику для работы с задачами.
type TaskService struct {
	repo *repository.TaskRepository
	log  *logger.Logger
}

// NewTaskService создает новый экземпляр TaskService
func NewTaskService(repo *repository.TaskRepository, log *logger.Logger) *TaskService {
	return &TaskService{
		repo: repo,
		log:  log.WithComponent("service").WithFunction("TaskService"),
	}
}

// CreateTask создает новую задачу с применением бизнес-логики и валидации
func (t *TaskService) CreateTask(req models.CreateTaskRequest) (*models.Task, error) {
	const op = "CreateTask"

	t.log.LogRequest(op, req)
	start := time.Now()

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
	duration := time.Since(start).Microseconds()

	t.log.LogResponse(op, task)
	t.log.LogQueryResult(op, duration, 1)
	return task, nil
}

// GetTaskByID возвращает задачу по ее ID
func (t *TaskService) GetTaskByID(id int) (*models.Task, error) {
	const op = "GetTaskByID"

	t.log.LogRequest(op, map[string]interface{}{"id": id})

	start := time.Now()
	if id <= 0 {
		err := errors.New("invalid task id")
		t.log.ErrorWithContext("validation failed", err, op, "task_id", id)
		return nil, err
	}
	task, err := t.repo.GetTaskByID(id)
	if err != nil {
		t.log.ErrorWithContext("failed to get task", err, op)
		return nil, err
	}

	duration := time.Since(start).Milliseconds()

	t.log.LogResponse(op, task)
	t.log.LogQueryResult(op, duration, 1)

	return task, nil
}

// GetAllTasks возвращает слайс всех задач
func (t *TaskService) GetAllTasks() ([]models.Task, error) {
	const op = "GetAllTasks"
	t.log.LogRequest(op, nil)

	start := time.Now()

	tasks, err := t.repo.GetAllTasks()
	if err != nil {
		t.log.ErrorWithContext("failed to get tasks", err, op)
		return nil, err
	}

	duration := time.Since(start).Milliseconds()

	t.log.LogResponse(op, tasks)
	t.log.LogQueryResult(op, duration, int64(len(tasks)))

	return tasks, nil
}

// CompleteTask помечает задачу как выполненную
func (t *TaskService) CompleteTask(id int) (*models.Task, error) {
	const op = "CompleteTask"

	t.log.LogRequest(op, map[string]interface{}{"id": id})

	start := time.Now()

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
	}

	duration := time.Since(start).Milliseconds()

	t.log.LogResponse(op, taskCompleted)
	t.log.LogQueryResult(op, duration, 1)
	return taskCompleted, nil
}

// DeleteTask удаляет задачу по id
func (t *TaskService) DeleteTask(id int) error {
	const op = "DeleteTask"

	t.log.LogRequest(op, map[string]interface{}{"id": id})

	start := time.Now()
	if id <= 0 {
		err := errors.New("invalid id")
		t.log.ErrorWithContext("validation error", err, op, "task_id", id)
		return err
	}
	task, err := t.repo.GetTaskByID(id)
	if err != nil {
		err := errors.New("failed to find task")
		t.log.ErrorWithContext("task not found", err, op, "task_id", id)
	}

	err = t.repo.DeleteTask(id)
	if err != nil {
		t.log.ErrorWithContext("failed to delete task", err, op, "task_id", id)
		return err
	}

	duration := time.Since(start).Milliseconds()

	t.log.LogResponse(op, map[string]interface{}{"deleted": true, "task_id": id, "task_title": task.Title})
	t.log.LogQueryResult(op, duration, 1)
	return nil
}
