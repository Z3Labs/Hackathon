package logic

import (
	"context"

	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAppListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAppListLogic(ctx context.Context, svcCtx *svc.ServiceContext) GetAppListLogic {
	return GetAppListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAppListLogic) GetAppList(req *types.GetAppListReq) (resp *types.GetAppListResp, err error) {
	// todo: add your logic here and delete this line

	return
}
