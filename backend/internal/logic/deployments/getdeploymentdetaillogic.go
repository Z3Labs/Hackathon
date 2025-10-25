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
	var nodeDeployments []types.NodeDeployment
	for _, machine := range deployment.NodeDeployments {
		nodeDeployments = append(nodeDeployments, types.NodeDeployment{
			Id:               machine.Id,
			Ip:               machine.Ip,
			NodeDeployStatus: string(machine.NodeDeployStatus),
			ReleaseLog:       machine.ReleaseLog,
		})
	}

	// 构建响应
	deploymentDetail := types.Deployment{
		Id:              deployment.Id,
		AppName:         deployment.AppName,
		Status:          string(deployment.Status),
		PackageVersion:  deployment.PackageVersion,
		ConfigPath:      deployment.ConfigPath,
		GrayMachineId:   deployment.GrayMachineId,
		NodeDeployments: nodeDeployments,
		CreatedAt:       deployment.CreatedTime,
		UpdatedAt:       deployment.UpdatedTime,
	}

	// 查询诊断报告
	var reportResp *types.Report

	report, err := l.svcCtx.ReportModel.FindByDeploymentId(l.ctx, deployment.Id)
	if err == nil && report != nil {
		// 找到报告，转换为响应类型
		reportResp = &types.Report{
			Id:           report.Id,
			DeploymentId: report.DeploymentId,
			Content:      report.Content,
			Status:       string(report.Status),
			CreatedAt:    report.CreatedTime.Unix(),
			UpdatedAt:    report.UpdatedTime.Unix(),
		}
	}

	l.Infof("[GetDeploymentDetail] Successfully retrieved deployment detail: %s", req.Id)

	return &types.GetDeploymentDetailResp{
		Deployment: deploymentDetail,
		Report:     reportResp,
	}, nil
}
