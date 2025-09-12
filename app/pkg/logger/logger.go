package logger

import (
	"io"
	"log"
	"os"
	"runtime/debug"
	"strings"
)

func InitFile(logFilePath string) {
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	multiWriter := io.MultiWriter(logFile, os.Stdout)
	log.SetOutput(multiWriter)
}

func Debug(text ...string) {
	log.Printf("[DEBUG] %s at %s", strings.Join(text, ", "), string(debug.Stack()))
}
