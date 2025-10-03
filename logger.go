package wiston

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Logger struct {
	target string
	file   *os.File
}

// NewLogger creates a new Logger that writes to stdout if the target is "stdout", otherwise it writes to the given file path.
func NewLogger(target string) *Logger {
	l := &Logger{target: target}

	if target != "stdout" {
		f, err := os.OpenFile(target, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Can't open log file %s: %v", target, err)
		}
		l.file = f
	}

	return l
}

// Info logs a message at info level.
func (l *Logger) Info(msg string) {
	l.write("INFO", msg)
}

// Success logs a message at success level.
func (l *Logger) Success(msg string) {
	l.write("SUCCESS", msg)
}

// Warn logs a message at warn level.
func (l *Logger) Warn(msg string) {
	l.write("WARN", msg)
}

// Error logs a message at error level.
func (l *Logger) Error(msg string) {
	l.write("ERROR", msg)
}

// write logs a message with level.
func (l *Logger) write(level, msg string) {
	finalMsg := fmt.Sprintf("%s [%s] %s\n", time.Now().Format(time.RFC3339), level, msg)
	if l.target == "stdout" {
		fmt.Print(finalMsg)
	} else {
		if l.file != nil {
			if _, err := l.file.WriteString(finalMsg); err != nil {
				log.Printf("failed to write log: %v", err)
			}
		}

	}
}

// Close closes the underlying file if the logger writes to a file.
func (l *Logger) Close() {
	if l.file != nil {
		l.file.Close()
		l.file = nil
	}
}
