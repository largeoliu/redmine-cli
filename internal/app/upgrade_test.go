package app

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	apperrors "github.com/largeoliu/redmine-cli/internal/errors"
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

func TestNewUpgradeCommandExecution(t *testing.T) {
	origGet := httpGetFunc
	httpGetFunc = func(url string) (*http.Response, error) {
		return nil, errors.New("test error")
	}
	defer func() { httpGetFunc = origGet }()

	cmd := newUpgradeCommand()
	cmd.SetArgs([]string{"--check"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error from --check with network failure, got nil")
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

func TestFetchLatestRelease(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		releaseJSON := `{"tag_name":"v1.2.3","assets":[{"name":"redmine-cli_1.2.3_linux_amd64.tar.gz","browser_download_url":"https://example.com/file.tar.gz"}]}`
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(releaseJSON))
		}))
		defer server.Close()

		origGet := httpGetFunc
		httpGetFunc = func(url string) (*http.Response, error) {
			return server.Client().Get(server.URL)
		}
		defer func() { httpGetFunc = origGet }()

		release, err := fetchLatestRelease()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if release.Tag != "v1.2.3" {
			t.Errorf("expected tag v1.2.3, got %s", release.Tag)
		}
		if len(release.Assets) != 1 {
			t.Fatalf("expected 1 asset, got %d", len(release.Assets))
		}
		if release.Assets[0].Name != "redmine-cli_1.2.3_linux_amd64.tar.gz" {
			t.Errorf("unexpected asset name: %s", release.Assets[0].Name)
		}
	})

	t.Run("network error", func(t *testing.T) {
		origGet := httpGetFunc
		httpGetFunc = func(url string) (*http.Response, error) {
			return nil, errors.New("network error")
		}
		defer func() { httpGetFunc = origGet }()

		_, err := fetchLatestRelease()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var appErr *apperrors.Error
		if !apperrors.As(err, &appErr) {
			t.Fatalf("expected *errors.Error, got %T", err)
		}
		if appErr.Category != "network" {
			t.Errorf("expected network error, got %s", appErr.Category)
		}
	})

	t.Run("non-200 status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		origGet := httpGetFunc
		httpGetFunc = func(url string) (*http.Response, error) {
			return server.Client().Get(server.URL)
		}
		defer func() { httpGetFunc = origGet }()

		_, err := fetchLatestRelease()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var appErr *apperrors.Error
		if !apperrors.As(err, &appErr) {
			t.Fatalf("expected *errors.Error, got %T", err)
		}
		if appErr.Category != "network" {
			t.Errorf("expected network error, got %s", appErr.Category)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		origGet := httpGetFunc
		httpGetFunc = func(url string) (*http.Response, error) {
			return server.Client().Get(server.URL)
		}
		defer func() { httpGetFunc = origGet }()

		_, err := fetchLatestRelease()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var appErr *apperrors.Error
		if !apperrors.As(err, &appErr) {
			t.Fatalf("expected *errors.Error, got %T", err)
		}
		if appErr.Category != "internal" {
			t.Errorf("expected internal error, got %s", appErr.Category)
		}
	})

	t.Run("download error", func(t *testing.T) {
		origVersion := version
		version = "0.1.0"
		defer func() { version = origVersion }()

		goos := runtime.GOOS
		goarch := runtime.GOARCH
		var ext string
		switch goos {
		case "windows":
			ext = ".zip"
		default:
			ext = ".tar.gz"
		}
		assetName := fmt.Sprintf("redmine-cli_2.0.0_%s_%s%s", goos, goarch, ext)

		apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(githubRelease{
				Tag: "v2.0.0",
				Assets: []githubAsset{
					{Name: assetName, DownloadURL: "http://download.example.com/file"},
				},
			})
		}))
		defer apiServer.Close()

		callCount := 0
		origGet := httpGetFunc
		httpGetFunc = func(url string) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				return apiServer.Client().Get(apiServer.URL)
			}
			return nil, errors.New("download network error")
		}
		defer func() { httpGetFunc = origGet }()

		err := runUpgrade(false, "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("replace binary error", func(t *testing.T) {
		origVersion := version
		version = "0.1.0"
		defer func() { version = origVersion }()

		goos := runtime.GOOS
		goarch := runtime.GOARCH
		var ext string
		switch goos {
		case "windows":
			ext = ".zip"
		default:
			ext = ".tar.gz"
		}
		assetName := fmt.Sprintf("redmine-cli_2.0.0_%s_%s%s", goos, goarch, ext)

		var archiveData []byte
		var aerr error
		if goos == "windows" {
			archiveData, aerr = createZipArchive("redmine.exe", "new binary")
		} else {
			archiveData, aerr = createTarGzArchive("redmine", "new binary")
		}
		if aerr != nil {
			t.Fatalf("failed to create archive: %v", aerr)
		}

		apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(githubRelease{
				Tag: "v2.0.0",
				Assets: []githubAsset{
					{Name: assetName, DownloadURL: "http://download.example.com/file"},
				},
			})
		}))
		defer apiServer.Close()

		downloadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(archiveData)
		}))
		defer downloadServer.Close()

		callCount := 0
		origGet := httpGetFunc
		httpGetFunc = func(url string) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				return apiServer.Client().Get(apiServer.URL)
			}
			return downloadServer.Client().Get(downloadServer.URL)
		}
		defer func() { httpGetFunc = origGet }()

		origExe := osExecutableFunc
		osExecutableFunc = func() (string, error) { return "/nonexistent/path/redmine", nil }
		defer func() { osExecutableFunc = origExe }()

		origEval := filepathEvalSymlinks
		filepathEvalSymlinks = func(path string) (string, error) { return path, nil }
		defer func() { filepathEvalSymlinks = origEval }()

		err := runUpgrade(false, "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func createTarGzArchive(binaryName, content string) ([]byte, error) {
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gzw)

	hdr := &tar.Header{
		Name: binaryName,
		Mode: 0755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return nil, err
	}
	if _, err := tw.Write([]byte(content)); err != nil {
		return nil, err
	}
	tw.Close()
	gzw.Close()
	return buf.Bytes(), nil
}

