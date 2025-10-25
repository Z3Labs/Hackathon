package machines

import (
	"context"
	"fmt"

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
	// 检查机器是否存在
	existingMachine, err := l.svcCtx.MachineModel.FindById(l.ctx, req.Id)
	if err != nil {
		l.Errorf("[DeleteMachine] MachineModel.FindById error:%v", err)
		return nil, fmt.Errorf("machine not found")
	}

	// 检查机器是否正在被使用（这里可以添加更复杂的业务逻辑检查）
	// 例如：检查是否有正在进行的部署任务使用此机器
	// 暂时跳过这个检查，直接删除

	// 删除机器
	err = l.svcCtx.MachineModel.Delete(l.ctx, req.Id)
	if err != nil {
		l.Errorf("[DeleteMachine] MachineModel.Delete error:%v", err)
		return nil, fmt.Errorf("delete machine failed")
	}

	l.Infof("[DeleteMachine] Successfully deleted machine:%s, IP:%s", req.Id, existingMachine.Ip)

	return &types.DeleteMachineResp{
		Success: true,
	}, nil
}
