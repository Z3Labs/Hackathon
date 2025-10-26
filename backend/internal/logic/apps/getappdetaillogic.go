package apps

import (
	"context"
	"errors"

	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAppDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAppDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) GetAppDetailLogic {
	return GetAppDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAppDetailLogic) GetAppDetail(req *types.GetAppDetailReq) (resp *types.GetAppDetailResp, err error) {
	// 根据ID查询应用
	application, err := l.svcCtx.ApplicationModel.FindById(l.ctx, req.Id)
	if err != nil {
		l.Errorf("[GetAppDetail] ApplicationModel.FindById error:%v", err)
		return nil, errors.New("应用不存在")
	}

	// 转换机器信息
	var machines []types.Machine
	for _, machine := range application.Machines {
		machines = append(machines, types.Machine{
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
		})
	}

	// 构建响应
	app := types.Application{
		Id:               application.Id,
		Name:             application.Name,
		DeployPath:       application.DeployPath,
		StartCmd:         application.StartCmd,
		StopCmd:          application.StopCmd,
		CurrentVersion:   application.CurrentVersion,
		MachineCount:     application.MachineCount,
		HealthCount:      application.HealthCount,
		ErrorCount:       application.ErrorCount,
		AlertCount:       application.AlertCount,
		Machines:         machines,
		RollbackPolicy:   convertRollbackPolicy(application.RollbackPolicy),
		REDMetricsConfig: convertREDMetrics(application.REDMetricsConfig),
		CreatedAt:        application.CreatedTime.Unix(),
		UpdatedAt:        application.UpdatedTime.Unix(),
	}

	l.Infof("[GetAppDetail] Successfully retrieved app detail: %s", req.Id)

	return &types.GetAppDetailResp{
		Application: app,
	}, nil
}
