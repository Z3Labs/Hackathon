package apps

import (
	"context"
	"errors"
	"time"

	"github.com/Z3Labs/Hackathon/backend/internal/model"
	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAppLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateAppLogic(ctx context.Context, svcCtx *svc.ServiceContext) UpdateAppLogic {
	return UpdateAppLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateAppLogic) UpdateApp(req *types.UpdateAppReq) (resp *types.UpdateAppResp, err error) {
	// 检查应用是否存在
	existingApp, err := l.svcCtx.ApplicationModel.FindById(l.ctx, req.Id)
	if err != nil {
		l.Errorf("[UpdateApp] ApplicationModel.FindById error:%v", err)
		return nil, errors.New("应用不存在")
	}

	// 更新应用信息
	existingApp.Name = req.Name
	if req.Repo != "" {
		existingApp.Repo = req.Repo
	}
	existingApp.DeployPath = req.DeployPath
	if req.ConfigPath != "" {
		existingApp.ConfigPath = req.ConfigPath
	}
	existingApp.StartCmd = req.StartCmd
	existingApp.StopCmd = req.StopCmd
	existingApp.UpdatedTime = time.Now()

	// 更新回滚策略配置
	if req.RollbackPolicy != nil {
		existingApp.RollbackPolicy = convertTypesToModelRollbackPolicy(req.RollbackPolicy)
	}

	// 更新RED指标配置
	if req.REDMetricsConfig != nil {
		existingApp.REDMetricsConfig = convertTypesToModelREDMetrics(req.REDMetricsConfig)
	}

	// 如果提供了机器ID列表，更新机器关联
	if req.MachineIds != nil {
		machines := make([]model.Machine, 0)
		for _, machineId := range req.MachineIds {
			machine, err := l.svcCtx.MachineModel.FindById(l.ctx, machineId)
			if err != nil {
				l.Errorf("[UpdateApp] MachineModel.FindById error:%v, machineId:%s", err, machineId)
				continue
			}
			machines = append(machines, *machine)
		}
		existingApp.Machines = machines

		// 更新机器统计
		existingApp.MachineCount = len(machines)
		healthCount := 0
		errorCount := 0
		alertCount := 0
		for _, m := range machines {
			if m.HealthStatus == "healthy" {
				healthCount++
			}
			if m.ErrorStatus == "error" {
				errorCount++
			}
			if m.AlertStatus == "alert" {
				alertCount++
			}
		}
		existingApp.HealthCount = healthCount
		existingApp.ErrorCount = errorCount
		existingApp.AlertCount = alertCount
	}

	// 保存到数据库
	err = l.svcCtx.ApplicationModel.Update(l.ctx, existingApp)
	if err != nil {
		l.Errorf("[UpdateApp] ApplicationModel.Update error:%v", err)
		return nil, errors.New("更新应用失败")
	}

	l.Infof("[UpdateApp] Successfully updated app: %s, ID: %s", req.Name, req.Id)

	return &types.UpdateAppResp{
		Success: true,
	}, nil
}
