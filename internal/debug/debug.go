package debug

import (
	"log"
	"os"
)

var (
	isDebug  bool
	logger   *log.Logger
	exitFunc = os.Exit // replacable for testing
)

func init() {
	logger = log.New(os.Stderr, "[DEBUG] ", log.LstdFlags)
}

// Enable enables debug mode
func Enable(enabled bool) {
	isDebug = enabled
}

// Print prints a debug message if debug mode is enabled
func Print(format string, args ...interface{}) {
	if isDebug {
		logger.Printf(format, args...)
	}
}

// Printf is an alias for Print
func Printf(format string, args ...interface{}) {
	Print(format, args...)
}

// Println prints a debug message with a newline if debug mode is enabled
func Println(args ...interface{}) {
	if isDebug {
		logger.Println(args...)
	}
}

// Fatal prints a debug message and exits if debug mode is enabled
func Fatal(args ...interface{}) {
	if isDebug {
		logger.Fatal(args...)
	} else {
		exitFunc(1)
	}
}

// Fatalf prints a formatted debug message and exits if debug mode is enabled
func Fatalf(format string, args ...interface{}) {
	if isDebug {
		logger.Fatalf(format, args...)
	} else {
		exitFunc(1)
	}
}
