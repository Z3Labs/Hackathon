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

	// 创建部署对象
	deployment := &model.Deployment{
		Id:              deploymentId,
		AppName:         req.AppName,
		Status:          string(model.DeploymentStatusPending),
		PackageVersion:  req.PackageVersion,
		ConfigPath:      req.ConfigPath,
		GrayStrategy:    req.GrayStrategy,
		StartTime:       0,
		EndTime:         0,
		ReleaseMachines: []model.DeploymentMachine{},
		ReleaseLog:      "",
		CreatedTime:     time.Now().Unix(),
		UpdatedTime:     time.Now().Unix(),
	}

	// 保存到数据库
	err = l.svcCtx.DeploymentModel.Insert(l.ctx, deployment)
	if err != nil {
		l.Errorf("[CreateDeployment] DeploymentModel.Insert error:%v", err)
		return nil, errors.New("创建部署失败")
	}

	l.Infof("[CreateDeployment] Successfully created deployment: %s, ID: %s", req.AppName, deploymentId)

	return &types.CreateDeploymentResp{
		Id: deploymentId,
	}, nil
}
