package app

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewUpgradeCommand(t *testing.T) {
	cmd := newUpgradeCommand()

	if cmd == nil {
		t.Fatal("expected upgrade command, got nil")
	}

	if cmd.Use != "upgrade" {
		t.Errorf("expected Use 'upgrade', got %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description, got empty")
	}

	if cmd.RunE == nil {
		t.Error("expected RunE function, got nil")
	}

	checkFlag, err := cmd.Flags().GetBool("check")
	if err != nil {
		t.Fatalf("expected --check flag, got error: %v", err)
	}
	if checkFlag != false {
		t.Errorf("expected --check default false, got %v", checkFlag)
	}

	versionFlag, err := cmd.Flags().GetString("version")
	if err != nil {
		t.Fatalf("expected --version flag, got error: %v", err)
	}
	if versionFlag != "" {
		t.Errorf("expected --version default empty, got %s", versionFlag)
	}
}

func TestUpgradeCommandRegistration(t *testing.T) {
	ctx := context.Background()
	root := NewRootCommand(ctx)

	upgradeCmd, _, err := root.Find([]string{"upgrade"})
	if err != nil {
		t.Fatalf("failed to find upgrade command: %v", err)
	}

	if upgradeCmd == nil {
		t.Fatal("expected upgrade command, got nil")
	}

	if upgradeCmd.Name() != "upgrade" {
		t.Errorf("expected command name 'upgrade', got %s", upgradeCmd.Name())
	}

	commands := root.Commands()
	var found bool
	for _, cmd := range commands {
		if cmd.Name() == "upgrade" {
			found = true
			break
		}
	}
	if !found {
		t.Error("upgrade command not registered with root command")
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    int
	}{
		{"equal versions", "1.0.0", "1.0.0", 0},
		{"current less than latest", "1.0.0", "1.1.0", -1},
		{"current greater than latest", "2.0.0", "1.1.0", 1},
		{"different lengths shorter", "1.0", "1.0.1", -1},
		{"dev version", "dev", "0.1.0", -1},
		{"major version difference", "0.1.0", "1.0.0", -1},
		{"patch difference", "1.0.0", "1.0.1", -1},
		{"minor version greater", "1.2.0", "1.1.0", 1},
		{"different lengths longer", "1.0.1", "1.0", 1},
		{"zero versions", "0.0.0", "0.0.0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVersions(tt.current, tt.latest)
			if got != tt.want {
				t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.current, tt.latest, got, tt.want)
			}
		})
	}
}

func TestFindAsset(t *testing.T) {
	assets := []githubAsset{
		{Name: "redmine-cli_1.0.0_linux_amd64.tar.gz", DownloadURL: "https://example.com/linux.tar.gz"},
		{Name: "redmine-cli_1.0.0_darwin_arm64.tar.gz", DownloadURL: "https://example.com/darwin.tar.gz"},
		{Name: "redmine-cli_1.0.0_windows_amd64.zip", DownloadURL: "https://example.com/windows.zip"},
		{Name: "redmine-cli_1.0.0_linux_arm64.tar.gz", DownloadURL: "https://example.com/linux-arm64.tar.gz"},
	}

	tests := []struct {
		name    string
		version string
		goos    string
		goarch  string
		want    *githubAsset
	}{
		{
			name:    "linux amd64 tar.gz",
			version: "1.0.0",
			goos:    "linux",
			goarch:  "amd64",
			want:    &assets[0],
		},
		{
			name:    "darwin arm64 tar.gz",
			version: "1.0.0",
			goos:    "darwin",
			goarch:  "arm64",
			want:    &assets[1],
		},
		{
			name:    "windows amd64 zip",
			version: "1.0.0",
			goos:    "windows",
			goarch:  "amd64",
			want:    &assets[2],
		},
		{
			name:    "no matching asset",
			version: "1.0.0",
			goos:    "freebsd",
			goarch:  "amd64",
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findAsset(assets, tt.version, tt.goos, tt.goarch)
			if got != tt.want {
				if got == nil && tt.want != nil {
					t.Errorf("findAsset() = nil, want %v", *tt.want)
				} else if got != nil && tt.want == nil {
					t.Errorf("findAsset() = %v, want nil", *got)
				} else if got != nil && tt.want != nil && got.Name != tt.want.Name {
					t.Errorf("findAsset() = %v, want %v", got.Name, tt.want.Name)
				}
			}
		})
	}
}

