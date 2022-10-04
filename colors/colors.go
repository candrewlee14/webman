package colors

import "os"

// Check if color output is enabled looking for a NO_COLOR environment variable
// that, when present and not an empty string (regardless of its value), prevents the addition of ANSI color.
func Enabled() bool {
	_, e := os.LookupEnv("NO_COLOR")
	return e
}
