package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/N0F1X3d/todo/pkg/kafka"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 📁 создаём файл логов
	logFile, err := os.OpenFile(
		"/logs/events.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666,
	)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	// graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	brokers := []string{"kafka:9092"}
	topic := "task-events"
	groupID := "event-logger-group"

	consumer := kafka.NewConsumer(brokers, topic, groupID)
	defer consumer.Close()

	log.Println("event-logger-service started...")

	consumer.Start(ctx, func(event kafka.TaskEvent) error {
		log.Printf("EVENT: %+v\n", event)
		return nil
	})
}