func createZipArchive(binaryName, content string) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create(binaryName)
	if err != nil {
		return nil, err
	}
	if _, err := w.Write([]byte(content)); err != nil {
		return nil, err
	}
	zw.Close()
	return buf.Bytes(), nil
}

func createTarGzArchiveWithEntries(entries []struct{ name, content string; typeflag byte }) ([]byte, error) {
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gzw)

	for _, e := range entries {
		hdr := &tar.Header{
			Name:     e.name,
			Mode:     0755,
			Size:     int64(len(e.content)),
			Typeflag: e.typeflag,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return nil, err
		}
		if e.typeflag == tar.TypeReg {
			if _, err := tw.Write([]byte(e.content)); err != nil {
				return nil, err
			}
		}
	}
	tw.Close()
	gzw.Close()
	return buf.Bytes(), nil
}

func createZipArchiveWithEntries(entries []struct{ name, content string; isDir bool }) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	for _, e := range entries {
		if e.isDir {
			w, err := zw.Create(e.name + "/")
			if err != nil {
				return nil, err
			}
			_ = w
		} else {
			w, err := zw.Create(e.name)
			if err != nil {
				return nil, err
			}
			if _, err := w.Write([]byte(e.content)); err != nil {
				return nil, err
			}
		}
	}
	zw.Close()
	return buf.Bytes(), nil
}

