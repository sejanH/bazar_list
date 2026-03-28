package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarning
	LevelError
)

// Logger handles logging operations
type Logger struct {
	level LogLevel
	debug bool
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	// Check for debug environment variable
	debug := os.Getenv("BAZARLIST_DEBUG") == "true"

	level := LevelInfo
	if debug {
		level = LevelDebug
	}

	return &Logger{
		level: level,
		debug: debug,
	}
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= LevelDebug {
		l.log("DEBUG", format, args...)
	}
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= LevelInfo {
		l.log("INFO", format, args...)
	}
}

// Warning logs a warning message
func (l *Logger) Warning(format string, args ...interface{}) {
	if l.level <= LevelWarning {
		l.log("WARN", format, args...)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= LevelError {
		l.log("ERROR", format, args...)
	}
}

// log is the internal logging function
func (l *Logger) log(level string, format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf("[%s] [%s] %s", timestamp, level, fmt.Sprintf(format, args...))
	log.Println(message)
}

// IsDebugEnabled returns whether debug mode is enabled
func (l *Logger) IsDebugEnabled() bool {
	return l.debug
}
