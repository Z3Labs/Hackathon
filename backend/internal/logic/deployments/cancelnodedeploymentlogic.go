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

type CancelNodeDeploymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelNodeDeploymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelNodeDeploymentLogic {
	return &CancelNodeDeploymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelNodeDeploymentLogic) CancelNodeDeployment(req *types.CancelNodeDeploymentReq) (resp *types.CancelNodeDeploymentResp, err error) {
	deployment, err := l.svcCtx.DeploymentModel.FindById(l.ctx, req.Id)
	if err != nil {
		l.Errorf("[CancelNodeDeployment] DeploymentModel.FindById error:%v", err)
		return nil, errors.New("发布记录不存在")
	}

	if deployment.Status == model.DeploymentStatusCanceled {
		l.Errorf("[CancelNodeDeployment] Deployment already canceled: %s", deployment.Status)
		return nil, errors.New("发布单已被取消")
	}

	if deployment.Status == model.DeploymentStatusRolledBack {
		l.Errorf("[CancelNodeDeployment] Cannot cancel rolled back deployment: %s", deployment.Status)
		return nil, errors.New("已回滚的发布单无法执行取消操作")
	}

	if deployment.Status == model.DeploymentStatusSuccess {
		l.Errorf("[CancelNodeDeployment] Cannot cancel successful deployment: %s", deployment.Status)
		return nil, errors.New("已成功的发布单无法执行取消操作")
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
		l.Errorf("[CancelNodeDeployment] No valid machines found for cancel")
		return nil, errors.New("没有找到有效的机器进行取消")
	}

	cancelCount := 0
	for i := range deployment.NodeDeployments {
		if nodeDeploymentIdMap[deployment.NodeDeployments[i].Id] {
			currentStatus := deployment.NodeDeployments[i].NodeDeployStatus
			if currentStatus == model.NodeDeploymentStatusPending || 
			   currentStatus == model.NodeDeploymentStatusDeploying {
				deployment.NodeDeployments[i].NodeDeployStatus = model.NodeDeploymentStatusFailed
				deployment.NodeDeployments[i].ReleaseLog = "用户手动取消"
				cancelCount++
			}
		}
	}

	if cancelCount == 0 {
		l.Errorf("[CancelNodeDeployment] No machines were canceled")
		return nil, errors.New("没有机器被取消，只能取消待发布或发布中的设备")
	}

	allCanceled := true
	for _, machine := range deployment.NodeDeployments {
		if machine.NodeDeployStatus == model.NodeDeploymentStatusPending ||
		   machine.NodeDeployStatus == model.NodeDeploymentStatusDeploying ||
		   machine.NodeDeployStatus == model.NodeDeploymentStatusSuccess {
			allCanceled = false
			break
		}
	}

	if allCanceled {
		deployment.Status = model.DeploymentStatusCanceled
	}

	deployment.UpdatedTime = time.Now().Unix()

	err = l.svcCtx.DeploymentModel.Update(l.ctx, deployment)
	if err != nil {
		l.Errorf("[CancelNodeDeployment] DeploymentModel.Update error:%v", err)
		return nil, errors.New("取消指定机器失败")
	}

	l.Infof("[CancelNodeDeployment] Successfully canceled %d machines: %v for deployment: %s", cancelCount, validMachineIds, req.Id)

	return &types.CancelNodeDeploymentResp{
		Success: true,
	}, nil
}
