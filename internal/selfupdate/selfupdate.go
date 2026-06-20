// Package selfupdate checks GitHub for newer agent-init releases and replaces
// the running binary in place. It is the engine behind `agent-init upgrade`.
//
// The flow is deliberately conservative: a download is never installed without
// first verifying its SHA-256 against the release's checksums.txt, and the
// binary is swapped atomically (write-temp-then-rename) so a failure can't leave
// a half-written executable on disk.
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
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// DefaultRepo is the GitHub "owner/name" releases are pulled from.
const DefaultRepo = "Lillevang/agent-init"

// Release is the subset of a GitHub release this package consumes.
type Release struct {
	Tag    string
	Assets []Asset
}

// Asset is one downloadable file attached to a release.
type Asset struct {
	Name string
	URL  string
}

// Source fetches release metadata and downloads release assets. It is the seam
// the CLI fills with a GitHub client and tests fill with an in-memory fake.
type Source interface {
	Latest(ctx context.Context) (Release, error)
	Download(ctx context.Context, url string) ([]byte, error)
}

// Updater drives a check or an upgrade against a Source. GOOS/GOARCH and ExePath
// are fields rather than hard-coded calls so tests can target arbitrary
// platforms and a throwaway file instead of the real executable.
type Updater struct {
	Source  Source
	Out     io.Writer
	GOOS    string
	GOARCH  string
	ExePath func() (string, error)
}

// NewUpdater returns an Updater wired to the current platform and the running
// executable's path.
func NewUpdater(src Source, out io.Writer) *Updater {
	return &Updater{
		Source:  src,
		Out:     out,
		GOOS:    runtime.GOOS,
		GOARCH:  runtime.GOARCH,
		ExePath: defaultExePath,
	}
}

// CheckResult reports the outcome of a version check.
type CheckResult struct {
	Current        string
	Latest         string
	NewerAvailable bool
}

// Check fetches the latest release and compares it to current without
// downloading anything.
func (u *Updater) Check(ctx context.Context, current string) (CheckResult, error) {
	rel, err := u.Source.Latest(ctx)
	if err != nil {
		return CheckResult{}, fmt.Errorf("fetching latest release: %w", err)
	}
	return CheckResult{
		Current:        current,
		Latest:         rel.Tag,
		NewerAvailable: compareVersions(current, rel.Tag) < 0,
	}, nil
}

// UpgradeOptions tunes an upgrade run.
type UpgradeOptions struct {
	Current string
	// Force installs the latest release even when current is already newest (or
	// unparseable, e.g. a dev build).
	Force bool
	// DryRun downloads and verifies the release but stops before replacing the
	// binary.
	DryRun bool
}

// Upgrade fetches the latest release, verifies its checksum, and atomically
// replaces the running binary. It is a no-op (with a notice) when the current
// version is already the newest and Force is unset.
func (u *Updater) Upgrade(ctx context.Context, opts UpgradeOptions) error {
	rel, err := u.Source.Latest(ctx)
	if err != nil {
		return fmt.Errorf("fetching latest release: %w", err)
	}
	if !opts.Force && compareVersions(opts.Current, rel.Tag) >= 0 {
		_, _ = fmt.Fprintf(u.Out, "agent-init is already up to date (%s).\n", opts.Current)
		return nil
	}

	assetName := u.assetName()
	asset, ok := findAsset(rel.Assets, assetName)
	if !ok {
		return fmt.Errorf("release %s has no asset for %s/%s (expected %q)", rel.Tag, u.GOOS, u.GOARCH, assetName)
	}
	sums, ok := findAsset(rel.Assets, "checksums.txt")
	if !ok {
		return fmt.Errorf("release %s has no checksums.txt; refusing to install an unverified binary", rel.Tag)
	}

	_, _ = fmt.Fprintf(u.Out, "Downloading %s (%s)...\n", assetName, rel.Tag)
	archive, err := u.Source.Download(ctx, asset.URL)
	if err != nil {
		return fmt.Errorf("downloading %s: %w", assetName, err)
	}
	sumData, err := u.Source.Download(ctx, sums.URL)
	if err != nil {
		return fmt.Errorf("downloading checksums.txt: %w", err)
	}
	if err := verifyChecksum(assetName, archive, sumData); err != nil {
		return err
	}

	bin, err := extractBinary(archive, assetName, u.binaryName())
	if err != nil {
		return fmt.Errorf("extracting %s: %w", u.binaryName(), err)
	}

	target, err := u.resolveExePath()
	if err != nil {
		return err
	}
	if opts.DryRun {
		_, _ = fmt.Fprintf(u.Out, "Verified %s. Would replace %s (dry-run).\n", assetName, target)
		return nil
	}
	if err := replaceBinary(target, bin); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(u.Out, "Upgraded %s -> %s (%s).\n", opts.Current, rel.Tag, target)
	return nil
}

// assetName is the release archive for this platform, matching the names cut by
// the release workflow (agent-init-<os>-<arch>.tar.gz, .zip on Windows).
func (u *Updater) assetName() string {
	base := fmt.Sprintf("agent-init-%s-%s", u.GOOS, u.GOARCH)
	if u.GOOS == "windows" {
		return base + ".zip"
	}
	return base + ".tar.gz"
}

// binaryName is the executable file packed inside the archive.
func (u *Updater) binaryName() string {
	base := fmt.Sprintf("agent-init-%s-%s", u.GOOS, u.GOARCH)
	if u.GOOS == "windows" {
		return base + ".exe"
	}
	return base
}

func (u *Updater) resolveExePath() (string, error) {
	fn := u.ExePath
	if fn == nil {
		fn = defaultExePath
	}
	return fn()
}

