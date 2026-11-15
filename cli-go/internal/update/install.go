package update

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type InstallMethod string

const (
	InstallSelf InstallMethod = "self"
	InstallGo   InstallMethod = "go"
)

const goUpdateCommand = "git pull && make install (in your es-cli/cli-go checkout)"

const maxReleaseArtifactBytes int64 = 256 << 20

func UpdateCommand(method InstallMethod) string {
	if method == InstallGo {
		return goUpdateCommand
	}
	return binaryName + " update"
}

func DetectInstallMethod(exePath string, getenv func(string) string, homeDir string) InstallMethod {
	var goBins []string
	if gobin := getenv("GOBIN"); gobin != "" {
		goBins = append(goBins, gobin)
	}
	for _, p := range filepath.SplitList(getenv("GOPATH")) {
		if p != "" {
			goBins = append(goBins, filepath.Join(p, "bin"))
		}
	}
	if homeDir != "" {
		goBins = append(goBins, filepath.Join(homeDir, "go", "bin"))
	}
	exeDir, err := os.Stat(filepath.Dir(exePath))
	if err != nil {
		return InstallSelf
	}
	for _, dir := range goBins {
		if fi, err := os.Stat(dir); err == nil && os.SameFile(exeDir, fi) {
			return InstallGo
		}
	}
	return InstallSelf
}

func ExecutablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("finding the running executable: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return "", fmt.Errorf("resolving %s: %w", exe, err)
	}
	return resolved, nil
}

func RemoveOldExecutable(exePath string) {
	_ = os.Remove(exePath + ".old")
}

type Installer struct {
	GOOS     string
	GOARCH   string
	ExePath  string
	Progress io.Writer
}

func ArchiveName(goos, goarch string) string {
	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}
	return fmt.Sprintf("%s_%s_%s%s", projectName, goos, goarch, ext)
}

func executableName(goos string) string {
	if goos == "windows" {
		return binaryName + ".exe"
	}
	return binaryName
}

func (in *Installer) Install(ctx context.Context, version string) error {
	progress := in.Progress
	if progress == nil {
		progress = io.Discard
	}
	if err := checkWritable(in.GOOS, filepath.Dir(in.ExePath)); err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "es-update-*")
	if err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archive := ArchiveName(in.GOOS, in.GOARCH)
	releaseBase := fmt.Sprintf("%s/%s/releases/download/v%s/", downloadBaseURL, Repo, version)
	archivePath := filepath.Join(tmpDir, "release-archive")

	fmt.Fprintf(progress, "Downloading %s...\n", archive)
	if err := downloadFile(ctx, releaseBase+archive, archivePath); err != nil {
		return fmt.Errorf("downloading %s: %w", archive, err)
	}
	fmt.Fprintln(progress, "Verifying checksum...")
	var sums bytes.Buffer
	if err := download(ctx, releaseBase+"checksums.txt", &sums, 1<<20); err != nil {
		return fmt.Errorf("downloading checksums.txt: %w", err)
	}
	if err := verifyChecksum(archivePath, archive, sums.Bytes()); err != nil {
		return err
	}

	newBinary, err := extractBinary(archivePath, in.GOOS, tmpDir)
	if err != nil {
		return fmt.Errorf("extracting %s: %w", archive, err)
	}
	fmt.Fprintf(progress, "Installing to %s...\n", in.ExePath)
	if err := replaceExecutable(in.GOOS, newBinary, in.ExePath); err != nil {
		return fmt.Errorf("replacing %s: %w", in.ExePath, err)
	}
	return nil
}

func checkWritable(goos, dir string) error {
	f, err := os.CreateTemp(dir, ".es-update-*")
	if err != nil {
		hint := "re-run with sudo, or reinstall es with the install script into a writable directory (INSTALL_DIR=~/.local/bin)"
		if goos == "windows" {
			hint = "re-run from a terminal opened as Administrator, or move es.exe to a directory you can write to"
		}
		return fmt.Errorf("cannot write to %s: %s", dir, hint)
	}
	name := f.Name()
	_ = f.Close()
	_ = os.Remove(name)
	return nil
}

