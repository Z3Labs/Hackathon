package machines

import (
	"context"
	"fmt"
	"time"

	"github.com/Z3Labs/Hackathon/backend/internal/model"
	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"
	"github.com/Z3Labs/Hackathon/backend/internal/utils"

	"github.com/google/uuid"
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
	// 检查IP是否已存在
	cond := &model.MachineCond{
		Ip: req.Ip,
	}
	existingMachines, err := l.svcCtx.MachineModel.Search(l.ctx, cond)
	if err != nil {
		l.Errorf("[CreateMachine] MachineModel.Search error:%v", err)
		return nil, fmt.Errorf("query machine failed")
	}
	if len(existingMachines) > 0 {
		return nil, fmt.Errorf("IP address %s already exists", req.Ip)
	}

	// 生成机器ID
	machineId := uuid.New().String()

	// 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		l.Errorf("[CreateMachine] HashPassword error:%v", err)
		return nil, fmt.Errorf("encrypt password failed")
	}

	// 创建机器对象
	machine := &model.Machine{
		Id:           machineId,
		Name:         req.Name,
		Ip:           req.Ip,
		Port:         req.Port,
		Username:     req.Username,
		Password:     hashedPassword,
		Description:  req.Description,
		HealthStatus: model.HealthStatusHealthy, // 默认健康状态
		ErrorStatus:  model.ErrorStatusNormal,   // 默认正常状态
		AlertStatus:  model.AlertStatusNormal,   // 默认正常状态
		CreatedTime:  time.Now(),
		UpdatedTime:  time.Now(),
	}

	// 保存到数据库
	err = l.svcCtx.MachineModel.Insert(l.ctx, machine)
	if err != nil {
		l.Errorf("[CreateMachine] MachineModel.Insert error:%v", err)
		return nil, fmt.Errorf("save machine failed")
	}

	l.Infof("[CreateMachine] Successfully created machine:%s, IP:%s", machineId, req.Ip)

	return &types.CreateMachineResp{
		Id: machineId,
	}, nil
}
