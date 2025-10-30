package svc

import (
	"context"
	"log"
	"time"

	"github.com/Z3Labs/Hackathon/backend/common/qiniu"
	"github.com/Z3Labs/Hackathon/backend/internal/config"
	"github.com/Z3Labs/Hackathon/backend/internal/model"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ServiceContext struct {
	Config           config.Config
	ApplicationModel model.ApplicationModel
	DeploymentModel  model.DeploymentModel
	MachineModel     model.MachineModel
	ReportModel      model.ReportModel
	QiniuClient      *qiniu.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	var qiniuClient *qiniu.Client
	if c.Qiniu.AccessKey != "" && c.Qiniu.SecretKey != "" && c.Qiniu.Bucket != "" {
		qiniuClient = qiniu.NewClient(c.Qiniu.AccessKey, c.Qiniu.SecretKey, c.Qiniu.Bucket, c.Qiniu.DownloadHost)
	}

	return &ServiceContext{
		Config:           c,
		ApplicationModel: model.NewApplicationModel(c.Mongo.URL, c.Mongo.Database),
		DeploymentModel:  model.NewDeploymentModel(c.Mongo.URL, c.Mongo.Database),
		MachineModel:     model.NewMachineModel(c.Mongo.URL, c.Mongo.Database),
		ReportModel:      model.NewReportModel(c.Mongo.URL, c.Mongo.Database),
		QiniuClient:      qiniuClient,
	}
}
func NewUTServiceContext(c config.Config) *ServiceContext {
	var qiniuClient *qiniu.Client
	if c.Qiniu.AccessKey != "" && c.Qiniu.SecretKey != "" && c.Qiniu.Bucket != "" {
		qiniuClient = qiniu.NewClient(c.Qiniu.AccessKey, c.Qiniu.SecretKey, c.Qiniu.Bucket, c.Qiniu.DownloadHost)
	}

	svc := &ServiceContext{
		Config:           c,
		ApplicationModel: model.NewApplicationModel(c.Mongo.URL, c.Mongo.Database),
		DeploymentModel:  model.NewDeploymentModel(c.Mongo.URL, c.Mongo.Database),
		MachineModel:     model.NewMachineModel(c.Mongo.URL, c.Mongo.Database),
		ReportModel:      model.NewReportModel(c.Mongo.URL, c.Mongo.Database),
		QiniuClient:      qiniuClient,
	}

	// 清空测试数据库中的所有集合
	cleanTestDatabase(c.Mongo.URL, c.Mongo.Database)

	return svc
}

// cleanTestDatabase 清空测试数据库中的所有数据
func cleanTestDatabase(url, database string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		log.Printf("连接 MongoDB 失败: %v", err)
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database(database)

	// 清空所有集合
	collections := []string{
		model.CollectionApplication,
		model.CollectionDeployment,
		model.CollectionMachine,
		model.CollectionReport,
	}

	for _, collection := range collections {
		if err := db.Collection(collection).Drop(ctx); err != nil {
			log.Printf("清空集合 %s 失败: %v", collection, err)
		} else {
			log.Printf("清空集合 %s 成功", collection)
		}
	}
}
