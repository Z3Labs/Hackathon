package apps

import (
	"github.com/Z3Labs/Hackathon/backend/internal/model"
	"github.com/Z3Labs/Hackathon/backend/internal/types"
)

func convertRollbackPolicy(policy *model.RollbackPolicy) *types.RollbackPolicy {
	if policy == nil {
		return nil
	}

	var alertRules []types.PrometheusAlert
	for _, rule := range policy.AlertRules {
		alertRules = append(alertRules, types.PrometheusAlert{
			Name:        rule.Name,
			AlertExpr:   rule.AlertExpr,
			Duration:    rule.Duration,
			Severity:    rule.Severity,
			Labels:      rule.Labels,
			Annotations: rule.Annotations,
		})
	}

	return &types.RollbackPolicy{
		Enabled:       policy.Enabled,
		AlertRules:    alertRules,
		AutoRollback:  policy.AutoRollback,
		NotifyChannel: policy.NotifyChannel,
	}
}

func convertREDMetrics(metrics *model.REDMetrics) *types.REDMetrics {
	if metrics == nil {
		return nil
	}

	return &types.REDMetrics{
		Enabled:         metrics.Enabled,
		RateMetric:      convertMetricDefinition(metrics.RateMetric),
		ErrorMetric:     convertMetricDefinition(metrics.ErrorMetric),
		DurationMetric:  convertMetricDefinition(metrics.DurationMetric),
		HealthThreshold: convertHealthThreshold(metrics.HealthThreshold),
	}
}

func convertMetricDefinition(metric *model.MetricDefinition) *types.MetricDefinition {
	if metric == nil {
		return nil
	}

	return &types.MetricDefinition{
		MetricName:  metric.MetricName,
		PromQL:      metric.PromQL,
		Labels:      metric.Labels,
		Description: metric.Description,
	}
}

func convertHealthThreshold(threshold *model.HealthThreshold) *types.HealthThreshold {
	if threshold == nil {
		return nil
	}

	return &types.HealthThreshold{
		RateMin:        threshold.RateMin,
		ErrorRateMax:   threshold.ErrorRateMax,
		DurationP99Max: threshold.DurationP99Max,
		DurationP95Max: threshold.DurationP95Max,
	}
}
