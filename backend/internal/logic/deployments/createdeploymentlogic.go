package deployments

import (
	"context"
	"errors"
	"time"

	"github.com/Z3Labs/Hackathon/backend/internal/model"
	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateDeploymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateDeploymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) CreateDeploymentLogic {
	return CreateDeploymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateDeploymentLogic) CreateDeployment(req *types.CreateDeploymentReq) (resp *types.CreateDeploymentResp, err error) {
	// 生成部署ID
	deploymentId := primitive.NewObjectID().Hex()

	// 根据应用名称查找应用信息
	application, err := l.svcCtx.ApplicationModel.Search(l.ctx, &model.ApplicationCond{
		Name: req.AppName,
	})
	if err != nil {
		l.Errorf("[CreateDeployment] ApplicationModel.Search error:%v", err)
		return nil, errors.New("查找应用信息失败")
	}

	if len(application) == 0 {
		l.Errorf("[CreateDeployment] Application not found: %s", req.AppName)
		return nil, errors.New("应用不存在")
	}

	// 从应用信息中提取机器列表并转换为 DeploymentMachine 格式
	var nodeDeployments []model.NodeDeployment
	for _, machine := range application[0].Machines {
		deploymentMachine := model.NodeDeployment{
			Id:               machine.Id,
			Ip:               machine.Ip,
			NodeDeployStatus: model.NodeDeploymentStatusPending,
		}
		nodeDeployments = append(nodeDeployments, deploymentMachine)
	}

	// 创建部署对象
	deployment := &model.Deployment{
		Id:              deploymentId,
		AppName:         req.AppName,
		Status:          model.DeploymentStatusPending,
		PackageVersion:  req.PackageVersion,
		ConfigPath:      req.ConfigPath,
		GrayStrategy:    req.GrayStrategy,
		NodeDeployments: nodeDeployments,
		CreatedTime:     time.Now().Unix(),
		UpdatedTime:     time.Now().Unix(),
	}

	// 保存到数据库
	err = l.svcCtx.DeploymentModel.Insert(l.ctx, deployment)
	if err != nil {
		l.Errorf("[CreateDeployment] DeploymentModel.Insert error:%v", err)
		return nil, errors.New("创建部署失败")
	}

	l.Infof("[CreateDeployment] Successfully created deployment: %s, ID: %s, machines count: %d", req.AppName, deploymentId, len(nodeDeployments))

	return &types.CreateDeploymentResp{
		Id: deploymentId,
	}, nil
}
