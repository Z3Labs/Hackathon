package machines

import (
	"context"

	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMachineDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMachineDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) GetMachineDetailLogic {
	return GetMachineDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMachineDetailLogic) GetMachineDetail(req *types.GetMachineDetailReq) (resp *types.GetMachineDetailResp, err error) {
	// todo: add your logic here and delete this line

	return
}
