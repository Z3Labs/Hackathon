package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
)

type (
	Application struct {
		Id               string          `bson:"_id"              json:"id,omitempty"`       // mongo id
		Name             string          `bson:"name"             json:"name"`               // 应用名称
		DeployPath       string          `bson:"deployPath"       json:"deploy_path"`        // 部署路径
		StartCmd         string          `bson:"startCmd"         json:"start_cmd"`          // 启动命令
		StopCmd          string          `bson:"stopCmd"          json:"stop_cmd"`           // 停止命令
		CurrentVersion   string          `bson:"currentVersion"   json:"currentVersion"`     // 当前版本
		MachineCount     int             `bson:"machineCount"     json:"machine_count"`      // 机器总数量
		HealthCount      int             `bson:"healthCount"      json:"health_count"`       // 健康机器数量
		ErrorCount       int             `bson:"errorCount"       json:"error_count"`        // 异常机器数量
		AlertCount       int             `bson:"alertCount"       json:"alert_count"`        // 告警机器数量
		Machines         []Machine       `bson:"machines"         json:"machines"`           // 机器列表
		RollbackPolicy   *RollbackPolicy `bson:"rollbackPolicy"   json:"rollback_policy"`    // 回滚策略配置
		REDMetricsConfig *REDMetrics     `bson:"redMetricsConfig" json:"red_metrics_config"` // RED指标配置
		CreatedTime      time.Time       `bson:"createdTime"      json:"createdTime"`        // 创建时间
		UpdatedTime      time.Time       `bson:"updatedTime"      json:"updatedTime"`        // 更新时间
	}

	RollbackPolicy struct {
		Enabled       bool              `bson:"enabled"       json:"enabled"`        // 是否启用自动回滚
		AlertRules    []PrometheusAlert `bson:"alertRules"    json:"alert_rules"`    // Prometheus 告警规则列表
		AutoRollback  bool              `bson:"autoRollback"  json:"auto_rollback"`  // 是否自动执行回滚
		NotifyChannel string            `bson:"notifyChannel" json:"notify_channel"` // 通知渠道(如webhook、钉钉、企业微信等)
	}

	PrometheusAlert struct {
		Name        string            `bson:"name"        json:"name"`        // 告警名称
		AlertExpr   string            `bson:"alertExpr"   json:"alert_expr"`  // Prometheus PromQL 表达式
		Duration    string            `bson:"duration"    json:"duration"`    // 持续时长(如 "1m", "5m")
		Severity    string            `bson:"severity"    json:"severity"`    // 告警级别(critical, warning, info)
		Labels      map[string]string `bson:"labels"      json:"labels"`      // 自定义标签
		Annotations map[string]string `bson:"annotations" json:"annotations"` // 告警注解(描述信息)
	}

	REDMetrics struct {
		Enabled         bool              `bson:"enabled"         json:"enabled"`          // 是否启用 RED 监控
		RateMetric      *MetricDefinition `bson:"rateMetric"      json:"rate_metric"`      // Rate - 请求速率
		ErrorMetric     *MetricDefinition `bson:"errorMetric"     json:"error_metric"`     // Error - 错误率
		DurationMetric  *MetricDefinition `bson:"durationMetric"  json:"duration_metric"`  // Duration - 响应时长
		HealthThreshold *HealthThreshold  `bson:"healthThreshold" json:"health_threshold"` // 健康度阈值
	}

	MetricDefinition struct {
		MetricName  string            `bson:"metricName"  json:"metric_name"` // Prometheus 指标名称
		PromQL      string            `bson:"promql"      json:"promql"`      // PromQL 查询语句
		Labels      map[string]string `bson:"labels"      json:"labels"`      // 指标标签过滤
		Description string            `bson:"description" json:"description"` // 指标描述
	}

	HealthThreshold struct {
		RateMin        float64 `bson:"rateMin"        json:"rate_min"`         // 最低请求速率(req/s),低于此值告警
		ErrorRateMax   float64 `bson:"errorRateMax"   json:"error_rate_max"`   // 最大错误率(%),超过此值告警
		DurationP99Max float64 `bson:"durationP99Max" json:"duration_p99_max"` // P99 响应时长上限(ms)
		DurationP95Max float64 `bson:"durationP95Max" json:"duration_p95_max"` // P95 响应时长上限(ms)
	}

	ApplicationModel interface {
		Insert(ctx context.Context, application *Application) error
		Update(ctx context.Context, application *Application) error
		Delete(ctx context.Context, id string) error
		FindById(ctx context.Context, id string) (*Application, error)
		Search(ctx context.Context, cond *ApplicationCond) ([]*Application, error)
		Count(ctx context.Context, cond *ApplicationCond) (int64, error)
	}

	defaultApplicationModel struct {
		model *mon.Model
	}

	ApplicationCond struct {
		Id         string
		Ids        []string
		Name       string
		Status     string
		Pagination *Pagination
	}
)

func NewApplicationModel(url, db string) ApplicationModel {
	return &defaultApplicationModel{
		model: mon.MustNewModel(url, db, CollectionApplication),
	}
}

func (c *ApplicationCond) genCond() bson.M {
	filter := bson.M{}

	if c.Id != "" {
		filter["_id"] = c.Id
	} else if len(c.Ids) > 0 {
		filter["_id"] = bson.M{"$in": c.Ids}
	}

	if c.Name != "" {
		filter["name"] = bson.M{"$regex": c.Name, "$options": "i"}
	}

	return filter
}

func (m *defaultApplicationModel) Insert(ctx context.Context, application *Application) error {
	application.CreatedTime = time.Now()
	application.UpdatedTime = time.Now()

	_, err := m.model.InsertOne(ctx, application)
	return err
}

func (m *defaultApplicationModel) Update(ctx context.Context, application *Application) error {
	application.UpdatedTime = time.Now()

	_, err := m.model.UpdateOne(
		ctx,
		bson.M{"_id": application.Id},
		bson.M{"$set": application},
	)
	return err
}

func (m *defaultApplicationModel) Delete(ctx context.Context, id string) error {
	_, err := m.model.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (m *defaultApplicationModel) FindById(ctx context.Context, id string) (*Application, error) {
	var application Application
	err := m.model.FindOne(ctx, &application, bson.M{"_id": id})
	if err != nil {
		return nil, err
	}
	return &application, nil
}

func (m *defaultApplicationModel) Search(ctx context.Context, cond *ApplicationCond) ([]*Application, error) {
	var result []*Application
	filter := cond.genCond()

	// 根据是否有分页参数决定查询方式
	var err error
	if cond.Pagination.IsEmpty() {
		err = m.model.Find(ctx, &result, filter)
	} else {
		findOptions := cond.Pagination.ToFindOptions()
		err = m.model.Find(ctx, &result, filter, findOptions)
	}

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *defaultApplicationModel) Count(ctx context.Context, cond *ApplicationCond) (int64, error) {
	count, err := m.model.CountDocuments(ctx, cond.genCond())
	return count, err
}
