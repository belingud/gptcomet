package debug

import (
	"bytes"
	"os"
	"regexp"
	"testing"
)

func matchLogFormat(log, message string) bool {
	pattern := `^\[DEBUG\] \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2} ` + regexp.QuoteMeta(message) + `\n$`
	matched, err := regexp.MatchString(pattern, log)
	return err == nil && matched
}

func TestEnable(t *testing.T) {
	Enable(true)
	if !isDebug {
		t.Error("Expected debug mode to be enabled")
	}

	Enable(false)
	if isDebug {
		t.Error("Expected debug mode to be disabled")
	}
}

func TestPrint(t *testing.T) {
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	defer logger.SetOutput(os.Stderr)

	// Test with debug disabled
	Enable(false)
	Print("test message")
	if buf.Len() > 0 {
		t.Error("Expected no output when debug is disabled")
	}

	// Test with debug enabled
	Enable(true)
	Print("test message %s", "123")
	if !matchLogFormat(buf.String(), "test message 123") {
		t.Errorf("Unexpected output: %q", buf.String())
	}
}

func TestPrintf(t *testing.T) {
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	defer logger.SetOutput(os.Stderr)

	Enable(true)
	Printf("formatted %s", "message")
	if !matchLogFormat(buf.String(), "formatted message") {
		t.Errorf("Unexpected output: %q", buf.String())
	}
}

func TestPrintln(t *testing.T) {
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	defer logger.SetOutput(os.Stderr)

	Enable(true)
	Println("line", "message")
	if !matchLogFormat(buf.String(), "line message") {
		t.Errorf("Unexpected output: %q", buf.String())
	}
}

func TestFatal(t *testing.T) {
	called := false
	exitFunc = func(code int) {
		called = true
		if code != 1 {
			t.Errorf("Expected exit code 1, got %d", code)
		}
	}

	defer func() {
		exitFunc = os.Exit
	}()

	isDebug = false
	Fatal("This is a test")
	if !called {
		t.Error("exitFunc was not called")
	}
}

func TestFatalf(t *testing.T) {
	called := false
	exitFunc = func(code int) {
		called = true
		if code != 1 {
			t.Errorf("Expected exit code 1, got %d", code)
		}
	}

	defer func() {
		exitFunc = os.Exit
	}()

	isDebug = false
	Fatalf("This is a formatted test: %d", 123)
	if !called {
		t.Error("exitFunc was not called")
	}
}
