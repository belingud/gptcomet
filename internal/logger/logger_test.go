package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogger_LogLevels(t *testing.T) {
	tests := []struct {
		name       string
		level      Level
		logFunc    func(*Logger, string)
		shouldLog  bool
		wantPrefix string
	}{
		{
			name:       "Debug level logs debug messages",
			level:      DebugLevel,
			logFunc:    func(l *Logger, msg string) { l.Debug(msg) },
			shouldLog:  true,
			wantPrefix: "[DEBUG]",
		},
		{
			name:       "Info level skips debug messages",
			level:      InfoLevel,
			logFunc:    func(l *Logger, msg string) { l.Debug(msg) },
			shouldLog:  false,
			wantPrefix: "",
		},
		{
			name:       "Info level logs info messages",
			level:      InfoLevel,
			logFunc:    func(l *Logger, msg string) { l.Info(msg) },
			shouldLog:  true,
			wantPrefix: "[INFO]",
		},
		{
			name:       "Warn level skips info messages",
			level:      WarnLevel,
			logFunc:    func(l *Logger, msg string) { l.Info(msg) },
			shouldLog:  false,
			wantPrefix: "",
		},
		{
			name:       "Warn level logs warn messages",
			level:      WarnLevel,
			logFunc:    func(l *Logger, msg string) { l.Warn(msg) },
			shouldLog:  true,
			wantPrefix: "[WARN]",
		},
		{
			name:       "Error level logs error messages",
			level:      ErrorLevel,
			logFunc:    func(l *Logger, msg string) { l.Error(msg) },
			shouldLog:  true,
			wantPrefix: "[ERROR]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(&buf, tt.level, false)

			testMsg := "test message"
			tt.logFunc(logger, testMsg)

			output := buf.String()
			if tt.shouldLog {
				if !strings.Contains(output, tt.wantPrefix) {
					t.Errorf("expected prefix %q in output, got: %s", tt.wantPrefix, output)
				}
				if !strings.Contains(output, testMsg) {
					t.Errorf("expected message %q in output, got: %s", testMsg, output)
				}
			} else {
				if output != "" {
					t.Errorf("expected no output, got: %s", output)
				}
			}
		})
	}
}

func TestLogger_IconSupport(t *testing.T) {
	tests := []struct {
		name       string
		enableIcon bool
		level      Level
		logFunc    func(*Logger, string)
		wantIcon   string
	}{
		{
			name:       "Debug with icon",
			enableIcon: true,
			level:      DebugLevel,
			logFunc:    func(l *Logger, msg string) { l.Debug(msg) },
			wantIcon:   "🔍",
		},
		{
			name:       "Debug without icon",
			enableIcon: false,
			level:      DebugLevel,
			logFunc:    func(l *Logger, msg string) { l.Debug(msg) },
			wantIcon:   "",
		},
		{
			name:       "Info with icon",
			enableIcon: true,
			level:      InfoLevel,
			logFunc:    func(l *Logger, msg string) { l.Info(msg) },
			wantIcon:   "ℹ️",
		},
		{
			name:       "Warn with icon",
			enableIcon: true,
			level:      WarnLevel,
			logFunc:    func(l *Logger, msg string) { l.Warn(msg) },
			wantIcon:   "⚠️",
		},
		{
			name:       "Error with icon",
			enableIcon: true,
			level:      ErrorLevel,
			logFunc:    func(l *Logger, msg string) { l.Error(msg) },
			wantIcon:   "❌",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(&buf, tt.level, tt.enableIcon)

			tt.logFunc(logger, "test message")
			output := buf.String()

			if tt.enableIcon {
				if !strings.Contains(output, tt.wantIcon) {
					t.Errorf("expected icon %q in output, got: %s", tt.wantIcon, output)
				}
			} else if tt.wantIcon != "" {
				if strings.Contains(output, tt.wantIcon) {
					t.Errorf("expected no icon in output, got: %s", output)
				}
			}
		})
	}
}

