package machines

import (
	"context"

	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"
	"github.com/Z3Labs/Hackathon/backend/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMachineHostnameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMachineHostnameLogic(ctx context.Context, svcCtx *svc.ServiceContext) GetMachineHostnameLogic {
	return GetMachineHostnameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMachineHostnameLogic) GetMachineHostname(req *types.GetMachineHostnameReq) (resp *types.GetMachineHostnameResp, err error) {
	// 测试SSH连接并获取hostname
	success, hostname, message, err := utils.TestSSHConnectionAndGetHostname(req.Ip, req.Port, req.Username, req.Password)

	if err != nil || !success {
		l.Infof("[GetMachineHostname] Failed to get hostname for %s:%d, message:%s, error:%v", req.Ip, req.Port, message, err)
		return &types.GetMachineHostnameResp{
			Success:  false,
			Hostname: "",
			Message:  message,
		}, nil
	}

	l.Infof("[GetMachineHostname] Successfully got hostname for %s:%d, hostname:%s", req.Ip, req.Port, hostname)

	return &types.GetMachineHostnameResp{
		Success:  true,
		Hostname: hostname,
		Message:  message,
	}, nil
}
