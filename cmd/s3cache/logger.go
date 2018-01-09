package main

import (
	"os"

	"github.com/zenreach/hatchet"
)

const logBuffer = 10

func newLogger() hatchet.Logger {
	return hatchet.Buffer(hatchet.JSON(os.Stdout), logBuffer)
}