func TestRunUpgrade(t *testing.T) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	var ext string
	switch goos {
	case "windows":
		ext = ".zip"
	default:
		ext = ".tar.gz"
	}

	assetName := fmt.Sprintf("redmine-cli_2.0.0_%s_%s%s", goos, goarch, ext)

	t.Run("check only mode", func(t *testing.T) {
		releaseJSON := `{"tag_name":"v2.0.0","assets":[]}`
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(releaseJSON))
		}))
		defer server.Close()

		origGet := httpGetFunc
		httpGetFunc = func(url string) (*http.Response, error) {
			return server.Client().Get(server.URL)
		}
		defer func() { httpGetFunc = origGet }()

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := runUpgrade(true, "")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "Current version:") {
			t.Errorf("expected 'Current version:' in output, got %s", output)
		}
		if !strings.Contains(output, "Latest version:") {
			t.Errorf("expected 'Latest version:' in output, got %s", output)
		}
	})

	t.Run("already up to date", func(t *testing.T) {
		origVersion := version
		version = "2.0.0"
		defer func() { version = origVersion }()

		releaseJSON := `{"tag_name":"v2.0.0","assets":[]}`
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(releaseJSON))
		}))
		defer server.Close()

		origGet := httpGetFunc
		httpGetFunc = func(url string) (*http.Response, error) {
			return server.Client().Get(server.URL)
		}
		defer func() { httpGetFunc = origGet }()

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := runUpgrade(false, "")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "Already up to date") {
			t.Errorf("expected 'Already up to date' in output, got %s", output)
		}
	})

	t.Run("fetch latest release error", func(t *testing.T) {
		origGet := httpGetFunc
		httpGetFunc = func(url string) (*http.Response, error) {
			return nil, errors.New("network error")
		}
		defer func() { httpGetFunc = origGet }()

		err := runUpgrade(false, "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("asset not found", func(t *testing.T) {
		origVersion := version
		version = "0.1.0"
		defer func() { version = origVersion }()

		releaseJSON := `{"tag_name":"v2.0.0","assets":[{"name":"redmine-cli_2.0.0_freebsd_amd64.tar.gz","browser_download_url":"https://example.com/file.tar.gz"}]}`
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(releaseJSON))
		}))
		defer server.Close()

		origGet := httpGetFunc
		httpGetFunc = func(url string) (*http.Response, error) {
			return server.Client().Get(server.URL)
		}
		defer func() { httpGetFunc = origGet }()

		err := runUpgrade(false, "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var appErr *apperrors.Error
		if !apperrors.As(err, &appErr) {
			t.Fatalf("expected *errors.Error, got %T", err)
		}
		if appErr.Category != "validation" {
			t.Errorf("expected validation error, got %s", appErr.Category)
		}
	})

	t.Run("target version flag", func(t *testing.T) {
		origVersion := version
		version = "2.0.0"
		defer func() { version = origVersion }()

		releaseJSON := `{"tag_name":"v3.0.0","assets":[]}`
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(releaseJSON))
		}))
		defer server.Close()

		origGet := httpGetFunc
		httpGetFunc = func(url string) (*http.Response, error) {
			return server.Client().Get(server.URL)
		}
		defer func() { httpGetFunc = origGet }()

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := runUpgrade(true, "v2.0.0")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "2.0.0") {
			t.Errorf("expected '2.0.0' in output, got %s", output)
		}
	})

	t.Run("successful upgrade", func(t *testing.T) {
		origVersion := version
		version = "0.1.0"
		defer func() { version = origVersion }()

		tmpDir := t.TempDir()
		fakeExe := filepath.Join(tmpDir, "redmine")
		os.WriteFile(fakeExe, []byte("old"), 0755)

		var archiveData []byte
		var err error
		if goos == "windows" {
			archiveData, err = createZipArchive("redmine.exe", "new binary")
		} else {
			archiveData, err = createTarGzArchive("redmine", "new binary")
		}
		if err != nil {
			t.Fatalf("failed to create archive: %v", err)
		}

		apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(githubRelease{
				Tag: "v2.0.0",
				Assets: []githubAsset{
					{Name: assetName, DownloadURL: "http://download.example.com/file"},
				},
			})
		}))
		defer apiServer.Close()

		downloadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(archiveData)
		}))
		defer downloadServer.Close()

		callCount := 0
		origGet := httpGetFunc
		httpGetFunc = func(url string) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				return apiServer.Client().Get(apiServer.URL)
			}
			return downloadServer.Client().Get(downloadServer.URL)
		}
		defer func() { httpGetFunc = origGet }()

		origExe := osExecutableFunc
		osExecutableFunc = func() (string, error) { return fakeExe, nil }
		defer func() { osExecutableFunc = origExe }()

		origEval := filepathEvalSymlinks
		filepathEvalSymlinks = func(path string) (string, error) { return path, nil }
		defer func() { filepathEvalSymlinks = origEval }()

		oldStdout := os.Stdout
		rOut, wOut, _ := os.Pipe()
		os.Stdout = wOut

		err = runUpgrade(false, "")

		wOut.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, rOut)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "Successfully upgraded") {
			t.Errorf("expected 'Successfully upgraded' in output, got %s", output)
		}

		gotContent, _ := os.ReadFile(fakeExe)
		if string(gotContent) != "new binary" {
			t.Errorf("expected 'new binary', got %q", string(gotContent))
		}
	})

	t.Run("os executable error", func(t *testing.T) {
		origVersion := version
		version = "0.1.0"
		defer func() { version = origVersion }()

		var archiveData []byte
		var err error
		if goos == "windows" {
			archiveData, err = createZipArchive("redmine.exe", "new binary")
		} else {
			archiveData, err = createTarGzArchive("redmine", "new binary")
		}
		if err != nil {
			t.Fatalf("failed to create archive: %v", err)
		}

		apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(githubRelease{
				Tag: "v2.0.0",
				Assets: []githubAsset{
					{Name: assetName, DownloadURL: "http://download.example.com/file"},
				},
			})
		}))
		defer apiServer.Close()

		downloadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(archiveData)
		}))
		defer downloadServer.Close()

		callCount := 0
		origGet := httpGetFunc
		httpGetFunc = func(url string) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				return apiServer.Client().Get(apiServer.URL)
			}
			return downloadServer.Client().Get(downloadServer.URL)
		}
		defer func() { httpGetFunc = origGet }()

		origExe := osExecutableFunc
		osExecutableFunc = func() (string, error) { return "", errors.New("exe error") }
		defer func() { osExecutableFunc = origExe }()

		err = runUpgrade(false, "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var appErr *apperrors.Error
		if !apperrors.As(err, &appErr) {
			t.Fatalf("expected *errors.Error, got %T", err)
		}
		if appErr.Category != "internal" {
			t.Errorf("expected internal error, got %s", appErr.Category)
		}
	})

	t.Run("eval symlinks error", func(t *testing.T) {
		origVersion := version
		version = "0.1.0"
		defer func() { version = origVersion }()

		var archiveData []byte
		var err error
		if goos == "windows" {
			archiveData, err = createZipArchive("redmine.exe", "new binary")
		} else {
			archiveData, err = createTarGzArchive("redmine", "new binary")
		}
		if err != nil {
			t.Fatalf("failed to create archive: %v", err)
		}

		apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(githubRelease{
				Tag: "v2.0.0",
				Assets: []githubAsset{
					{Name: assetName, DownloadURL: "http://download.example.com/file"},
				},
			})
		}))
		defer apiServer.Close()

		downloadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(archiveData)
		}))
		defer downloadServer.Close()

		callCount := 0
		origGet := httpGetFunc
		httpGetFunc = func(url string) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				return apiServer.Client().Get(apiServer.URL)
			}
			return downloadServer.Client().Get(downloadServer.URL)
		}
		defer func() { httpGetFunc = origGet }()

		origExe := osExecutableFunc
		osExecutableFunc = func() (string, error) { return "/some/path", nil }
		defer func() { osExecutableFunc = origExe }()

		origEval := filepathEvalSymlinks
		filepathEvalSymlinks = func(path string) (string, error) { return "", errors.New("symlink error") }
		defer func() { filepathEvalSymlinks = origEval }()

		err = runUpgrade(false, "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var appErr *apperrors.Error
		if !apperrors.As(err, &appErr) {
			t.Fatalf("expected *errors.Error, got %T", err)
		}
		if appErr.Category != "internal" {
			t.Errorf("expected internal error, got %s", appErr.Category)
		}
	})
}

