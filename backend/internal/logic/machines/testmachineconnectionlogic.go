package machines

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/Z3Labs/Hackathon/backend/internal/model"
	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TestMachineConnectionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTestMachineConnectionLogic(ctx context.Context, svcCtx *svc.ServiceContext) TestMachineConnectionLogic {
	return TestMachineConnectionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TestMachineConnectionLogic) TestMachineConnection(req *types.TestMachineConnectionReq) (resp *types.TestMachineConnectionResp, err error) {
	// 查询机器信息
	machine, err := l.svcCtx.MachineModel.FindById(l.ctx, req.Id)
	if err != nil {
		l.Errorf("[TestMachineConnection] MachineModel.FindById error:%v", err)
		return nil, fmt.Errorf("machine not found")
	}

	// 测试网络连接
	success, message := l.testNetworkConnection(machine.Ip, machine.Port)

	// 更新机器健康状态
	if success {
		// 连接成功，更新为健康状态
		machine.HealthStatus = model.HealthStatusHealthy
		machine.ErrorStatus = model.ErrorStatusNormal
	} else {
		// 连接失败，更新为不健康状态
		machine.HealthStatus = model.HealthStatusUnhealthy
		machine.ErrorStatus = model.ErrorStatusError
	}

	// 保存状态更新
	err = l.svcCtx.MachineModel.Update(l.ctx, machine)
	if err != nil {
		l.Errorf("[TestMachineConnection] MachineModel.Update error:%v", err)
		// 不返回错误，因为连接测试本身可能成功
	}

	l.Infof("[TestMachineConnection] Machine connection test completed:%s, result:%s", req.Id, message)

	return &types.TestMachineConnectionResp{
		Success: success,
		Message: message,
	}, nil
}

// testNetworkConnection 测试网络连接
func (l *TestMachineConnectionLogic) testNetworkConnection(ip string, port int) (bool, string) {
	// 设置连接超时
	timeout := 5 * time.Second

	// 尝试建立TCP连接
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), timeout)
	if err != nil {
		return false, fmt.Sprintf("unable to connect to %s:%d, error: %v", ip, port, err)
	}

	// 关闭连接
	conn.Close()

	return true, fmt.Sprintf("successfully connected to %s:%d", ip, port)
}
