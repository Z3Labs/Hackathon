package machines

import (
	"context"

	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateMachineLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateMachineLogic(ctx context.Context, svcCtx *svc.ServiceContext) UpdateMachineLogic {
	return UpdateMachineLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateMachineLogic) UpdateMachine(req *types.UpdateMachineReq) (resp *types.UpdateMachineResp, err error) {
	// todo: add your logic here and delete this line

	return
}