func TestDownloadAndExtract(t *testing.T) {
	t.Run("successful download tar.gz", func(t *testing.T) {
		archiveData, err := createTarGzArchive("redmine", "binary content")
		if err != nil {
			t.Fatalf("failed to create archive: %v", err)
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(archiveData)
		}))
		defer server.Close()

		origGet := httpGetFunc
		httpGetFunc = func(url string) (*http.Response, error) {
			return server.Client().Get(server.URL)
		}
		defer func() { httpGetFunc = origGet }()

		rc, err := downloadAndExtract(server.URL, "linux")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer rc.Close()

		var buf bytes.Buffer
		io.Copy(&buf, rc)
		if buf.String() != "binary content" {
			t.Errorf("expected 'binary content', got %q", buf.String())
		}
	})

	t.Run("successful download zip", func(t *testing.T) {
		archiveData, err := createZipArchive("redmine.exe", "windows binary")
		if err != nil {
			t.Fatalf("failed to create archive: %v", err)
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(archiveData)
		}))
		defer server.Close()

		origGet := httpGetFunc
		httpGetFunc = func(url string) (*http.Response, error) {
			return server.Client().Get(server.URL)
		}
		defer func() { httpGetFunc = origGet }()

		rc, err := downloadAndExtract(server.URL, "windows")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer rc.Close()

		var buf bytes.Buffer
		io.Copy(&buf, rc)
		if buf.String() != "windows binary" {
			t.Errorf("expected 'windows binary', got %q", buf.String())
		}
	})

	t.Run("network error", func(t *testing.T) {
		origGet := httpGetFunc
		httpGetFunc = func(url string) (*http.Response, error) {
			return nil, errors.New("network error")
		}
		defer func() { httpGetFunc = origGet }()

		_, err := downloadAndExtract("http://example.com/file", "linux")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var appErr *apperrors.Error
		if !apperrors.As(err, &appErr) {
			t.Fatalf("expected *errors.Error, got %T", err)
		}
		if appErr.Category != "network" {
			t.Errorf("expected network error, got %s", appErr.Category)
		}
	})

	t.Run("non-200 status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer server.Close()

		origGet := httpGetFunc
		httpGetFunc = func(url string) (*http.Response, error) {
			return server.Client().Get(server.URL)
		}
		defer func() { httpGetFunc = origGet }()

		_, err := downloadAndExtract(server.URL, "linux")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var appErr *apperrors.Error
		if !apperrors.As(err, &appErr) {
			t.Fatalf("expected *errors.Error, got %T", err)
		}
		if appErr.Category != "network" {
			t.Errorf("expected network error, got %s", appErr.Category)
		}
	})

	t.Run("invalid archive", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("not a valid archive"))
		}))
		defer server.Close()

		origGet := httpGetFunc
		httpGetFunc = func(url string) (*http.Response, error) {
			return server.Client().Get(server.URL)
		}
		defer func() { httpGetFunc = origGet }()

		_, err := downloadAndExtract(server.URL, "linux")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestCleanupReadCloser(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "testfile")
	os.WriteFile(tmpFile, []byte("test"), 0644)

	f, _ := os.Open(tmpFile)
	cleanupDir := t.TempDir()
	os.WriteFile(filepath.Join(cleanupDir, "marker"), []byte("marker"), 0644)

	c := &cleanupReadCloser{ReadCloser: f, cleanupDir: cleanupDir}

	if _, err := os.Stat(filepath.Join(cleanupDir, "marker")); err != nil {
		t.Fatal("cleanup dir should exist before Close")
	}

	err := c.Close()
	if err != nil {
		t.Fatalf("unexpected error on Close: %v", err)
	}

	if _, err := os.Stat(cleanupDir); !os.IsNotExist(err) {
		t.Error("cleanup dir should be removed after Close")
	}
}

