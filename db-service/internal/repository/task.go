package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/N0F1X3d/todo/db-service/internal/models"
	"github.com/N0F1X3d/todo/pkg/logger"
	"github.com/redis/go-redis/v9"
)

//go:generate mockery --name=TaskRepositoryInterface --filename=task_repository_interface.go --output=../../mocks --case=underscore
type TaskRepositoryInterface interface {
	CreateTask(req models.CreateTaskRequest) (*models.Task, error)
	GetTaskByID(id int) (*models.Task, error)
	GetAllTasks() ([]models.Task, error)
	CompleteTask(id int) (*models.Task, error)
	DeleteTask(id int) error
}

// TaskRepository предоставляет методы для работы с PostgreSQL
// Реализует паттерн Repository для абстракции доступа к данным
type TaskRepository struct {
	db          *sql.DB
	log         *logger.Logger
	redisClient *redis.Client
	cacheTTL    time.Duration
}

func NewTaskRepository(db *sql.DB, log *logger.Logger, redisClient *redis.Client, cacheTTL time.Duration) *TaskRepository {
	return &TaskRepository{
		db:          db,
		log:         log.WithComponent("repository").WithFunction("TaskRepository"),
		redisClient: redisClient,
		cacheTTL:    cacheTTL,
	}
}

func (r *TaskRepository) cacheKey(id int) string {
	return fmt.Sprintf("task:%d", id)
}

func (r *TaskRepository) cacheEnabled() bool {
	return r != nil && r.redisClient != nil && r.cacheTTL > 0
}

func (r *TaskRepository) setTaskCache(ctx context.Context, task *models.Task) {
	if !r.cacheEnabled() || task == nil {
		return
	}

	data, err := json.Marshal(task)
	if err != nil {
		r.log.Warn("failed to marshal task for cache", "function", "setTaskCache", "task_id", task.ID, "error", err)
		return
	}

	if err := r.redisClient.Set(ctx, r.cacheKey(task.ID), data, r.cacheTTL).Err(); err != nil {
		r.log.Warn("failed to set task cache", "function", "setTaskCache", "task_id", task.ID, "error", err)
	}
}

func (r *TaskRepository) getTaskFromCache(ctx context.Context, id int) (*models.Task, bool) {
	if !r.cacheEnabled() {
		return nil, false
	}

	res, err := r.redisClient.Get(ctx, r.cacheKey(id)).Result()
	if err != nil {
		if err != redis.Nil {
			r.log.Warn("failed to get task from cache", "function", "getTaskFromCache", "task_id", id, "error", err)
		}
		return nil, false
	}

	var task models.Task
	if err := json.Unmarshal([]byte(res), &task); err != nil {
		r.log.Warn("failed to unmarshal task from cache", "function", "getTaskFromCache", "task_id", id, "error", err)
		return nil, false
	}

	return &task, true
}

func (r *TaskRepository) deleteTaskCache(ctx context.Context, id int) {
	if !r.cacheEnabled() {
		return
	}

	if err := r.redisClient.Del(ctx, r.cacheKey(id)).Err(); err != nil {
		r.log.Warn("failed to delete task cache", "function", "deleteTaskCache", "task_id", id, "error", err)
	}
}

// CreateTask создает новую задачу в базе данных
func (r *TaskRepository) CreateTask(req models.CreateTaskRequest) (*models.Task, error) {
	const op = "CreateTask"
	r.log.LogRequest(op, req)
	start := time.Now()

	var task models.Task

	query := `INSERT INTO tasks (title, description) VALUES ($1, $2)
			  RETURNING id, title, description, completed, created_at, updated_at`

	logQuery(r.log, op, query, req.Title, req.Description)

	err := r.db.QueryRow(query, req.Title, req.Description).Scan(
		&task.ID, &task.Title, &task.Description, &task.Completed, &task.CreatedAt, &task.UpdatedAt,
	)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		r.log.ErrorWithContext("failed to create task", err, op, "title", req.Title, "description", req.Description, "duration", duration)
		return nil, err
	}

	// Кэшируем только что созданную задачу
	r.setTaskCache(context.Background(), &task)

	r.log.LogResponse(op, task)
	logQueryResult(r.log, op, duration, 1)
	return &task, nil
}

