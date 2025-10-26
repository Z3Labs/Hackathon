package deployments

import (
	"context"
	"fmt"

	"github.com/robfig/cron/v3"
)

type DeploymentCron struct {
	cron              *cron.Cron
	deploymentManager *DeploymentManager
	rollbackManager   *RollbackManager
	alertMonitor      *AlertMonitor
}

func NewDeploymentCron(deploymentManager *DeploymentManager, rollbackManager *RollbackManager, alertMonitor *AlertMonitor) *DeploymentCron {
	return &DeploymentCron{
		cron:              cron.New(),
		deploymentManager: deploymentManager,
		rollbackManager:   rollbackManager,
		alertMonitor:      alertMonitor,
	}
}

func (dc *DeploymentCron) Start() error {
	_, err := dc.cron.AddFunc("@every 1m", func() {
		ctx := context.Background()
		if err := dc.deploymentManager.ContinueDeployingDeployments(ctx); err != nil {
			fmt.Printf("continue deploying deployments error: %v\n", err)
		}

		if err := dc.rollbackManager.ContinueRollingBackDeployments(ctx); err != nil {
			fmt.Printf("continue rolling back deployments error: %v\n", err)
		}
		if dc.alertMonitor != nil {
			if err := dc.alertMonitor.CheckAlerts(ctx); err != nil {
				fmt.Printf("check deployment alerts error: %v\n", err)
			}
		}
	})
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	dc.cron.Start()
	fmt.Println("Deployment cron job started, will process deployments every minute")
	return nil
}

func (dc *DeploymentCron) Stop() {
	dc.cron.Stop()
	fmt.Println("Deployment cron job stopped")
}
