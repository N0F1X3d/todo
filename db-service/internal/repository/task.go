package repository

import (
	"database/sql"
	"time"

	"github.com/N0F1X3d/todo/db-service/internal/models"
	"github.com/N0F1X3d/todo/db-service/pkg/logger"
)

// TaskRepository предоставляет методы для работы с PostgreSQL
// Реализует паттерн Repository для абстракции доступа к данным
type TaskRepository struct {
	db  *sql.DB
	log *logger.Logger
}

func NewTaskRepository(db *sql.DB, log *logger.Logger) *TaskRepository {
	return &TaskRepository{
		db:  db,
		log: log.WithComponent("repository").WithFunction("TaskRepository"),
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

	r.log.LogQuery(op, query, req.Title, req.Description)

	err := r.db.QueryRow(query, req.Title, req.Description).Scan(
		&task.ID, &task.Title, &task.Description, &task.Completed, &task.CreatedAt, &task.UpdatedAt,
	)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		r.log.ErrorWithContext("failed to create task", err, op, "title", req.Title, "description", req.Description, "duration", duration)
		return nil, err
	}
	r.log.LogResponse(op, task)
	r.log.LogQueryResult(op, duration, 1)
	return &task, nil
}

// GetTaskById возвращает задачу по ее id
func (r *TaskRepository) GetTaskByID(id int) (*models.Task, error) {
	const op = "GetTaskByID"
	r.log.LogRequest(op, map[string]interface{}{"id": id})
	start := time.Now()

	var task models.Task

	query := `SELECT id, title, description, completed, created_at, updated_at
			  FROM tasks WHERE id = $1`
	r.log.LogQuery(op, query, id)

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

	r.log.LogResponse(op, task)
	r.log.LogQueryResult(op, duration, 1)
	return &task, nil
}

// GetAllTasks возвращает список всех задач
func (r *TaskRepository) GetAllTasks() ([]models.Task, error) {
	const op = "GetAllTasks"
	r.log.LogRequest(op, nil)

	start := time.Now()

	query := `SELECT id, title, description, completed, created_at, updated_at FROM tasks`

	r.log.LogQuery(op, query)

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
	r.log.LogQueryResult(op, duration, int64(len(tasks)))
	return tasks, nil
}

// UpdateTask обновляет задачу в базе данных по переданному id задачи completed -> true
func (r *TaskRepository) DoneTask(id int) (*models.Task, error) {
	const op = "UpdateTask"
	r.log.LogRequest(op, map[string]interface{}{"id": id})
	start := time.Now()

	var task models.Task
	query := `UPDATE tasks
			  SET completed = true, updated_at = CURRENT_TIMESTAMP
			  WHERE id = $1
			  RETURNING id, title, description, completed, created_at, updated_at`
	r.log.LogQuery(op, query, id)

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

	r.log.LogResponse(op, task)
	r.log.LogQueryResult(op, duration, 1)
	return &task, nil
}

// DeleteTask удаляет задачу из базы данных по ее id
func (r *TaskRepository) DeleteTask(id int) error {
	const op = "DeleteTask"
	r.log.LogRequest(op, map[string]interface{}{"id": id})
	start := time.Now()

	query := `DELETE FROM tasks WHERE id = $1`

	r.log.LogQuery(op, query, id)
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

	r.log.LogResponse(op, map[string]interface{}{"deleted": true, "id": id})
	r.log.LogQueryResult(op, duration, rowsAffected)
	return nil
}
