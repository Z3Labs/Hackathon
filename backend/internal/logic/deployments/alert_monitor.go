package deployments

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Z3Labs/Hackathon/backend/internal/clients/prom"
	"github.com/Z3Labs/Hackathon/backend/internal/model"
	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeploymentAlert struct {
	DeploymentID  string
	AppName       string
	AlertRule     model.PrometheusAlert
	StartTime     time.Time
	LastCheckTime time.Time
	FiringStart   *time.Time
	IsFiring      bool
}

type AlertMonitor struct {
	svcCtx       *svc.ServiceContext
	promClient   prom.VMClient
	mu           sync.RWMutex
	activeAlerts map[string][]*DeploymentAlert
}

func NewAlertMonitor(svcCtx *svc.ServiceContext, promClient prom.VMClient) *AlertMonitor {
	return &AlertMonitor{
		svcCtx:       svcCtx,
		promClient:   promClient,
		activeAlerts: make(map[string][]*DeploymentAlert),
	}
}

func (am *AlertMonitor) StartMonitoring(ctx context.Context, deployment *model.Deployment, app *model.Application) error {
	if app.RollbackPolicy == nil || !app.RollbackPolicy.Enabled {
		return nil
	}

	am.mu.Lock()
	defer am.mu.Unlock()

	alerts := make([]*DeploymentAlert, 0, len(app.RollbackPolicy.AlertRules))
	for _, rule := range app.RollbackPolicy.AlertRules {
		alerts = append(alerts, &DeploymentAlert{
			DeploymentID:  deployment.Id,
			AppName:       deployment.AppName,
			AlertRule:     rule,
			StartTime:     time.Now(),
			LastCheckTime: time.Now(),
			IsFiring:      false,
		})
	}

	am.activeAlerts[deployment.Id] = alerts
	logx.Infof("Started monitoring deployment %s with %d alert rules", deployment.Id, len(alerts))
	return nil
}

func (am *AlertMonitor) StopMonitoring(deploymentID string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	delete(am.activeAlerts, deploymentID)
	logx.Infof("Stopped monitoring deployment %s", deploymentID)
}

func (am *AlertMonitor) CheckAlerts(ctx context.Context) error {
	am.mu.RLock()
	alertsToCheck := make(map[string][]*DeploymentAlert)
	for k, v := range am.activeAlerts {
		alertsToCheck[k] = v
	}
	am.mu.RUnlock()

	now := time.Now()

	for deploymentID, alerts := range alertsToCheck {
		deployment, err := am.svcCtx.DeploymentModel.FindById(ctx, deploymentID)
		if err != nil {
			logx.Errorf("Failed to find deployment %s: %v", deploymentID, err)
			continue
		}

		if am.shouldStopMonitoring(deployment, now) {
			am.StopMonitoring(deploymentID)
			continue
		}

		for _, alert := range alerts {
			if err := am.checkSingleAlert(ctx, deployment, alert, now); err != nil {
				logx.Errorf("Failed to check alert %s for deployment %s: %v", 
					alert.AlertRule.Name, deploymentID, err)
			}
		}
	}

	return nil
}

func (am *AlertMonitor) shouldStopMonitoring(deployment *model.Deployment, now time.Time) bool {
	if deployment.Status != model.DeploymentStatusDeploying {
		endTime := time.Unix(deployment.UpdatedTime, 0)
		if now.Sub(endTime) > 30*time.Minute {
			return true
		}
	}
	return false
}

