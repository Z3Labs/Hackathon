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
		Name:       req.Name,
		Pagination: model.NewPaginationWithDefaultSort(req.Page, req.PageSize),
	}

	// 获取总数（不分页）
	countCond := &model.ApplicationCond{
		Name: req.Name,
	}
	total, err := l.svcCtx.ApplicationModel.Count(l.ctx, countCond)
	if err != nil {
		l.Errorf("[GetAppList] ApplicationModel.Count error:%v", err)
		return nil, errors.New("获取应用列表失败")
	}

	// 获取分页应用列表
	applications, err := l.svcCtx.ApplicationModel.Search(l.ctx, cond)

	// 转换为响应格式
	var apps []types.Application
	for _, app := range applications {
		// 转换机器信息
		var machines []types.Machine
		for _, machine := range app.Machines {
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

		apps = append(apps, types.Application{
			Id:               app.Id,
			Name:             app.Name,
			DeployPath:       app.DeployPath,
			ConfigPath:       app.ConfigPath,
			StartCmd:         app.StartCmd,
			StopCmd:          app.StopCmd,
			CurrentVersion:   app.CurrentVersion,
			MachineCount:     app.MachineCount,
			HealthCount:      app.HealthCount,
			ErrorCount:       app.ErrorCount,
			AlertCount:       app.AlertCount,
			Machines:         machines,
			RollbackPolicy:   convertRollbackPolicy(app.RollbackPolicy),
			REDMetricsConfig: convertREDMetrics(app.REDMetricsConfig),
			CreatedAt:        app.CreatedTime.Unix(),
			UpdatedAt:        app.UpdatedTime.Unix(),
		})
	}

	l.Infof("[GetAppList] Successfully retrieved app list, total: %d, page: %d, pageSize: %d", total, req.Page, req.PageSize)

	return &types.GetAppListResp{
		Apps:     apps,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}
