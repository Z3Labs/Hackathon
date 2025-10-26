package alert

import (
	"context"

	"github.com/Z3Labs/Hackathon/backend/internal/clients/diagnosis"
	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AlertCallBackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAlertCallBackLogic(ctx context.Context, svcCtx *svc.ServiceContext) AlertCallBackLogic {
	return AlertCallBackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AlertCallBackLogic) AlertCallBack(req *types.PostAlertCallbackReq) error {

	// TODO 其它逻辑
	l.Infof("AlertCallBack req: %#v", req)
	return nil
	// 必要入参：
	// Annotations["description"] 告警描述信息
	// Labels["hostname"] 指标异常的主机名
	// Labels["deploymentId"] 发布任务ID
	//
	// 以下参数从deploymentId查询：
	// RepoAddress 检查的github 仓库地址
	// Tag 发布Tag
	deployId := req.Labels["deploymentId"]
	if deployId == "" {
		l.Errorf("AlertCallBack missing deploymentId")
		return nil
	}
	deploy, err := l.svcCtx.DeploymentModel.FindById(l.ctx, deployId)
	if err != nil {
		l.Errorf("DeploymentModel.FindById error: %v", err)
		return err
	}
	req.Tag = deploy.PackageVersion
	application, err := l.svcCtx.ApplicationModel.FindById(l.ctx, deploy.AppId)
	if err != nil {
		l.Errorf("ApplicationModel.FindById error: %v", err)
		return err
	}
	req.RepoAddress = application.Repo
	// 耗时操作，约需要30～60s
	_, err = diagnosis.New(context.Background(), l.svcCtx, l.svcCtx.Config.AI).GenerateReport(req)
	if err != nil {
		l.Errorf("GenerateReport error: %v", err)
		return err
	}

	return nil
}
