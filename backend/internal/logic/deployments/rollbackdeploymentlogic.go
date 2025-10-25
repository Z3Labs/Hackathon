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

type RollbackDeploymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRollbackDeploymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) RollbackDeploymentLogic {
	return RollbackDeploymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RollbackDeploymentLogic) RollbackDeployment(req *types.RollbackDeploymentReq) (resp *types.RollbackDeploymentResp, err error) {
	deployment, err := l.svcCtx.DeploymentModel.FindById(l.ctx, req.Id)
	if err != nil {
		l.Errorf("[RollbackDeployment] DeploymentModel.FindById error:%v", err)
		return nil, errors.New("发布记录不存在")
	}

	if deployment.Status != model.DeploymentStatusDeploying {
		l.Errorf("[RollbackDeployment] Invalid status for rollback: %s", deployment.Status)
		return nil, errors.New("只能回滚发布中的发布单")
	}

	for _, machine := range deployment.NodeDeployments {
		if machine.NodeDeployStatus == model.NodeDeploymentStatusDeploying {
			l.Errorf("[RollbackDeployment] Machine %s is still deploying", machine.Id)
			return nil, errors.New("存在发布中的设备，无法回滚")
		}
	}

	deployment.Status = model.DeploymentStatusRolledBack
	deployment.UpdatedTime = time.Now().Unix()

	err = l.svcCtx.DeploymentModel.Update(l.ctx, deployment)
	if err != nil {
		l.Errorf("[RollbackDeployment] DeploymentModel.Update error:%v", err)
		return nil, errors.New("回滚发布失败")
	}

	l.Infof("[RollbackDeployment] Successfully rolled back deployment: %s", req.Id)

	return &types.RollbackDeploymentResp{
		Success: true,
	}, nil
}