func defaultExePath() (string, error) {
	p, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locating current executable: %w", err)
	}
	// Resolve symlinks so we replace the real binary, not a symlink pointing at
	// it. If resolution fails, fall back to the reported path.
	if resolved, err := filepath.EvalSymlinks(p); err == nil {
		return resolved, nil
	}
	return p, nil
}

func findAsset(assets []Asset, name string) (Asset, bool) {
	for _, a := range assets {
		if a.Name == name {
			return a, true
		}
	}
	return Asset{}, false
}

// verifyChecksum confirms data hashes to the entry for name in a sha256sum-style
// checksums file. Lines look like "<hex>␣␣<name>"; the binary-mode "*name"
// prefix is tolerated.
func verifyChecksum(name string, data, sums []byte) error {
	want := ""
	for _, line := range strings.Split(string(sums), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		if fields[1] == name || fields[1] == "*"+name {
			want = strings.ToLower(fields[0])
			break
		}
	}
	if want == "" {
		return fmt.Errorf("checksums.txt has no entry for %s", name)
	}
	sum := sha256.Sum256(data)
	got := hex.EncodeToString(sum[:])
	if got != want {
		return fmt.Errorf("checksum mismatch for %s: got %s, want %s", name, got, want)
	}
	return nil
}

func extractBinary(archive []byte, assetName, binName string) ([]byte, error) {
	if strings.HasSuffix(assetName, ".zip") {
		return extractFromZip(archive, binName)
	}
	return extractFromTarGz(archive, binName)
}

func extractFromTarGz(archive []byte, binName string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, fmt.Errorf("opening gzip: %w", err)
	}
	defer func() { _ = gz.Close() }()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading tar: %w", err)
		}
		if filepath.Base(hdr.Name) == binName {
			return io.ReadAll(tr)
		}
	}
	return nil, fmt.Errorf("%s not found in archive", binName)
}

func extractFromZip(archive []byte, binName string) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, fmt.Errorf("opening zip: %w", err)
	}
	for _, f := range zr.File {
		if filepath.Base(f.Name) == binName {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("opening %s in zip: %w", binName, err)
			}
			defer func() { _ = rc.Close() }()
			return io.ReadAll(rc)
		}
	}
	return nil, fmt.Errorf("%s not found in archive", binName)
}

// replaceBinary installs data at target atomically. It writes a temp file in the
// same directory (so the final rename stays on one filesystem), then renames it
// over the target. When a direct rename fails — Windows can't replace a running
// executable — it moves the old binary aside first and rolls back on error.
func replaceBinary(target string, data []byte) error {
	dir := filepath.Dir(target)
	tmp, err := os.CreateTemp(dir, ".agent-init-upgrade-*")
	if err != nil {
		return fmt.Errorf("creating temp file in %s: %w (the install directory may need elevated write access)", dir, err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("writing new binary: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("finalizing new binary: %w", err)
	}
	if err := os.Chmod(tmpName, 0o755); err != nil {
		return fmt.Errorf("setting permissions on new binary: %w", err)
	}

	if err := os.Rename(tmpName, target); err == nil {
		return nil
	}
	// Fallback: move the existing binary aside, then move the new one in.
	old := target + ".old"
	_ = os.Remove(old)
	if err := os.Rename(target, old); err != nil {
		return fmt.Errorf("replacing %s: %w (the install directory may need elevated write access)", target, err)
	}
	if err := os.Rename(tmpName, target); err != nil {
		_ = os.Rename(old, target) // roll back
		return fmt.Errorf("installing new binary at %s: %w", target, err)
	}
	_ = os.Remove(old)
	return nil
}

// compareVersions orders two version strings as semver (-1, 0, 1). A version
// that doesn't parse (e.g. "dev") is treated as older than any real release, so
// a dev build always sees a release as newer.
func compareVersions(a, b string) int {
	pa, oka := parseSemver(a)
	pb, okb := parseSemver(b)
	switch {
	case oka && okb:
		return pa.compare(pb)
	case oka && !okb:
		return 1
	case !oka && okb:
		return -1
	default:
		return strings.Compare(a, b)
	}
}

type semver struct {
	major, minor, patch int
	pre                 string
}

// parseSemver accepts vX, vX.Y, vX.Y.Z (with optional leading v and optional
// -prerelease / +build suffixes). Anything else fails to parse.
func parseSemver(s string) (semver, bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	s = strings.TrimPrefix(s, "V")
	if s == "" {
		return semver{}, false
	}
	if i := strings.IndexByte(s, '+'); i >= 0 {
		s = s[:i]
	}
	pre := ""
	if i := strings.IndexByte(s, '-'); i >= 0 {
		pre = s[i+1:]
		s = s[:i]
	}
	parts := strings.Split(s, ".")
	if len(parts) == 0 || len(parts) > 3 {
		return semver{}, false
	}
	var nums [3]int
	for i := range parts {
		n, err := strconv.Atoi(parts[i])
		if err != nil || n < 0 {
			return semver{}, false
		}
		nums[i] = n
	}
	return semver{nums[0], nums[1], nums[2], pre}, true
}

func (a semver) compare(b semver) int {
	if c := cmpInt(a.major, b.major); c != 0 {
		return c
	}
	if c := cmpInt(a.minor, b.minor); c != 0 {
		return c
	}
	if c := cmpInt(a.patch, b.patch); c != 0 {
		return c
	}
	// A release outranks a prerelease of the same core version.
	switch {
	case a.pre == "" && b.pre == "":
		return 0
	case a.pre == "":
		return 1
	case b.pre == "":
		return -1
	default:
		return strings.Compare(a.pre, b.pre)
	}
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
