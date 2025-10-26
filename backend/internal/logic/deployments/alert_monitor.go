package deployments

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Z3Labs/Hackathon/backend/internal/clients/prom"
	"github.com/Z3Labs/Hackathon/backend/internal/logic/alert"
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
	alert        alert.AlertCallBackLogic
	activeAlerts map[string][]*DeploymentAlert
}

func NewAlertMonitor(svcCtx *svc.ServiceContext, promClient prom.VMClient) *AlertMonitor {
	return &AlertMonitor{
		svcCtx:       svcCtx,
		promClient:   promClient,
		activeAlerts: make(map[string][]*DeploymentAlert),
		alert:        alert.NewAlertCallBackLogic(context.Background(), svcCtx),
	}
}

func (am *AlertMonitor) StartMonitoring(ctx context.Context, deployment *model.Deployment, app *model.Application) error {
	// 检查是否启用了回滚策略
	if app.RollbackPolicy == nil || !app.RollbackPolicy.Enabled {
		logx.Infof("Rollback policy not enabled for app %s, skipping alert monitoring", deployment.AppName)
		return nil
	}

	// 只有发布中和回滚中的状态才需要启动告警监控
	if deployment.Status != model.DeploymentStatusDeploying && deployment.Status != model.DeploymentStatusRollingBack {
		logx.Infof("Deployment %s status is %s, no need to start alert monitoring", deployment.Id, deployment.Status)
		return nil
	}

	am.mu.Lock()
	defer am.mu.Unlock()

	// 检查是否已经在监控中
	if _, exists := am.activeAlerts[deployment.Id]; exists {
		logx.Infof("Deployment %s is already being monitored", deployment.Id)
		return nil
	}

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
	logx.Infof("Started monitoring deployment %s (status: %s) with %d alert rules", deployment.Id, deployment.Status, len(alerts))
	return nil
}

func (am *AlertMonitor) StopMonitoring(deploymentID string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	if alerts, exists := am.activeAlerts[deploymentID]; exists {
		firingCount := 0
		for _, alert := range alerts {
			if alert.IsFiring {
				firingCount++
			}
		}
		delete(am.activeAlerts, deploymentID)
		logx.Infof("Stopped monitoring deployment %s (had %d alerts, %d firing)", deploymentID, len(alerts), firingCount)
	} else {
		logx.Infof("Deployment %s was not being monitored", deploymentID)
	}
}

func (am *AlertMonitor) CheckAlerts(ctx context.Context) error {
	am.mu.RLock()
	alertsToCheck := make(map[string][]*DeploymentAlert)
	for k, v := range am.activeAlerts {
		alertsToCheck[k] = v
	}
	am.mu.RUnlock()

	now := time.Now()
	monitoredDeployments := len(alertsToCheck)

	if monitoredDeployments > 0 {
		logx.Debugf("Checking alerts for %d deployments", monitoredDeployments)
	}

	for deploymentID, alerts := range alertsToCheck {
		deployment, err := am.svcCtx.DeploymentModel.FindById(ctx, deploymentID)
		if err != nil {
			logx.Errorf("Failed to find deployment %s: %v", deploymentID, err)
			// 如果找不到发布单，停止监控
			am.StopMonitoring(deploymentID)
			continue
		}

		// 检查发布单状态是否还需要监控
		if !am.IsDeploymentInMonitoringStatus(deployment) {
			if am.shouldStopMonitoring(deployment, now) {
				am.StopMonitoring(deploymentID)
				continue
			}
		}

		// 检查单个告警规则
		for _, alert := range alerts {
			if err := am.checkSingleAlert(ctx, deployment, alert, now); err != nil {
				logx.Errorf("Failed to check alert %s for deployment %s: %v",
					alert.AlertRule.Name, deploymentID, err)
			}
		}
	}
	am.cleanAlert(ctx)
	return nil
}

func (am *AlertMonitor) cleanAlert(ctx context.Context) {
	am.mu.Lock()
	defer am.mu.Unlock()

	for deploymentID, _ := range am.activeAlerts {
		if deploy, err := am.svcCtx.DeploymentModel.FindById(ctx, deploymentID); err == nil && !am.IsDeploymentInMonitoringStatus(deploy) {
			delete(am.activeAlerts, deploymentID)
		}
	}

}
func (am *AlertMonitor) shouldStopMonitoring(deployment *model.Deployment, now time.Time) bool {
	// 对于发布中和回滚中的状态，需要持续监控，不应该停止
	switch deployment.Status {
	case model.DeploymentStatusDeploying, model.DeploymentStatusRollingBack:
		return false
	}

	// 对于其他状态（成功、失败、已回滚、已取消），在状态变更30分钟后停止监控
	endTime := time.Unix(deployment.UpdatedTime, 0)
	if now.Sub(endTime) > 30*time.Minute {
		logx.Infof("Stopping monitoring for deployment %s due to status %s being unchanged for 30 minutes",
			deployment.Id, deployment.Status)
		return true
	}

	return false
}

