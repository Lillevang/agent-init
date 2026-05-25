//go:build !linux && !darwin

package scaffold

import "os"

// isTerminal returns false on unsupported platforms (including Windows), so
// color output is deliberately disabled there. Modern Windows terminals can
// support ANSI, but issue #58 only requested the dependency-free Unix TTY gate.
func isTerminal(file *os.File) bool {
	return false
}
