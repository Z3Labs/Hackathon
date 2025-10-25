// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package deployments

import (
	"context"

	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelNodeDeploymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 取消发布中的设备
func NewCancelNodeDeploymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelNodeDeploymentLogic {
	return &CancelNodeDeploymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CancelNodeDeploymentLogic) CancelNodeDeployment(req *types.CancelNodeDeploymentReq) (resp *types.CancelNodeDeploymentResp, err error) {
	// todo: add your logic here and delete this line

	return
}
