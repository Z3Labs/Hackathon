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

type CancelDeploymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelDeploymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) CancelDeploymentLogic {
	return CancelDeploymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelDeploymentLogic) CancelDeployment(req *types.CancelDeploymentReq) (resp *types.CancelDeploymentResp, err error) {
	deployment, err := l.svcCtx.DeploymentModel.FindById(l.ctx, req.Id)
	if err != nil {
		l.Errorf("[CancelDeployment] DeploymentModel.FindById error:%v", err)
		return nil, errors.New("发布记录不存在")
	}

	if deployment.Status != model.DeploymentStatusPending && deployment.Status != model.DeploymentStatusDeploying {
		l.Errorf("[CancelDeployment] Invalid status for cancel: %s", deployment.Status)
		return nil, errors.New("只能取消待发布或发布中的发布单")
	}

	deployment.Status = model.DeploymentStatusFailed
	deployment.UpdatedTime = time.Now().Unix()

	err = l.svcCtx.DeploymentModel.Update(l.ctx, deployment)
	if err != nil {
		l.Errorf("[CancelDeployment] DeploymentModel.Update error:%v", err)
		return nil, errors.New("取消发布失败")
	}

	l.Infof("[CancelDeployment] Successfully cancelled deployment: %s", req.Id)

	return &types.CancelDeploymentResp{
		Success: true,
	}, nil
}
