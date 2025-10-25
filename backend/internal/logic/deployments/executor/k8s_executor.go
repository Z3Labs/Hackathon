package executor

import (
	"context"
	"fmt"
	"time"
)

type K8sExecutor struct {
	config ExecutorConfig
}

func NewK8sExecutor(config ExecutorConfig) *K8sExecutor {
	return &K8sExecutor{
		config: config,
	}
}

func (k *K8sExecutor) Deploy(ctx context.Context) error {
	if err := k.updateDeployment(ctx); err != nil {
		return err
	}

	if err := k.waitForReady(ctx); err != nil {
		return err
	}

	return nil
}

func (k *K8sExecutor) Rollback(ctx context.Context) error {
	if k.config.PrevVersion == "" {
		return fmt.Errorf("no previous version to rollback to")
	}

	prevImageURL := k.buildImageURL(k.config.PrevVersion)
	if err := k.setDeploymentImage(ctx, prevImageURL); err != nil {
		return fmt.Errorf("failed to rollback deployment: %w", err)
	}

	if err := k.waitForReady(ctx); err != nil {
		return err
	}

	return nil
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
