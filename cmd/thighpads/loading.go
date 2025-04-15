package main

import (
	"fmt"
	"strings"
	"time"
)

// LoadingBar represents a configurable loading bar
type LoadingBar struct {
	progress      int
	total         int
	width         int
	fill          string
	empty         string
	leftBracket   string
	rightBracket  string
	showPercent   bool
	showDuration  bool
	message       string
	startTime     time.Time
	isDone        bool
	spinnerIndex  int
	spinnerFrames []string
}

// NewLoadingBar creates a new loading bar with default settings
func NewLoadingBar(total int, message string) *LoadingBar {
	return &LoadingBar{
		progress:      0,
		total:         total,
		width:         30,
		fill:          "█",
		empty:         "░",
		leftBracket:   "[",
		rightBracket:  "]",
		showPercent:   true,
		showDuration:  true,
		message:       message,
		startTime:     time.Now(),
		isDone:        false,
		spinnerIndex:  0,
		spinnerFrames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}
}

// Update updates the loading bar with new progress value
func (lb *LoadingBar) Update(progress int) {
	lb.progress = progress
	if lb.progress >= lb.total {
		lb.isDone = true
	}
}

// Increment increases the loading bar progress by the given amount
func (lb *LoadingBar) Increment(amount int) {
	lb.progress += amount
	if lb.progress >= lb.total {
		lb.isDone = true
	}
}

// Complete marks the loading bar as complete
func (lb *LoadingBar) Complete() {
	lb.progress = lb.total
	lb.isDone = true
}

// Render returns the string representation of the loading bar
func (lb *LoadingBar) Render() string {
	percentage := float64(lb.progress) / float64(lb.total)
	filled := int(percentage * float64(lb.width))

	// Ensure filled doesn't exceed width
	if filled > lb.width {
		filled = lb.width
	}

	// Create the bar
	bar := strings.Repeat(lb.fill, filled) + strings.Repeat(lb.empty, lb.width-filled)

	// Add brackets
	bar = lb.leftBracket + bar + lb.rightBracket

	// Add percentage
	if lb.showPercent {
		bar += fmt.Sprintf(" %.0f%%", percentage*100)
	}

	// Add duration
	if lb.showDuration {
		elapsed := time.Since(lb.startTime).Round(time.Second)
		bar += fmt.Sprintf(" (%s)", elapsed)
	}

	// Add spinner or checkmark
	var indicator string
	if lb.isDone {
		indicator = "✓"
	} else {
		indicator = lb.spinnerFrames[lb.spinnerIndex]
		lb.spinnerIndex = (lb.spinnerIndex + 1) % len(lb.spinnerFrames)
	}

	// Construct final output
	output := fmt.Sprintf("%s %s %s", indicator, lb.message, bar)

	return output
}

// PrintProgress renders and prints the loading bar
func (lb *LoadingBar) PrintProgress() {
	fmt.Printf("\r%s", lb.Render())
}

// SimulateProgress simulates progress with the given duration
func (lb *LoadingBar) SimulateProgress(duration time.Duration) {
	steps := lb.total
	interval := duration / time.Duration(steps)

	for i := 0; i < steps; i++ {
		lb.Increment(1)
		lb.PrintProgress()
		time.Sleep(interval)
	}

	// Ensure we complete the progress bar
	lb.Complete()
	lb.PrintProgress()
	fmt.Println()
}

// GetIndeterminateSpinner returns a single character for indeterminate progress
func GetIndeterminateSpinner(index int) string {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	return frames[index%len(frames)]
}

// PrintIndeterminateProgress prints a spinner with a message
func PrintIndeterminateProgress(message string, doneChannel chan bool) {
	i := 0
	for {
		select {
		case <-doneChannel:
			fmt.Printf("\r%s %s ✓%s\n", " ", message, strings.Repeat(" ", 10))
			return
		default:
			fmt.Printf("\r%s %s%s", GetIndeterminateSpinner(i), message, strings.Repeat(" ", 5))
			i++
			time.Sleep(100 * time.Millisecond)
		}
	}
}
