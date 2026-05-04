package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
)

// CheckAndUpdate checks for a new version and applies the update if available.
// It downloads the binary from the server, verifies the checksum, and replaces
// the current binary atomically.
func CheckAndUpdate(currentVersion, serverURL, binaryPath string, logger *zap.Logger) error {
	client := &http.Client{Timeout: 30 * time.Second}

	version, err := fetchVersion(client, serverURL)
	if err != nil {
		return fmt.Errorf("failed to fetch version: %w", err)
	}

	if version.Version == currentVersion {
		return nil
	}

	logger.Info("New version available",
		zap.String("current", currentVersion),
		zap.String("latest", version.Version))

	tempPath := binaryPath + ".new"
	if err := downloadBinary(client, serverURL, tempPath); err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}

	if version.Checksum != "" {
		checksum, err := fileChecksum(tempPath)
		if err != nil {
			os.Remove(tempPath)
			return fmt.Errorf("failed to compute checksum: %w", err)
		}
		if checksum != version.Checksum {
			os.Remove(tempPath)
			return fmt.Errorf("checksum mismatch: expected %s, got %s", version.Checksum, checksum)
		}
	}

	if err := os.Chmod(tempPath, 0755); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	if err := os.Rename(tempPath, binaryPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	logger.Info("Update applied successfully", zap.String("version", version.Version))
	return nil
}

type versionInfo struct {
	Version  string `json:"version"`
	Checksum string `json:"checksum"`
}

func fetchVersion(client *http.Client, serverURL string) (versionInfo, error) {
	url := fmt.Sprintf("%s/client/version", serverURL)
	resp, err := client.Get(url)
	if err != nil {
		return versionInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return versionInfo{}, fmt.Errorf("status %d", resp.StatusCode)
	}

	var info versionInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return versionInfo{}, err
	}

	return info, nil
}

func downloadBinary(client *http.Client, serverURL, destPath string) error {
	url := fmt.Sprintf("%s/client/binary?os=linux&arch=amd64", serverURL)
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func fileChecksum(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}
