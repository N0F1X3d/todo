package kafka

import "time"

type TaskEvent struct {
	Action        string    `json:"action"`
	DBRequestTime time.Time `json:"db_request_time"`
}
