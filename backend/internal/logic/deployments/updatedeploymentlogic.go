package deployments

import (
	"context"

	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateDeploymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateDeploymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) UpdateDeploymentLogic {
	return UpdateDeploymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateDeploymentLogic) UpdateDeployment(req *types.UpdateDeploymentReq) (resp *types.UpdateDeploymentResp, err error) {
	// todo: add your logic here and delete this line

	return
}
