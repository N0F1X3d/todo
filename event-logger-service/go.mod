module github.com/N0F1X3d/todo/event-logger-service

go 1.25.3

require github.com/N0F1X3d/todo/pkg v0.0.0

require (
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/segmentio/kafka-go v0.4.50 // indirect
)

replace github.com/N0F1X3d/todo/pkg => ../pkg
