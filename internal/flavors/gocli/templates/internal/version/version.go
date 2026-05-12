package version

// String returns a human-readable build identifier.
func String(commit, date string) string {
	return "commit=" + commit + " buildDate=" + date
}