func (am *AlertMonitor) checkSingleAlert(ctx context.Context, deployment *model.Deployment, 
	alert *DeploymentAlert, now time.Time) error {
	
	alert.LastCheckTime = now

	queryExpr := alert.AlertRule.AlertExpr
	results, err := am.promClient.QueryInstant(queryExpr)
	if err != nil {
		return fmt.Errorf("failed to query prometheus: %w", err)
	}

	isFiring := am.isAlertFiring(results, alert.AlertRule)

	if isFiring {
		if !alert.IsFiring {
			alert.IsFiring = true
			firingStart := now
			alert.FiringStart = &firingStart
			logx.Infof("Alert %s started firing for deployment %s", 
				alert.AlertRule.Name, deployment.Id)
		} else {
			duration, err := time.ParseDuration(alert.AlertRule.Duration)
			if err != nil {
				return fmt.Errorf("invalid duration %s: %w", alert.AlertRule.Duration, err)
			}

			if now.Sub(*alert.FiringStart) >= duration {
				if err := am.triggerAlert(ctx, deployment, alert, results); err != nil {
					return fmt.Errorf("failed to trigger alert: %w", err)
				}
			}
		}
	} else {
		if alert.IsFiring {
			logx.Infof("Alert %s stopped firing for deployment %s", 
				alert.AlertRule.Name, deployment.Id)
			alert.IsFiring = false
			alert.FiringStart = nil
		}
	}

	return nil
}

func (am *AlertMonitor) isAlertFiring(results []prom.InstantQueryResult, rule model.PrometheusAlert) bool {
	if len(results) == 0 {
		return false
	}

	for _, result := range results {
		if am.matchesLabels(result.Metric, rule.Labels) {
			if result.Value.Value > 0 {
				return true
			}
		}
	}

	return false
}

func (am *AlertMonitor) matchesLabels(metric map[string]string, ruleLabels map[string]string) bool {
	if len(ruleLabels) == 0 {
		return true
	}

	for key, expectedValue := range ruleLabels {
		if actualValue, exists := metric[key]; !exists || actualValue != expectedValue {
			return false
		}
	}

	return true
}

func (am *AlertMonitor) triggerAlert(ctx context.Context, deployment *model.Deployment, 
	alert *DeploymentAlert, results []prom.InstantQueryResult) error {
	
	logx.Infof("Triggering alert %s for deployment %s after duration threshold", 
		alert.AlertRule.Name, deployment.Id)

	now := time.Now()
	desc := fmt.Sprintf("Alert %s has been firing for %s", 
		alert.AlertRule.Name, alert.AlertRule.Duration)
	
	if len(alert.AlertRule.Annotations) > 0 {
		if d, ok := alert.AlertRule.Annotations["description"]; ok {
			desc = d
		}
	}

	alertReq := &types.PostAlertCallbackReq{
		Key:       fmt.Sprintf("%s-%s-%d", deployment.Id, alert.AlertRule.Name, now.Unix()),
		Status:    "firing",
		Desc:      desc,
		StartsAt:  alert.FiringStart.Format(time.RFC3339),
		ReceiveAt: now.Format(time.RFC3339),
		Severity:  alert.AlertRule.Severity,
		Alertname: alert.AlertRule.Name,
		Labels: map[string]string{
			"deployment_id": deployment.Id,
			"app_name":      deployment.AppName,
		},
		Annotations: alert.AlertRule.Annotations,
	}

	if len(results) > 0 {
		alertReq.Values = results[0].Value.Value
	}

	app, err := am.svcCtx.ApplicationModel.FindById(ctx, deployment.AppName)
	if err != nil {
		return fmt.Errorf("failed to find application: %w", err)
	}

	if app.RollbackPolicy != nil && app.RollbackPolicy.AutoRollback {
		logx.Infof("Auto rollback enabled for deployment %s, initiating rollback", deployment.Id)
	}

	return nil
}

func (am *AlertMonitor) formatMetrics(results []prom.InstantQueryResult) string {
	if len(results) == 0 {
		return ""
	}

	output := ""
	for i, result := range results {
		if i > 0 {
			output += ", "
		}
		output += fmt.Sprintf("{%v}=%f", result.Metric, result.Value.Value)
	}
	return output
}

func (am *AlertMonitor) GetActiveAlertsCount() int {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return len(am.activeAlerts)
}
