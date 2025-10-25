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

type RollbackNodeDeploymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRollbackNodeDeploymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) RollbackNodeDeploymentLogic {
	return RollbackNodeDeploymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RollbackNodeDeploymentLogic) RollbackNodeDeployment(req *types.RollbackNodeDeploymentReq) (resp *types.RollbackNodeDeploymentResp, err error) {
	// 查找发布单
	deployment, err := l.svcCtx.DeploymentModel.FindById(l.ctx, req.Id)
	if err != nil {
		l.Errorf("[RollbackNodeDeployment] DeploymentModel.FindById error:%v", err)
		return nil, errors.New("发布记录不存在")
	}

	// 检查发布单状态，只有发布中或成功的发布单才能回滚
	if deployment.Status != model.DeploymentStatusDeploying && deployment.Status != model.DeploymentStatusSuccess {
		l.Errorf("[RollbackNodeDeployment] Invalid status for rollback: %s", deployment.Status)
		return nil, errors.New("只能回滚发布中或成功的发布单")
	}

	// 验证指定的机器ID是否存在
	nodeDeploymentIdMap := make(map[string]bool)
	for _, id := range req.NodeDeploymentIds {
		nodeDeploymentIdMap[id] = true
	}

	// 检查指定的设备是否正在发布中
	for _, machine := range deployment.NodeDeployments {
		if nodeDeploymentIdMap[machine.Id] && machine.NodeDeployStatus == model.NodeDeploymentStatusDeploying {
			l.Errorf("[RollbackNodeDeployment] Machine %s is still deploying", machine.Id)
			return nil, errors.New("指定的设备正在发布中，无法回滚")
		}
	}

	// 检查指定的机器是否存在于发布单中
	var validMachineIds []string
	for _, machine := range deployment.NodeDeployments {
		if nodeDeploymentIdMap[machine.Id] {
			validMachineIds = append(validMachineIds, machine.Id)
		}
	}

	if len(validMachineIds) == 0 {
		l.Errorf("[RollbackNodeDeployment] No valid machines found for rollback")
		return nil, errors.New("没有找到有效的机器进行回滚")
	}

	// 执行指定机器的回滚
	rollbackCount := 0
	for i := range deployment.NodeDeployments {
		if nodeDeploymentIdMap[deployment.NodeDeployments[i].Id] {
			if deployment.NodeDeployments[i].NodeDeployStatus == model.NodeDeploymentStatusSuccess {
				deployment.NodeDeployments[i].NodeDeployStatus = model.NodeDeploymentStatusRolledBack
				rollbackCount++
			}
		}
	}

	if rollbackCount == 0 {
		l.Errorf("[RollbackNodeDeployment] No machines were successfully rolled back")
		return nil, errors.New("没有机器被成功回滚")
	}

	// 检查是否所有设备都已回滚，如果是则更新发布单状态
	allRolledBack := true
	for _, machine := range deployment.NodeDeployments {
		if machine.NodeDeployStatus != model.NodeDeploymentStatusRolledBack &&
			machine.NodeDeployStatus != model.NodeDeploymentStatusFailed {
			allRolledBack = false
			break
		}
	}

	if allRolledBack {
		deployment.Status = model.DeploymentStatusRolledBack
	} else {
		// 部分回滚，保持原状态或设置为部分成功
		if deployment.Status == model.DeploymentStatusSuccess {
			deployment.Status = model.DeploymentStatusDeploying // 保持发布中状态
		}
	}

	deployment.UpdatedTime = time.Now().Unix()

	// 更新发布单
	err = l.svcCtx.DeploymentModel.Update(l.ctx, deployment)
	if err != nil {
		l.Errorf("[RollbackNodeDeployment] DeploymentModel.Update error:%v", err)
		return nil, errors.New("回滚指定机器失败")
	}

	l.Infof("[RollbackNodeDeployment] Successfully rolled back %d machines: %v for deployment: %s", rollbackCount, validMachineIds, req.Id)

	return &types.RollbackNodeDeploymentResp{
		Success: true,
	}, nil
}
