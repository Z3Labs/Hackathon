package svc

import (
	"github.com/Z3Labs/Hackathon/backend/common/qiniu"
	"github.com/Z3Labs/Hackathon/backend/internal/config"
	"github.com/Z3Labs/Hackathon/backend/internal/model"
)

type ServiceContext struct {
	Config           config.Config
	ApplicationModel model.ApplicationModel
	DeploymentModel  model.DeploymentModel
	MachineModel     model.MachineModel
	ReportModel      model.ReportModel
	QiniuClient      *qiniu.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	var qiniuClient *qiniu.Client
	if c.Qiniu.AccessKey != "" && c.Qiniu.SecretKey != "" && c.Qiniu.Bucket != "" {
		qiniuClient = qiniu.NewClient(c.Qiniu.AccessKey, c.Qiniu.SecretKey, c.Qiniu.Bucket, c.Qiniu.DownloadHost)
	}

	return &ServiceContext{
		Config:           c,
		ApplicationModel: model.NewApplicationModel(c.Mongo.URL, c.Mongo.Database),
		DeploymentModel:  model.NewDeploymentModel(c.Mongo.URL, c.Mongo.Database),
		MachineModel:     model.NewMachineModel(c.Mongo.URL, c.Mongo.Database),
		ReportModel:      model.NewReportModel(c.Mongo.URL, c.Mongo.Database),
		QiniuClient:      qiniuClient,
	}
}
func NewUTServiceContext(c config.Config) *ServiceContext {
	var qiniuClient *qiniu.Client
	if c.Qiniu.AccessKey != "" && c.Qiniu.SecretKey != "" && c.Qiniu.Bucket != "" {
		qiniuClient = qiniu.NewClient(c.Qiniu.AccessKey, c.Qiniu.SecretKey, c.Qiniu.Bucket, c.Qiniu.DownloadHost)
	}

	svc := &ServiceContext{
		Config:           c,
		ApplicationModel: model.NewApplicationModel(c.Mongo.URL, c.Mongo.Database),
		DeploymentModel:  model.NewDeploymentModel(c.Mongo.URL, c.Mongo.Database),
		MachineModel:     model.NewMachineModel(c.Mongo.URL, c.Mongo.Database),
		ReportModel:      model.NewReportModel(c.Mongo.URL, c.Mongo.Database),
		QiniuClient:      qiniuClient,
	}
	return svc
}
