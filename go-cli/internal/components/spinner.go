package components

import (
	"fmt"
	"github.com/briandowns/spinner"
	"time"
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
}

// Succeed shows a success message and stops the spinner
func (s *Spinner) Succeed(message string) {
	s.s.Stop()
	fmt.Printf("✓ %s\n", message)
}

// Fail shows a failure message and stops the spinner
func (s *Spinner) Fail(message string) {
	s.s.Stop()
	fmt.Printf("✗ %s\n", message)
}

// Finish stops the spinner
func (s *Spinner) Finish() {
	s.s.Stop()
} 