package machines

import (
	"context"

	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TestMachineConnectionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTestMachineConnectionLogic(ctx context.Context, svcCtx *svc.ServiceContext) TestMachineConnectionLogic {
	return TestMachineConnectionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TestMachineConnectionLogic) TestMachineConnection(req *types.TestMachineConnectionReq) (resp *types.TestMachineConnectionResp, err error) {
	// todo: add your logic here and delete this line

	return
}
