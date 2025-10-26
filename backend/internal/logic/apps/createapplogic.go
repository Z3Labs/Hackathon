package apps

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

type CreateAppLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateAppLogic(ctx context.Context, svcCtx *svc.ServiceContext) CreateAppLogic {
	return CreateAppLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateAppLogic) CreateApp(req *types.CreateAppReq) (resp *types.CreateAppResp, err error) {
	// 生成应用ID
	appId := primitive.NewObjectID().Hex()

	// 创建应用对象
	application := &model.Application{
		Id:             appId,
		Name:           req.Name,
		Repo:           req.Repo,
		DeployPath:     req.DeployPath,
		ConfigPath:     req.ConfigPath,
		StartCmd:       req.StartCmd,
		StopCmd:        req.StopCmd,
		CurrentVersion: "--",
		CreatedTime:    time.Now(),
		UpdatedTime:    time.Now(),
	}

	// 保存到数据库
	err = l.svcCtx.ApplicationModel.Insert(l.ctx, application)
	if err != nil {
		l.Errorf("[CreateApp] ApplicationModel.Insert error:%v", err)
		return nil, errors.New("创建应用失败")
	}

	l.Infof("[CreateApp] Successfully created app: %s, ID: %s", req.Name, appId)

	return &types.CreateAppResp{
		Id: appId,
	}, nil
}
