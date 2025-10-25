package executor

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/Z3Labs/Hackathon/backend/internal/model"
)

func TestNewAnsibleExecutor(t *testing.T) {
	tests := []struct {
		name         string
		config       ExecutorConfig
		envPlaybook  string
		wantPlaybook string
	}{
		{
			name: "默认 playbook 路径",
			config: ExecutorConfig{
				Host:        "192.168.1.100",
				Service:     "test-service",
				Version:     "v1.0.0",
				PrevVersion: "v0.9.0",
				PackageURL:  "http://example.com/package.tar.gz",
				MD5:         "abc123",
			},
			envPlaybook:  "",
			wantPlaybook: "/workspace/backend/playbooks/deploy.yml",
		},
		{
			name: "环境变量指定 playbook 路径",
			config: ExecutorConfig{
				Host:        "192.168.1.100",
				Service:     "test-service",
				Version:     "v1.0.0",
				PrevVersion: "v0.9.0",
			},
			envPlaybook:  "/custom/path/playbook.yml",
			wantPlaybook: "/custom/path/playbook.yml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envPlaybook != "" {
				os.Setenv("ANSIBLE_PLAYBOOK_PATH", tt.envPlaybook)
				defer os.Unsetenv("ANSIBLE_PLAYBOOK_PATH")
			}

			executor := NewAnsibleExecutor(tt.config)

			if executor == nil {
				t.Fatal("NewAnsibleExecutor() returned nil")
			}

			if executor.playbookPath != tt.wantPlaybook {
				t.Errorf("playbookPath = %v, want %v", executor.playbookPath, tt.wantPlaybook)
			}

			if executor.status == nil {
				t.Fatal("status is nil")
			}

			if executor.status.Host != tt.config.Host {
				t.Errorf("status.Host = %v, want %v", executor.status.Host, tt.config.Host)
			}

			if executor.status.Service != tt.config.Service {
				t.Errorf("status.Service = %v, want %v", executor.status.Service, tt.config.Service)
			}

			if executor.status.CurrentVersion != tt.config.Version {
				t.Errorf("status.CurrentVersion = %v, want %v", executor.status.CurrentVersion, tt.config.Version)
			}

			if executor.status.Platform != model.PlatformPhysical {
				t.Errorf("status.Platform = %v, want %v", executor.status.Platform, model.PlatformPhysical)
			}

			if executor.status.State != model.NodeStatusPending {
				t.Errorf("status.State = %v, want %v", executor.status.State, model.NodeStatusPending)
			}
		})
	}
}

func TestAnsibleExecutor_GetStatus(t *testing.T) {
	config := ExecutorConfig{
		Host:        "192.168.1.100",
		Service:     "test-service",
		Version:     "v1.0.0",
		PrevVersion: "v0.9.0",
	}

	executor := NewAnsibleExecutor(config)
	ctx := context.Background()

	status, err := executor.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	if status == nil {
		t.Fatal("GetStatus() returned nil status")
	}

	if status.Host != config.Host {
		t.Errorf("status.Host = %v, want %v", status.Host, config.Host)
	}

	if status.Service != config.Service {
		t.Errorf("status.Service = %v, want %v", status.Service, config.Service)
	}
}

func TestAnsibleExecutor_Deploy_StatusUpdates(t *testing.T) {
	if os.Getenv("GO_TEST_PROCESS") == "1" {
		os.Exit(0)
		return
	}

	config := ExecutorConfig{
		Host:        "192.168.1.100",
		Service:     "test-service",
		Version:     "v1.0.0",
		PrevVersion: "v0.9.0",
		PackageURL:  "http://example.com/package.tar.gz",
		MD5:         "abc123",
	}

	executor := NewAnsibleExecutor(config)

	oldExecCommand := execCommand
	execCommand = mockExecCommand
	defer func() { execCommand = oldExecCommand }()

	ctx := context.Background()
	err := executor.Deploy(ctx)

	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}

	if executor.status.State != model.NodeStatusSuccess {
		t.Errorf("status.State = %v, want %v", executor.status.State, model.NodeStatusSuccess)
	}

	if executor.status.CurrentVersion != config.Version {
		t.Errorf("status.CurrentVersion = %v, want %v", executor.status.CurrentVersion, config.Version)
	}

	if executor.status.DeployingVersion != "" {
		t.Errorf("status.DeployingVersion = %v, want empty", executor.status.DeployingVersion)
	}
}

