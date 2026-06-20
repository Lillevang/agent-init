package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// fakeSource serves a fixed release and a URL->bytes map for downloads.
type fakeSource struct {
	rel   Release
	files map[string][]byte
	err   error
}

func (f *fakeSource) Latest(_ context.Context) (Release, error) { return f.rel, f.err }

func (f *fakeSource) Download(_ context.Context, url string) ([]byte, error) {
	b, ok := f.files[url]
	if !ok {
		return nil, fmt.Errorf("no file at %s", url)
	}
	return b, nil
}

func TestCompareVersions(t *testing.T) {
	t.Parallel()
	cases := []struct {
		a, b string
		want int
	}{
		{"v1.0.0", "v1.0.1", -1},
		{"v1.2.0", "v1.1.9", 1},
		{"v1.2.3", "v1.2.3", 0},
		{"1.2.3", "v1.2.3", 0}, // leading v optional
		{"v2.0.0", "v1.9.9", 1},
		{"v1.0.0", "v1.0", 0},         // missing patch == .0
		{"v1.0.0-rc1", "v1.0.0", -1},  // prerelease < release
		{"v1.0.0", "v1.0.0-rc1", 1},   // release > prerelease
		{"v1.0.0+build", "v1.0.0", 0}, // build metadata ignored
		{"dev", "v1.0.0", -1},         // unparseable current is older
		{"v1.0.0", "dev", 1},          // unparseable latest is older
	}
	for _, tc := range cases {
		if got := compareVersions(tc.a, tc.b); got != tc.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestAssetAndBinaryNames(t *testing.T) {
	t.Parallel()
	cases := []struct {
		goos, goarch       string
		wantAsset, wantBin string
	}{
		{"linux", "amd64", "agent-init-linux-amd64.tar.gz", "agent-init-linux-amd64"},
		{"darwin", "arm64", "agent-init-darwin-arm64.tar.gz", "agent-init-darwin-arm64"},
		{"windows", "amd64", "agent-init-windows-amd64.zip", "agent-init-windows-amd64.exe"},
	}
	for _, tc := range cases {
		u := &Updater{GOOS: tc.goos, GOARCH: tc.goarch}
		if got := u.assetName(); got != tc.wantAsset {
			t.Errorf("assetName(%s/%s) = %q, want %q", tc.goos, tc.goarch, got, tc.wantAsset)
		}
		if got := u.binaryName(); got != tc.wantBin {
			t.Errorf("binaryName(%s/%s) = %q, want %q", tc.goos, tc.goarch, got, tc.wantBin)
		}
	}
}

func TestVerifyChecksum(t *testing.T) {
	t.Parallel()
	data := []byte("the archive bytes")
	sum := sha256.Sum256(data)
	good := fmt.Sprintf("%s  agent-init-linux-amd64.tar.gz\n", hex.EncodeToString(sum[:]))

	if err := verifyChecksum("agent-init-linux-amd64.tar.gz", data, []byte(good)); err != nil {
		t.Errorf("verifyChecksum(matching) = %v, want nil", err)
	}
	// Tolerate the sha256sum binary-mode "*" prefix.
	star := fmt.Sprintf("%s *agent-init-linux-amd64.tar.gz\n", hex.EncodeToString(sum[:]))
	if err := verifyChecksum("agent-init-linux-amd64.tar.gz", data, []byte(star)); err != nil {
		t.Errorf("verifyChecksum(star prefix) = %v, want nil", err)
	}
	if err := verifyChecksum("agent-init-linux-amd64.tar.gz", []byte("tampered"), []byte(good)); err == nil {
		t.Error("verifyChecksum(mismatch) = nil, want error")
	}
	if err := verifyChecksum("missing.tar.gz", data, []byte(good)); err == nil {
		t.Error("verifyChecksum(no entry) = nil, want error")
	}
}

func makeTarGz(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(content))}); err != nil {
		t.Fatalf("tar header: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("tar write: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar close: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}
	return buf.Bytes()
}

