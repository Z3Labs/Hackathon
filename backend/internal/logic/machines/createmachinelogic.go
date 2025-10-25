package machines

import (
	"context"

	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateMachineLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateMachineLogic(ctx context.Context, svcCtx *svc.ServiceContext) CreateMachineLogic {
	return CreateMachineLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateMachineLogic) CreateMachine(req *types.CreateMachineReq) (resp *types.CreateMachineResp, err error) {
	// todo: add your logic here and delete this line

	return
}
