// Package app provides the CLI application commands and logic.
package app

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/largeoliu/redmine-cli/internal/errors"
)

type githubRelease struct {
	Tag    string         `json:"tag_name"`
	Assets []githubAsset  `json:"assets"`
}

type githubAsset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
}

func newUpgradeCommand() *cobra.Command {
	var checkOnly bool
	var targetVersion string

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade redmine CLI to the latest version",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runUpgrade(checkOnly, targetVersion)
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "Only check for updates, don't upgrade")
	cmd.Flags().StringVar(&targetVersion, "version", "", "Upgrade to a specific version instead of latest")

	return cmd
}

func runUpgrade(checkOnly bool, targetVersion string) error {
	release, err := fetchLatestRelease()
	if err != nil {
		return err
	}

	latestVersion := strings.TrimPrefix(release.Tag, "v")

	currentVersion := strings.TrimPrefix(version, "v")

	if targetVersion != "" {
		latestVersion = strings.TrimPrefix(targetVersion, "v")
	}

	if checkOnly {
		fmt.Printf("Current version: %s\n", currentVersion)
		fmt.Printf("Latest version:  %s\n", latestVersion)
		return nil
	}

	cmp := compareVersions(currentVersion, latestVersion)
	if cmp >= 0 {
		fmt.Printf("%s Already up to date (v%s)\n", green("✓"), currentVersion)
		return nil
	}

	goos := runtime.GOOS
	goarch := runtime.GOARCH

	asset := findAsset(release.Assets, latestVersion, goos, goarch)
	if asset == nil {
		return errors.NewValidation(
			fmt.Sprintf("No matching release found for %s/%s", goos, goarch),
			errors.WithHint("Visit https://github.com/largeoliu/redmine-cli/releases to download manually"),
		)
	}

	fmt.Printf("Upgrading from v%s to v%s...\n", currentVersion, latestVersion)

	binary, err := downloadAndExtract(asset.DownloadURL, goos)
	if err != nil {
		return err
	}
	defer binary.Close()

	exePath, err := os.Executable()
	if err != nil {
		return errors.NewInternal("failed to get current executable path", errors.WithCause(err))
	}

	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return errors.NewInternal("failed to resolve executable path", errors.WithCause(err))
	}

	if err := replaceBinary(exePath, binary); err != nil {
		return err
	}

	fmt.Printf("%s Successfully upgraded from v%s to v%s\n", green("✓"), currentVersion, latestVersion)
	return nil
}

func fetchLatestRelease() (*githubRelease, error) {
	url := "https://api.github.com/repos/largeoliu/redmine-cli/releases/latest"

	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.NewNetwork("failed to fetch release info", errors.WithCause(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewNetwork(
			fmt.Sprintf("GitHub API returned status %d", resp.StatusCode),
			errors.WithHint("Check your internet connection or try again later"),
		)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, errors.NewInternal("failed to parse release info", errors.WithCause(err))
	}

	return &release, nil
}

func findAsset(assets []githubAsset, version, goos, goarch string) *githubAsset {
	var ext string
	switch goos {
	case "windows":
		ext = ".zip"
	default:
		ext = ".tar.gz"
	}

	prefix := fmt.Sprintf("redmine-cli_%s_%s_%s", version, goos, goarch)

	for i := range assets {
		name := assets[i].Name
		if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, ext) {
			return &assets[i]
		}
	}

	return nil
}

func compareVersions(current, latest string) int {
	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")

	maxLen := len(currentParts)
	if len(latestParts) > maxLen {
		maxLen = len(latestParts)
	}

	for i := 0; i < maxLen; i++ {
		var c, l int
		if i < len(currentParts) {
			c, _ = strconv.Atoi(currentParts[i])
		}
		if i < len(latestParts) {
			l, _ = strconv.Atoi(latestParts[i])
		}
		if c < l {
			return -1
		}
		if c > l {
			return 1
		}
	}

	return 0
}

