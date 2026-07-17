package selfupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// maxDownloadBytes caps a single asset download. The release archives are a few
// MiB; 32 MiB is ~10x the largest plausible archive and stops a misbehaving or
// hostile endpoint from streaming unbounded data into memory.
const maxDownloadBytes = 32 << 20 // 32 MiB

// GitHubSource implements Source against the GitHub releases REST API.
type GitHubSource struct {
	Repo       string // "owner/name"
	APIBaseURL string // default "https://api.github.com"
	HTTPClient *http.Client
	// Token is sent as a bearer credential to lift the unauthenticated rate
	// limit. Optional; release assets are public.
	Token string
}

// NewGitHubSource returns a source for repo, reading an optional token from
// GITHUB_TOKEN or GH_TOKEN so authenticated users avoid the low anonymous rate
// limit.
func NewGitHubSource(repo string) *GitHubSource {
	return &GitHubSource{
		Repo:       repo,
		APIBaseURL: "https://api.github.com",
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		Token:      firstEnv("GITHUB_TOKEN", "GH_TOKEN"),
	}
}

func (g *GitHubSource) client() *http.Client {
	if g.HTTPClient != nil {
		return g.HTTPClient
	}
	return http.DefaultClient
}

// Latest fetches the repo's latest published release.
func (g *GitHubSource) Latest(ctx context.Context) (Release, error) {
	url := fmt.Sprintf("%s/repos/%s/releases/latest", strings.TrimRight(g.APIBaseURL, "/"), g.Repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Release{}, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if g.Token != "" {
		req.Header.Set("Authorization", "Bearer "+g.Token)
	}
	resp, err := g.client().Do(req)
	if err != nil {
		return Release{}, fmt.Errorf("requesting latest release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return Release{}, fmt.Errorf("github API %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	var payload struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return Release{}, fmt.Errorf("decoding release: %w", err)
	}
	if payload.TagName == "" {
		return Release{}, fmt.Errorf("github API returned a release with no tag")
	}
	rel := Release{Tag: payload.TagName}
	for _, a := range payload.Assets {
		rel.Assets = append(rel.Assets, Asset{Name: a.Name, URL: a.URL})
	}
	return rel, nil
}

// Download fetches the bytes at url, capped at maxDownloadBytes.
func (g *GitHubSource) Download(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	if g.Token != "" {
		req.Header.Set("Authorization", "Bearer "+g.Token)
	}
	resp, err := g.client().Do(req)
	if err != nil {
		return nil, fmt.Errorf("requesting %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("downloading %s: %s", url, resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxDownloadBytes+1))
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", url, err)
	}
	if len(data) > maxDownloadBytes {
		return nil, fmt.Errorf("download from %s exceeds %d bytes", url, maxDownloadBytes)
	}
	return data, nil
}

func firstEnv(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}
