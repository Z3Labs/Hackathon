package apps

import (
	"context"
	"errors"

	"github.com/Z3Labs/Hackathon/backend/internal/model"
	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAppListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAppListLogic(ctx context.Context, svcCtx *svc.ServiceContext) GetAppListLogic {
	return GetAppListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAppListLogic) GetAppList(req *types.GetAppListReq) (resp *types.GetAppListResp, err error) {
	// 构建查询条件
	cond := &model.ApplicationCond{
		Name: req.Name,
	}

	// 获取总数
	total, err := l.svcCtx.ApplicationModel.Count(l.ctx, cond)
	if err != nil {
		l.Errorf("[GetAppList] ApplicationModel.Count error:%v", err)
		return nil, errors.New("获取应用列表失败")
	}

	// 获取应用列表
	applications, err := l.svcCtx.ApplicationModel.Search(l.ctx, cond)
	if err != nil {
		l.Errorf("[GetAppList] ApplicationModel.Search error:%v", err)
		return nil, errors.New("获取应用列表失败")
	}

	// 转换为响应格式
	var apps []types.Application
	for _, app := range applications {
		// 转换机器信息
		var machines []types.Machine
		for _, machine := range app.Machines {
			machines = append(machines, types.Machine{
				Id:           machine.Id,
				Ip:           machine.Ip,
				Port:         machine.Port,
				HealthStatus: string(machine.HealthStatus),
				ErrorStatus:  string(machine.ErrorStatus),
				AlertStatus:  string(machine.AlertStatus),
			})
		}

		apps = append(apps, types.Application{
			Id:             app.Id,
			Name:           app.Name,
			DeployPath:     app.DeployPath,
			StartCmd:       app.StartCmd,
			StopCmd:        app.StopCmd,
			CurrentVersion: app.CurrentVersion,
			MachineCount:   app.MachineCount,
			HealthCount:    app.HealthCount,
			ErrorCount:     app.ErrorCount,
			AlertCount:     app.AlertCount,
			Machines:       machines,
			CreatedAt:      app.CreatedTime.Unix(),
			UpdatedAt:      app.UpdatedTime.Unix(),
		})
	}

	// 实现分页逻辑
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize

	var pagedApps []types.Application
	if start < len(apps) {
		if end > len(apps) {
			end = len(apps)
		}
		pagedApps = apps[start:end]
	}

	l.Infof("[GetAppList] Successfully retrieved app list, total: %d, page: %d, pageSize: %d", total, req.Page, req.PageSize)

	return &types.GetAppListResp{
		Apps:     pagedApps,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}
