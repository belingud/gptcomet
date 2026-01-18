package ui

import "github.com/charmbracelet/lipgloss"

// Style definitions for progress display
var (
	// Progress bar styles
	progressBarStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")). // Cyan
		Width(40)

	// Stage status styles
	stageDoneStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("70")) // Green

	stageRunningStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("228")) // Yellow

	stagePendingStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")) // Gray

	stageErrorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")) // Red

	// General text styles
	infoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")) // Cyan

	successStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("70")) // Green
)

// Stage status icons
const (
	iconDone    = "✓"
	iconRunning = "→"
	iconPending = "⌀"
	iconError   = "✗"
)
