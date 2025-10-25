package machines

import (
	"context"
	"fmt"

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
	// 查询机器详情
	machine, err := l.svcCtx.MachineModel.FindById(l.ctx, req.Id)
	if err != nil {
		l.Errorf("[GetMachineDetail] MachineModel.FindById error:%v", err)
		return nil, fmt.Errorf("machine not found")
	}

	// 转换为API响应格式
	machineResp := types.Machine{
		Id:           machine.Id,
		Name:         machine.Name,
		Ip:           machine.Ip,
		Port:         machine.Port,
		Username:     machine.Username,
		Password:     machine.Password,
		Description:  machine.Description,
		HealthStatus: string(machine.HealthStatus),
		ErrorStatus:  string(machine.ErrorStatus),
		AlertStatus:  string(machine.AlertStatus),
		CreatedAt:    machine.CreatedTime.Unix(),
		UpdatedAt:    machine.UpdatedTime.Unix(),
	}

	l.Infof("[GetMachineDetail] Successfully retrieved machine detail:%s", req.Id)

	return &types.GetMachineDetailResp{
		Machine: machineResp,
	}, nil
}
