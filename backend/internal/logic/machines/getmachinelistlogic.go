package machines

import (
	"context"

	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMachineListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMachineListLogic(ctx context.Context, svcCtx *svc.ServiceContext) GetMachineListLogic {
	return GetMachineListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMachineListLogic) GetMachineList(req *types.GetMachineListReq) (resp *types.GetMachineListResp, err error) {
	// todo: add your logic here and delete this line

	return
}