func makeZip(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create(name)
	if err != nil {
		t.Fatalf("zip create: %v", err)
	}
	if _, err := w.Write(content); err != nil {
		t.Fatalf("zip write: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return buf.Bytes()
}

func TestExtractBinary(t *testing.T) {
	t.Parallel()
	want := []byte("new binary contents")

	gz := makeTarGz(t, "agent-init-linux-amd64", want)
	got, err := extractBinary(gz, "agent-init-linux-amd64.tar.gz", "agent-init-linux-amd64")
	if err != nil {
		t.Fatalf("extractBinary(tar.gz) error = %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("tar.gz extract = %q, want %q", got, want)
	}

	z := makeZip(t, "agent-init-windows-amd64.exe", want)
	got, err = extractBinary(z, "agent-init-windows-amd64.zip", "agent-init-windows-amd64.exe")
	if err != nil {
		t.Fatalf("extractBinary(zip) error = %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("zip extract = %q, want %q", got, want)
	}

	if _, err := extractBinary(gz, "agent-init-linux-amd64.tar.gz", "not-present"); err == nil {
		t.Error("extractBinary(missing entry) = nil, want error")
	}
}

// buildRelease wires a fakeSource for linux/amd64 with a valid archive and
// checksums file for the given tag and binary content.
func buildRelease(t *testing.T, tag string, binContent []byte) *fakeSource {
	t.Helper()
	archive := makeTarGz(t, "agent-init-linux-amd64", binContent)
	sum := sha256.Sum256(archive)
	checksums := fmt.Sprintf("%s  agent-init-linux-amd64.tar.gz\n", hex.EncodeToString(sum[:]))
	const archiveURL = "https://example.test/agent-init-linux-amd64.tar.gz"
	const sumURL = "https://example.test/checksums.txt"
	return &fakeSource{
		rel: Release{
			Tag: tag,
			Assets: []Asset{
				{Name: "agent-init-linux-amd64.tar.gz", URL: archiveURL},
				{Name: "checksums.txt", URL: sumURL},
			},
		},
		files: map[string][]byte{
			archiveURL: archive,
			sumURL:     []byte(checksums),
		},
	}
}

func newTestUpdater(src Source, out *bytes.Buffer, target string) *Updater {
	return &Updater{
		Source:  src,
		Out:     out,
		GOOS:    "linux",
		GOARCH:  "amd64",
		ExePath: func() (string, error) { return target, nil },
	}
}

func TestUpgradeReplacesBinary(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "agent-init")
	if err := os.WriteFile(target, []byte("OLD BINARY"), 0o755); err != nil {
		t.Fatal(err)
	}
	newBin := []byte("BRAND NEW BINARY")
	src := buildRelease(t, "v2.0.0", newBin)
	var out bytes.Buffer
	u := newTestUpdater(src, &out, target)

	if err := u.Upgrade(context.Background(), UpgradeOptions{Current: "v1.0.0"}); err != nil {
		t.Fatalf("Upgrade error = %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, newBin) {
		t.Errorf("binary after upgrade = %q, want %q", got, newBin)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("upgraded binary is not executable: mode %v", info.Mode())
	}
}

func TestUpgradeAlreadyUpToDate(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "agent-init")
	if err := os.WriteFile(target, []byte("CURRENT"), 0o755); err != nil {
		t.Fatal(err)
	}
	src := buildRelease(t, "v1.0.0", []byte("SHOULD NOT INSTALL"))
	var out bytes.Buffer
	u := newTestUpdater(src, &out, target)

	if err := u.Upgrade(context.Background(), UpgradeOptions{Current: "v1.0.0"}); err != nil {
		t.Fatalf("Upgrade error = %v", err)
	}
	got, _ := os.ReadFile(target)
	if string(got) != "CURRENT" {
		t.Errorf("binary changed when already up to date: %q", got)
	}
	if !bytes.Contains(out.Bytes(), []byte("up to date")) {
		t.Errorf("expected up-to-date notice, got %q", out.String())
	}
}

func TestUpgradeForceReinstallsSameVersion(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "agent-init")
	if err := os.WriteFile(target, []byte("OLD"), 0o755); err != nil {
		t.Fatal(err)
	}
	newBin := []byte("REINSTALLED")
	src := buildRelease(t, "v1.0.0", newBin)
	var out bytes.Buffer
	u := newTestUpdater(src, &out, target)

	if err := u.Upgrade(context.Background(), UpgradeOptions{Current: "v1.0.0", Force: true}); err != nil {
		t.Fatalf("Upgrade(force) error = %v", err)
	}
	got, _ := os.ReadFile(target)
	if !bytes.Equal(got, newBin) {
		t.Errorf("force did not reinstall: %q", got)
	}
}

func TestUpgradeDryRunDoesNotReplace(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "agent-init")
	if err := os.WriteFile(target, []byte("OLD"), 0o755); err != nil {
		t.Fatal(err)
	}
	src := buildRelease(t, "v2.0.0", []byte("NEW"))
	var out bytes.Buffer
	u := newTestUpdater(src, &out, target)

	if err := u.Upgrade(context.Background(), UpgradeOptions{Current: "v1.0.0", DryRun: true}); err != nil {
		t.Fatalf("Upgrade(dry-run) error = %v", err)
	}
	got, _ := os.ReadFile(target)
	if string(got) != "OLD" {
		t.Errorf("dry-run replaced the binary: %q", got)
	}
	if !bytes.Contains(out.Bytes(), []byte("dry-run")) {
		t.Errorf("expected dry-run notice, got %q", out.String())
	}
}

func TestUpgradeChecksumMismatchLeavesBinary(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "agent-init")
	if err := os.WriteFile(target, []byte("OLD"), 0o755); err != nil {
		t.Fatal(err)
	}
	src := buildRelease(t, "v2.0.0", []byte("NEW"))
	// Corrupt the published checksum so verification fails.
	src.files["https://example.test/checksums.txt"] = []byte("deadbeef  agent-init-linux-amd64.tar.gz\n")
	var out bytes.Buffer
	u := newTestUpdater(src, &out, target)

	err := u.Upgrade(context.Background(), UpgradeOptions{Current: "v1.0.0"})
	if err == nil {
		t.Fatal("Upgrade(bad checksum) = nil, want error")
	}
	got, _ := os.ReadFile(target)
	if string(got) != "OLD" {
		t.Errorf("checksum failure replaced the binary: %q", got)
	}
}

