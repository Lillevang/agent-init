package scaffold

import (
	"bytes"
	"os"
	"testing"
)

func TestColorDisabledForNonTTYOutputs(t *testing.T) {
	t.Parallel()
	var buffer bytes.Buffer
	if colorEnabled(&buffer) {
		t.Fatal("colorEnabled(bytes.Buffer) = true, want false")
	}

	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}
	defer func() { _ = devNull.Close() }()
	if colorEnabled(devNull) {
		t.Fatalf("colorEnabled(%s) = true, want false", os.DevNull)
	}
}

func TestColorDisabledByEnvironment(t *testing.T) {
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}
	defer func() { _ = devNull.Close() }()

	alwaysTerminal := func(*os.File) bool { return true }
	for _, tt := range []struct {
		name string
		env  map[string]string
	}{
		{name: "NO_COLOR", env: map[string]string{"NO_COLOR": "1"}},
		{name: "TERM dumb", env: map[string]string{"TERM": "dumb"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			getenv := func(key string) string { return tt.env[key] }
			if colorEnabledWith(devNull, getenv, alwaysTerminal) {
				t.Fatalf("colorEnabledWith(%s) = true, want false", tt.name)
			}
		})
	}
}

func TestColorEnabledForTerminalFile(t *testing.T) {
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}
	defer func() { _ = devNull.Close() }()

	if !colorEnabledWith(devNull, func(string) string { return "" }, func(*os.File) bool { return true }) {
		t.Fatal("colorEnabledWith terminal file = false, want true")
	}
}
