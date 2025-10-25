package apps

import (
	"context"
	"fmt"

	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAppVersionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAppVersionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAppVersionsLogic {
	return &GetAppVersionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAppVersionsLogic) GetAppVersions(req *types.GetAppVersionsReq) (resp *types.GetAppVersionsResp, err error) {
	if req.AppName == "" {
		return nil, fmt.Errorf("应用名称不能为空")
	}

	if l.svcCtx.QiniuClient == nil {
		l.Error("[GetAppVersions] 七牛云配置未设置，请检查环境变量 QINIU_ACCESS_KEY, QINIU_SECRET_KEY, QINIU_BUCKET")
		return nil, fmt.Errorf("七牛云配置未设置，请联系管理员配置环境变量")
	}

	versions, err := l.svcCtx.QiniuClient.GetAppVersions(l.ctx, req.AppName)
	if err != nil {
		l.Errorf("[GetAppVersions] QiniuClient.GetAppVersions error: %v", err)
		return nil, fmt.Errorf("获取版本列表失败: %v", err)
	}

	return &types.GetAppVersionsResp{
		Versions: versions,
	}, nil
}
