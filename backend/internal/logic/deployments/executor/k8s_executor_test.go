package executor

import (
	"context"
	"errors"
	"testing"

	"github.com/Z3Labs/Hackathon/backend/internal/model"
)

func TestNewK8sExecutor(t *testing.T) {
	config := ExecutorConfig{
		Host:        "k8s-cluster",
		Service:     "test-service",
		Version:     "v1.0.0",
		PrevVersion: "v0.9.0",
		Namespace:   "default",
		Deployment:  "test-deployment",
		ImageURL:    "registry.example.com/test-service",
	}

	executor := NewK8sExecutor(config)

	if executor == nil {
		t.Fatal("NewK8sExecutor() returned nil")
	}

	if executor.status == nil {
		t.Fatal("status is nil")
	}

	if executor.status.Host != config.Host {
		t.Errorf("status.Host = %v, want %v", executor.status.Host, config.Host)
	}

	if executor.status.Service != config.Service {
		t.Errorf("status.Service = %v, want %v", executor.status.Service, config.Service)
	}

	if executor.status.CurrentVersion != config.Version {
		t.Errorf("status.CurrentVersion = %v, want %v", executor.status.CurrentVersion, config.Version)
	}

	if executor.status.Platform != model.PlatformK8s {
		t.Errorf("status.Platform = %v, want %v", executor.status.Platform, model.PlatformK8s)
	}

	if executor.status.State != model.NodeStatusPending {
		t.Errorf("status.State = %v, want %v", executor.status.State, model.NodeStatusPending)
	}
}

func TestK8sExecutor_GetStatus(t *testing.T) {
	config := ExecutorConfig{
		Host:    "k8s-cluster",
		Service: "test-service",
		Version: "v1.0.0",
	}

	executor := NewK8sExecutor(config)
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

func TestK8sExecutor_buildImageURL(t *testing.T) {
	tests := []struct {
		name     string
		config   ExecutorConfig
		version  string
		wantURL  string
	}{
		{
			name: "使用配置的 ImageURL",
			config: ExecutorConfig{
				Service:  "test-service",
				ImageURL: "registry.example.com/test-service",
			},
			version: "v1.0.0",
			wantURL: "registry.example.com/test-service:v1.0.0",
		},
		{
			name: "使用 Service 名称作为镜像名",
			config: ExecutorConfig{
				Service:  "my-service",
				ImageURL: "",
			},
			version: "v2.0.0",
			wantURL: "my-service:v2.0.0",
		},
		{
			name: "带有 latest 标签",
			config: ExecutorConfig{
				Service:  "test-service",
				ImageURL: "docker.io/myorg/test-service",
			},
			version: "latest",
			wantURL: "docker.io/myorg/test-service:latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewK8sExecutor(tt.config)
			gotURL := executor.buildImageURL(tt.version)

			if gotURL != tt.wantURL {
				t.Errorf("buildImageURL() = %v, want %v", gotURL, tt.wantURL)
			}
		})
	}
}

func TestK8sExecutor_Deploy_NotImplemented(t *testing.T) {
	config := ExecutorConfig{
		Host:       "k8s-cluster",
		Service:    "test-service",
		Version:    "v1.0.0",
		Namespace:  "default",
		Deployment: "test-deployment",
		ImageURL:   "registry.example.com/test-service",
	}

	executor := NewK8sExecutor(config)
	ctx := context.Background()

	err := executor.Deploy(ctx)

	if err == nil {
		t.Fatal("Deploy() expected error for unimplemented K8s client, got nil")
	}

	if executor.status.State != model.NodeStatusFailed {
		t.Errorf("status.State = %v, want %v", executor.status.State, model.NodeStatusFailed)
	}

	if executor.status.LastError == "" {
		t.Error("status.LastError is empty, expected error message")
	}
}

func TestK8sExecutor_Rollback_NoPreviousVersion(t *testing.T) {
	config := ExecutorConfig{
		Host:        "k8s-cluster",
		Service:     "test-service",
		Version:     "v1.0.0",
		PrevVersion: "",
	}

	executor := NewK8sExecutor(config)
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

func TestK8sExecutor_Rollback_NotImplemented(t *testing.T) {
	config := ExecutorConfig{
		Host:        "k8s-cluster",
		Service:     "test-service",
		Version:     "v1.0.0",
		PrevVersion: "v0.9.0",
		Namespace:   "default",
		Deployment:  "test-deployment",
		ImageURL:    "registry.example.com/test-service",
	}

	executor := NewK8sExecutor(config)
	ctx := context.Background()

	err := executor.Rollback(ctx)

	if err == nil {
		t.Fatal("Rollback() expected error for unimplemented K8s client, got nil")
	}

	if executor.status.State != model.NodeStatusFailed {
		t.Errorf("status.State = %v, want %v", executor.status.State, model.NodeStatusFailed)
	}
}

func TestK8sExecutor_updateDeployment(t *testing.T) {
	config := ExecutorConfig{
		Host:     "k8s-cluster",
		Service:  "test-service",
		Version:  "v1.0.0",
		ImageURL: "registry.example.com/test-service",
	}

	executor := NewK8sExecutor(config)
	ctx := context.Background()

	err := executor.updateDeployment(ctx)

	if err == nil {
		t.Fatal("updateDeployment() expected error for unimplemented K8s client, got nil")
	}

	expectedErr := "K8s client not implemented yet"
	if err.Error() != expectedErr {
		t.Errorf("updateDeployment() error = %v, want %v", err.Error(), expectedErr)
	}
}

func TestK8sExecutor_setDeploymentImage(t *testing.T) {
	config := ExecutorConfig{
		Host:    "k8s-cluster",
		Service: "test-service",
		Version: "v1.0.0",
	}

	executor := NewK8sExecutor(config)
	ctx := context.Background()

	err := executor.setDeploymentImage(ctx, "test-image:v1.0.0")

	if err == nil {
		t.Fatal("setDeploymentImage() expected error for unimplemented K8s client, got nil")
	}
}

func TestK8sExecutor_checkDeploymentReady(t *testing.T) {
	config := ExecutorConfig{
		Host:    "k8s-cluster",
		Service: "test-service",
		Version: "v1.0.0",
	}

	executor := NewK8sExecutor(config)
	ctx := context.Background()

	ready, err := executor.checkDeploymentReady(ctx)

	if err == nil {
		t.Fatal("checkDeploymentReady() expected error for unimplemented K8s client, got nil")
	}

	if ready {
		t.Error("checkDeploymentReady() ready = true, want false")
	}
}

func TestK8sExecutor_waitForReady_ImmediateError(t *testing.T) {
	config := ExecutorConfig{
		Host:    "k8s-cluster",
		Service: "test-service",
		Version: "v1.0.0",
	}

	executor := NewK8sExecutor(config)
	ctx := context.Background()

	err := executor.waitForReady(ctx)

	if err == nil {
		t.Fatal("waitForReady() expected error for unimplemented K8s client, got nil")
	}
}

func TestK8sExecutor_ContextCancellation(t *testing.T) {
	config := ExecutorConfig{
		Host:    "k8s-cluster",
		Service: "test-service",
		Version: "v1.0.0",
	}

	executor := NewK8sExecutor(config)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := executor.waitForReady(ctx)

	if err == nil {
		t.Fatal("waitForReady() expected error, got nil")
	}

	if !errors.Is(err, context.Canceled) && err.Error() != "K8s client not implemented yet" {
		t.Logf("waitForReady() with canceled context error = %v", err)
	}
}