func TestExtractTarGz(t *testing.T) {
	t.Run("successful extraction", func(t *testing.T) {
		tmpDir := t.TempDir()
		archivePath := filepath.Join(tmpDir, "test.tar.gz")

		archiveData, err := createTarGzArchive("redmine", "#!/bin/sh\necho redmine")
		if err != nil {
			t.Fatalf("failed to create archive: %v", err)
		}
		os.WriteFile(archivePath, archiveData, 0644)

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

		if string(gotContent) != "#!/bin/sh\necho redmine" {
			t.Errorf("extracted content = %q, want %q", string(gotContent), "#!/bin/sh\necho redmine")
		}

		info, err := os.Stat(binaryPath)
		if err != nil {
			t.Fatalf("failed to stat extracted file: %v", err)
		}

		if info.Mode().Perm()&0111 == 0 {
			t.Error("expected extracted file to be executable")
		}
	})

	t.Run("archive with directory entries", func(t *testing.T) {
		tmpDir := t.TempDir()
		archivePath := filepath.Join(tmpDir, "test.tar.gz")

		entries := []struct {
			name     string
			content  string
			typeflag byte
		}{
			{"subdir/", "", tar.TypeDir},
			{"subdir/redmine", "binary content", tar.TypeReg},
		}
		archiveData, err := createTarGzArchiveWithEntries(entries)
		if err != nil {
			t.Fatalf("failed to create archive: %v", err)
		}
		os.WriteFile(archivePath, archiveData, 0644)

		destDir := t.TempDir()
		binaryPath, err := extractTarGz(archivePath, destDir)
		if err != nil {
			t.Fatalf("extractTarGz() error = %v", err)
		}

		gotContent, _ := os.ReadFile(binaryPath)
		if string(gotContent) != "binary content" {
			t.Errorf("extracted content = %q, want %q", string(gotContent), "binary content")
		}
	})

	t.Run("binary not found in archive", func(t *testing.T) {
		tmpDir := t.TempDir()
		archivePath := filepath.Join(tmpDir, "test.tar.gz")

		entries := []struct {
			name     string
			content  string
			typeflag byte
		}{
			{"other-file", "some content", tar.TypeReg},
		}
		archiveData, err := createTarGzArchiveWithEntries(entries)
		if err != nil {
			t.Fatalf("failed to create archive: %v", err)
		}
		os.WriteFile(archivePath, archiveData, 0644)

		destDir := t.TempDir()
		_, err = extractTarGz(archivePath, destDir)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "binary not found") {
			t.Errorf("expected 'binary not found' error, got %v", err)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		destDir := t.TempDir()
		_, err := extractTarGz("/nonexistent/path.tar.gz", destDir)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("invalid gzip data", func(t *testing.T) {
		tmpDir := t.TempDir()
		archivePath := filepath.Join(tmpDir, "test.tar.gz")
		os.WriteFile(archivePath, []byte("not gzip data"), 0644)

		destDir := t.TempDir()
		_, err := extractTarGz(archivePath, destDir)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestExtractZip(t *testing.T) {
	t.Run("successful extraction", func(t *testing.T) {
		tmpDir := t.TempDir()
		archivePath := filepath.Join(tmpDir, "test.zip")

		archiveData, err := createZipArchive("redmine.exe", "windows binary")
		if err != nil {
			t.Fatalf("failed to create archive: %v", err)
		}
		os.WriteFile(archivePath, archiveData, 0644)

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

		if string(gotContent) != "windows binary" {
			t.Errorf("extracted content = %q, want %q", string(gotContent), "windows binary")
		}

		info, err := os.Stat(binaryPath)
		if err != nil {
			t.Fatalf("failed to stat extracted file: %v", err)
		}

		if info.Mode().Perm()&0111 == 0 {
			t.Error("expected extracted file to be executable")
		}
	})

	t.Run("archive with directory entries", func(t *testing.T) {
		tmpDir := t.TempDir()
		archivePath := filepath.Join(tmpDir, "test.zip")

		entries := []struct {
			name    string
			content string
			isDir   bool
		}{
			{"subdir", "", true},
			{"subdir/redmine.exe", "binary content", false},
		}
		archiveData, err := createZipArchiveWithEntries(entries)
		if err != nil {
			t.Fatalf("failed to create archive: %v", err)
		}
		os.WriteFile(archivePath, archiveData, 0644)

		destDir := t.TempDir()
		binaryPath, err := extractZip(archivePath, destDir)
		if err != nil {
			t.Fatalf("extractZip() error = %v", err)
		}

		gotContent, _ := os.ReadFile(binaryPath)
		if string(gotContent) != "binary content" {
			t.Errorf("extracted content = %q, want %q", string(gotContent), "binary content")
		}
	})

	t.Run("binary not found in archive", func(t *testing.T) {
		tmpDir := t.TempDir()
		archivePath := filepath.Join(tmpDir, "test.zip")

		entries := []struct {
			name    string
			content string
			isDir   bool
		}{
			{"other-file.txt", "some content", false},
		}
		archiveData, err := createZipArchiveWithEntries(entries)
		if err != nil {
			t.Fatalf("failed to create archive: %v", err)
		}
		os.WriteFile(archivePath, archiveData, 0644)

		destDir := t.TempDir()
		_, err = extractZip(archivePath, destDir)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "binary not found") {
			t.Errorf("expected 'binary not found' error, got %v", err)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		destDir := t.TempDir()
		_, err := extractZip("/nonexistent/path.zip", destDir)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestReplaceBinary(t *testing.T) {
	t.Run("successful replacement", func(t *testing.T) {
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
	})

	t.Run("cannot create temp file", func(t *testing.T) {
		err := replaceBinary("/nonexistent/dir/redmine", strings.NewReader("data"))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var appErr *apperrors.Error
		if !apperrors.As(err, &appErr) {
			t.Fatalf("expected *errors.Error, got %T", err)
		}
		if appErr.Category != "internal" {
			t.Errorf("expected internal error, got %s", appErr.Category)
		}
	})

	t.Run("write error", func(t *testing.T) {
		tmpDir := t.TempDir()
		currentPath := filepath.Join(tmpDir, "redmine")

		err := replaceBinary(currentPath, &upgradeErrorReader{})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var appErr *apperrors.Error
		if !apperrors.As(err, &appErr) {
			t.Fatalf("expected *errors.Error, got %T", err)
		}
		if appErr.Category != "internal" {
			t.Errorf("expected internal error, got %s", appErr.Category)
		}
	})

	t.Run("rename error - target is directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		currentPath := filepath.Join(tmpDir, "redmine")
		os.WriteFile(currentPath, []byte("old"), 0755)

		targetDir := filepath.Join(tmpDir, "target_dir")
		os.Mkdir(targetDir, 0755)

		err := replaceBinary(targetDir, strings.NewReader("new"))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var appErr *apperrors.Error
		if !apperrors.As(err, &appErr) {
			t.Fatalf("expected *apperrors.Error, got %T", err)
		}
		if appErr.Category != "internal" {
			t.Errorf("expected internal error, got %s", appErr.Category)
		}
		if appErr.Hint == "" {
			t.Error("expected hint in error")
		}
		if len(appErr.Actions) == 0 {
			t.Error("expected actions in error")
		}
	})
}

type upgradeErrorReader struct{}

func (r *upgradeErrorReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestDownloadAndExtract_MkdirTempError(t *testing.T) {
	archiveData, err := createTarGzArchive("redmine", "binary content")
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(archiveData)
	}))
	defer server.Close()

	origGet := httpGetFunc
	httpGetFunc = func(url string) (*http.Response, error) {
		return server.Client().Get(server.URL)
	}
	defer func() { httpGetFunc = origGet }()

	origMkdirTemp := osMkdirTempFunc
	osMkdirTempFunc = func(dir, pattern string) (string, error) {
		return "", errors.New("mkdirtemp error")
	}
	defer func() { osMkdirTempFunc = origMkdirTemp }()

	_, err = downloadAndExtract(server.URL, "linux")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var appErr *apperrors.Error
	if !apperrors.As(err, &appErr) {
		t.Fatalf("expected *apperrors.Error, got %T", err)
	}
	if appErr.Category != "internal" {
		t.Errorf("expected internal error, got %s", appErr.Category)
	}
}

func TestDownloadAndExtract_CreateError(t *testing.T) {
	archiveData, err := createTarGzArchive("redmine", "binary content")
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(archiveData)
	}))
	defer server.Close()

	origGet := httpGetFunc
	httpGetFunc = func(url string) (*http.Response, error) {
		return server.Client().Get(server.URL)
	}
	defer func() { httpGetFunc = origGet }()

	origCreate := osCreateFunc
	osCreateFunc = func(name string) (*os.File, error) {
		return nil, errors.New("create error")
	}
	defer func() { osCreateFunc = origCreate }()

	_, err = downloadAndExtract(server.URL, "linux")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var appErr *apperrors.Error
	if !apperrors.As(err, &appErr) {
		t.Fatalf("expected *apperrors.Error, got %T", err)
	}
	if appErr.Category != "internal" {
		t.Errorf("expected internal error, got %s", appErr.Category)
	}
}

func TestDownloadAndExtract_WriteError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("some data"))
	}))
	defer server.Close()

	origGet := httpGetFunc
	httpGetFunc = func(url string) (*http.Response, error) {
		return server.Client().Get(server.URL)
	}
	defer func() { httpGetFunc = origGet }()

	origCreate := osCreateFunc
	osCreateFunc = func(name string) (*os.File, error) {
		return os.Open(os.DevNull)
	}
	defer func() { osCreateFunc = origCreate }()

	_, err := downloadAndExtract(server.URL, "linux")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDownloadAndExtract_OpenBinaryError(t *testing.T) {
	archiveData, err := createTarGzArchive("redmine", "binary content")
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(archiveData)
	}))
	defer server.Close()

	origGet := httpGetFunc
	httpGetFunc = func(url string) (*http.Response, error) {
		return server.Client().Get(server.URL)
	}
	defer func() { httpGetFunc = origGet }()

	openCount := 0
	origOpen := osOpenFunc
	osOpenFunc = func(name string) (*os.File, error) {
		openCount++
		if openCount > 1 {
			return nil, errors.New("open error")
		}
		return origOpen(name)
	}
	defer func() { osOpenFunc = origOpen }()

	_, err = downloadAndExtract(server.URL, "linux")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var appErr *apperrors.Error
	if !apperrors.As(err, &appErr) {
		t.Fatalf("expected *apperrors.Error, got %T", err)
	}
	if appErr.Category != "internal" {
		t.Errorf("expected internal error, got %s", appErr.Category)
	}
}

