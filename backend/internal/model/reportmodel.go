package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewReportId() string {
	return primitive.NewObjectID().Hex()
}

type (
	// Report 存储 AI 生成的诊断报告
	Report struct {
		Id           string       `bson:"_id,omitempty" json:"id,omitempty"`
		DeploymentId string       `bson:"deploymentId"  json:"deploymentId"` // 关联的部署ID
		Content      string       `bson:"content"       json:"content"`      // AI 生成的报告
		Status       ReportStatus `bson:"status"        json:"status"`       // 报告生成状态
		CreatedTime  time.Time    `bson:"createdTime"   json:"createdTime"`
		UpdatedTime  time.Time    `bson:"updatedTime"   json:"updatedTime"`
	}

	ReportModel interface {
		Insert(ctx context.Context, report *Report) error
		FindById(ctx context.Context, id string) (*Report, error)
		FindByDeploymentId(ctx context.Context, deploymentId string) ([]*Report, error)
		Update(ctx context.Context, report *Report) error
		DeleteByDeploymentId(ctx context.Context, deploymentId string) error
	}

	defaultReportModel struct {
		model *mon.Model
	}
)

func NewReportModel(url, db string) ReportModel {
	return &defaultReportModel{
		model: mon.MustNewModel(url, db, CollectionReport),
	}
}

func (m *defaultReportModel) Insert(ctx context.Context, report *Report) error {
	if report.Id == "" {
		report.Id = NewReportId()
	}
	report.CreatedTime = time.Now()
	report.UpdatedTime = time.Now()

	_, err := m.model.InsertOne(ctx, report)
	return err
}

func (m *defaultReportModel) FindById(ctx context.Context, id string) (*Report, error) {
	var report Report
	err := m.model.FindOne(ctx, &report, bson.M{"_id": id})
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (m *defaultReportModel) FindByDeploymentId(ctx context.Context, deploymentId string) ([]*Report, error) {
	var reports []*Report
	
	// 使用数据库排序，按创建时间倒序（最新的在前）
	opts := options.Find().SetSort(bson.D{{Key: "createdTime", Value: -1}})
	
	err := m.model.Find(ctx, &reports, bson.M{"deploymentId": deploymentId}, opts)
	if err != nil {
		return nil, err
	}
	
	return reports, nil
}

func (m *defaultReportModel) Update(ctx context.Context, report *Report) error {
	report.UpdatedTime = time.Now()

	_, err := m.model.UpdateOne(
		ctx,
		bson.M{"_id": report.Id},
		bson.M{"$set": report},
	)
	return err
}

func (m *defaultReportModel) DeleteByDeploymentId(ctx context.Context, deploymentId string) error {
	_, err := m.model.DeleteOne(ctx, bson.M{"deploymentId": deploymentId})
	return err
}
