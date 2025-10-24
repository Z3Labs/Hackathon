package apps

import (
	"context"
	"errors"
	"time"

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
	existingApp.DeployPath = req.DeployPath
	existingApp.StartCmd = req.StartCmd
	existingApp.StopCmd = req.StopCmd
	existingApp.Version = req.Version
	existingApp.UpdatedTime = time.Now()

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
