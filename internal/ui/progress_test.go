package ui

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewProgress(t *testing.T) {
	p := NewProgress(true)

	if p == nil {
		t.Fatal("NewProgress() should return non-nil Progress")
	}

	if p.verbose != true {
		t.Errorf("NewProgress(true) verbose = %v, want true", p.verbose)
	}

	if p.current != -1 {
		t.Errorf("NewProgress() current = %d, want -1", p.current)
	}

	if len(p.stages) != 0 {
		t.Errorf("NewProgress() stages = %v, want empty slice", p.stages)
	}
}

func TestProgress_AddStage(t *testing.T) {
	p := NewProgress(false)
	p.AddStage("Test Stage")

	if len(p.stages) != 1 {
		t.Fatalf("AddStage() stages length = %d, want 1", len(p.stages))
	}

	stage := p.stages[0]
	if stage.Name != "Test Stage" {
		t.Errorf("AddStage() stage.Name = %q, want %q", stage.Name, "Test Stage")
	}

	if stage.Status != StageStatusPending {
		t.Errorf("AddStage() stage.Status = %q, want %q", stage.Status, StageStatusPending)
	}
}

func TestProgress_AddStages(t *testing.T) {
	p := NewProgress(false)
	p.AddStages("Stage 1", "Stage 2", "Stage 3")

	if len(p.stages) != 3 {
		t.Fatalf("AddStages() stages length = %d, want 3", len(p.stages))
	}

	names := []string{"Stage 1", "Stage 2", "Stage 3"}
	for i, name := range names {
		if p.stages[i].Name != name {
			t.Errorf("AddStages() stage[%d].Name = %q, want %q", i, p.stages[i].Name, name)
		}
	}
}

func TestProgress_Start(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
	}{
		{"verbose mode", true},
		{"silent mode", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProgress(tt.verbose)
			p.AddStages("Stage 1", "Stage 2")

			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			p.Start("Stage 1")

			// Restore stdout
			w.Close()
			os.Stdout = old

			// Read output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Check stage status
			if p.stages[0].Status != StageStatusRunning {
				t.Errorf("Start() stage.Status = %q, want %q", p.stages[0].Status, StageStatusRunning)
			}

			if p.current != 0 {
				t.Errorf("Start() current = %d, want 0", p.current)
			}

			if p.stages[0].StartTime.IsZero() {
				t.Error("Start() StartTime should not be zero")
			}

			// Check output in verbose mode
			if tt.verbose {
				if !strings.Contains(output, "[1/2]") {
					t.Errorf("Start() output should contain stage number, got: %q", output)
				}
				if !strings.Contains(output, "Stage 1") {
					t.Errorf("Start() output should contain stage name, got: %q", output)
				}
			} else {
				if output != "" {
					t.Errorf("Start() in silent mode should not output, got: %q", output)
				}
			}
		})
	}
}

func TestProgress_StartWithNewLine(t *testing.T) {
	p := NewProgress(true)
	p.AddStages("Stage 1", "Stage 2")

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	p.StartWithNewLine("Stage 1")

	// Restore stdout
	w.Close()
	os.Stdout = old

	// Read output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Should contain newline
	if !strings.Contains(output, "\n") {
		t.Errorf("StartWithNewLine() output should contain newline, got: %q", output)
	}
}

func TestProgress_Complete(t *testing.T) {
	p := NewProgress(true)
	p.AddStages("Stage 1", "Stage 2")
	p.Start("Stage 1")

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	p.Complete("Stage 1")

	// Restore stdout
	w.Close()
	os.Stdout = old

	// Read output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check stage status
	if p.stages[0].Status != StageStatusDone {
		t.Errorf("Complete() stage.Status = %q, want %q", p.stages[0].Status, StageStatusDone)
	}

	if p.stages[0].EndTime.IsZero() {
		t.Error("Complete() EndTime should not be zero")
	}

	if p.stages[0].Duration == 0 {
		t.Error("Complete() Duration should be greater than 0")
	}

	// Check output
	if !strings.Contains(output, "✓") {
		t.Errorf("Complete() output should contain checkmark, got: %q", output)
	}

	if !strings.Contains(output, "Stage 1") {
		t.Errorf("Complete() output should contain stage name, got: %q", output)
	}

	// Should contain time in seconds (e.g., "0.01s")
	if !strings.Contains(output, "s") {
		t.Errorf("Complete() output should contain duration, got: %q", output)
	}
}

func TestProgress_CompleteInNewLine(t *testing.T) {
	p := NewProgress(true)
	p.AddStages("Stage 1", "Stage 2")
	p.Start("Stage 1")

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	p.CompleteInNewLine("Stage 1")

	// Restore stdout
	w.Close()
	os.Stdout = old

	// Read output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check output format - should be simple format without [1/2]
	if strings.Contains(output, "[1/2]") {
		t.Errorf("CompleteInNewLine() should not contain stage numbers, got: %q", output)
	}

	// Should contain checkmark and stage name
	if !strings.Contains(output, "✓") {
		t.Errorf("CompleteInNewLine() output should contain checkmark, got: %q", output)
	}

	if !strings.Contains(output, "Stage 1") {
		t.Errorf("CompleteInNewLine() output should contain stage name, got: %q", output)
	}
}

