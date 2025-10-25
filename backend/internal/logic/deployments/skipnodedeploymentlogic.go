// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package deployments

import (
	"context"

	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SkipNodeDeploymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 跳过指定设备
func NewSkipNodeDeploymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SkipNodeDeploymentLogic {
	return &SkipNodeDeploymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SkipNodeDeploymentLogic) SkipNodeDeployment(req *types.SkipNodeDeploymentReq) (resp *types.SkipNodeDeploymentResp, err error) {
	// todo: add your logic here and delete this line

	return
}