func TestExtractTarGz_TrNextError(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "test.tar.gz")

	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gzw)

	hdr1 := &tar.Header{
		Name:     "other-file",
		Mode:     0644,
		Size:     int64(len("some data")),
		Typeflag: tar.TypeReg,
	}
	tw.WriteHeader(hdr1)
	tw.Write([]byte("some data"))

	tw.Flush()
	gzw.Flush()

	tw.Close()
	gzw.Close()

	data := buf.Bytes()
	corruptOffset := len(data) - 20
	if corruptOffset < 0 {
		corruptOffset = 0
	}
	for i := corruptOffset; i < len(data); i++ {
		data[i] = 0xFF
	}
	os.WriteFile(archivePath, data, 0644)

	destDir := t.TempDir()
	_, err := extractTarGz(archivePath, destDir)
	if err == nil {
		t.Fatal("expected error from corrupted tar.gz, got nil")
	}
}

func TestExtractTarGz_OpenFileError(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "test.tar.gz")

	archiveData, err := createTarGzArchive("redmine", "binary content")
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}
	os.WriteFile(archivePath, archiveData, 0644)

	origOpenFile := osOpenFileFunc
	osOpenFileFunc = func(name string, flag int, perm os.FileMode) (*os.File, error) {
		return nil, errors.New("openfile error")
	}
	defer func() { osOpenFileFunc = origOpenFile }()

	destDir := t.TempDir()
	_, err = extractTarGz(archivePath, destDir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExtractTarGz_IoCopyError(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "test.tar.gz")

	archiveData, err := createTarGzArchive("redmine", "binary content")
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}
	os.WriteFile(archivePath, archiveData, 0644)

	origIoCopy := ioCopyFunc
	ioCopyFunc = func(dst io.Writer, src io.Reader) (int64, error) {
		return 0, errors.New("io copy error")
	}
	defer func() { ioCopyFunc = origIoCopy }()

	destDir := t.TempDir()
	_, err = extractTarGz(archivePath, destDir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExtractZip_OpenFileError(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "test.zip")

	archiveData, err := createZipArchive("redmine.exe", "binary content")
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}
	os.WriteFile(archivePath, archiveData, 0644)

	origOpenFile := osOpenFileFunc
	osOpenFileFunc = func(name string, flag int, perm os.FileMode) (*os.File, error) {
		return nil, errors.New("openfile error")
	}
	defer func() { osOpenFileFunc = origOpenFile }()

	destDir := t.TempDir()
	_, err = extractZip(archivePath, destDir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExtractZip_IoCopyError(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "test.zip")

	archiveData, err := createZipArchive("redmine.exe", "binary content")
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}
	os.WriteFile(archivePath, archiveData, 0644)

	origIoCopy := ioCopyFunc
	ioCopyFunc = func(dst io.Writer, src io.Reader) (int64, error) {
		return 0, errors.New("io copy error")
	}
	defer func() { ioCopyFunc = origIoCopy }()

	destDir := t.TempDir()
	_, err = extractZip(archivePath, destDir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExtractZip_FileOpenError(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "test.zip")

	archiveData, err := createZipArchive("redmine.exe", "binary content")
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}
	os.WriteFile(archivePath, archiveData, 0644)

	origIoCopy := ioCopyFunc
	ioCopyFunc = func(dst io.Writer, src io.Reader) (int64, error) {
		return 0, errors.New("simulated io.Copy error")
	}
	defer func() { ioCopyFunc = origIoCopy }()

	destDir := t.TempDir()
	_, err = extractZip(archivePath, destDir)
	if err == nil {
		t.Fatal("expected error from f.Open() failure, got nil")
	}
}

