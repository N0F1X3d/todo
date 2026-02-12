package dto

import (
	"errors"
	"strings"

	pb "github.com/N0F1X3d/todo/pkg/proto"
)

// CreateTaskRequest - запрос на создание задачи
type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Validate проверяет корректность запроса
func (r *CreateTaskRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("title is required")
	}
	if len(r.Title) > 255 {
		return errors.New("title too long, maximum 255 characters")
	}
	if len(r.Description) > 1000 {
		return errors.New("description too long, maximum 1000 characters")
	}
	return nil
}

// ToProto конвертирует в protobuf сообщение
func (r *CreateTaskRequest) ToProto() *pb.CreateTaskRequest {
	return &pb.CreateTaskRequest{
		Title:       r.Title,
		Description: r.Description,
	}
}

// DeleteTaskRequest - запрос на удаление задачи
type DeleteTaskRequest struct {
	ID int32 `json:"id"`
}

// Validate проверяет корректность запроса
func (r *DeleteTaskRequest) Validate() error {
	if r.ID <= 0 {
		return errors.New("id must be positive integer")
	}
	return nil
}

// CompleteTaskRequest - запрос на выполнение задачи
type CompleteTaskRequest struct {
	ID int32 `json:"id"`
}

// Validate проверяет корректность запроса
func (r *CompleteTaskRequest) Validate() error {
	if r.ID <= 0 {
		return errors.New("id must be positive integer")
	}
	return nil
}
