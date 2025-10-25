package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
)

type (
	Application struct {
		Id             string     `bson:"_id"            json:"id,omitempty"`     // mongo id
		Name           string     `bson:"name"           json:"name"`             // 应用名称
		DeployPath     string     `bson:"deployPath"     json:"deploy_path"`      // 部署路径
		StartCmd       string     `bson:"startCmd"       json:"start_cmd"`        // 启动命令
		StopCmd        string     `bson:"stopCmd"        json:"stop_cmd"`         // 停止命令
		CurrentVersion string     `bson:"currentVersion" json:"currentVersion"`   // 当前版本
		MachineCount   int        `bson:"machineCount"   json:"machine_count"`    // 机器总数量
		HealthCount    int        `bson:"healthCount"    json:"health_count"`     // 健康机器数量
		ErrorCount     int        `bson:"errorCount"     json:"error_count"`      // 异常机器数量
		AlertCount     int        `bson:"alertCount"     json:"alert_count"`      // 告警机器数量
		Machines       []Machine  `bson:"machines"       json:"machines"`         // 机器列表
		RollBackPolicy string     `bson:"rollBackPolicy" json:"roll_back_policy"` // 回滚策略(prom alert语句)
		REDMetrics     REDMetrics `bson:"redMetrics"    json:"red_metrics"`       // RED指标配置
		CreatedTime    time.Time  `bson:"createdTime"    json:"createdTime"`      // 创建时间
		UpdatedTime    time.Time  `bson:"updatedTime"    json:"updatedTime"`      // 更新时间
	}

	REDMetrics struct {
		RateQuery     string `bson:"rateQuery"     json:"rate_query"`     // 请求速率查询语句
		ErrorQuery    string `bson:"errorQuery"    json:"error_query"`    // 错误率查询语句
		DurationQuery string `bson:"durationQuery" json:"duration_query"` // 延迟查询语句
	}

	ApplicationModel interface {
		Insert(ctx context.Context, application *Application) error
		Update(ctx context.Context, application *Application) error
		Delete(ctx context.Context, id string) error
		FindById(ctx context.Context, id string) (*Application, error)
		Search(ctx context.Context, cond *ApplicationCond) ([]*Application, error)
		Count(ctx context.Context, cond *ApplicationCond) (int64, error)
	}

	defaultApplicationModel struct {
		model *mon.Model
	}

	ApplicationCond struct {
		Id         string
		Ids        []string
		Name       string
		Status     string
		Pagination *Pagination
	}
)

func NewApplicationModel(url, db string) ApplicationModel {
	return &defaultApplicationModel{
		model: mon.MustNewModel(url, db, CollectionApplication),
	}
}

func (c *ApplicationCond) genCond() bson.M {
	filter := bson.M{}

	if c.Id != "" {
		filter["_id"] = c.Id
	} else if len(c.Ids) > 0 {
		filter["_id"] = bson.M{"$in": c.Ids}
	}

	if c.Name != "" {
		filter["name"] = bson.M{"$regex": c.Name, "$options": "i"}
	}

	return filter
}

func (m *defaultApplicationModel) Insert(ctx context.Context, application *Application) error {
	application.CreatedTime = time.Now()
	application.UpdatedTime = time.Now()

	_, err := m.model.InsertOne(ctx, application)
	return err
}

func (m *defaultApplicationModel) Update(ctx context.Context, application *Application) error {
	application.UpdatedTime = time.Now()

	_, err := m.model.UpdateOne(
		ctx,
		bson.M{"_id": application.Id},
		bson.M{"$set": application},
	)
	return err
}

func (m *defaultApplicationModel) Delete(ctx context.Context, id string) error {
	_, err := m.model.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (m *defaultApplicationModel) FindById(ctx context.Context, id string) (*Application, error) {
	var application Application
	err := m.model.FindOne(ctx, &application, bson.M{"_id": id})
	if err != nil {
		return nil, err
	}
	return &application, nil
}

func (m *defaultApplicationModel) Search(ctx context.Context, cond *ApplicationCond) ([]*Application, error) {
	var result []*Application
	filter := cond.genCond()

	// 根据是否有分页参数决定查询方式
	var err error
	if cond.Pagination.IsEmpty() {
		err = m.model.Find(ctx, &result, filter)
	} else {
		findOptions := cond.Pagination.ToFindOptions()
		err = m.model.Find(ctx, &result, filter, findOptions)
	}

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *defaultApplicationModel) Count(ctx context.Context, cond *ApplicationCond) (int64, error) {
	count, err := m.model.CountDocuments(ctx, cond.genCond())
	return count, err
}
