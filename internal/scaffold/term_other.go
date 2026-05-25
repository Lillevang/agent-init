//go:build !linux && !darwin

package scaffold

import "os"

// isTerminal returns false on unsupported platforms (like Windows), disabling
// colorized output. This is a deliberate choice to avoid a dependency on a
// term library (like golang.org/x/term) for a cosmetic feature that was not
// originally requested for Windows in issue #58. Modern Windows Terminal
// supports ANSI, but this keeps the implementation minimal and dependency-free.
func isTerminal(file *os.File) bool {
	return false
}
