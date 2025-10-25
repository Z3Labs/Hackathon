package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
)

type (
	Deployment struct {
		Id              string           `bson:"_id,omitempty"   json:"id,omitempty"`
		AppName         string           `bson:"appName"         json:"app_name"`         // 应用名称
		Status          DeploymentStatus `bson:"status"          json:"status"`           // 发布状态
		PackageVersion  string           `bson:"packageVersion"  json:"package_version"`  // 包版本
		ConfigPath      string           `bson:"configPath"      json:"config_path"`      // 配置文件路径
		GrayMachineId   string           `bson:"grayMachineId"   json:"gray_machine_id"`  // 灰度设备ID
		Platform        PlatformType     `bson:"platform"        json:"platform"`         // 平台类型
		Package         PackageInfo      `bson:"package"         json:"package"`          // 包信息
		Pacer           PacerConfig      `bson:"pacer"           json:"pacer"`            // 批量部署控制
		NodeDeployments []NodeDeployment `bson:"nodeDeployments" json:"node_deployments"` // 发布机器列表
		CreatedTime     int64            `bson:"createdTime"     json:"createdTime"`      // 创建时间戳
		UpdatedTime     int64            `bson:"updatedTime"     json:"updatedTime"`      // 更新时间戳
	}

	PackageInfo struct {
		URL       string    `bson:"url"       json:"url"`
		SHA256    string    `bson:"sha256"    json:"sha256"`
		Size      int64     `bson:"size"      json:"size"`
		CreatedAt time.Time `bson:"createdAt" json:"created_at"`
	}

	PacerConfig struct {
		BatchSize       int `bson:"batchSize"       json:"batch_size"`
		IntervalSeconds int `bson:"intervalSeconds" json:"interval_seconds"`
	}

	// 发布机器信息（嵌套结构体）
	NodeDeployment struct {
		Id               string               `bson:"id"            json:"id"`             // 机器唯一标识
		Ip               string               `bson:"ip"            json:"ip"`             // IP地址
		NodeDeployStatus NodeDeploymentStatus `bson:"releaseStatus" json:"release_status"` // 节点发布状态
		ReleaseLog       string               `bson:"releaseLog"    json:"release_log"`    // 发布日志
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