func TestAnsibleExecutor_Deploy_Failure(t *testing.T) {
	if os.Getenv("GO_TEST_PROCESS") == "1" {
		os.Exit(1)
		return
	}

	config := ExecutorConfig{
		Host:        "192.168.1.100",
		Service:     "test-service",
		Version:     "v1.0.0",
		PrevVersion: "v0.9.0",
		PackageURL:  "http://example.com/package.tar.gz",
		MD5:         "abc123",
	}

	executor := NewAnsibleExecutor(config)

	oldExecCommand := execCommand
	execCommand = mockExecCommandError
	defer func() { execCommand = oldExecCommand }()

	ctx := context.Background()
	err := executor.Deploy(ctx)

	if err == nil {
		t.Fatal("Deploy() expected error, got nil")
	}

	if executor.status.State != model.NodeStatusFailed {
		t.Errorf("status.State = %v, want %v", executor.status.State, model.NodeStatusFailed)
	}

	if executor.status.LastError == "" {
		t.Error("status.LastError is empty, expected error message")
	}
}

func TestAnsibleExecutor_Rollback_Success(t *testing.T) {
	if os.Getenv("GO_TEST_PROCESS") == "1" {
		os.Exit(0)
		return
	}

	config := ExecutorConfig{
		Host:        "192.168.1.100",
		Service:     "test-service",
		Version:     "v1.0.0",
		PrevVersion: "v0.9.0",
		PackageURL:  "http://example.com/package.tar.gz",
		MD5:         "abc123",
	}

	executor := NewAnsibleExecutor(config)

	oldExecCommand := execCommand
	execCommand = mockExecCommand
	defer func() { execCommand = oldExecCommand }()

	ctx := context.Background()
	err := executor.Rollback(ctx)

	if err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}

	if executor.status.State != model.NodeStatusRolledBack {
		t.Errorf("status.State = %v, want %v", executor.status.State, model.NodeStatusRolledBack)
	}

	if executor.status.CurrentVersion != config.PrevVersion {
		t.Errorf("status.CurrentVersion = %v, want %v", executor.status.CurrentVersion, config.PrevVersion)
	}

	if executor.status.DeployingVersion != "" {
		t.Errorf("status.DeployingVersion = %v, want empty", executor.status.DeployingVersion)
	}
}

func TestAnsibleExecutor_Rollback_NoPreviousVersion(t *testing.T) {
	config := ExecutorConfig{
		Host:        "192.168.1.100",
		Service:     "test-service",
		Version:     "v1.0.0",
		PrevVersion: "",
	}

	executor := NewAnsibleExecutor(config)
	ctx := context.Background()

	err := executor.Rollback(ctx)

	if err == nil {
		t.Fatal("Rollback() expected error for empty PrevVersion, got nil")
	}

	expectedErr := "no previous version to rollback to"
	if err.Error() != expectedErr {
		t.Errorf("Rollback() error = %v, want %v", err.Error(), expectedErr)
	}
}

func TestAnsibleExecutor_Rollback_Failure(t *testing.T) {
	if os.Getenv("GO_TEST_PROCESS") == "1" {
		os.Exit(1)
		return
	}

	config := ExecutorConfig{
		Host:        "192.168.1.100",
		Service:     "test-service",
		Version:     "v1.0.0",
		PrevVersion: "v0.9.0",
		PackageURL:  "http://example.com/package.tar.gz",
		MD5:         "abc123",
	}

	executor := NewAnsibleExecutor(config)

	oldExecCommand := execCommand
	execCommand = mockExecCommandError
	defer func() { execCommand = oldExecCommand }()

	ctx := context.Background()
	err := executor.Rollback(ctx)

	if err == nil {
		t.Fatal("Rollback() expected error, got nil")
	}

	if executor.status.State != model.NodeStatusFailed {
		t.Errorf("status.State = %v, want %v", executor.status.State, model.NodeStatusFailed)
	}

	if executor.status.LastError == "" {
		t.Error("status.LastError is empty, expected error message")
	}
}

func TestAnsibleExecutor_getPlaybookDir(t *testing.T) {
	tests := []struct {
		name         string
		playbookPath string
		wantDir      string
	}{
		{
			name:         "标准路径",
			playbookPath: "/workspace/backend/playbooks/deploy.yml",
			wantDir:      "/workspace/backend/playbooks",
		},
		{
			name:         "自定义路径",
			playbookPath: "/custom/path/playbook.yml",
			wantDir:      "/custom/path",
		},
		{
			name:         "根目录",
			playbookPath: "/playbook.yml",
			wantDir:      "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &AnsibleExecutor{
				playbookPath: tt.playbookPath,
			}

			gotDir := executor.getPlaybookDir()
			if gotDir != tt.wantDir {
				t.Errorf("getPlaybookDir() = %v, want %v", gotDir, tt.wantDir)
			}
		})
	}
}

func mockExecCommand(ctx context.Context, command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestAnsibleExecutor_Deploy_StatusUpdates", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_TEST_PROCESS=1"}
	return cmd
}

func mockExecCommandError(ctx context.Context, command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestAnsibleExecutor_Deploy_Failure", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_TEST_PROCESS=1"}
	return cmd
}
