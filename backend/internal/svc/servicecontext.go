package svc

import (
	"context"

	"github.com/Z3Labs/Hackathon/backend/internal/config"
	"github.com/Z3Labs/Hackathon/backend/internal/model"
)

type ServiceContext struct {
	Config           config.Config
	ApplicationModel model.ApplicationModel
	DeploymentModel  model.DeploymentModel
	MachineModel     model.MachineModel
	ReportModel      model.ReportModel
	ReleasePlanModel model.ReleasePlanModel
	NodeStatusModel  model.NodeStatusModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:           c,
		ApplicationModel: model.NewApplicationModel(c.Mongo.URL, c.Mongo.Database),
		DeploymentModel:  model.NewDeploymentModel(c.Mongo.URL, c.Mongo.Database),
		MachineModel:     model.NewMachineModel(c.Mongo.URL, c.Mongo.Database),
		ReportModel:      model.NewReportModel(c.Mongo.URL, c.Mongo.Database),
		ReleasePlanModel: model.NewReleasePlanModel(c.Mongo.URL, c.Mongo.Database),
		NodeStatusModel:  model.NewNodeStatusModel(c.Mongo.URL, c.Mongo.Database),
	}
}
func NewUTServiceContext(c config.Config) *ServiceContext {
	svc := &ServiceContext{
		Config:           c,
		ApplicationModel: model.NewApplicationModel(c.Mongo.URL, c.Mongo.Database),
		DeploymentModel:  model.NewDeploymentModel(c.Mongo.URL, c.Mongo.Database),
		MachineModel:     model.NewMachineModel(c.Mongo.URL, c.Mongo.Database),
		ReportModel:      model.NewReportModel(c.Mongo.URL, c.Mongo.Database),
		ReleasePlanModel: model.NewReleasePlanModel(c.Mongo.URL, c.Mongo.Database),
		NodeStatusModel:  model.NewNodeStatusModel(c.Mongo.URL, c.Mongo.Database),
	}
	svc.ReleasePlanModel.DeleteMany(context.Background(), &model.ReleasePlanCond{})
	svc.NodeStatusModel.DeleteMany(context.Background(), &model.NodeStatusCond{})
	return svc
}
