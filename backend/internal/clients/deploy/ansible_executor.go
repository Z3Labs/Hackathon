package deploy

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Z3Labs/Hackathon/backend/internal/model"
)

type AnsibleExecutor struct {
	config       ExecutorConfig
	status       *model.NodeStatusRecord
	playbookPath string
}

func NewAnsibleExecutor(config ExecutorConfig) *AnsibleExecutor {
	playbookPath := "/workspace/backend/playbooks/deploy.yml"
	if path := os.Getenv("ANSIBLE_PLAYBOOK_PATH"); path != "" {
		playbookPath = path
	}

	return &AnsibleExecutor{
		config:       config,
		playbookPath: playbookPath,
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

	extraVars := fmt.Sprintf("host=%s service_name=%s deploy_version=%s package_url=%s package_sha256=%s prev_version=%s",
		a.config.Host,
		a.config.Service,
		a.config.Version,
		a.config.PackageURL,
		a.config.SHA256,
		a.config.PrevVersion,
	)

	cmd := exec.CommandContext(ctx, "ansible-playbook",
		a.playbookPath,
		"-e", extraVars,
		"-v",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		a.status.State = model.NodeStatusFailed
		a.status.LastError = fmt.Sprintf("ansible-playbook execution failed: %v", err)
		a.status.UpdatedAt = time.Now()
		return fmt.Errorf("failed to execute ansible-playbook: %w", err)
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
	a.status.UpdatedAt = time.Now()

	extraVars := fmt.Sprintf("host=%s service_name=%s deploy_version=%s package_url=%s package_sha256=%s prev_version=%s rollback=true",
		a.config.Host,
		a.config.Service,
		a.config.Version,
		a.config.PackageURL,
		a.config.SHA256,
		a.config.PrevVersion,
	)

	cmd := exec.CommandContext(ctx, "ansible-playbook",
		a.playbookPath,
		"-e", extraVars,
		"-v",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		a.status.State = model.NodeStatusFailed
		a.status.LastError = fmt.Sprintf("rollback failed: %v", err)
		a.status.UpdatedAt = time.Now()
		return fmt.Errorf("failed to execute rollback: %w", err)
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

func (a *AnsibleExecutor) getPlaybookDir() string {
	return filepath.Dir(a.playbookPath)
}
