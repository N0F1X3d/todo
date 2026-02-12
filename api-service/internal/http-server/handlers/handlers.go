package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/N0F1X3d/todo/api-service/internal/clients/grpcclient"
	"github.com/N0F1X3d/todo/api-service/internal/dto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TaskHandler struct {
	grpcClient *grpcclient.TaskClient
}

func NewTaskHandler(client *grpcclient.TaskClient) *TaskHandler {
	return &TaskHandler{
		grpcClient: client,
	}
}

// POST /create
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
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

	task, err := h.grpcClient.CreateTask(req.Title, req.Description)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	resp := dto.TaskResponseFromProto(task)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GET /list
func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tasks, err := h.grpcClient.GetAllTasks()
	if err != nil {
		http.Error(w, "Failed to get tasks", http.StatusInternalServerError)
		return
	}

	resp := dto.TaskListResponseFromProto(tasks)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// DELETE /delete
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
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

	err := h.grpcClient.DeleteTask(req.ID)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Task deleted successfully",
	})
}

// PUT /done
func (h *TaskHandler) CompleteTask(w http.ResponseWriter, r *http.Request) {
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

	task, err := h.grpcClient.CompleteTask(req.ID)
	if err != nil {
		handleGrpcError(w, err)
		return
	}

	resp := dto.TaskResponseFromProto(task)

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
