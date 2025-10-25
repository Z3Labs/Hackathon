package executor

import (
	"context"
	"fmt"
	"sync"

	"github.com/Z3Labs/Hackathon/backend/internal/model"
)

type MockExecutor struct {
	config         ExecutorConfig
	deployCalled   bool
	rollbackCalled bool
	deployError    error
	rollbackError  error
	mu             sync.Mutex
}

func NewMockExecutor(config ExecutorConfig) *MockExecutor {
	return &MockExecutor{
		config: config,
	}
}

func (m *MockExecutor) SetDeployError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deployError = err
}

func (m *MockExecutor) SetRollbackError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rollbackError = err
}

func (m *MockExecutor) Deploy(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deployCalled = true
	return m.deployError
}

func (m *MockExecutor) Rollback(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rollbackCalled = true
	return m.rollbackError
}

func (m *MockExecutor) GetStatus(ctx context.Context) (*model.NodeDeployStatusRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return &model.NodeDeployStatusRecord{
		Host:             m.config.Host,
		Service:          m.config.Service,
		CurrentVersion:   m.config.Version,
		DeployingVersion: "",
		PrevVersion:      m.config.PrevVersion,
		Platform:         model.PlatformMock,
		State:            model.NodeStatusSuccess,
	}, nil
}

func (m *MockExecutor) DeployCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.deployCalled
}

func (m *MockExecutor) RollbackCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.rollbackCalled
}

type MockExecutorFactory struct {
	executors map[string]*MockExecutor
	mu        sync.Mutex
}

func NewMockExecutorFactory() ExecutorFactoryInterface {
	return &MockExecutorFactory{
		executors: make(map[string]*MockExecutor),
	}
}

func (f *MockExecutorFactory) CreateExecutor(ctx context.Context, config ExecutorConfig) (Executor, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	key := fmt.Sprintf("%s-%s", config.Host, config.Service)
	executor := NewMockExecutor(config)
	f.executors[key] = executor
	return executor, nil
}

func (f *MockExecutorFactory) GetExecutor(host, service string) *MockExecutor {
	f.mu.Lock()
	defer f.mu.Unlock()
	key := fmt.Sprintf("%s-%s", host, service)
	return f.executors[key]
}
