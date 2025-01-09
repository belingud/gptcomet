package cmd

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
	"strings"

	"github.com/spf13/cobra"
)

// GithubRelease represents the GitHub release API response structure
type GithubRelease struct {
	TagName string `json:"tag_name"`
}

// NewUpdateCmd creates a new cobra command for handling version updates
// It automatically downloads and installs the latest version if available
func NewUpdateCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update gptcomet to latest version",
		Long: `Update gptcomet to the latest version from GitHub releases.
For Unix-like systems, it installs to ~/.local/bin/
For Windows, it installs to %USERPROFILE%\.gptcomet\`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return checkUpdate(version)
		},
	}
}

// checkUpdate checks GitHub for a newer version and installs it if available
func checkUpdate(currentVersion string) error {
	fmt.Println("Checking for updates...")
	resp, err := http.Get("https://api.github.com/repos/belingud/gptcomet/releases/latest")
	if err != nil {
		return fmt.Errorf("failed to check updates: %v", err)
	}
	defer resp.Body.Close()

	var release GithubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	if latestVersion == currentVersion {
		fmt.Println("You are using the latest version:", currentVersion)
		return nil
	}

	fmt.Printf("Found new version: %s (current: %s)\n", latestVersion, currentVersion)
	fmt.Println("Starting update...")

	if err := installUpdate(latestVersion, release.TagName); err != nil {
		return fmt.Errorf("❌ Update failed: %v", err)
	}

	return nil
}

// installUpdate downloads and installs the specified version
// Parameters:
//   - version: The version to install (without 'v' prefix)
//   - tag: The complete tag name (with 'v' prefix)
//
// Returns an error if the installation fails
func installUpdate(version, tag string) error {
	// Create temporary directory for downloads
	tempDir, err := os.MkdirTemp("", "gptcomet-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Build download URL based on platform and architecture
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Select appropriate file extension for the platform
	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}

	fileName := fmt.Sprintf("gptcomet_%s_%s_%s%s", version, osName, arch, ext)
	downloadURL := fmt.Sprintf("https://github.com/belingud/gptcomet/releases/download/%s/%s", tag, fileName)

	// Download the release archive
	fmt.Printf("Downloading %s...\n", downloadURL)
	archivePath := filepath.Join(tempDir, fileName)
	if err := downloadFile(downloadURL, archivePath); err != nil {
		return fmt.Errorf("download failed: %v", err)
	}

	// Extract based on archive type
	if ext == ".zip" {
		if err := unzip(archivePath, tempDir); err != nil {
			return fmt.Errorf("extract failed: %v", err)
		}
	} else {
		if err := untargz(archivePath, tempDir); err != nil {
			return fmt.Errorf("extract failed: %v", err)
		}
	}

	// Install the extracted binary
	var installDir string
	if runtime.GOOS == "windows" {
		installDir = filepath.Join(os.Getenv("USERPROFILE"), ".gptcomet")
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %v", err)
		}
		installDir = filepath.Join(homeDir, ".local", "bin")
	}

	// Create install directory if it doesn't exist
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %v", err)
	}

	// Copy the binary to the install directory
	exeSuffix := ""
	if runtime.GOOS == "windows" {
		exeSuffix = ".exe"
	}

	srcPath := filepath.Join(tempDir, "gptcomet"+exeSuffix)
	dstPath := filepath.Join(installDir, "gptcomet"+exeSuffix)

	// Copy the new version to a temporary file
	tempDstPath := dstPath + ".tmp"
	if err := copyFile(srcPath, tempDstPath); err != nil {
		os.Remove(tempDstPath) // Clean up temp file
		return fmt.Errorf("failed to copy new version: %v", err)
	}

	// Replace the existing binary with the new version on Windows
	if runtime.GOOS == "windows" {
		if err := os.Remove(dstPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove existing file: %v (try again or download new version manually)", err)
		}
		if err := os.Rename(tempDstPath, dstPath); err != nil {
			return fmt.Errorf("failed to install new version: %v", err)
		}
	} else {
		// Atomic replacement on Unix systems
		if err := os.Rename(tempDstPath, dstPath); err != nil {
			os.Remove(tempDstPath) // Clean up temp file
			return fmt.Errorf("failed to install new version: %v", err)
		}
	}

	// Create symlink for Unix-like systems
	if runtime.GOOS != "windows" {
		gmsgPath := filepath.Join(installDir, "gmsg")

		// Remove existing gmsg file or symlink
		_ = os.Remove(gmsgPath)

		// Create symlink from gmsg to gptcomet
		if err := os.Symlink(dstPath, gmsgPath); err != nil {
			return fmt.Errorf("failed to create gmsg symlink: %v", err)
		}

		fmt.Printf("Created symlink: %s -> %s\n", gmsgPath, dstPath)
	}

	fmt.Printf("✅ Successfully installed gptcomet %s to %s\n", version, installDir)
	return nil
}

// downloadFile downloads a file from url to the specified destination
// Parameters:
//   - url: The URL to download from
//   - dst: The local path to save the file to
//
// Returns an error if the download fails
func downloadFile(url string, dst string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the total size of the file
	totalSize := resp.ContentLength

	// Create a progress writer
	pw := &progressWriter{total: totalSize}

	// Copy the response body to the file with progress
	_, err = io.Copy(out, io.TeeReader(resp.Body, pw))
	return err
}

// progressWriter wraps an io.Writer and displays download progress
type progressWriter struct {
	total      int64
	current    int64
	lastUpdate int64
}

// Write implements io.Writer and updates the progress bar
func (pw *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.current += int64(n)

	// Update progress every 100KB or when complete
	if pw.current-pw.lastUpdate > 102400 || pw.current == pw.total {
		pw.lastUpdate = pw.current
		pw.displayProgress()
	}
	return n, nil
}

// displayProgress shows the current download progress
func (pw *progressWriter) displayProgress() {
	const width = 50
	progress := float64(pw.current) / float64(pw.total)
	bar := strings.Repeat("=", int(progress*width)) + strings.Repeat(" ", width-int(progress*width))
	fmt.Printf("\r[%s] %.2f%%", bar, progress*100)
	if pw.current == pw.total {
		fmt.Println()
	}
}

// unzip extracts a zip archive to the specified destination
// Parameters:
//   - src: Path to the zip file
//   - dst: Directory to extract to
//
// Returns an error if the extraction fails
func unzip(src, dst string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}

		path := filepath.Join(dst, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), 0755)
			outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				rc.Close()
				return err
			}

			_, err = io.Copy(outFile, rc)
			outFile.Close()
			if err != nil {
				rc.Close()
				return err
			}
		}
		rc.Close()
	}
	return nil
}

// untargz extracts a tar.gz archive to the specified destination
// Parameters:
//   - src: Path to the tar.gz file
//   - dst: Directory to extract to
//
// Returns an error if the extraction fails
func untargz(src, dst string) error {
	gzipFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer gzipFile.Close()

	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		path := filepath.Join(dst, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}

// copyFile copies a file from src to dst preserving file mode
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	// Get original file mode
	info, err := in.Stat()
	if err != nil {
		return err
	}

	// Create new file with same permissions
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	// Ensure all data is written to disk
	return out.Sync()
}
