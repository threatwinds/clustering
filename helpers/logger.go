package helpers

import (
	"github.com/threatwinds/logger"
)

var Logger = logger.NewLogger(&logger.Config{
	Format: "json",
	Level:  200,
	Retries: 3,
	Wait: 1,
	Output: "stdout",
	StatusMap: map[int][]string{
		100: {"node not found"},
	},
})