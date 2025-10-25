// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package deployments

import (
	"context"

	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeployNodeDeploymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 发布指定设备
func NewDeployNodeDeploymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeployNodeDeploymentLogic {
	return &DeployNodeDeploymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeployNodeDeploymentLogic) DeployNodeDeployment(req *types.DeployNodeDeploymentReq) (resp *types.DeployNodeDeploymentResp, err error) {
	// todo: add your logic here and delete this line

	return
}