func TestExtractZip_FOpenError(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "test.zip")

	archiveData, err := createZipArchive("redmine.exe", "binary content")
	if err != nil {
		t.Fatalf("failed to create archive: %v", err)
	}
	os.WriteFile(archivePath, archiveData, 0644)

	f, err := os.Open(archivePath)
	if err != nil {
		t.Fatalf("failed to open archive: %v", err)
	}
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		t.Fatalf("failed to stat archive: %v", err)
	}

	zr, err := zip.NewReader(f, fi.Size())
	if err != nil {
		f.Close()
		t.Fatalf("failed to read zip: %v", err)
	}

	if len(zr.File) > 0 {
		zr.File[0].CompressedSize64 = 0
		zr.File[0].UncompressedSize64 = 999
		zr.File[0].Method = 99
	}

	origZipOpenReader := zipOpenReaderFunc
	zipOpenReaderFunc = func(name string) (*zip.ReadCloser, error) {
		return &zip.ReadCloser{Reader: *zr}, nil
	}
	defer func() { zipOpenReaderFunc = origZipOpenReader }()

	destDir := t.TempDir()
	_, err = extractZip(archivePath, destDir)
	if err == nil {
		t.Fatal("expected error from f.Open() failure, got nil")
	}
}

func TestReplaceBinary_ChmodError(t *testing.T) {
	tmpDir := t.TempDir()
	currentPath := filepath.Join(tmpDir, "redmine")
	os.WriteFile(currentPath, []byte("old"), 0755)

	origChmod := osChmodFunc
	osChmodFunc = func(f *os.File, mode os.FileMode) error {
		return errors.New("chmod error")
	}
	defer func() { osChmodFunc = origChmod }()

	err := replaceBinary(currentPath, strings.NewReader("new"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var appErr *apperrors.Error
	if !apperrors.As(err, &appErr) {
		t.Fatalf("expected *apperrors.Error, got %T", err)
	}
	if appErr.Category != "internal" {
		t.Errorf("expected internal error, got %s", appErr.Category)
	}
}
