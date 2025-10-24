package logic

import (
	"context"

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
	// todo: add your logic here and delete this line

	return
}