func (am *AlertMonitor) checkSingleAlert(ctx context.Context, deployment *model.Deployment,
	alert *DeploymentAlert, now time.Time) error {

	alert.LastCheckTime = now

	// todo 只查询发布中的机器的指标
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

	now := time.Now()
	firingDuration := now.Sub(*alert.FiringStart)
	logx.Infof("Triggering alert %s for deployment %s (status: %s) after firing for %v",
		alert.AlertRule.Name, deployment.Id, deployment.Status, firingDuration)

	// 构建告警描述，包含发布单状态信息
	desc := fmt.Sprintf("Alert %s has been firing for %s (deployment status: %s)",
		alert.AlertRule.Name, alert.AlertRule.Duration, deployment.Status)

	if len(alert.AlertRule.Annotations) > 0 {
		if d, ok := alert.AlertRule.Annotations["description"]; ok {
			desc = d
		}
	}
	// 带上非 pending 状态的节点列表
	hostNames := make([]string, 0)
	for _, node := range deployment.NodeDeployments {
		if model.NodeStatus(node.NodeDeployStatus) != model.NodeStatusPending {
			hostNames = append(hostNames, node.Name)
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
			"deploymentId": deployment.Id,
			"appName":      deployment.AppName,
			"hostname":     strings.Join(hostNames, ","),
		},
		Annotations: alert.AlertRule.Annotations,
	}
	for _, res := range results {
		for k, v := range res.Metric {
			alertReq.Labels[k] = v
		}
	}

	if len(results) > 0 {
		alertReq.Values = results[0].Value.Value
	}
	am.alert.AlertCallBack(alertReq)
	app, err := am.svcCtx.ApplicationModel.FindById(ctx, deployment.AppId)
	if err != nil {
		return fmt.Errorf("failed to find application: %w", err)
	}
	// 告警告警回调
	// 根据发布单状态决定是否触发回滚
	shouldRollback := false
	if app.RollbackPolicy != nil && app.RollbackPolicy.AutoRollback {
		// 只有在发布中状态才自动回滚，回滚中的状态不再触发回滚
		if deployment.Status == model.DeploymentStatusDeploying {
			shouldRollback = true
			logx.Infof("Auto rollback triggered for deployment %s (status: %s) due to alert %s",
				deployment.Id, deployment.Status, alert.AlertRule.Name)
		} else if deployment.Status == model.DeploymentStatusRollingBack {
			logx.Infof("Deployment %s is already rolling back, skipping auto rollback for alert %s",
				deployment.Id, alert.AlertRule.Name)
		}
	}

	// 这里可以添加实际的回滚逻辑
	if shouldRollback {
		// TODO: 实现自动回滚逻辑
		am.svcCtx.DeploymentModel.UpdateStatus(ctx, deployment.Id, model.DeploymentStatusRollingBack)
		logx.Errorf("Auto rollback logic needs to be implemented for deployment %s", deployment.Id)
	}

	return nil
}

func (am *AlertMonitor) GetActiveAlertsCount() int {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return len(am.activeAlerts)
}

// GetMonitoringStatus 获取指定发布单的监控状态
func (am *AlertMonitor) GetMonitoringStatus(deploymentID string) bool {
	am.mu.RLock()
	defer am.mu.RUnlock()
	_, exists := am.activeAlerts[deploymentID]
	return exists
}

// GetFiringAlertsCount 获取正在告警的告警数量
func (am *AlertMonitor) GetFiringAlertsCount() int {
	am.mu.RLock()
	defer am.mu.RUnlock()

	count := 0
	for _, alerts := range am.activeAlerts {
		for _, alert := range alerts {
			if alert.IsFiring {
				count++
			}
		}
	}
	return count
}

// RestartMonitoring 重启指定发布单的监控（用于状态变更时）
func (am *AlertMonitor) RestartMonitoring(ctx context.Context, deployment *model.Deployment, app *model.Application) error {
	// 先停止现有监控
	am.StopMonitoring(deployment.Id)

	// 重新启动监控
	return am.StartMonitoring(ctx, deployment, app)
}

// IsDeploymentInMonitoringStatus 检查发布单状态是否需要告警监控
func (am *AlertMonitor) IsDeploymentInMonitoringStatus(deployment *model.Deployment) bool {
	return deployment.Status == model.DeploymentStatusDeploying ||
		deployment.Status == model.DeploymentStatusRollingBack
}
