package grpcclient

import (
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
