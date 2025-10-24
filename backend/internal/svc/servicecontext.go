package svc

import (
	"github.com/Z3Labs/Hackathon/backend/internal/config"
	"github.com/Z3Labs/Hackathon/backend/internal/model"
)

type ServiceContext struct {
	Config           config.Config
	ApplicationModel model.ApplicationModel
	DeploymentModel  model.DeploymentModel
	MachineModel     model.MachineModel
	MetricModel      model.MetricModel
	ReportModel      model.ReportModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:           c,
		ApplicationModel: model.NewApplicationModel(c.Mongo.URL, c.Mongo.Database),
		DeploymentModel:  model.NewDeploymentModel(c.Mongo.URL, c.Mongo.Database),
		MachineModel:     model.NewMachineModel(c.Mongo.URL, c.Mongo.Database),
		MetricModel:      model.NewMetricModel(c.Mongo.URL, c.Mongo.Database),
		ReportModel:      model.NewReportModel(c.Mongo.URL, c.Mongo.Database),
	}
}
