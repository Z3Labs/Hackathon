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

	// 必要入参：
	// RepoAddress 检查的github 仓库地址
	// Annotations["description"] 告警描述信息
	// Labels["hostname"] 指标异常的主机名
	_, err := diagnosis.New(context.Background(), l.svcCtx, l.svcCtx.Config.AI).GenerateReport(req)
	if err != nil {
		l.Errorf("GenerateReport error: %v", err)
		return err
	}

	return nil
}
