package executor

import (
	"context"
	"testing"

	"github.com/Z3Labs/Hackathon/backend/internal/model"
)

func TestExecutorFactory_CreateExecutor(t *testing.T) {
	factory := NewExecutorFactory()
	ctx := context.Background()

	tests := []struct {
		name        string
		config      ExecutorConfig
		wantType    string
		wantErr     bool
		errContains string
	}{
		{
			name: "创建 Ansible Executor - Physical 平台",
			config: ExecutorConfig{
				Platform:   string(model.PlatformPhysical),
				Host:       "192.168.1.100",
				Service:    "test-service",
				Version:    "v1.0.0",
				PrevVersion: "v0.9.0",
				PackageURL: "http://example.com/package.tar.gz",
				SHA256:     "abc123",
			},
			wantType: "*executor.AnsibleExecutor",
			wantErr:  false,
		},
		{
			name: "创建 K8s Executor - K8s 平台",
			config: ExecutorConfig{
				Platform:   string(model.PlatformK8s),
				Host:       "k8s-cluster",
				Service:    "test-service",
				Version:    "v1.0.0",
				PrevVersion: "v0.9.0",
				Namespace:  "default",
				Deployment: "test-deployment",
				ImageURL:   "registry.example.com/test-service",
			},
			wantType: "*executor.K8sExecutor",
			wantErr:  false,
		},
		{
			name: "不支持的平台类型",
			config: ExecutorConfig{
				Platform: "unsupported",
				Host:     "test-host",
				Service:  "test-service",
				Version:  "v1.0.0",
			},
			wantErr:     true,
			errContains: "unsupported platform",
		},
		{
			name: "空平台类型",
			config: ExecutorConfig{
				Platform: "",
				Host:     "test-host",
				Service:  "test-service",
				Version:  "v1.0.0",
			},
			wantErr:     true,
			errContains: "unsupported platform",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := factory.CreateExecutor(ctx, tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateExecutor() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("CreateExecutor() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("CreateExecutor() unexpected error = %v", err)
				return
			}

			if executor == nil {
				t.Error("CreateExecutor() returned nil executor")
				return
			}

			gotType := getTypeName(executor)
			if gotType != tt.wantType {
				t.Errorf("CreateExecutor() type = %v, want %v", gotType, tt.wantType)
			}
		})
	}
}

func TestExecutorFactory_CreateExecutor_InterfaceCompliance(t *testing.T) {
	factory := NewExecutorFactory()
	ctx := context.Background()

	config := ExecutorConfig{
		Platform: string(model.PlatformPhysical),
		Host:     "test-host",
		Service:  "test-service",
		Version:  "v1.0.0",
	}

	executor, err := factory.CreateExecutor(ctx, config)
	if err != nil {
		t.Fatalf("CreateExecutor() error = %v", err)
	}

	var _ Executor = executor
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func getTypeName(i interface{}) string {
	if i == nil {
		return "<nil>"
	}
	switch i.(type) {
	case *AnsibleExecutor:
		return "*executor.AnsibleExecutor"
	case *K8sExecutor:
		return "*executor.K8sExecutor"
	default:
		return "unknown"
	}
}
