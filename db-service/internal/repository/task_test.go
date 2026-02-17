package repository_test

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/N0F1X3d/todo/db-service/internal/models"
	"github.com/N0F1X3d/todo/db-service/internal/repository"
	"github.com/N0F1X3d/todo/pkg/logger"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	_ "github.com/lib/pq"
)

var (
	testRepo *repository.TaskRepository
	testDB   *sql.DB

	mr  *miniredis.Miniredis
	rdb *redis.Client
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	// Подключение к тестовой БД
	connStr := "user=test_user password=test_password dbname=todo_test host=localhost port=5433 sslmode=disable"

	var err error
	testDB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to test database:", err)
	}

	// Проверка подключения
	if err := testDB.Ping(); err != nil {
		log.Fatal("Failed to ping test database:", err)
	}

	// Поднимаем in-memory Redis для тестов
	mr, err = miniredis.Run()
	if err != nil {
		log.Fatal("Failed to start miniredis:", err)
	}

	rdb = redis.NewClient(&redis.Options{Addr: mr.Addr()})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("Failed to ping redis:", err)
	}

	// Инициализация логгера (репозиторий требует)
	testLogger := logger.New("db-service", "test-logs")

	// Создание репозитория с Redis и TTL (кэш включён)
	testRepo = repository.NewTaskRepository(testDB, testLogger, rdb, 5*time.Minute)

	// Чистим всё перед стартом
	cleanupAll()
}

func teardown() {
	cleanupAll()

	if rdb != nil {
		_ = rdb.Close()
	}
	if mr != nil {
		mr.Close()
	}
	if testDB != nil {
		_ = testDB.Close()
	}
}

func cleanupAll() {
	cleanupDatabase()
	cleanupRedis()
}

func cleanupDatabase() {
	_, err := testDB.Exec("DELETE FROM tasks")
	if err != nil {
		log.Fatal("Failed to clean up database:", err)
	}
}

func cleanupRedis() {
	if rdb == nil {
		return
	}
	_ = rdb.FlushDB(context.Background()).Err()
}

func TestCreateTask(t *testing.T) {
	cleanupAll()

	req := models.CreateTaskRequest{
		Title:       "Test Task",
		Description: "Test Description",
	}

	task, err := testRepo.CreateTask(req)
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	if task.Title != req.Title {
		t.Errorf("Expected title %s, got %s", req.Title, task.Title)
	}
	if task.Description != req.Description {
		t.Errorf("Expected description %s, got %s", req.Description, task.Description)
	}
	if task.Completed != false {
		t.Errorf("Expected completed false, got %v", task.Completed)
	}
	if task.ID == 0 {
		t.Error("Expected non-zero ID")
	}
}

func TestGetTaskByID(t *testing.T) {
	cleanupAll()

	// Сначала создаем задачу
	req := models.CreateTaskRequest{
		Title:       "Get Test Task",
		Description: "Get Test Description",
	}
	createdTask, err := testRepo.CreateTask(req)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Потом получаем её (может вернуться из кэша, т.к. CreateTask кэширует)
	task, err := testRepo.GetTaskByID(createdTask.ID)
	if err != nil {
		t.Fatalf("GetTaskByID failed: %v", err)
	}

	if task.ID != createdTask.ID {
		t.Errorf("Expected ID %d, got %d", createdTask.ID, task.ID)
	}
	if task.Title != req.Title {
		t.Errorf("Expected title %s, got %s", req.Title, task.Title)
	}
}

func TestGetTaskByID_NotFound(t *testing.T) {
	cleanupAll()

	_, err := testRepo.GetTaskByID(99999) // Несуществующий ID
	if err == nil {
		t.Error("Expected error for non-existent task, got nil")
	}
	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
}

func TestGetAllTasks(t *testing.T) {
	cleanupAll()

	// Создаем несколько задач
	tasksToCreate := []models.CreateTaskRequest{
		{Title: "Task 1", Description: "Desc 1"},
		{Title: "Task 2", Description: "Desc 2"},
	}

	for _, req := range tasksToCreate {
		_, err := testRepo.CreateTask(req)
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}
	}

	// Получаем все задачи
	tasks, err := testRepo.GetAllTasks()
	if err != nil {
		t.Fatalf("GetAllTasks failed: %v", err)
	}

	if len(tasks) != len(tasksToCreate) {
		t.Errorf("Expected %d tasks, got %d", len(tasksToCreate), len(tasks))
	}
}

func TestCompleteTask(t *testing.T) {
	cleanupAll()

	// Создаем задачу
	req := models.CreateTaskRequest{
		Title:       "Complete Test Task",
		Description: "Complete Test Description",
	}
	createdTask, err := testRepo.CreateTask(req)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Чтобы updated_at точно изменился (иногда может совпасть по точности времени)
	time.Sleep(5 * time.Millisecond)

	// Отмечаем как выполненную
	completedTask, err := testRepo.CompleteTask(createdTask.ID)
	if err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}

	if !completedTask.Completed {
		t.Error("Expected task to be completed")
	}

	// Более корректная проверка, чем Equal(): time может быть с разной точностью/округлением
	if !completedTask.UpdatedAt.After(createdTask.UpdatedAt) {
		t.Errorf("Expected updated_at to be after previous updated_at. before=%v after=%v",
			createdTask.UpdatedAt, completedTask.UpdatedAt)
	}
}

func TestDeleteTask(t *testing.T) {
	cleanupAll()

	// Создаем задачу
	req := models.CreateTaskRequest{
		Title:       "Delete Test Task",
		Description: "Delete Test Description",
	}
	createdTask, err := testRepo.CreateTask(req)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Удаляем задачу
	err = testRepo.DeleteTask(createdTask.ID)
	if err != nil {
		t.Fatalf("DeleteTask failed: %v", err)
	}

	// Проверяем что задача удалена (и из БД, и из кэша)
	_, err = testRepo.GetTaskByID(createdTask.ID)
	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows after delete, got %v", err)
	}
}

// Дополнительный тест: проверяем, что кэш действительно отдаёт данные
// даже если строку в БД удалить напрямую (то есть проверяем интеграцию с Redis).
func TestGetTaskByID_UsesCache_WhenRowDeleted(t *testing.T) {
	cleanupAll()

	req := models.CreateTaskRequest{
		Title:       "Cache Test Task",
		Description: "Cache Test Description",
	}

	createdTask, err := testRepo.CreateTask(req)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// На всякий случай дернем GetTaskByID (если вдруг в будущем CreateTask перестанет кэшировать)
	_, err = testRepo.GetTaskByID(createdTask.ID)
	if err != nil {
		t.Fatalf("GetTaskByID failed: %v", err)
	}

	// Удаляем строку из БД напрямую (имитируем ситуацию, когда БД "не содержит" запись)
	_, err = testDB.Exec("DELETE FROM tasks WHERE id = $1", createdTask.ID)
	if err != nil {
		t.Fatalf("Failed to delete task row directly: %v", err)
	}

	// Если кэш работает — задача вернётся из Redis, несмотря на отсутствие в БД
	task, err := testRepo.GetTaskByID(createdTask.ID)
	if err != nil {
		t.Fatalf("Expected task from cache, got error: %v", err)
	}
	if task.ID != createdTask.ID {
		t.Fatalf("Expected ID %d, got %d", createdTask.ID, task.ID)
	}
	if task.Title != req.Title {
		t.Fatalf("Expected title %q, got %q", req.Title, task.Title)
	}
	if task.Description != req.Description {
		t.Fatalf("Expected description %q, got %q", req.Description, task.Description)
	}
}