func TestProgress_Error(t *testing.T) {
	p := NewProgress(true)
	p.AddStages("Stage 1", "Stage 2")
	p.Start("Stage 1")

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	err := errors.New("test error")

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	p.Error("Stage 1", err)

	// Restore stdout
	w.Close()
	os.Stdout = old

	// Read output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	_ = buf.String() // We don't check output in Error()

	// Check stage status
	if p.stages[0].Status != StageStatusError {
		t.Errorf("Error() stage.Status = %q, want %q", p.stages[0].Status, StageStatusError)
	}

	if p.stages[0].EndTime.IsZero() {
		t.Error("Error() EndTime should not be zero")
	}

	if p.stages[0].Duration == 0 {
		t.Error("Error() Duration should be greater than 0")
	}
}

func TestProgress_MultipleStages(t *testing.T) {
	p := NewProgress(true)
	p.AddStages("Fetch diff", "Generate message", "Format output")

	// Start first stage
	p.Start("Fetch diff")
	if p.current != 0 {
		t.Errorf("After Start() current = %d, want 0", p.current)
	}

	// Complete first stage
	p.Complete("Fetch diff")
	if p.stages[0].Status != StageStatusDone {
		t.Errorf("Stage 1 status = %q, want %q", p.stages[0].Status, StageStatusDone)
	}

	// Start second stage
	p.Start("Generate message")
	if p.current != 1 {
		t.Errorf("After second Start() current = %d, want 1", p.current)
	}

	// Complete second stage
	p.Complete("Generate message")
	if p.stages[1].Status != StageStatusDone {
		t.Errorf("Stage 2 status = %q, want %q", p.stages[1].Status, StageStatusDone)
	}

	// Start third stage
	p.Start("Format output")
	if p.current != 2 {
		t.Errorf("After third Start() current = %d, want 2", p.current)
	}

	p.Complete("Format output")
	if p.stages[2].Status != StageStatusDone {
		t.Errorf("Stage 3 status = %q, want %q", p.stages[2].Status, StageStatusDone)
	}
}

func TestStageNumber(t *testing.T) {
	p := NewProgress(false)
	p.AddStages("Stage 1", "Stage 2", "Stage 3")

	tests := []struct {
		name     string
		stage    string
		want     int
	}{
		{"first stage", "Stage 1", 0},
		{"second stage", "Stage 2", 1},
		{"third stage", "Stage 3", 2},
		{"non-existent stage", "Stage 4", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stageNumber(p, tt.stage); got != tt.want {
				t.Errorf("stageNumber(%q) = %d, want %d", tt.stage, got, tt.want)
			}
		})
	}
}

func TestProgress_SilentMode(t *testing.T) {
	p := NewProgress(false)
	p.AddStages("Stage 1", "Stage 2")

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	p.Start("Stage 1")
	p.Complete("Stage 1")

	// Restore stdout
	w.Close()
	os.Stdout = old

	// Read output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// In silent mode, there should be no output
	if output != "" {
		t.Errorf("In silent mode, there should be no output, got: %q", output)
	}

	// But stage status should still be updated
	if p.stages[0].Status != StageStatusDone {
		t.Errorf("Stage status should be Done even in silent mode, got %q", p.stages[0].Status)
	}
}

func TestProgress_DurationCalculation(t *testing.T) {
	p := NewProgress(true)
	p.AddStage("Test Stage")

	p.Start("Test Stage")

	// Sleep for a measurable duration
	sleepDuration := 50 * time.Millisecond
	time.Sleep(sleepDuration)

	p.Complete("Test Stage")

	// Check duration is approximately correct (within 10ms tolerance)
	tolerance := 10 * time.Millisecond
	expectedMin := sleepDuration - tolerance
	expectedMax := sleepDuration + tolerance

	if p.stages[0].Duration < expectedMin || p.stages[0].Duration > expectedMax {
		t.Errorf("Duration = %v, want between %v and %v", p.stages[0].Duration, expectedMin, expectedMax)
	}
}

func TestProgress_Render(t *testing.T) {
	p := NewProgress(true)
	p.AddStages("Stage 1", "Stage 2")
	p.Start("Stage 1")

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	p.render()

	// Restore stdout
	w.Close()
	os.Stdout = old

	// Read output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// render() should only show running stages
	if !strings.Contains(output, "Stage 1") {
		t.Errorf("render() should show running stage, got: %q", output)
	}
}

func TestProgress_ConcurrentOperations(t *testing.T) {
	p := NewProgress(true)
	p.AddStages("Stage 1", "Stage 2", "Stage 3")

	// Test that we can start and complete stages in order
	for i := 0; i < len(p.stages); i++ {
		stageName := p.stages[i].Name
		p.Start(stageName)
		if p.current != i {
			t.Errorf("After starting stage %d, current = %d", i, p.current)
		}
		p.Complete(stageName)
		if p.stages[i].Status != StageStatusDone {
			t.Errorf("Stage %d not completed", i)
		}
	}
}
