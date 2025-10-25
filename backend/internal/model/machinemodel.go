package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
)

type (
	Machine struct {
		Id           string       `bson:"_id"          json:"id,omitempty"`  // mongo id
		Name         string       `bson:"name"         json:"name"`          // 机器名称
		Ip           string       `bson:"ip"           json:"ip"`            // IP地址
		Port         int          `bson:"port"         json:"port"`          // 端口号
		Username     string       `bson:"username"     json:"username"`      // SSH用户名
		Password     string       `bson:"password"     json:"password"`      // SSH密码
		Description  string       `bson:"description"  json:"description"`   // 机器描述
		HealthStatus HealthStatus `bson:"healthStatus" json:"health_status"` // 健康状态
		ErrorStatus  ErrorStatus  `bson:"errorStatus"  json:"error_status"`  // 异常状态
		AlertStatus  AlertStatus  `bson:"alertStatus"  json:"alert_status"`  // 告警状态
		CreatedTime  time.Time    `bson:"createdTime"  json:"createdTime"`   // 创建时间戳
		UpdatedTime  time.Time    `bson:"updatedTime"  json:"updatedTime"`   // 更新时间戳
	}

	MachineModel interface {
		Insert(ctx context.Context, machine *Machine) error
		Update(ctx context.Context, machine *Machine) error
		Delete(ctx context.Context, id string) error
		FindById(ctx context.Context, id string) (*Machine, error)
		Search(ctx context.Context, cond *MachineCond) ([]*Machine, error)
		Count(ctx context.Context, cond *MachineCond) (int64, error)
	}

	defaultMachineModel struct {
		model *mon.Model
	}

	MachineCond struct {
		Id           string
		Ids          []string
		Name         string
		Ip           string
		HealthStatus string
		ErrorStatus  string
		AlertStatus  string
	}
)

func NewMachineModel(url, db string) MachineModel {
	return &defaultMachineModel{
		model: mon.MustNewModel(url, db, CollectionMachine),
	}
}

func (c *MachineCond) genCond() bson.M {
	filter := bson.M{}

	if c.Id != "" {
		filter["_id"] = c.Id
	} else if len(c.Ids) > 0 {
		filter["_id"] = bson.M{"$in": c.Ids}
	}

	if c.Name != "" {
		filter["name"] = bson.M{"$regex": c.Name, "$options": "i"} // 支持模糊查询
	}

	if c.Ip != "" {
		filter["ip"] = c.Ip
	}

	if c.HealthStatus != "" {
		filter["healthStatus"] = c.HealthStatus
	}

	if c.ErrorStatus != "" {
		filter["errorStatus"] = c.ErrorStatus
	}

	if c.AlertStatus != "" {
		filter["alertStatus"] = c.AlertStatus
	}

	return filter
}

func (m *defaultMachineModel) Insert(ctx context.Context, machine *Machine) error {
	machine.CreatedTime = time.Now()
	machine.UpdatedTime = time.Now()

	_, err := m.model.InsertOne(ctx, machine)
	return err
}

func (m *defaultMachineModel) Update(ctx context.Context, machine *Machine) error {
	machine.UpdatedTime = time.Now()

	_, err := m.model.UpdateOne(
		ctx,
		bson.M{"_id": machine.Id},
		bson.M{"$set": machine},
	)
	return err
}

func (m *defaultMachineModel) Delete(ctx context.Context, id string) error {
	_, err := m.model.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (m *defaultMachineModel) FindById(ctx context.Context, id string) (*Machine, error) {
	var machine Machine
	err := m.model.FindOne(ctx, &machine, bson.M{"_id": id})
	if err != nil {
		return nil, err
	}
	return &machine, nil
}

func (m *defaultMachineModel) Search(ctx context.Context, cond *MachineCond) ([]*Machine, error) {
	var result []*Machine
	filter := cond.genCond()

	err := m.model.Find(ctx, &result, filter)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *defaultMachineModel) Count(ctx context.Context, cond *MachineCond) (int64, error) {
	count, err := m.model.CountDocuments(ctx, cond.genCond())
	return count, err
}
