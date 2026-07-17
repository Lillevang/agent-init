package selfupdate

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGitHubSourceLatest(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"tag_name": "v3.1.4",
			"assets": [
				{"name": "agent-init-linux-amd64.tar.gz", "browser_download_url": "https://dl.test/a.tar.gz"},
				{"name": "checksums.txt", "browser_download_url": "https://dl.test/checksums.txt"}
			]
		}`))
	}))
	defer srv.Close()

	g := &GitHubSource{Repo: "Lillevang/agent-init", APIBaseURL: srv.URL, HTTPClient: srv.Client()}
	rel, err := g.Latest(context.Background())
	if err != nil {
		t.Fatalf("Latest error = %v", err)
	}
	if gotPath != "/repos/Lillevang/agent-init/releases/latest" {
		t.Errorf("requested path = %q", gotPath)
	}
	if rel.Tag != "v3.1.4" {
		t.Errorf("tag = %q, want v3.1.4", rel.Tag)
	}
	if len(rel.Assets) != 2 || rel.Assets[0].Name != "agent-init-linux-amd64.tar.gz" {
		t.Errorf("assets = %+v", rel.Assets)
	}
}

func TestGitHubSourceLatestNon200(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	g := &GitHubSource{Repo: "x/y", APIBaseURL: srv.URL, HTTPClient: srv.Client()}
	if _, err := g.Latest(context.Background()); err == nil {
		t.Fatal("Latest(404) = nil, want error")
	}
}

func TestGitHubSourceDownload(t *testing.T) {
	t.Parallel()
	want := []byte("archive-bytes")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(want)
	}))
	defer srv.Close()

	g := &GitHubSource{Repo: "x/y", APIBaseURL: srv.URL, HTTPClient: srv.Client()}
	got, err := g.Download(context.Background(), srv.URL+"/asset")
	if err != nil {
		t.Fatalf("Download error = %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("Download = %q, want %q", got, want)
	}
}
