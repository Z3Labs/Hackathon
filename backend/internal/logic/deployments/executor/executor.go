package executor

import (
	"context"
	"fmt"

	"github.com/Z3Labs/Hackathon/backend/internal/model"
)

type Executor interface {
	Deploy(ctx context.Context) error
	Rollback(ctx context.Context) error
	GetStatus(ctx context.Context) (*model.NodeDeployStatusRecord, error)
}

type ExecutorConfig struct {
	Platform    string
	Host        string
	Service     string
	Version     string
	PrevVersion string
	PackageURL  string
	SHA256      string
	Namespace   string
	Deployment  string
	ImageURL    string
}

type ExecutorFactory struct {
}

func NewExecutorFactory() *ExecutorFactory {
	return &ExecutorFactory{}
}

func (f *ExecutorFactory) CreateExecutor(ctx context.Context, config ExecutorConfig) (Executor, error) {
	switch config.Platform {
	case string(model.PlatformPhysical):
		return NewAnsibleExecutor(config), nil
	case string(model.PlatformK8s):
		return NewK8sExecutor(config), nil
	default:
		return nil, fmt.Errorf("unsupported platform: %s", config.Platform)
	}
}
