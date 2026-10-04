package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
)

type entry struct {
	name    string
	body    string
	symlink bool
}

func tarGz(t *testing.T, entries ...entry) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		hdr := &tar.Header{Name: e.name, Mode: 0o755, Size: int64(len(e.body)), Typeflag: tar.TypeReg}
		if e.symlink {
			hdr = &tar.Header{Name: e.name, Linkname: "/etc/passwd", Mode: 0o777, Typeflag: tar.TypeSymlink}
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if !e.symlink {
			if _, err := tw.Write([]byte(e.body)); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func zipArchive(t *testing.T, entries ...entry) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range entries {
		hdr := &zip.FileHeader{Name: e.name, Method: zip.Deflate}
		hdr.SetMode(0o755)
		body := e.body
		if e.symlink {
			hdr.SetMode(os.ModeSymlink | 0o777)
			body = "C:/Windows/System32/cmd.exe"
		}
		w, err := zw.CreateHeader(hdr)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func writeArchive(t *testing.T, data []byte) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "archive")
	if err := os.WriteFile(p, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestExtractBinaryUsesFixedOutputName(t *testing.T) {
	dest := t.TempDir()
	archive := writeArchive(t, tarGz(t, entry{name: "LICENSE", body: "mit"}, entry{name: "./es", body: "new binary"}))
	got, err := extractBinary(archive, "linux", dest)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dest, "es"); got != want {
		t.Fatalf("extracted to %q, want %q", got, want)
	}
	if data, _ := os.ReadFile(got); string(data) != "new binary" {
		t.Fatalf("extracted %q", data)
	}
	if _, err := os.Stat(filepath.Join(dest, "LICENSE")); !os.IsNotExist(err) {
		t.Fatalf("extracted a file other than the binary: %v", err)
	}
}

func TestExtractBinaryFromWindowsZip(t *testing.T) {
	dest := t.TempDir()
	archive := writeArchive(t, zipArchive(t, entry{name: "README.md", body: "docs"}, entry{name: "es.exe", body: "new exe"}))
	got, err := extractBinary(archive, "windows", dest)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dest, "es.exe"); got != want {
		t.Fatalf("extracted to %q, want %q", got, want)
	}
	if data, _ := os.ReadFile(got); string(data) != "new exe" {
		t.Fatalf("extracted %q", data)
	}
}

func TestExtractBinaryRefusesUnsafeArchives(t *testing.T) {
	for _, tc := range []struct {
		name    string
		goos    string
		entries []entry
		want    string
	}{
		{"traversal to binary", "linux", []entry{{name: "../../es", body: "x"}}, "unsafe path"},
		{"traversal elsewhere", "linux", []entry{{name: "../evil", body: "x"}, {name: "es", body: "x"}}, "unsafe path"},
		{"absolute path", "linux", []entry{{name: "/usr/local/bin/es", body: "x"}}, "unsafe path"},
		{"binary not at root", "linux", []entry{{name: "nested/es", body: "x"}}, "not at the archive root"},
		{"symlink binary", "linux", []entry{{name: "es", symlink: true}}, "not a regular file"},
		{"missing binary", "linux", []entry{{name: "README.md", body: "x"}}, "not found"},
		{"zip traversal", "windows", []entry{{name: `..\..\es.exe`, body: "x"}}, "unsafe path"},
		{"zip drive path", "windows", []entry{{name: "C:/es.exe", body: "x"}}, "unsafe path"},
		{"zip symlink binary", "windows", []entry{{name: "es.exe", symlink: true}}, "not a regular file"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := tarGz(t, tc.entries...)
			if tc.goos == "windows" {
				data = zipArchive(t, tc.entries...)
			}
			root := t.TempDir()
			dest := filepath.Join(root, "dest")
			if err := os.Mkdir(dest, 0o755); err != nil {
				t.Fatal(err)
			}
			_, err := extractBinary(writeArchive(t, data), tc.goos, dest)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("extractBinary error = %v, want it to mention %q", err, tc.want)
			}
			if files, _ := os.ReadDir(root); len(files) != 1 {
				t.Fatalf("extraction wrote outside dest: %v", files)
			}
		})
	}
}

