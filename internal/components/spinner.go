package components

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorReset  = "\033[0m" // Reset to default color
)

// Spinner wraps the spinner functionality
type Spinner struct {
	s *spinner.Spinner
}

// NewSpinner creates a new spinner instance
func NewSpinner(message string) *Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + message
	s.Start()
	return &Spinner{s: s}
}

// SetMessage updates the spinner message
func (s *Spinner) SetMessage(message string) {
	s.s.Suffix = " " + message
	s.s.Start()
}

// Succeed shows a success message and stops the spinner
func (s *Spinner) Succeed(message string) {
	s.s.Stop()
	fmt.Printf(colorGreen+"✓"+colorReset+" %s\n", message)
}

// Warn shows a warning message and stops the spinner
func (s *Spinner) Warn(message string) {
	s.s.Stop()
	fmt.Printf(colorYellow+"⚠"+colorReset+" %s\n", message)
}

// Fail shows a failure message and stops the spinner
func (s *Spinner) Fail(message string) {
	s.s.Stop()
	fmt.Printf(colorRed+"✗"+colorReset+" %s\n", message)
}

// Finish stops the spinner
func (s *Spinner) Finish() {
	s.s.Stop()
}
