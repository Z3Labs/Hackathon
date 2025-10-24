package deploy

import (
	"context"
	"fmt"
	"time"

	"github.com/Z3Labs/Hackathon/backend/internal/model"
)

type K8sExecutor struct {
	config ExecutorConfig
	status *model.NodeStatusRecord
}

func NewK8sExecutor(config ExecutorConfig) *K8sExecutor {
	return &K8sExecutor{
		config: config,
		status: &model.NodeStatusRecord{
			Host:             config.Host,
			Service:          config.Service,
			CurrentVersion:   config.Version,
			DeployingVersion: config.Version,
			PrevVersion:      config.PrevVersion,
			Platform:         model.PlatformK8s,
			State:            model.NodeStatusPending,
		},
	}
}

func (k *K8sExecutor) Deploy(ctx context.Context) error {
	k.status.State = model.NodeStatusDeploying
	k.status.UpdatedAt = time.Now()

	if err := k.updateDeployment(ctx); err != nil {
		k.status.State = model.NodeStatusFailed
		k.status.LastError = fmt.Sprintf("deploy failed: %v", err)
		return err
	}

	if err := k.waitForReady(ctx); err != nil {
		k.status.State = model.NodeStatusFailed
		k.status.LastError = fmt.Sprintf("wait for ready failed: %v", err)
		return err
	}

	k.status.State = model.NodeStatusSuccess
	k.status.CurrentVersion = k.config.Version
	k.status.PrevVersion = k.config.Version
	k.status.DeployingVersion = ""
	k.status.UpdatedAt = time.Now()

	return nil
}

func (k *K8sExecutor) Rollback(ctx context.Context) error {
	if k.config.PrevVersion == "" {
		return fmt.Errorf("no previous version to rollback to")
	}

	k.status.State = model.NodeStatusDeploying

	prevImageURL := k.buildImageURL(k.config.PrevVersion)
	if err := k.setDeploymentImage(ctx, prevImageURL); err != nil {
		k.status.State = model.NodeStatusFailed
		k.status.LastError = fmt.Sprintf("rollback failed: %v", err)
		return fmt.Errorf("failed to rollback deployment: %w", err)
	}

	if err := k.waitForReady(ctx); err != nil {
		k.status.State = model.NodeStatusFailed
		k.status.LastError = fmt.Sprintf("wait for ready after rollback failed: %v", err)
		return err
	}

	k.status.State = model.NodeStatusRolledBack
	k.status.CurrentVersion = k.config.PrevVersion
	k.status.DeployingVersion = ""
	k.status.UpdatedAt = time.Now()

	return nil
}

func (k *K8sExecutor) GetStatus(ctx context.Context) (*model.NodeStatusRecord, error) {
	return k.status, nil
}

func (k *K8sExecutor) updateDeployment(ctx context.Context) error {
	imageURL := k.buildImageURL(k.config.Version)
	return k.setDeploymentImage(ctx, imageURL)
}

func (k *K8sExecutor) setDeploymentImage(ctx context.Context, imageURL string) error {
	return fmt.Errorf("K8s client not implemented yet")
}

func (k *K8sExecutor) waitForReady(ctx context.Context) error {
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timeout waiting for deployment to be ready")
		case <-ticker.C:
			ready, err := k.checkDeploymentReady(ctx)
			if err != nil {
				return err
			}
			if ready {
				return nil
			}
		}
	}
}

func (k *K8sExecutor) checkDeploymentReady(ctx context.Context) (bool, error) {
	return false, fmt.Errorf("K8s client not implemented yet")
}

func (k *K8sExecutor) buildImageURL(version string) string {
	if k.config.ImageURL != "" {
		return fmt.Sprintf("%s:%s", k.config.ImageURL, version)
	}
	return fmt.Sprintf("%s:%s", k.config.Service, version)
}
