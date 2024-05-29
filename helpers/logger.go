package helpers

import (
	"net/http"
	"sync"

	"github.com/threatwinds/logger"
)

var loggerInstance *logger.Logger
var loggerOnce sync.Once

func NewLogger(level int) *logger.Logger {
	loggerOnce.Do(func() {
		loggerInstance = logger.NewLogger(&logger.Config{
			Level:   level,
			Format:  "text",
			Retries: 3,
			Wait:    5,
			StatusMap: map[int][]string{
				100:                       {"node not found"},
				http.StatusGatewayTimeout: {"timeout"},
			},
		})
	})

	return loggerInstance
}

func Logger() *logger.Logger {
	return loggerInstance
}