func TestUpgradeNoAssetForPlatform(t *testing.T) {
	t.Parallel()
	src := buildRelease(t, "v2.0.0", []byte("NEW")) // only has linux/amd64
	var out bytes.Buffer
	u := newTestUpdater(src, &out, filepath.Join(t.TempDir(), "agent-init"))
	u.GOOS = "plan9"
	u.GOARCH = "mips"

	if err := u.Upgrade(context.Background(), UpgradeOptions{Current: "v1.0.0"}); err == nil {
		t.Fatal("Upgrade(no matching asset) = nil, want error")
	}
}

func TestUpgradeMissingChecksums(t *testing.T) {
	t.Parallel()
	src := buildRelease(t, "v2.0.0", []byte("NEW"))
	// Drop the checksums asset entirely.
	src.rel.Assets = src.rel.Assets[:1]
	var out bytes.Buffer
	u := newTestUpdater(src, &out, filepath.Join(t.TempDir(), "agent-init"))

	if err := u.Upgrade(context.Background(), UpgradeOptions{Current: "v1.0.0"}); err == nil {
		t.Fatal("Upgrade(no checksums.txt) = nil, want error")
	}
}

func TestCheck(t *testing.T) {
	t.Parallel()
	src := buildRelease(t, "v2.0.0", []byte("NEW"))
	u := newTestUpdater(src, &bytes.Buffer{}, "")

	res, err := u.Check(context.Background(), "v1.0.0")
	if err != nil {
		t.Fatalf("Check error = %v", err)
	}
	if !res.NewerAvailable || res.Latest != "v2.0.0" {
		t.Errorf("Check = %+v, want NewerAvailable and Latest v2.0.0", res)
	}

	res, err = u.Check(context.Background(), "v2.0.0")
	if err != nil {
		t.Fatalf("Check error = %v", err)
	}
	if res.NewerAvailable {
		t.Errorf("Check(current==latest) reported newer available")
	}
}
