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
func (p *Progress) Start(name string) {
	for i, stage := range p.stages {
		if stage.Name == name {
			stage.Status = StageStatusRunning
			stage.StartTime = time.Now()
			p.current = i
			if p.verbose {
				p.render()
			}
			return
		}
	}
}

// Complete marks a stage as complete
func (p *Progress) Complete(name string) {
	for _, stage := range p.stages {
		if stage.Name == name {
			stage.Status = StageStatusDone
			stage.EndTime = time.Now()
			stage.Duration = stage.EndTime.Sub(stage.StartTime)
			if p.verbose {
				p.render()
			}
			return
		}
	}
}

// Error marks a stage as failed with an error
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

// getStageIcon returns the appropriate icon for a stage status
func (p *Progress) getStageIcon(status StageStatus) string {
	switch status {
	case StageStatusDone:
		return iconDone
	case StageStatusRunning:
		return iconRunning
	case StageStatusError:
		return iconError
	default:
		return iconPending
	}
}
