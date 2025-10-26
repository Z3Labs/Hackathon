package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var execCommand = exec.CommandContext

type AnsibleExecutor struct {
	config       ExecutorConfig
	playbookPath string
}

func NewAnsibleExecutor(config ExecutorConfig) *AnsibleExecutor {
	playbookPath := "/etc/playbook/deploy.yml"
	if path := os.Getenv("ANSIBLE_PLAYBOOK_PATH"); path != "" {
		playbookPath = path
	}

	return &AnsibleExecutor{
		config:       config,
		playbookPath: playbookPath,
	}
}

func (a *AnsibleExecutor) Deploy(ctx context.Context) error {
	extraVars := fmt.Sprintf("ansible_user=root service_name=%s deploy_version=%s package_url=%s package_md5sum=%s prev_version=%s",
		a.config.Service,
		a.config.Version,
		a.config.PackageURL,
		a.config.MD5,
		a.config.PrevVersion,
	)

	args := []string{a.playbookPath}
	if a.config.IP != "" {
		args = append(args, "-i", a.config.IP+",")
	}
	args = append(args, "-e", extraVars, "-v")

	cmd := execCommand(ctx, "ansible-playbook", args...)
	fmt.Printf("Running ansible-playbook, cmd is: %v\n", cmd.String())

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute ansible-playbook: %w", err)
	}

	return nil
}

func (a *AnsibleExecutor) Rollback(ctx context.Context) error {
	if a.config.PrevVersion == "" {
		return fmt.Errorf("no previous version to rollback to")
	}

	extraVars := fmt.Sprintf("ansible_user=root service_name=%s deploy_version=%s package_url=%s package_md5sum=%s prev_version=%s rollback=true",
		a.config.Service,
		a.config.Version,
		a.config.PackageURL,
		a.config.MD5,
		a.config.PrevVersion,
	)

	args := []string{a.playbookPath}
	if a.config.IP != "" {
		args = append(args, "-i", a.config.IP+",")
	}
	args = append(args, "-e", extraVars, "-v")

	cmd := execCommand(ctx, "ansible-playbook", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute rollback: %w", err)
	}

	return nil
}

func (a *AnsibleExecutor) getPlaybookDir() string {
	return filepath.Dir(a.playbookPath)
}
