//go:build !linux && !darwin

package scaffold

import "os"

func isTerminal(file *os.File) bool {
	return false
}
