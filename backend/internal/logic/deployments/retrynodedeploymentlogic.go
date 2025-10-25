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

type RetryNodeDeploymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRetryNodeDeploymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetryNodeDeploymentLogic {
	return &RetryNodeDeploymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RetryNodeDeploymentLogic) RetryNodeDeployment(req *types.RetryNodeDeploymentReq) (resp *types.RetryNodeDeploymentResp, err error) {
	deployment, err := l.svcCtx.DeploymentModel.FindById(l.ctx, req.Id)
	if err != nil {
		l.Errorf("[RetryNodeDeployment] DeploymentModel.FindById error:%v", err)
		return nil, errors.New("发布记录不存在")
	}

	if deployment.Status == model.DeploymentStatusCanceled {
		l.Errorf("[RetryNodeDeployment] Cannot retry on canceled deployment: %s", deployment.Status)
		return nil, errors.New("已取消的发布单无法执行重试操作")
	}

	if deployment.Status == model.DeploymentStatusRolledBack {
		l.Errorf("[RetryNodeDeployment] Cannot retry on rolled back deployment: %s", deployment.Status)
		return nil, errors.New("已回滚的发布单无法执行重试操作")
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
		l.Errorf("[RetryNodeDeployment] No valid machines found for retry")
		return nil, errors.New("没有找到有效的机器进行重试")
	}

	retryCount := 0
	for i := range deployment.NodeDeployments {
		if nodeDeploymentIdMap[deployment.NodeDeployments[i].Id] {
			if deployment.NodeDeployments[i].NodeDeployStatus == model.NodeDeploymentStatusFailed {
				deployment.NodeDeployments[i].NodeDeployStatus = model.NodeDeploymentStatusDeploying
				deployment.NodeDeployments[i].ReleaseLog = ""
				retryCount++
			}
		}
	}

	if retryCount == 0 {
		l.Errorf("[RetryNodeDeployment] No failed machines found for retry")
		return nil, errors.New("没有失败的机器可以重试")
	}

	if deployment.Status == model.DeploymentStatusFailed {
		deployment.Status = model.DeploymentStatusDeploying
	}

	deployment.UpdatedTime = time.Now().Unix()

	err = l.svcCtx.DeploymentModel.Update(l.ctx, deployment)
	if err != nil {
		l.Errorf("[RetryNodeDeployment] DeploymentModel.Update error:%v", err)
		return nil, errors.New("重试指定机器失败")
	}

	l.Infof("[RetryNodeDeployment] Successfully retried %d machines: %v for deployment: %s", retryCount, validMachineIds, req.Id)

	return &types.RetryNodeDeploymentResp{
		Success: true,
	}, nil
}
