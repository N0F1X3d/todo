package dto

import (
	pb "github.com/N0F1X3d/todo/pkg/proto"
)

// TaskResponse - ответ с информацией о задаче
type TaskResponse struct {
	ID          int32  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// TaskResponseFromProto создает DTO из protobuf сообщения
func TaskResponseFromProto(protoTask *pb.TaskResponse) *TaskResponse {
	if protoTask == nil {
		return nil
	}

	return &TaskResponse{
		ID:          protoTask.Id,
		Title:       protoTask.Title,
		Description: protoTask.Description,
		Completed:   protoTask.Completed,
		CreatedAt:   protoTask.CreatedAt,
		UpdatedAt:   protoTask.UpdatedAt,
	}
}

// TaskListResponse - список задач
type TaskListResponse []*TaskResponse

// TaskListResponseFromProto создает список DTO из protobuf сообщения
func TaskListResponseFromProto(protoTasks []*pb.TaskResponse) TaskListResponse {
	if protoTasks == nil {
		return TaskListResponse{}
	}

	tasks := make([]*TaskResponse, 0, len(protoTasks))
	for _, protoTask := range protoTasks {
		tasks = append(tasks, TaskResponseFromProto(protoTask))
	}
	return tasks
}

// DeleteTaskResponse - ответ на удаление задачи
type DeleteTaskResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// CompleteTaskResponse - ответ на выполнение задачи
type CompleteTaskResponse struct {
	ID        int32  `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
	Message   string `json:"message,omitempty"`
}

// CompleteTaskResponseFromProto создает DTO из protobuf сообщения
func CompleteTaskResponseFromProto(protoTask *pb.TaskResponse) *CompleteTaskResponse {
	if protoTask == nil {
		return nil
	}

	return &CompleteTaskResponse{
		ID:        protoTask.Id,
		Title:     protoTask.Title,
		Completed: protoTask.Completed,
		Message:   "Task marked as done",
	}
}
