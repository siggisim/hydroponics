package main

import (
	"os"

	"github.com/zenreach/hatchet"
)

const logBuffer = 10

func newLogger(level string) hatchet.Logger {
	logger := hatchet.Buffer(hatchet.JSON(os.Stdout), logBuffer)
	minValue := hatchet.LevelValue(level)
	if minValue > 0 {
		logger = hatchet.Filter(logger, func(log map[string]interface{}) bool {
			l := hatchet.L(log)
			return l.LevelValue() >= minValue
		})
	}
	return logger
}
