// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package deployments

import (
	"context"

	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RetryNodeDeploymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 重试失败的设备
func NewRetryNodeDeploymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetryNodeDeploymentLogic {
	return &RetryNodeDeploymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RetryNodeDeploymentLogic) RetryNodeDeployment(req *types.RetryNodeDeploymentReq) (resp *types.RetryNodeDeploymentResp, err error) {
	// todo: add your logic here and delete this line

	return
}
