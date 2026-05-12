package version

import "testing"

func TestString(t *testing.T) {
	got := String("abc", "2026-01-01")
	want := "commit=abc buildDate=2026-01-01"
	if got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}