func TestExtractTarGz(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "test.tar.gz")

	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}

	gzw := gzip.NewWriter(f)
	tw := tar.NewWriter(gzw)

	content := []byte("#!/bin/sh\necho redmine")
	hdr := &tar.Header{
		Name: "redmine",
		Mode: 0755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("failed to write tar header: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("failed to write tar content: %v", err)
	}

	tw.Close()
	gzw.Close()
	f.Close()

	destDir := t.TempDir()
	binaryPath, err := extractTarGz(archivePath, destDir)
	if err != nil {
		t.Fatalf("extractTarGz() error = %v", err)
	}

	if !strings.HasSuffix(binaryPath, "redmine") {
		t.Errorf("expected binary path to end with 'redmine', got %s", binaryPath)
	}

	gotContent, err := os.ReadFile(binaryPath)
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}

	if string(gotContent) != string(content) {
		t.Errorf("extracted content = %q, want %q", string(gotContent), string(content))
	}

	info, err := os.Stat(binaryPath)
	if err != nil {
		t.Fatalf("failed to stat extracted file: %v", err)
	}

	if info.Mode().Perm()&0111 == 0 {
		t.Error("expected extracted file to be executable")
	}
}

func TestExtractZip(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "test.zip")

	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}

	zw := zip.NewWriter(f)

	content := []byte("redmine windows binary")
	w, err := zw.Create("redmine.exe")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	if _, err := w.Write(content); err != nil {
		t.Fatalf("failed to write zip content: %v", err)
	}

	zw.Close()
	f.Close()

	destDir := t.TempDir()
	binaryPath, err := extractZip(archivePath, destDir)
	if err != nil {
		t.Fatalf("extractZip() error = %v", err)
	}

	if !strings.HasSuffix(binaryPath, "redmine.exe") {
		t.Errorf("expected binary path to end with 'redmine.exe', got %s", binaryPath)
	}

	gotContent, err := os.ReadFile(binaryPath)
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}

	if string(gotContent) != string(content) {
		t.Errorf("extracted content = %q, want %q", string(gotContent), string(content))
	}

	info, err := os.Stat(binaryPath)
	if err != nil {
		t.Fatalf("failed to stat extracted file: %v", err)
	}

	if info.Mode().Perm()&0111 == 0 {
		t.Error("expected extracted file to be executable")
	}
}

func TestReplaceBinary(t *testing.T) {
	tmpDir := t.TempDir()
	currentPath := filepath.Join(tmpDir, "redmine")

	originalContent := []byte("old binary")
	if err := os.WriteFile(currentPath, originalContent, 0755); err != nil {
		t.Fatalf("failed to create current binary: %v", err)
	}

	newContent := []byte("new binary content")
	err := replaceBinary(currentPath, strings.NewReader(string(newContent)))
	if err != nil {
		t.Fatalf("replaceBinary() error = %v", err)
	}

	gotContent, err := os.ReadFile(currentPath)
	if err != nil {
		t.Fatalf("failed to read replaced binary: %v", err)
	}

	if string(gotContent) != string(newContent) {
		t.Errorf("replaced content = %q, want %q", string(gotContent), string(newContent))
	}

	info, err := os.Stat(currentPath)
	if err != nil {
		t.Fatalf("failed to stat replaced binary: %v", err)
	}

	if info.Mode().Perm()&0111 == 0 {
		t.Error("expected replaced binary to be executable")
	}
}