func downloadAndExtract(url, goos string) (io.ReadCloser, error) {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return nil, errors.NewNetwork("failed to download release archive", errors.WithCause(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewNetwork(
			fmt.Sprintf("download failed with status %d", resp.StatusCode),
		)
	}

	tmpDir, err := os.MkdirTemp("", "redmine-upgrade-*")
	if err != nil {
		return nil, errors.NewInternal("failed to create temp directory", errors.WithCause(err))
	}

	archivePath := filepath.Join(tmpDir, "archive")
	f, err := os.Create(archivePath)
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, errors.NewInternal("failed to create temp file", errors.WithCause(err))
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.RemoveAll(tmpDir)
		return nil, errors.NewNetwork("failed to write archive", errors.WithCause(err))
	}
	f.Close()

	var binaryPath string

	switch goos {
	case "windows":
		binaryPath, err = extractZip(archivePath, tmpDir)
	default:
		binaryPath, err = extractTarGz(archivePath, tmpDir)
	}

	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, errors.NewInternal("failed to extract archive", errors.WithCause(err))
	}

	binaryFile, err := os.Open(binaryPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, errors.NewInternal("failed to open extracted binary", errors.WithCause(err))
	}

	return &cleanupReadCloser{ReadCloser: binaryFile, cleanupDir: tmpDir}, nil
}

type cleanupReadCloser struct {
	io.ReadCloser
	cleanupDir string
}

func (c *cleanupReadCloser) Close() error {
	err := c.ReadCloser.Close()
	os.RemoveAll(c.cleanupDir)
	return err
}

func extractTarGz(archivePath, destDir string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		name := filepath.Base(hdr.Name)
		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		if name == "redmine" || name == "redmine.exe" {
			outPath := filepath.Join(destDir, name)
			outFile, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(outFile, io.LimitReader(tr, 100*1024*1024)); err != nil {
				outFile.Close()
				return "", err
			}
			outFile.Close()
			return outPath, nil
		}
	}

	return "", fmt.Errorf("binary not found in archive")
}

func extractZip(archivePath, destDir string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, f := range r.File {
		name := filepath.Base(f.Name)
		if f.FileInfo().IsDir() {
			continue
		}

		if name == "redmine" || name == "redmine.exe" {
			outPath := filepath.Join(destDir, name)
			outFile, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				return "", err
			}
			rc, err := f.Open()
			if err != nil {
				outFile.Close()
				return "", err
			}
			if _, err := io.Copy(outFile, io.LimitReader(rc, 100*1024*1024)); err != nil {
				rc.Close()
				outFile.Close()
				return "", err
			}
			rc.Close()
			outFile.Close()
			return outPath, nil
		}
	}

	return "", fmt.Errorf("binary not found in archive")
}

func replaceBinary(currentPath string, newBinary io.Reader) error {
	dir := filepath.Dir(currentPath)
	tmpFile, err := os.CreateTemp(dir, "redmine-upgrade-*")
	if err != nil {
		return errors.NewInternal("failed to create temp file for replacement", errors.WithCause(err))
	}
	tmpPath := tmpFile.Name()

	if _, err := io.Copy(tmpFile, newBinary); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return errors.NewInternal("failed to write new binary", errors.WithCause(err))
	}

	if err := tmpFile.Chmod(0755); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return errors.NewInternal("failed to set permissions on new binary", errors.WithCause(err))
	}

	tmpFile.Close()

	if err := os.Rename(tmpPath, currentPath); err != nil {
		os.Remove(tmpPath)
		return errors.NewInternal(
			"failed to replace binary",
			errors.WithCause(err),
			errors.WithHint("You may need to run the upgrade with elevated permissions"),
			errors.WithActions(
				fmt.Sprintf("sudo mv %s %s", tmpPath, currentPath),
				"Or run: sudo redmine upgrade",
			),
		)
	}

	return nil
}
