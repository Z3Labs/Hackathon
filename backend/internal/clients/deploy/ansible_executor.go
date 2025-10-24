package deploy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Z3Labs/Hackathon/backend/internal/model"
)

type AnsibleExecutor struct {
	config ExecutorConfig
	status *model.NodeStatusRecord
}

func NewAnsibleExecutor(config ExecutorConfig) *AnsibleExecutor {
	return &AnsibleExecutor{
		config: config,
		status: &model.NodeStatusRecord{
			Host:             config.Host,
			Service:          config.Service,
			CurrentVersion:   config.Version,
			DeployingVersion: config.Version,
			PrevVersion:      config.PrevVersion,
			Platform:         model.PlatformPhysical,
			State:            model.NodeStatusPending,
		},
	}
}

func (a *AnsibleExecutor) Deploy(ctx context.Context) error {
	a.status.State = model.NodeStatusDeploying
	a.status.UpdatedAt = time.Now()

	releaseDir := fmt.Sprintf("/opt/releases/%s/%s", a.config.Service, a.config.Version)
	if err := os.MkdirAll(releaseDir, 0755); err != nil {
		return fmt.Errorf("failed to create release directory: %w", err)
	}

	packagePath := filepath.Join(releaseDir, fmt.Sprintf("%s.tar.gz", a.config.Service))
	if err := a.downloadPackage(ctx, a.config.PackageURL, packagePath); err != nil {
		return fmt.Errorf("failed to download package: %w", err)
	}

	if err := a.verifyChecksum(packagePath, a.config.SHA256); err != nil {
		return fmt.Errorf("checksum verification failed: %w", err)
	}

	if err := a.extractPackage(releaseDir, packagePath); err != nil {
		return fmt.Errorf("failed to extract package: %w", err)
	}

	if err := a.createSystemdService(); err != nil {
		return fmt.Errorf("failed to create systemd service: %w", err)
	}

	if err := a.reloadSystemd(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	if err := a.restartService(); err != nil {
		return fmt.Errorf("failed to restart service: %w", err)
	}

	a.status.State = model.NodeStatusSuccess
	a.status.CurrentVersion = a.config.Version
	a.status.PrevVersion = a.config.Version
	a.status.DeployingVersion = ""
	a.status.UpdatedAt = time.Now()

	return nil
}

func (a *AnsibleExecutor) Rollback(ctx context.Context) error {
	if a.config.PrevVersion == "" {
		return fmt.Errorf("no previous version to rollback to")
	}

	a.status.State = model.NodeStatusDeploying

	serviceName := fmt.Sprintf("%s@%s", a.config.Service, a.config.PrevVersion)
	cmd := exec.CommandContext(ctx, "systemctl", "restart", serviceName)
	if err := cmd.Run(); err != nil {
		a.status.State = model.NodeStatusFailed
		a.status.LastError = fmt.Sprintf("rollback failed: %v", err)
		return fmt.Errorf("failed to restart service for rollback: %w", err)
	}

	a.status.State = model.NodeStatusRolledBack
	a.status.CurrentVersion = a.config.PrevVersion
	a.status.DeployingVersion = ""
	a.status.UpdatedAt = time.Now()

	return nil
}

func (a *AnsibleExecutor) GetStatus(ctx context.Context) (*model.NodeStatusRecord, error) {
	return a.status, nil
}

func (a *AnsibleExecutor) downloadPackage(ctx context.Context, url, destPath string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download package: status %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (a *AnsibleExecutor) verifyChecksum(filePath, expectedSHA256 string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	actualSHA256 := hex.EncodeToString(hash.Sum(nil))
	if actualSHA256 != expectedSHA256 {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedSHA256, actualSHA256)
	}

	return nil
}

func (a *AnsibleExecutor) extractPackage(destDir, packagePath string) error {
	cmd := exec.Command("tar", "-xzf", packagePath, "-C", destDir)
	return cmd.Run()
}

func (a *AnsibleExecutor) createSystemdService() error {
	serviceContent := fmt.Sprintf(`[Unit]
Description=%s service version %s
After=network.target

[Service]
Type=simple
ExecStart=/opt/releases/%s/%s/%s
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
`, a.config.Service, a.config.Version, a.config.Service, a.config.Version, a.config.Service)

	servicePath := fmt.Sprintf("/etc/systemd/system/%s@%s.service", a.config.Service, a.config.Version)
	return os.WriteFile(servicePath, []byte(serviceContent), 0644)
}

func (a *AnsibleExecutor) reloadSystemd() error {
	cmd := exec.Command("systemctl", "daemon-reload")
	return cmd.Run()
}

func (a *AnsibleExecutor) restartService() error {
	serviceName := fmt.Sprintf("%s@%s", a.config.Service, a.config.Version)
	cmd := exec.Command("systemctl", "restart", serviceName)
	return cmd.Run()
}
