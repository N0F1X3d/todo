package grpcclient

import (
	"context"
	"fmt"
	"time"

	"github.com/N0F1X3d/todo/pkg/logger"
	pb "github.com/N0F1X3d/todo/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TaskClient struct {
	conn   *grpc.ClientConn
	client pb.TaskServiceClient
	log    *logger.Logger
}

func NewTaskClient(addr string, log *logger.Logger) (*TaskClient, error) {
	const op = "NewTaskClient"
	log = log.WithComponent("grpc-client").WithFunction("TaskClient")
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.ErrorWithContext("failed to create client", err, op)
		return nil, err
	}

	log.Info("grpc clinet created, address:", addr, op)

	return &TaskClient{
		conn:   conn,
		client: pb.NewTaskServiceClient(conn),
		log:    log,
	}, nil
}

func (c *TaskClient) Close() error {
	c.log.Info("closing grpc connection")
	return c.conn.Close()
}

// CreateTask создает новую задачу
func (c *TaskClient) CreateTask(title, description string) (*pb.TaskResponse, error) {
	const op = "CreateTask"

	log := c.log.WithFunction(op)

	log.LogRequest(op, map[string]interface{}{
		"title":       title,
		"description": description,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.CreateTask(ctx, &pb.CreateTaskRequest{
		Title:       title,
		Description: description,
	})
	if err != nil {
		log.ErrorWithContext("failed to create task", err, op)
		return nil, err
	}

	log.LogResponse(op, resp)
	return resp, nil
}

// GetAllTasks получает все задачи
func (c *TaskClient) GetAllTasks() ([]*pb.TaskResponse, error) {
	const op = "GetAllTasks"

	log := c.log.WithFunction(op)

	log.LogRequest(op, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.GetAllTasks(ctx, &pb.GetAllTasksRequest{})
	if err != nil {
		log.ErrorWithContext("failed to get all tasks", err, op)
		return nil, err
	}

	log.LogResponse(op, map[string]interface{}{"tasks_count": len(resp.Tasks)})
	return resp.Tasks, nil
}

// DeleteTask удаляет задачу по ID
func (c *TaskClient) DeleteTask(id int32) error {
	const op = "DeleteTask"

	log := c.log.WithFunction(op)

	log.LogRequest(op, map[string]interface{}{"id": id})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.DeleteTask(ctx, &pb.DeleteTaskRequest{
		Id: id,
	})
	if err != nil {
		log.ErrorWithContext("failed to delete task", err, op)
		return err
	}

	if !resp.Success {
		log.ErrorWithContext("delete task returned false", err, op)
		return fmt.Errorf("failed to delete task")
	}
	log.LogResponse(op, resp)
	return nil
}

// CompleteTask отмечает задачу выполненной
func (c *TaskClient) CompleteTask(id int32) (*pb.TaskResponse, error) {
	const op = "CompleteTask"

	log := c.log.WithFunction(op)

	log.LogRequest(op, map[string]interface{}{"id": id})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.CompleteTask(ctx, &pb.CompleteTaskRequest{
		Id: id,
	})

	if err != nil {
		c.log.ErrorWithContext("failed to complete task", err, op)
		return nil, err
	}

	log.LogResponse(op, resp)
	return resp, nil
}

// GetTaskByID возвращает задачу по ее ID
func (c *TaskClient) GetTaskByID(id int32) (*pb.TaskResponse, error) {
	const op = "GetTaskByID"

	log := c.log.WithFunction(op)

	log.LogRequest(op, map[string]interface{}{"id": id})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.GetTaskByID(ctx, &pb.GetTaskByIDRequest{
		Id: id,
	})

	if err != nil {
		c.log.ErrorWithContext("failed to get task by id", err, op)
		return nil, err
	}

	log.LogResponse(op, resp)
	return resp, nil
}
