package deployments

import (
	"context"
	"errors"
	"time"

	"github.com/Z3Labs/Hackathon/backend/common/qiniu"
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
		GrayMachineId:   req.GrayMachineId,
		NodeDeployments: nodeDeployments,
		CreatedTime:     time.Now().Unix(),
		UpdatedTime:     time.Now().Unix(),
	}
	pkg, err := pkgInfo(l.svcCtx.QiniuClient, req.AppName, req.PackageVersion)
	if err != nil {
		l.Errorf("[CreateDeployment] pkgInfo error:%v", err)
		return nil, errors.New("获取包信息失败")
	}
	deployment.Package = pkg
	// 保存到数据库
	err = l.svcCtx.DeploymentModel.Insert(l.ctx, deployment)
	if err != nil {
		l.Errorf("[CreateDeployment] DeploymentModel.Insert error:%v", err)
		return nil, errors.New("创建部署失败")
	}

	l.Infof("[CreateDeployment] Successfully created deployment: %s, ID: %s, machines count: %d", req.AppName, deploymentId, len(nodeDeployments))

	// 如果指定了灰度设备，立即发布到该设备
	if req.GrayMachineId != "" {
		// 验证灰度设备是否存在于 NodeDeployments 中
		grayMachineFound := false
		for i := range deployment.NodeDeployments {
			if deployment.NodeDeployments[i].Id == req.GrayMachineId {
				grayMachineFound = true
				// 设置灰度设备为发布中状态
				deployment.NodeDeployments[i].NodeDeployStatus = model.NodeDeploymentStatusDeploying
				break
			}
		}

		if grayMachineFound {
			// 更新发布单状态为发布中
			deployment.Status = model.DeploymentStatusDeploying
			deployment.UpdatedTime = time.Now().Unix()

			err = l.svcCtx.DeploymentModel.Update(l.ctx, deployment)
			if err != nil {
				l.Errorf("[CreateDeployment] Failed to update deployment for gray release: %v", err)
				// 不影响创建结果，继续返回
			} else {
				l.Infof("[CreateDeployment] Gray machine deployment started for machine: %s", req.GrayMachineId)
			}
		} else {
			l.Infof("[CreateDeployment] Gray machine ID %s not found in deployment machines", req.GrayMachineId)
		}
	}

	return &types.CreateDeploymentResp{
		Id: deploymentId,
	}, nil
}

func pkgInfo(kodo *qiniu.Client, app, version string) (model.PackageInfo, error) {
	// 先获取应用的所有版本
	versions, err := kodo.GetAppVersions(context.Background(), app)
	if err != nil {
		return model.PackageInfo{}, err
	}

	// 查找指定版本的文件
	var targetFile string
	for _, v := range versions {
		if v.Version == version {
			targetFile = app + "/" + v.FileName
			break
		}
	}

	if targetFile == "" {
		return model.PackageInfo{}, errors.New("版本文件不存在")
	}

	fileInfo, err := kodo.GetFileStat(context.Background(), targetFile)
	if err != nil {
		return model.PackageInfo{}, err
	}
	return model.PackageInfo{
		URL:       kodo.GetFileURL(context.Background(), targetFile, time.Now().Add(time.Hour*24*7).Unix()),
		MD5:       fileInfo.Md5,
		CreatedAt: time.Unix(fileInfo.PutTime, 0),
	}, nil
}
