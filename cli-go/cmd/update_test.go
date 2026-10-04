package cmd

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

// A traversing entry name must not place the binary outside destDir.
func TestExtractBinaryIgnoresEntryPath(t *testing.T) {
	root := t.TempDir()
	destDir := filepath.Join(root, "dest")
	if err := os.Mkdir(destDir, 0o755); err != nil {
		t.Fatal(err)
	}

	archivePath := filepath.Join(root, "es-cli.tar.gz")
	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	body := []byte("new binary")
	if err := tw.WriteHeader(&tar.Header{Name: "../../es-cli", Mode: 0o755, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(body); err != nil {
		t.Fatal(err)
	}
	for _, c := range []interface{ Close() error }{tw, gz, f} {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	}

	got, err := extractBinary(archivePath, destDir)
	if err != nil {
		t.Fatalf("extractBinary: %v", err)
	}
	if want := filepath.Join(destDir, "es"); got != want {
		t.Fatalf("extracted to %q, want %q", got, want)
	}
	if data, err := os.ReadFile(got); err != nil || string(data) != string(body) {
		t.Fatalf("extracted contents = %q, %v", data, err)
	}
	if _, err := os.Stat(filepath.Join(root, "es-cli")); !os.IsNotExist(err) {
		t.Fatalf("entry path escaped destDir: %v", err)
	}
}
