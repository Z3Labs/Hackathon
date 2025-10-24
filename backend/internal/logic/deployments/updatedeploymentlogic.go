package deployments

import (
	"context"
	"errors"
	"time"

	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateDeploymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateDeploymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) UpdateDeploymentLogic {
	return UpdateDeploymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateDeploymentLogic) UpdateDeployment(req *types.UpdateDeploymentReq) (resp *types.UpdateDeploymentResp, err error) {
	// 检查部署是否存在
	existingDeployment, err := l.svcCtx.DeploymentModel.FindById(l.ctx, req.Id)
	if err != nil {
		l.Errorf("[UpdateDeployment] DeploymentModel.FindById error:%v", err)
		return nil, errors.New("部署不存在")
	}

	// 更新部署信息
	existingDeployment.AppName = req.AppName
	existingDeployment.PackageVersion = req.PackageVersion
	existingDeployment.ConfigPath = req.ConfigPath
	existingDeployment.GrayStrategy = req.GrayStrategy
	existingDeployment.UpdatedTime = time.Now().Unix()

	// 保存到数据库
	err = l.svcCtx.DeploymentModel.Update(l.ctx, existingDeployment)
	if err != nil {
		l.Errorf("[UpdateDeployment] DeploymentModel.Update error:%v", err)
		return nil, errors.New("更新部署失败")
	}

	l.Infof("[UpdateDeployment] Successfully updated deployment: %s, ID: %s", req.AppName, req.Id)

	return &types.UpdateDeploymentResp{
		Success: true,
	}, nil
}
