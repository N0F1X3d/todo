package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/N0F1X3d/todo/api-service/internal/clients/grpcclient"
	"github.com/N0F1X3d/todo/api-service/internal/dto"
	"github.com/N0F1X3d/todo/pkg/kafka"
	"github.com/N0F1X3d/todo/pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TaskHandler struct {
	grpcClient *grpcclient.TaskClient
	producer   *kafka.Producer
	log        *logger.Logger
}

func NewTaskHandler(client *grpcclient.TaskClient, producer *kafka.Producer, log *logger.Logger) *TaskHandler {
	return &TaskHandler{
		grpcClient: client,
		producer:   producer,
		log:        log,
	}
}

// POST /create
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	const op = "CreateTask"
	ctx := r.Context()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dbRequestTime := time.Now()

	task, err := h.grpcClient.CreateTask(ctx, req.Title, req.Description)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	resp := dto.TaskResponseFromProto(task)

	event := kafka.TaskEvent{
		Action:        "create-task",
		DBRequestTime: dbRequestTime,
	}
	if err := h.producer.Send(ctx, "create", event); err != nil {
		h.log.ErrorWithContext("failed to send event", err, op)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GET /list
func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	const op = "ListTasks"
	ctx := r.Context()

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dbRequestTime := time.Now()

	tasks, err := h.grpcClient.GetAllTasks(ctx)
	if err != nil {
		http.Error(w, "Failed to get tasks", http.StatusInternalServerError)
		return
	}

	resp := dto.TaskListResponseFromProto(tasks)

	event := kafka.TaskEvent{
		Action:        "list-tasks",
		DBRequestTime: dbRequestTime,
	}
	if err := h.producer.Send(ctx, "list", event); err != nil {
		h.log.ErrorWithContext("failed to send event", err, op)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DELETE /delete
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	const op = "DeleteTask"
	ctx := r.Context()

	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.DeleteTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dbRequestTime := time.Now()

	err := h.grpcClient.DeleteTask(ctx, req.ID)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	event := kafka.TaskEvent{
		Action:        "delete-task",
		DBRequestTime: dbRequestTime,
	}
	if err := h.producer.Send(ctx, "delete", event); err != nil {
		h.log.ErrorWithContext("failed to send event", err, op)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Task deleted successfully",
	})
}

// PUT /done
func (h *TaskHandler) CompleteTask(w http.ResponseWriter, r *http.Request) {
	const op = "CompleteTask"
	ctx := r.Context()

	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.CompleteTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dbRequestTime := time.Now()

	task, err := h.grpcClient.CompleteTask(ctx, req.ID)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	resp := dto.TaskResponseFromProto(task)

	event := kafka.TaskEvent{
		Action:        "complete-task",
		DBRequestTime: dbRequestTime,
	}
	if err := h.producer.Send(ctx, "complete", event); err != nil {
		h.log.ErrorWithContext("failed to send event", err, op)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Общая обработка gRPC ошибок
func handleGrpcError(w http.ResponseWriter, err error) {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.InvalidArgument:
			http.Error(w, st.Message(), http.StatusBadRequest)
		case codes.NotFound:
			http.Error(w, st.Message(), http.StatusNotFound)
		case codes.FailedPrecondition:
			http.Error(w, st.Message(), http.StatusConflict)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	http.Error(w, "Internal server error", http.StatusInternalServerError)
}
