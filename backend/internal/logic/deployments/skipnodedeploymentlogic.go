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

type SkipNodeDeploymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSkipNodeDeploymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SkipNodeDeploymentLogic {
	return &SkipNodeDeploymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SkipNodeDeploymentLogic) SkipNodeDeployment(req *types.SkipNodeDeploymentReq) (resp *types.SkipNodeDeploymentResp, err error) {
	deployment, err := l.svcCtx.DeploymentModel.FindById(l.ctx, req.Id)
	if err != nil {
		l.Errorf("[SkipNodeDeployment] DeploymentModel.FindById error:%v", err)
		return nil, errors.New("发布记录不存在")
	}

	if deployment.Status == model.DeploymentStatusCanceled {
		l.Errorf("[SkipNodeDeployment] Cannot skip on canceled deployment: %s", deployment.Status)
		return nil, errors.New("已取消的发布单无法执行跳过操作")
	}

	if deployment.Status == model.DeploymentStatusRolledBack {
		l.Errorf("[SkipNodeDeployment] Cannot skip on rolled back deployment: %s", deployment.Status)
		return nil, errors.New("已回滚的发布单无法执行跳过操作")
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
		l.Errorf("[SkipNodeDeployment] No valid machines found for skip")
		return nil, errors.New("没有找到有效的机器进行跳过")
	}

	skipCount := 0
	for i := range deployment.NodeDeployments {
		if nodeDeploymentIdMap[deployment.NodeDeployments[i].Id] {
			currentStatus := deployment.NodeDeployments[i].NodeDeployStatus
			if currentStatus == model.NodeDeploymentStatusPending || 
			   currentStatus == model.NodeDeploymentStatusFailed {
				deployment.NodeDeployments[i].NodeDeployStatus = model.NodeDeploymentStatusSkipped
				skipCount++
			}
		}
	}

	if skipCount == 0 {
		l.Errorf("[SkipNodeDeployment] No machines were skipped")
		return nil, errors.New("没有机器被跳过，只能跳过待发布或失败的设备")
	}

	allCompleted := true
	for _, machine := range deployment.NodeDeployments {
		if machine.NodeDeployStatus == model.NodeDeploymentStatusPending ||
		   machine.NodeDeployStatus == model.NodeDeploymentStatusDeploying {
			allCompleted = false
			break
		}
	}

	if allCompleted {
		hasSuccess := false
		for _, machine := range deployment.NodeDeployments {
			if machine.NodeDeployStatus == model.NodeDeploymentStatusSuccess {
				hasSuccess = true
				break
			}
		}
		if hasSuccess {
			deployment.Status = model.DeploymentStatusSuccess
		} else {
			deployment.Status = model.DeploymentStatusFailed
		}
	}

	deployment.UpdatedTime = time.Now().Unix()

	err = l.svcCtx.DeploymentModel.Update(l.ctx, deployment)
	if err != nil {
		l.Errorf("[SkipNodeDeployment] DeploymentModel.Update error:%v", err)
		return nil, errors.New("跳过指定机器失败")
	}

	l.Infof("[SkipNodeDeployment] Successfully skipped %d machines: %v for deployment: %s", skipCount, validMachineIds, req.Id)

	return &types.SkipNodeDeploymentResp{
		Success: true,
	}, nil
}
