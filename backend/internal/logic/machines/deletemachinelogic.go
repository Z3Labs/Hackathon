package machines

import (
	"context"

	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteMachineLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteMachineLogic(ctx context.Context, svcCtx *svc.ServiceContext) DeleteMachineLogic {
	return DeleteMachineLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteMachineLogic) DeleteMachine(req *types.DeleteMachineReq) (resp *types.DeleteMachineResp, err error) {
	// todo: add your logic here and delete this line

	return
}
