package deployments

import (
	"context"

	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateDeploymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateDeploymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) CreateDeploymentLogic {
	return CreateDeploymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateDeploymentLogic) CreateDeployment(req *types.CreateDeploymentReq) (resp *types.CreateDeploymentResp, err error) {
	// todo: add your logic here and delete this line

	return
}
