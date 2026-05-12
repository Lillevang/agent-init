package api

import (
	"encoding/json"
	"net/http"
)

// NewRouter returns the HTTP handler for the service. The commit/buildDate
// values are surfaced via /healthz so deploys are self-identifying.
func NewRouter(commit, buildDate string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthz(commit, buildDate))
	return mux
}

func healthz(commit, buildDate string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":    "ok",
			"commit":    commit,
			"buildDate": buildDate,
		})
	}
}
