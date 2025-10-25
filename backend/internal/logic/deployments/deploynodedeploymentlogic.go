package deployments

import (
	"context"
	"errors"
	"time"

	"github.com/Z3Labs/Hackathon/backend/internal/model"
	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeployNodeDeploymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeployNodeDeploymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeployNodeDeploymentLogic {
	return &DeployNodeDeploymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeployNodeDeploymentLogic) DeployNodeDeployment(req *types.DeployNodeDeploymentReq) (resp *types.DeployNodeDeploymentResp, err error) {
	deployment, err := l.svcCtx.DeploymentModel.FindById(l.ctx, req.Id)
	if err != nil {
		l.Errorf("[DeployNodeDeployment] DeploymentModel.FindById error:%v", err)
		return nil, errors.New("发布记录不存在")
	}

	if deployment.Status == model.DeploymentStatusCanceled {
		l.Errorf("[DeployNodeDeployment] Cannot deploy on canceled deployment: %s", deployment.Status)
		return nil, errors.New("已取消的发布单无法执行发布操作")
	}

	if deployment.Status == model.DeploymentStatusRolledBack {
		l.Errorf("[DeployNodeDeployment] Cannot deploy on rolled back deployment: %s", deployment.Status)
		return nil, errors.New("已回滚的发布单无法执行发布操作")
	}

	nodeDeploymentIdMap := make(map[string]bool)
	for _, id := range req.NodeDeploymentIds {
		nodeDeploymentIdMap[id] = true
	}

	var validMachineIds []string
	for _, machine := range deployment.NodeDeployments {
		if nodeDeploymentIdMap[machine.Id] {
			validMachineIds = append(validMachineIds, machine.Id)
		}
	}

	if len(validMachineIds) == 0 {
		l.Errorf("[DeployNodeDeployment] No valid machines found for deployment")
		return nil, errors.New("没有找到有效的机器进行发布")
	}

	deployCount := 0
	for i := range deployment.NodeDeployments {
		if nodeDeploymentIdMap[deployment.NodeDeployments[i].Id] {
			currentStatus := deployment.NodeDeployments[i].NodeDeployStatus
			if currentStatus == model.NodeDeploymentStatusPending {
				deployment.NodeDeployments[i].NodeDeployStatus = model.NodeDeploymentStatusDeploying
				deployCount++
			}
		}
	}

	if deployCount == 0 {
		l.Errorf("[DeployNodeDeployment] No machines were set to deploying status")
		return nil, errors.New("没有机器被设置为发布中状态，只能发布待发布状态的设备")
	}

	if deployment.Status == model.DeploymentStatusPending {
		deployment.Status = model.DeploymentStatusDeploying
	}

	deployment.UpdatedTime = time.Now().Unix()

	err = l.svcCtx.DeploymentModel.Update(l.ctx, deployment)
	if err != nil {
		l.Errorf("[DeployNodeDeployment] DeploymentModel.Update error:%v", err)
		return nil, errors.New("发布指定机器失败")
	}

	l.Infof("[DeployNodeDeployment] Successfully deployed %d machines: %v for deployment: %s", deployCount, validMachineIds, req.Id)

	return &types.DeployNodeDeploymentResp{
		Success: true,
	}, nil
}
