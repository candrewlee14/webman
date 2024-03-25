package ui

import (
	"os"

	"github.com/mattn/go-isatty"
)

// Check if color output is enabled looking for a NO_COLOR environment variable
// that, when present and not an empty string (regardless of its value), prevents the addition of ANSI color.
// If NO_COLOR is unset checks the stdout file descriptor.
func AreAnsiCodesEnabled() bool {
	_, nocolor := os.LookupEnv("NO_COLOR")
	if nocolor {
		return false
	}
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}
