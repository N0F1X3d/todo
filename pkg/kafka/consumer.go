package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
	}
}

func (c *Consumer) Start(ctx context.Context, handler func(TaskEvent) error) {
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("error reading message: %v", err)
			continue
		}

		var event TaskEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("error unmarshalling event: %v", err)
			continue
		}

		log.Printf("message received: key=%s partition=%d offset=%d",
			string(msg.Key), msg.Partition, msg.Offset)

		if err := handler(event); err != nil {
			log.Printf("handler error: %v", err)
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
