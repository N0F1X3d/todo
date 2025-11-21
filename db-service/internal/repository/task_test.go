package repository_test

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/N0F1X3d/todo/db-service/internal/models"
	"github.com/N0F1X3d/todo/db-service/internal/repository"
	"github.com/N0F1X3d/todo/db-service/pkg/logger"

	_ "github.com/lib/pq"
)

var testRepo *repository.TaskRepository
var testDB *sql.DB

func TestMain(m *testing.M) {
	// Настройка тестовой БД
	setup()

	// Запуск тестов
	code := m.Run()

	// Очистка
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

	// Инициализация логгера (логи в тестах не нужны, но репозиторий требует)
	testLogger := logger.New("test-logs")

	// Создание репозитория
	testRepo = repository.NewTaskRepository(testDB, testLogger)

	// Очистка таблицы перед тестами
	cleanupDatabase()
}

func teardown() {
	cleanupDatabase()
	if testDB != nil {
		testDB.Close()
	}
}

func cleanupDatabase() {
	_, err := testDB.Exec("DELETE FROM tasks")
	if err != nil {
		log.Fatal("Failed to clean up database:", err)
	}
}

func TestCreateTask(t *testing.T) {
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
	// Сначала создаем задачу
	req := models.CreateTaskRequest{
		Title:       "Get Test Task",
		Description: "Get Test Description",
	}
	createdTask, err := testRepo.CreateTask(req)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Потом получаем её
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
	_, err := testRepo.GetTaskByID(99999) // Несуществующий ID

	if err == nil {
		t.Error("Expected error for non-existent task, got nil")
	}
	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
}

func TestGetAllTasks(t *testing.T) {
	// Очищаем базу
	cleanupDatabase()

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

func TestDoneTask(t *testing.T) {
	// Создаем задачу
	req := models.CreateTaskRequest{
		Title:       "Complete Test Task",
		Description: "Complete Test Description",
	}
	createdTask, err := testRepo.CreateTask(req)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Отмечаем как выполненную
	completedTask, err := testRepo.DoneTask(createdTask.ID)
	if err != nil {
		t.Fatalf("DoneTask failed: %v", err)
	}

	if !completedTask.Completed {
		t.Error("Expected task to be completed")
	}
	if completedTask.UpdatedAt == createdTask.UpdatedAt {
		t.Error("Expected updated_at to change after completion")
	}
}

func TestDeleteTask(t *testing.T) {
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

	// Проверяем что задача удалена
	_, err = testRepo.GetTaskByID(createdTask.ID)
	if err != sql.ErrNoRows {
		t.Error("Expected task to be deleted, but it still exists")
	}
}
