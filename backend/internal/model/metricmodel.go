package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
)

// 常见指标名称常量（用于异常检测）
const (
	MetricCPUUsage           = "node_cpu_seconds_total"
	MetricLoadAvg            = "node_load1"
	MetricMemoryUsage        = "node_memory_MemAvailable_bytes"
	MetricDiskUsage          = "node_filesystem_avail_bytes"
	MetricDiskIOWait         = "node_disk_io_time_seconds_total"
	MetricGoroutines         = "go_goroutines"
	MetricGCPauseDuration    = "go_gc_duration_seconds"
	MetricNetworkReceive     = "node_network_receive_bytes_total"
	MetricNetworkTransmit    = "node_network_transmit_bytes_total"
)

type (
	// Metric 直接对应 Prometheus 即时查询返回格式
	Metric struct {
		Id           string            `bson:"_id,omitempty" json:"id,omitempty"`
		DeploymentId string            `bson:"deploymentId" json:"deploymentId"` // 关联的部署ID

		// 直接对应 Prometheus 即时查询返回格式
		Metric       map[string]string `bson:"metric" json:"metric"`   // 包含 __name__ 和所有标签
		Value        []interface{}     `bson:"value" json:"value"`     // [timestamp(float64), "value"(string)]

		CreatedTime  time.Time         `bson:"createdTime" json:"createdTime"`
	}

	MetricModel interface {
		Insert(ctx context.Context, metric *Metric) error
		FindByDeploymentId(ctx context.Context, deploymentId string) ([]*Metric, error)
		DeleteByDeploymentId(ctx context.Context, deploymentId string) error
	}

	defaultMetricModel struct {
		model *mon.Model
	}
)

func NewMetricModel(url, db string) MetricModel {
	return &defaultMetricModel{
		model: mon.MustNewModel(url, db, "Metrics"),
	}
}

func (m *defaultMetricModel) Insert(ctx context.Context, metric *Metric) error {
	metric.CreatedTime = time.Now()
	
	_, err := m.model.InsertOne(ctx, metric)
	return err
}

func (m *defaultMetricModel) FindByDeploymentId(ctx context.Context, deploymentId string) ([]*Metric, error) {
	var result []*Metric
	filter := bson.M{"deploymentId": deploymentId}

	err := m.model.Find(ctx, &result, filter)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *defaultMetricModel) DeleteByDeploymentId(ctx context.Context, deploymentId string) error {
	_, err := m.model.DeleteMany(ctx, bson.M{"deploymentId": deploymentId})
	return err
}
