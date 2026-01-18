package ui

import (
	"fmt"
	"time"
)

// StageStatus represents the status of a progress stage
type StageStatus string

const (
	StageStatusPending StageStatus = "pending"
	StageStatusRunning StageStatus = "running"
	StageStatusDone    StageStatus = "done"
	StageStatusError   StageStatus = "error"
)

// ProgressStage represents a single stage in the progress
type ProgressStage struct {
	Name      string
	Status    StageStatus
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// Progress manages the display of progress information
type Progress struct {
	stages  []*ProgressStage
	current int
	verbose bool
}

// NewProgress creates a new Progress instance
func NewProgress(verbose bool) *Progress {
	return &Progress{
		stages:  make([]*ProgressStage, 0),
		current: -1,
		verbose: verbose,
	}
}

// AddStage adds a new stage to the progress
func (p *Progress) AddStage(name string) {
	stage := &ProgressStage{
		Name:   name,
		Status: StageStatusPending,
	}
	p.stages = append(p.stages, stage)
}

// AddStages adds multiple stages at once
func (p *Progress) AddStages(names ...string) {
	for _, name := range names {
		p.AddStage(name)
	}
}

// Start starts a specific stage by name
// This is used when we want to start a stage and show the stage number
// Example: [1/2] Fetching diff...
//
// This function is used to start a specific stage by name. It takes the name of the
// stage as a parameter and updates the status of the stage to "StageStatusRunning".
// It also sets the start time of the stage.
//
// Parameters:
// - name: The name of the stage to be started.
//
// Return: None.
func (p *Progress) Start(name string) {
	for i, stage := range p.stages {
		if stage.Name == name {
			stage.Status = StageStatusRunning
			stage.StartTime = time.Now()
			p.current = i
			if p.verbose {
				// Display start message
				fmt.Printf("[%d/%d] %s...", i+1, len(p.stages), stage.Name)
			}
			return
		}
	}
}

// StartWithNewLine starts a specific stage by name and prints a new line
// This is used when we want to start a stage and show the stage number in a new line
// Example: [1/2] Fetching diff...\n
//
// Parameters:
// - name: The name of the stage to be started.
//
// Return: None.
func (p *Progress) StartWithNewLine(name string) {
	p.Start(name)
	fmt.Println() // Add newline after start
}

// Complete marks a stage as complete
// Example: [1/2] Fetching diff... ✓ (0.07s)
//
// This function is used to mark a stage as complete. It takes the name of the
// stage as a parameter and updates the status of the stage to "StageStatusDone".
// It also calculates the duration of the stage and stores it.
//
// Parameters:
// - name: The name of the stage to be marked as complete.
//
// Return: None.
func (p *Progress) Complete(name string) {
	for _, stage := range p.stages {
		if stage.Name == name {
			stage.Status = StageStatusDone
			stage.EndTime = time.Now()
			stage.Duration = stage.EndTime.Sub(stage.StartTime)
			if p.verbose {
				// Display completion info immediately
				fmt.Printf("\r[%d/%d] %s... ✓ (%.2fs)\n",
					stageNumber(p, stage.Name)+1,
					len(p.stages),
					stage.Name,
					float64(stage.Duration.Milliseconds())/1000)
			}
			return
		}
	}
}

// CompleteInNewLine marks a stage as complete with simple format (no stage numbers)
// This is used when we want to complete a stage but don't want to show the stage number
// and we want to display it in a new line with a checkmark prefix
// Example: ✓ Generating review (0.00s)
//
// Parameters:
// - name: The name of the stage to be marked as complete.
//
// Return: None.
func (p *Progress) CompleteInNewLine(name string) {
	for _, stage := range p.stages {
		if stage.Name == name {
			stage.Status = StageStatusDone
			stage.EndTime = time.Now()
			stage.Duration = stage.EndTime.Sub(stage.StartTime)
			if p.verbose {
				// Display simple completion format
				fmt.Printf("✓ %s (%.2fs)\n",
					stage.Name,
					float64(stage.Duration.Milliseconds())/1000)
			}
			return
		}
	}
}

// Helper function to find stage number by name
// Returns the index of the stage with the given name
// Returns 0 if the stage is not found
//
// Parameters:
// - p: The progress object.
// - name: The name of the stage.
//
// Returns:
// - The index of the stage with the given name.
// - 0 if the stage is not found.
func stageNumber(p *Progress, name string) int {
	for i, stage := range p.stages {
		if stage.Name == name {
			return i
		}
	}
	return 0
}

// Error marks a stage as failed with an error
// This is used when we want to mark a stage as failed and show the stage number
// Example: [1/2] Fetching diff... ✗ (0.07s)
//
// Parameters:
// - name: The name of the stage.
// - err: The error that occurred.
func (p *Progress) Error(name string, err error) {
	for _, stage := range p.stages {
		if stage.Name == name {
			stage.Status = StageStatusError
			stage.EndTime = time.Now()
			stage.Duration = stage.EndTime.Sub(stage.StartTime)
			if p.verbose {
				p.render()
			}
			return
		}
	}
}

// render renders the progress display (only in verbose mode)
//
// This function is used to render the progress display when in verbose mode.
// It only renders the current running stage.
func (p *Progress) render() {
	if !p.verbose {
		return
	}

	// Only render the current running stage
	for i, stage := range p.stages {
		if stage.Status == StageStatusRunning {
			fmt.Printf("[%d/%d] %s...\n", i+1, len(p.stages), stage.Name)
			break
		}
	}
}
