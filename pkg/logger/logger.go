package logger

import (
	"log/slog"
	"os"
	"path/filepath"
)

type Logger struct {
	*slog.Logger
	serviceName string
}

// New создает логгер
func New(serviceName string, logDir string) *Logger {

	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic("failed to create log directory for: " + err.Error())
	}

	logFile, err := os.OpenFile(
		filepath.Join(logDir, serviceName+".log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666,
	)
	if err != nil {
		panic("failed to open log file: " + err.Error())
	}

	handler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{Level: slog.LevelDebug})

	baseLogger := slog.New(handler)
	logger := baseLogger.With("service", serviceName)
	return &Logger{
		Logger:      logger,
		serviceName: serviceName,
	}
}

// WithComponent добавляет компонент к логгеру (repository, service, server)
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		Logger:      l.Logger.With("component", component),
		serviceName: l.serviceName,
	}
}

// WithFunction добавляет функцию к логгеру
func (l *Logger) WithFunction(function string) *Logger {
	return &Logger{
		Logger:      l.Logger.With("function", function),
		serviceName: l.serviceName,
	}
}

// ErrorWithContext логирует ошибку с контекстом
func (l *Logger) ErrorWithContext(msg string, err error, function string, additionalFields ...any) {
	fields := append([]any{"error", err.Error(), "function", function}, additionalFields...)
	l.Error(msg, fields...)
}

// LogRequest логирует входящий запрос
func (l *Logger) LogRequest(function string, request interface{}) {
	l.Info("incoming request", "function", function, "request", request)
}

// LogResponse логирует исходящий ответ
func (l *Logger) LogResponse(function string, response interface{}) {
	l.Info("outgoing response", "function", function, "response", response)
}

func (l *Logger) Fatal(msg string, fields ...any) {
	l.Error(msg, fields...)
	os.Exit(1)
}
