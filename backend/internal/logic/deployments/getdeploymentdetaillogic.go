package deployments

import (
	"context"
	"errors"

	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDeploymentDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDeploymentDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) GetDeploymentDetailLogic {
	return GetDeploymentDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDeploymentDetailLogic) GetDeploymentDetail(req *types.GetDeploymentDetailReq) (resp *types.GetDeploymentDetailResp, err error) {
	// 根据ID查询部署
	deployment, err := l.svcCtx.DeploymentModel.FindById(l.ctx, req.Id)
	if err != nil {
		l.Errorf("[GetDeploymentDetail] DeploymentModel.FindById error:%v", err)
		return nil, errors.New("部署不存在")
	}

	// 转换发布机器信息
	var releaseMachines []types.DeploymentMachine
	for _, machine := range deployment.ReleaseMachines {
		releaseMachines = append(releaseMachines, types.DeploymentMachine{
			Id:            machine.Id,
			Ip:            machine.Ip,
			Port:          machine.Port,
			ReleaseStatus: string(machine.ReleaseStatus),
			HealthStatus:  string(machine.HealthStatus),
			ErrorStatus:   string(machine.ErrorStatus),
			AlertStatus:   string(machine.AlertStatus),
		})
	}

	// 构建响应
	deploymentDetail := types.Deployment{
		Id:              deployment.Id,
		AppName:         deployment.AppName,
		Status:          deployment.Status,
		PackageVersion:  deployment.PackageVersion,
		ConfigPath:      deployment.ConfigPath,
		GrayStrategy:    deployment.GrayStrategy,
		ReleaseMachines: releaseMachines,
		ReleaseLog:      deployment.ReleaseLog,
		CreatedAt:       deployment.CreatedTime,
		UpdatedAt:       deployment.UpdatedTime,
	}

	l.Infof("[GetDeploymentDetail] Successfully retrieved deployment detail: %s", req.Id)

	return &types.GetDeploymentDetailResp{
		Deployment: deploymentDetail,
	}, nil
}
