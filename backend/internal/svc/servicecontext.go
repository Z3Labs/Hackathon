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
	MetricModel      model.MetricModel
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
		MetricModel:      model.NewMetricModel(c.Mongo.URL, c.Mongo.Database),
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
		MetricModel:      model.NewMetricModel(c.Mongo.URL, c.Mongo.Database),
		ReportModel:      model.NewReportModel(c.Mongo.URL, c.Mongo.Database),
		ReleasePlanModel: model.NewReleasePlanModel(c.Mongo.URL, c.Mongo.Database),
		NodeStatusModel:  model.NewNodeStatusModel(c.Mongo.URL, c.Mongo.Database),
	}
	plans, _ := svc.ReleasePlanModel.Search(context.Background(), &model.ReleasePlanCond{})
	for _, plan := range plans {
		plan.Status = model.PlanStatusPending
		svc.ReleasePlanModel.Delete(context.Background(), plan.Id)
	}

	nodeDelpyRecs, _ := svc.DeploymentModel.Search(context.Background(), &model.DeploymentCond{})

	for _, rec := range nodeDelpyRecs {
		svc.DeploymentModel.Delete(context.Background(), rec.Id)

	}
	return svc
}
