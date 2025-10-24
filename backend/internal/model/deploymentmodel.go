package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
)

type (
	Deployment struct {
		Id              string              `bson:"_id,omitempty"   json:"id,omitempty"`
		AppName         string              `bson:"appName"         json:"app_name"`         // 应用名称
		Status          string              `bson:"status"          json:"status"`           // 发布状态: pending-待发布, deploying-发布中, success-成功, failed-失败, rolled_back-已回滚
		PackageVersion  string              `bson:"packageVersion"  json:"package_version"`  // 包版本
		ConfigPath      string              `bson:"configPath"      json:"config_path"`      // 配置文件路径
		GrayStrategy    string              `bson:"grayStrategy"    json:"gray_strategy"`    // 灰度策略: canary-金丝雀发布, blue-green-蓝绿发布, all-全量发布
		StartTime       int64               `bson:"startTime"       json:"start_time"`       // 开始时间戳
		EndTime         int64               `bson:"endTime"         json:"end_time"`         // 结束时间戳
		ReleaseMachines []DeploymentMachine `bson:"releaseMachines" json:"release_machines"` // 发布机器列表
		ReleaseLog      string              `bson:"releaseLog"      json:"release_log"`      // 发布日志
		CreatedTime     int64               `bson:"createdTime"     json:"createdTime"`      // 创建时间戳
		UpdatedTime     int64               `bson:"updatedTime"     json:"updatedTime"`      // 更新时间戳
	}

	// 发布机器信息（嵌套结构体）
	DeploymentMachine struct {
		Id            string        `bson:"id"            json:"id"`             // 机器唯一标识
		Ip            string        `bson:"ip"            json:"ip"`             // IP地址
		Port          int           `bson:"port"          json:"port"`           // 端口号
		ReleaseStatus ReleaseStatus `bson:"releaseStatus" json:"release_status"` // 发布状态: pending-待发布, deploying-发布中, success-成功, failed-失败
		HealthStatus  HealthStatus  `bson:"healthStatus"  json:"health_status"`  // 健康状态: healthy-健康, unhealthy-不健康
		ErrorStatus   ErrorStatus   `bson:"errorStatus"   json:"error_status"`   // 异常状态: normal-正常, error-异常
		AlertStatus   AlertStatus   `bson:"alertStatus"   json:"alert_status"`   // 告警状态: normal-正常, alert-告警
	}

	DeploymentModel interface {
		Insert(ctx context.Context, deployment *Deployment) error
		Update(ctx context.Context, deployment *Deployment) error
		Delete(ctx context.Context, id string) error
		FindById(ctx context.Context, id string) (*Deployment, error)
		Search(ctx context.Context, cond *DeploymentCond) ([]*Deployment, error)
		Count(ctx context.Context, cond *DeploymentCond) (int64, error)
	}

	defaultDeploymentModel struct {
		model *mon.Model
	}

	DeploymentCond struct {
		Id      string
		Ids     []string
		AppName string
		Status  string
	}
)

func NewDeploymentModel(url, db string) DeploymentModel {
	return &defaultDeploymentModel{
		model: mon.MustNewModel(url, db, CollectionDeployment),
	}
}

func (c *DeploymentCond) genCond() bson.M {
	filter := bson.M{}

	if c.Id != "" {
		filter["_id"] = c.Id
	} else if len(c.Ids) > 0 {
		filter["_id"] = bson.M{"$in": c.Ids}
	}

	if c.AppName != "" {
		filter["appName"] = bson.M{"$regex": c.AppName, "$options": "i"}
	}

	if c.Status != "" {
		filter["status"] = c.Status
	}

	return filter
}

func (m *defaultDeploymentModel) Insert(ctx context.Context, deployment *Deployment) error {
	deployment.CreatedTime = time.Now().Unix()
	deployment.UpdatedTime = time.Now().Unix()

	_, err := m.model.InsertOne(ctx, deployment)
	return err
}

func (m *defaultDeploymentModel) Update(ctx context.Context, deployment *Deployment) error {
	deployment.UpdatedTime = time.Now().Unix()

	_, err := m.model.UpdateOne(
		ctx,
		bson.M{"_id": deployment.Id},
		bson.M{"$set": deployment},
	)
	return err
}

func (m *defaultDeploymentModel) Delete(ctx context.Context, id string) error {
	_, err := m.model.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (m *defaultDeploymentModel) FindById(ctx context.Context, id string) (*Deployment, error) {
	var deployment Deployment
	err := m.model.FindOne(ctx, &deployment, bson.M{"_id": id})
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

func (m *defaultDeploymentModel) Search(ctx context.Context, cond *DeploymentCond) ([]*Deployment, error) {
	var result []*Deployment
	filter := cond.genCond()

	err := m.model.Find(ctx, &result, filter)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *defaultDeploymentModel) Count(ctx context.Context, cond *DeploymentCond) (int64, error) {
	count, err := m.model.CountDocuments(ctx, cond.genCond())
	return count, err
}
