package logic

import (
	"context"

	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAppDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAppDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) GetAppDetailLogic {
	return GetAppDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAppDetailLogic) GetAppDetail(req *types.GetAppDetailReq) (resp *types.GetAppDetailResp, err error) {
	// todo: add your logic here and delete this line

	return
}