// releaseServer serves a release archive and checksums.txt for v0.2.0 and
// points downloadBaseURL at it.
func releaseServer(t *testing.T, archiveName string, archive []byte, checksums string) *atomic.Int32 {
	t.Helper()
	var hits atomic.Int32
	prefix := "/" + Repo + "/releases/download/v0.2.0/"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		switch r.URL.Path {
		case prefix + archiveName:
			_, _ = w.Write(archive)
		case prefix + "checksums.txt":
			fmt.Fprint(w, checksums)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	orig := downloadBaseURL
	downloadBaseURL = srv.URL
	t.Cleanup(func() { downloadBaseURL = orig })
	return &hits
}

func checksumLine(name string, data []byte) string {
	return fmt.Sprintf("%x  %s\n", sha256.Sum256(data), name)
}

// installedExe creates a fake installed binary and returns its path.
func installedExe(t *testing.T, name string) string {
	t.Helper()
	exe := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(exe, []byte("old binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	return exe
}

func assertContent(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil || string(data) != want {
		t.Fatalf("%s = %q (%v), want %q", path, data, err, want)
	}
}

func TestInstallReplacesExecutable(t *testing.T) {
	archive := tarGz(t, entry{name: "es", body: "new binary"})
	name := ArchiveName("linux", "amd64")
	releaseServer(t, name, archive, checksumLine("es-cli_linux_arm64.tar.gz", []byte("other"))+checksumLine(name, archive))
	exe := installedExe(t, "es")

	in := &Installer{GOOS: "linux", GOARCH: "amd64", ExePath: exe}
	if err := in.Install(context.Background(), "0.2.0"); err != nil {
		t.Fatal(err)
	}
	assertContent(t, exe, "new binary")
	if runtime.GOOS != "windows" {
		if fi, _ := os.Stat(exe); fi.Mode().Perm() != 0o755 {
			t.Fatalf("installed mode %v, want 0755", fi.Mode().Perm())
		}
	}
	if files, _ := os.ReadDir(filepath.Dir(exe)); len(files) != 1 {
		t.Fatalf("temp files left next to the binary: %v", files)
	}
}

func TestInstallRefusesBadChecksums(t *testing.T) {
	archive := tarGz(t, entry{name: "es", body: "new binary"})
	name := ArchiveName("linux", "amd64")
	for label, sums := range map[string]string{
		"mismatch":      checksumLine(name, []byte("something else")),
		"missing entry": checksumLine("es-cli_darwin_amd64.tar.gz", archive),
		"malformed":     "abc123  " + name + "\n",
	} {
		t.Run(label, func(t *testing.T) {
			releaseServer(t, name, archive, sums)
			exe := installedExe(t, "es")
			in := &Installer{GOOS: "linux", GOARCH: "amd64", ExePath: exe}
			if err := in.Install(context.Background(), "0.2.0"); err == nil {
				t.Fatal("Install accepted a bad checksum")
			}
			assertContent(t, exe, "old binary")
		})
	}
}

func TestInstallFailsEarlyWhenDirectoryNotWritable(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("needs POSIX permissions and a non-root user")
	}
	archive := tarGz(t, entry{name: "es", body: "new binary"})
	name := ArchiveName("linux", "amd64")
	hits := releaseServer(t, name, archive, checksumLine(name, archive))
	exe := installedExe(t, "es")
	dir := filepath.Dir(exe)
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	in := &Installer{GOOS: "linux", GOARCH: "amd64", ExePath: exe}
	err := in.Install(context.Background(), "0.2.0")
	if err == nil || !strings.Contains(err.Error(), "cannot write to "+dir) || !strings.Contains(err.Error(), "sudo") {
		t.Fatalf("Install error = %v, want a not-writable message suggesting sudo", err)
	}
	if hits.Load() != 0 {
		t.Fatalf("downloaded %d files before failing", hits.Load())
	}
	assertContent(t, exe, "old binary")
}

func TestReplaceExecutableWindowsRenamesAside(t *testing.T) {
	exe := installedExe(t, "es.exe")
	// A leftover from an earlier update must not block this one.
	if err := os.WriteFile(exe+".old", []byte("older binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	newBinary := filepath.Join(t.TempDir(), "es.exe")
	if err := os.WriteFile(newBinary, []byte("new binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := replaceExecutable("windows", newBinary, exe); err != nil {
		t.Fatal(err)
	}
	assertContent(t, exe, "new binary")
	assertContent(t, exe+".old", "old binary")

	RemoveOldExecutable(exe)
	if _, err := os.Stat(exe + ".old"); !os.IsNotExist(err) {
		t.Fatalf("es.exe.old not removed on the next start: %v", err)
	}
	if files, _ := os.ReadDir(filepath.Dir(exe)); len(files) != 1 {
		t.Fatalf("unexpected files next to es.exe: %v", files)
	}
}

func TestInstallWindowsZipEndToEnd(t *testing.T) {
	archive := zipArchive(t, entry{name: "es.exe", body: "new exe"})
	name := ArchiveName("windows", "amd64")
	if name != "es-cli_windows_amd64.zip" {
		t.Fatalf("archive name %q", name)
	}
	releaseServer(t, name, archive, checksumLine(name, archive))
	exe := installedExe(t, "es.exe")

	in := &Installer{GOOS: "windows", GOARCH: "amd64", ExePath: exe}
	if err := in.Install(context.Background(), "0.2.0"); err != nil {
		t.Fatal(err)
	}
	assertContent(t, exe, "new exe")
	assertContent(t, exe+".old", "old binary")
}
