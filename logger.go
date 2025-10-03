package wiston

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Logger provides a simple logging facility that can write to either
// standard output (stdout) or a specified file.
type Logger struct {
	target string
	file   *os.File
}

// NewLogger creates and returns a new Logger instance.
// If the target is "stdout", logs will be printed to the console.
// Otherwise, the target is treated as a file path, and logs will be
// appended to that file. The file is created if it doesn't exist.
// The function will fatal log if the log file cannot be opened.
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

// Info logs a message with the INFO level.
func (l *Logger) Info(msg string) {
	l.write("INFO", msg)
}

// Success logs a message with the SUCCESS level.
func (l *Logger) Success(msg string) {
	l.write("SUCCESS", msg)
}

// Warn logs a message with the WARN level.
func (l *Logger) Warn(msg string) {
	l.write("WARN", msg)
}

// Error logs a message with the ERROR level.
func (l *Logger) Error(msg string) {
	l.write("ERROR", msg)
}

// write formats and writes a log message to the configured target.
// Each message is prefixed with a timestamp and the log level.
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

// Close closes the underlying log file if the logger is configured
// to write to a file. It is a no-op if the target is stdout.
func (l *Logger) Close() {
	if l.file != nil {
		l.file.Close()
		l.file = nil
	}
}
