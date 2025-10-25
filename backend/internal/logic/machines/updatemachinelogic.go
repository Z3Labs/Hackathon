package machines

import (
	"context"
	"fmt"

	"github.com/Z3Labs/Hackathon/backend/internal/model"
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
	// 检查机器是否存在
	existingMachine, err := l.svcCtx.MachineModel.FindById(l.ctx, req.Id)
	if err != nil {
		l.Errorf("[UpdateMachine] MachineModel.FindById error:%v", err)
		return nil, fmt.Errorf("machine not found")
	}

	// 如果IP地址发生变化，检查新IP是否已被其他机器使用
	if existingMachine.Ip != req.Ip {
		cond := &model.MachineCond{
			Ip: req.Ip,
		}
		machines, err := l.svcCtx.MachineModel.Search(l.ctx, cond)
		if err != nil {
			l.Errorf("[UpdateMachine] MachineModel.Search error:%v", err)
			return nil, fmt.Errorf("query machine failed")
		}
		// 过滤掉当前机器
		for _, machine := range machines {
			if machine.Id != req.Id {
				return nil, fmt.Errorf("IP address %s already in use by other machine", req.Ip)
			}
		}
	}

	// 更新机器信息
	machine := &model.Machine{
		Id:           req.Id,
		Name:         req.Name,
		Ip:           req.Ip,
		Port:         req.Port,
		Username:     req.Username,
		Password:     req.Password,
		Description:  req.Description,
		HealthStatus: existingMachine.HealthStatus, // 保持原有状态
		ErrorStatus:  existingMachine.ErrorStatus,  // 保持原有状态
		AlertStatus:  existingMachine.AlertStatus,  // 保持原有状态
		CreatedTime:  existingMachine.CreatedTime,  // 保持原有创建时间
		UpdatedTime:  existingMachine.UpdatedTime,  // 将在Update方法中更新
	}

	err = l.svcCtx.MachineModel.Update(l.ctx, machine)
	if err != nil {
		l.Errorf("[UpdateMachine] MachineModel.Update error:%v", err)
		return nil, fmt.Errorf("update machine failed")
	}

	l.Infof("[UpdateMachine] Successfully updated machine:%s, IP:%s", req.Id, req.Ip)

	return &types.UpdateMachineResp{
		Success: true,
	}, nil
}