func downloadFile(ctx context.Context, url, dst string) error {
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	if err := download(ctx, url, f, maxReleaseArtifactBytes); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func download(ctx context.Context, url string, dst io.Writer, limit int64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", projectName)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return copyLimited(dst, resp.Body, limit)
}

func copyLimited(dst io.Writer, src io.Reader, limit int64) error {
	written, err := io.CopyN(dst, src, limit+1)
	if err != nil && err != io.EOF {
		return err
	}
	if written > limit {
		return fmt.Errorf("exceeds the %d MiB size limit", limit>>20)
	}
	return nil
}

func verifyChecksum(archivePath, archiveName string, checksums []byte) error {
	var want []byte
	scanner := bufio.NewScanner(bytes.NewReader(checksums))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 && strings.TrimPrefix(fields[1], "*") == archiveName {
			sum, err := hex.DecodeString(fields[0])
			if err != nil || len(sum) != sha256.Size {
				return fmt.Errorf("checksums.txt has a malformed SHA-256 for %s", archiveName)
			}
			want = sum
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading checksums.txt: %w", err)
	}
	if want == nil {
		return fmt.Errorf("checksums.txt has no entry for %s; refusing to install", archiveName)
	}

	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("hashing %s: %w", archiveName, err)
	}
	if got := h.Sum(nil); !bytes.Equal(got, want) {
		return fmt.Errorf("SHA-256 mismatch for %s (expected %x, got %x); refusing to install", archiveName, want, got)
	}
	return nil
}

func extractBinary(archivePath, goos, destDir string) (string, error) {
	want := executableName(goos)
	out := filepath.Join(destDir, want)
	if goos == "windows" {
		return out, extractFromZip(archivePath, want, out)
	}
	return out, extractFromTarGz(archivePath, want, out)
}

func matchEntry(name, want string) (bool, error) {
	clean := path.Clean(strings.ReplaceAll(name, `\`, "/"))
	if path.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(clean, ":") {
		return false, fmt.Errorf("archive entry %q has an unsafe path", name)
	}
	if path.Base(clean) != want {
		return false, nil
	}
	if clean != want {
		return false, fmt.Errorf("archive entry %q is not at the archive root", name)
	}
	return true, nil
}

func extractFromTarGz(archivePath, want, out string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("opening gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("%s not found in archive", want)
		}
		if err != nil {
			return fmt.Errorf("reading tar: %w", err)
		}
		ok, err := matchEntry(hdr.Name, want)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		if hdr.Typeflag != tar.TypeReg {
			return fmt.Errorf("archive entry %q is not a regular file", hdr.Name)
		}
		if hdr.Size > maxReleaseArtifactBytes {
			return fmt.Errorf("archive entry %q exceeds the %d MiB size limit", hdr.Name, maxReleaseArtifactBytes>>20)
		}
		return writeBinary(out, tr)
	}
}

func extractFromZip(archivePath, want, out string) error {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("opening zip: %w", err)
	}
	defer zr.Close()
	for _, zf := range zr.File {
		ok, err := matchEntry(zf.Name, want)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		if !zf.Mode().IsRegular() {
			return fmt.Errorf("archive entry %q is not a regular file", zf.Name)
		}
		if zf.UncompressedSize64 > uint64(maxReleaseArtifactBytes) {
			return fmt.Errorf("archive entry %q exceeds the %d MiB size limit", zf.Name, maxReleaseArtifactBytes>>20)
		}
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		return writeBinary(out, rc)
	}
	return fmt.Errorf("%s not found in archive", want)
}

func writeBinary(out string, src io.Reader) error {
	f, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	if err := copyLimited(f, src, maxReleaseArtifactBytes); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func replaceExecutable(goos, newBinary, exePath string) error {
	dir := filepath.Dir(exePath)
	tmp, err := os.CreateTemp(dir, ".es-update-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	installed := false
	defer func() {
		if !installed {
			_ = os.Remove(tmpPath)
		}
	}()

	src, err := os.Open(newBinary)
	if err != nil {
		tmp.Close()
		return err
	}
	err = copyLimited(tmp, src, maxReleaseArtifactBytes)
	src.Close()
	if err == nil {
		err = tmp.Chmod(0o755)
	}
	if err == nil {
		err = tmp.Sync()
	}
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return err
	}

	if goos != "windows" {
		if err := os.Rename(tmpPath, exePath); err != nil {
			return err
		}
		installed = true
		return nil
	}

	old := exePath + ".old"
	if err := os.Remove(old); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("removing %s left by an earlier update: %w", old, err)
	}
	if err := os.Rename(exePath, old); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, exePath); err != nil {
		if rerr := os.Rename(old, exePath); rerr != nil {
			return fmt.Errorf("%w; restoring the previous binary also failed: %v (it is at %s)", err, rerr, old)
		}
		return err
	}
	installed = true
	return nil
}
