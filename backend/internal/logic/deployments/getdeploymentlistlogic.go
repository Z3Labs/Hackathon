package deployments

import (
	"context"
	"errors"

	"github.com/Z3Labs/Hackathon/backend/internal/model"
	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDeploymentListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDeploymentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) GetDeploymentListLogic {
	return GetDeploymentListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDeploymentListLogic) GetDeploymentList(req *types.GetDeploymentListReq) (resp *types.GetDeploymentListResp, err error) {
	// 构建查询条件
	cond := &model.DeploymentCond{
		AppName: req.AppName,
		Status:  req.Status,
	}

	// 获取总数
	total, err := l.svcCtx.DeploymentModel.Count(l.ctx, cond)
	if err != nil {
		l.Errorf("[GetDeploymentList] DeploymentModel.Count error:%v", err)
		return nil, errors.New("获取部署列表失败")
	}

	// 获取部署列表
	deployments, err := l.svcCtx.DeploymentModel.Search(l.ctx, cond)
	if err != nil {
		l.Errorf("[GetDeploymentList] DeploymentModel.Search error:%v", err)
		return nil, errors.New("获取部署列表失败")
	}

	// 转换为响应格式
	var deploymentList []types.Deployment
	for _, deployment := range deployments {
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

		deploymentList = append(deploymentList, types.Deployment{
			Id:              deployment.Id,
			AppName:         deployment.AppName,
			Status:          string(deployment.Status),
			PackageVersion:  deployment.PackageVersion,
			ConfigPath:      deployment.ConfigPath,
			GrayStrategy:    deployment.GrayStrategy,
			NodeDeployments: nodeDeployments,
			CreatedAt:       deployment.CreatedTime,
			UpdatedAt:       deployment.UpdatedTime,
		})
	}

	// 实现分页逻辑
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize

	var pagedDeployments []types.Deployment
	if start < len(deploymentList) {
		if end > len(deploymentList) {
			end = len(deploymentList)
		}
		pagedDeployments = deploymentList[start:end]
	}

	l.Infof("[GetDeploymentList] Successfully retrieved deployment list, total: %d, page: %d, pageSize: %d", total, req.Page, req.PageSize)

	return &types.GetDeploymentListResp{
		Deployments: pagedDeployments,
		Total:       total,
		Page:        req.Page,
		PageSize:    req.PageSize,
	}, nil
}