func TestLogger_FormatString(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, DebugLevel, false)

	logger.Debug("test %s %d", "message", 42)
	output := buf.String()

	if !strings.Contains(output, "test message 42") {
		t.Errorf("expected formatted message, got: %s", output)
	}
}

func TestLogger_CompatibilityFunctions(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, DebugLevel, false)

	// Save old std logger and restore after test
	oldStd := std
	std = logger
	defer func() { std = oldStd }()

	tests := []struct {
		name    string
		logFunc func()
		wantMsg string
	}{
		{
			name:    "Print function",
			logFunc: func() { Print("test %s", "message") },
			wantMsg: "test message",
		},
		{
			name:    "Printf function",
			logFunc: func() { Printf("test %s", "message") },
			wantMsg: "test message",
		},
		{
			name:    "Println function",
			logFunc: func() { Println("test", "message") },
			wantMsg: "test message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()
			output := buf.String()

			if !strings.Contains(output, tt.wantMsg) {
				t.Errorf("expected %q in output, got: %s", tt.wantMsg, output)
			}
		})
	}
}

func TestLogger_Fatal(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, ErrorLevel, false)

	exitCalled := false
	logger.exitFunc = func(code int) {
		exitCalled = true
		if code != 1 {
			t.Errorf("expected exit code 1, got %d", code)
		}
	}

	logger.Fatal("fatal error")

	if !exitCalled {
		t.Error("expected exit function to be called")
	}

	output := buf.String()
	if !strings.Contains(output, "fatal error") {
		t.Errorf("expected error message in output, got: %s", output)
	}
}

func TestPackageLevelFunctions(t *testing.T) {
	var buf bytes.Buffer
	oldStd := std
	std = New(&buf, DebugLevel, false)
	defer func() { std = oldStd }()

	tests := []struct {
		name    string
		logFunc func()
		wantMsg string
		wantLvl string
	}{
		{
			name:    "Debug function",
			logFunc: func() { Debug("debug message") },
			wantMsg: "debug message",
			wantLvl: "[DEBUG]",
		},
		{
			name:    "Info function",
			logFunc: func() { Info("info message") },
			wantMsg: "info message",
			wantLvl: "[INFO]",
		},
		{
			name:    "Warn function",
			logFunc: func() { Warn("warn message") },
			wantMsg: "warn message",
			wantLvl: "[WARN]",
		},
		{
			name:    "Error function",
			logFunc: func() { Error("error message") },
			wantMsg: "error message",
			wantLvl: "[ERROR]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()
			output := buf.String()

			if !strings.Contains(output, tt.wantMsg) {
				t.Errorf("expected %q in output, got: %s", tt.wantMsg, output)
			}
			if !strings.Contains(output, tt.wantLvl) {
				t.Errorf("expected %q in output, got: %s", tt.wantLvl, output)
			}
		})
	}
}

func TestSetLevel(t *testing.T) {
	var buf bytes.Buffer
	oldStd := std
	std = New(&buf, DebugLevel, false)
	defer func() { std = oldStd }()

	// Debug should be logged
	Debug("debug message")
	if buf.Len() == 0 {
		t.Error("expected debug message to be logged")
	}

	// Change level to Info
	buf.Reset()
	SetLevel(InfoLevel)

	// Debug should not be logged
	Debug("debug message")
	if buf.Len() != 0 {
		t.Error("expected debug message to be filtered")
	}

	// Info should be logged
	Info("info message")
	if buf.Len() == 0 {
		t.Error("expected info message to be logged")
	}
}

func TestEnableIcons(t *testing.T) {
	var buf bytes.Buffer
	oldStd := std
	std = New(&buf, DebugLevel, false)
	defer func() { std = oldStd }()

	// Without icons
	Debug("test")
	output := buf.String()
	if strings.Contains(output, "🔍") {
		t.Error("expected no icon in output")
	}

	// With icons
	buf.Reset()
	EnableIcons(true)
	Debug("test")
	output = buf.String()
	if !strings.Contains(output, "🔍") {
		t.Error("expected icon in output")
	}
}
