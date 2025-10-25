package alert

import (
	"context"

	"github.com/Z3Labs/Hackathon/backend/internal/clients/diagnosis"
	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AlertCallBackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAlertCallBackLogic(ctx context.Context, svcCtx *svc.ServiceContext) AlertCallBackLogic {
	return AlertCallBackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AlertCallBackLogic) AlertCallBack(req *types.PostAlertCallbackReq) error {

	// TODO 其它逻辑

	diagnosis.New(l.ctx, l.svcCtx, l.svcCtx.Config.AI).GenerateReport(req)

	return nil
}
