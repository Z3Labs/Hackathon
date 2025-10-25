package machines

import (
	"context"
	"fmt"

	"github.com/Z3Labs/Hackathon/backend/internal/model"
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
	// 构建查询条件
	cond := &model.MachineCond{}

	if req.Name != "" {
		cond.Name = req.Name
	}

	if req.Ip != "" {
		cond.Ip = req.Ip
	}
	if req.HealthStatus != "" {
		cond.HealthStatus = req.HealthStatus
	}
	if req.ErrorStatus != "" {
		cond.ErrorStatus = req.ErrorStatus
	}
	if req.AlertStatus != "" {
		cond.AlertStatus = req.AlertStatus
	}

	// 查询机器列表
	machines, err := l.svcCtx.MachineModel.Search(l.ctx, cond)
	if err != nil {
		l.Errorf("[GetMachineList] MachineModel.Search error:%v", err)
		return nil, fmt.Errorf("query machine list failed")
	}

	// 查询总数
	total, err := l.svcCtx.MachineModel.Count(l.ctx, cond)
	if err != nil {
		l.Errorf("[GetMachineList] MachineModel.Count error:%v", err)
		return nil, fmt.Errorf("query machine count failed")
	}

	// 转换为API响应格式
	var machineList []types.Machine
	for _, machine := range machines {
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
		machineList = append(machineList, machineResp)
	}

	l.Infof("[GetMachineList] Successfully retrieved machine list, total:%d", total)

	return &types.GetMachineListResp{
		Machines: machineList,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}