// GetTaskById возвращает задачу по ее id
func (r *TaskRepository) GetTaskByID(id int) (*models.Task, error) {
	const op = "GetTaskByID"
	r.log.LogRequest(op, map[string]interface{}{"id": id})
	start := time.Now()

	// Сначала пробуем получить задачу из кеша
	if taskFromCache, ok := r.getTaskFromCache(context.Background(), id); ok {
		duration := time.Since(start).Milliseconds()
		r.log.LogResponse(op, taskFromCache)
		logQueryResult(r.log, op, duration, 1)
		return taskFromCache, nil
	}

	var task models.Task

	query := `SELECT id, title, description, completed, created_at, updated_at
			  FROM tasks WHERE id = $1`
	logQuery(r.log, op, query, id)

	err := r.db.QueryRow(query, id).Scan(
		&task.ID, &task.Title, &task.Description, &task.Completed, &task.CreatedAt, &task.UpdatedAt,
	)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		if err == sql.ErrNoRows {
			r.log.Warn("task not found", "function", op, "id", id, "duration", duration)
		} else {
			r.log.ErrorWithContext("failed to get task", err, op, "id", id, "duration", duration)
		}
		return nil, err
	}

	// Обновляем кеш после успешного чтения из БД
	r.setTaskCache(context.Background(), &task)

	r.log.LogResponse(op, task)
	logQueryResult(r.log, op, duration, 1)
	return &task, nil
}

// GetAllTasks возвращает список всех задач
func (r *TaskRepository) GetAllTasks() ([]models.Task, error) {
	const op = "GetAllTasks"
	r.log.LogRequest(op, nil)

	start := time.Now()

	query := `SELECT id, title, description, completed, created_at, updated_at FROM tasks`

	logQuery(r.log, op, query)

	rows, err := r.db.Query(query)
	if err != nil {
		r.log.ErrorWithContext("failed to get all tasks", err, op)
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Completed, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			r.log.ErrorWithContext("failed to scan task", err, op)
			return nil, err
		}
		tasks = append(tasks, task)
	}

	duration := time.Since(start).Milliseconds()
	r.log.LogResponse(op, tasks)
	logQueryResult(r.log, op, duration, int64(len(tasks)))
	return tasks, nil
}

// UpdateTask обновляет задачу в базе данных по переданному id задачи completed -> true
func (r *TaskRepository) CompleteTask(id int) (*models.Task, error) {
	const op = "UpdateTask"
	r.log.LogRequest(op, map[string]interface{}{"id": id})
	start := time.Now()

	var task models.Task
	query := `UPDATE tasks
			  SET completed = true, updated_at = CURRENT_TIMESTAMP
			  WHERE id = $1
			  RETURNING id, title, description, completed, created_at, updated_at`
	logQuery(r.log, op, query, id)

	err := r.db.QueryRow(query, id).Scan(
		&task.ID, &task.Title, &task.Description, &task.Completed, &task.CreatedAt, &task.UpdatedAt,
	)
	duration := time.Since(start).Milliseconds()
	if err != nil {
		if err == sql.ErrNoRows {
			r.log.Warn("task not found", "function", op, "id", id, "duration", duration)
		} else {
			r.log.ErrorWithContext("failed to complete task", err, op, "id", id, "duration", duration)
		}
		return nil, err
	}

	// Обновляем кеш завершенной задачи (или добавляем, если ее не было)
	r.setTaskCache(context.Background(), &task)

	r.log.LogResponse(op, task)
	logQueryResult(r.log, op, duration, 1)
	return &task, nil
}

// DeleteTask удаляет задачу из базы данных по ее id
func (r *TaskRepository) DeleteTask(id int) error {
	const op = "DeleteTask"
	r.log.LogRequest(op, map[string]interface{}{"id": id})
	start := time.Now()

	query := `DELETE FROM tasks WHERE id = $1`

	logQuery(r.log, op, query, id)
	res, err := r.db.Exec(query, id)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		r.log.ErrorWithContext("failed to delete task", err, op, "id", id, "duration", duration)
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		r.log.ErrorWithContext("failed to get rows affected", err, op, "id", id, "duration", duration)
		return err
	}
	if rowsAffected == 0 {
		r.log.Warn("task not found for delete", "function", op, "id", id, "duration", duration)
		return sql.ErrNoRows
	}

	// Удаляем задачу из кеша
	r.deleteTaskCache(context.Background(), id)

	r.log.LogResponse(op, map[string]interface{}{"deleted": true, "id": id})
	logQueryResult(r.log, op, duration, rowsAffected)
	return nil
}
