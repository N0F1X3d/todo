package repository

import "github.com/N0F1X3d/todo/pkg/logger"

func logQuery(log *logger.Logger, function string, query string, args ...any) {
	log.Debug("sql query",
		"function", function,
		"query", query,
		"args", args,
	)
}

func logQueryResult(log *logger.Logger, function string, duration int64, rows int64) {
	log.Debug("query result",
		"function", function,
		"duration_ms", duration,
		"rows_affected", rows,
	)
}
