// Package logger provides structured logging functionality with multiple log levels.
//
// This package offers a standardized logging interface with support for debug, info,
// warning, and error messages. It includes features such as:
//
//   - Multiple log levels (Debug, Info, Warn, Error)
//   - Optional emoji icons for visual distinction
//   - Thread-safe operations
//   - Compatibility functions for migrating from debug package
//
// Basic usage:
//
//	logger.Debug("Starting operation %s", operationName)
//	logger.Info("Operation completed successfully")
//	logger.Warn("Deprecated feature used")
//	logger.Error("Operation failed: %v", err)
//
// Configuration:
//
//	logger.SetLevel(logger.InfoLevel)  // Set minimum log level
//	logger.EnableIcons(true)           // Enable emoji icons
//	logger.SetOutput(os.Stdout)        // Change output destination
//
// The package maintains a default logger instance that can be used through
// package-level functions, or you can create custom logger instances using New().
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

// Level represents the log level
type Level int

const (
	// DebugLevel is for detailed debug information
	DebugLevel Level = iota
	// InfoLevel is for general informational messages
	InfoLevel
	// WarnLevel is for warning messages
	WarnLevel
	// ErrorLevel is for error messages
	ErrorLevel
)

var (
	levelNames = map[Level]string{
		DebugLevel: "DEBUG",
		InfoLevel:  "INFO",
		WarnLevel:  "WARN",
		ErrorLevel: "ERROR",
	}
	levelPrefixes = map[Level]string{
		DebugLevel: "🔍",
		InfoLevel:  "ℹ️",
		WarnLevel:  "⚠️",
		ErrorLevel: "❌",
	}
)

// Logger provides structured logging with different log levels
type Logger struct {
	level      Level
	output     io.Writer
	mu         sync.Mutex
	logger     *log.Logger
	exitFunc   func(int) // replaceable for testing
	enableIcon bool
}

var (
	// std is the default logger instance
	std *Logger
)

// init initializes the default logger
func init() {
	std = New(os.Stderr, DebugLevel, true)
}

// New creates a new Logger instance
func New(output io.Writer, level Level, enableIcon bool) *Logger {
	return &Logger{
		level:      level,
		output:     output,
		logger:     log.New(output, "", log.LstdFlags),
		exitFunc:   os.Exit,
		enableIcon: enableIcon,
	}
}

// SetLevel sets the minimum log level
func SetLevel(level Level) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.level = level
}

// SetOutput sets the output destination
func SetOutput(output io.Writer) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.output = output
	std.logger = log.New(output, "", log.LstdFlags)
}

// EnableIcons enables or disables emoji icons in log messages
func EnableIcons(enable bool) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.enableIcon = enable
}

// log is the internal logging function
func (l *Logger) log(level Level, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	prefix := ""
	if l.enableIcon {
		prefix = levelPrefixes[level] + " "
	}
	prefix += fmt.Sprintf("[%s] ", levelNames[level])

	message := fmt.Sprintf(format, args...)
	l.logger.SetPrefix(prefix)
	l.logger.Output(3, message)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DebugLevel, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(InfoLevel, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WarnLevel, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ErrorLevel, format, args...)
}

// Fatal logs an error message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(ErrorLevel, format, args...)
	l.exitFunc(1)
}

// Package-level functions for convenience

// Debug logs a debug message using the default logger
func Debug(format string, args ...interface{}) {
	std.Debug(format, args...)
}

// Info logs an info message using the default logger
func Info(format string, args ...interface{}) {
	std.Info(format, args...)
}

// Warn logs a warning message using the default logger
func Warn(format string, args ...interface{}) {
	std.Warn(format, args...)
}

// Error logs an error message using the default logger
func Error(format string, args ...interface{}) {
	std.Error(format, args...)
}

// Fatal logs an error message and exits using the default logger
func Fatal(format string, args ...interface{}) {
	std.Fatal(format, args...)
}

// Compatibility functions for migrating from debug package

// Print is a compatibility function for debug.Print
func Print(format string, args ...interface{}) {
	Debug(format, args...)
}

// Printf is a compatibility function for debug.Printf
func Printf(format string, args ...interface{}) {
	Debug(format, args...)
}

// Println is a compatibility function for debug.Println
func Println(args ...interface{}) {
	message := fmt.Sprintln(args...)
	// Remove trailing newline added by Sprintln
	if len(message) > 0 && message[len(message)-1] == '\n' {
		message = message[:len(message)-1]
	}
	Debug("%s", message)
}
