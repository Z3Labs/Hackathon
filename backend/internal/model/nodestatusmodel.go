package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
)

type (
	NodeStatusRecord struct {
		Id               string       `bson:"_id,omitempty"    json:"id,omitempty"`
		Host             string       `bson:"host"             json:"host"`
		Service          string       `bson:"service"          json:"service"`
		CurrentVersion   string       `bson:"currentVersion"   json:"current_version"`
		DeployingVersion string       `bson:"deployingVersion" json:"deploying_version"`
		PrevVersion      string       `bson:"prevVersion"      json:"prev_version"`
		Platform         PlatformType `bson:"platform"         json:"platform"`
		State            NodeStatus   `bson:"state"            json:"state"`
		LastError        string       `bson:"lastError"        json:"last_error"`
		UpdatedAt        time.Time    `bson:"updatedAt"        json:"updated_at"`
		CreatedAt        time.Time    `bson:"createdAt"        json:"created_at"`
	}

	NodeStatusModel interface {
		Insert(ctx context.Context, nodeStatus *NodeStatusRecord) error
		Update(ctx context.Context, nodeStatus *NodeStatusRecord) error
		Delete(ctx context.Context, id string) error
		FindById(ctx context.Context, id string) (*NodeStatusRecord, error)
		FindByHostAndService(ctx context.Context, host, service string) (*NodeStatusRecord, error)
		Search(ctx context.Context, cond *NodeStatusCond) ([]*NodeStatusRecord, error)
		Count(ctx context.Context, cond *NodeStatusCond) (int64, error)
	}

	defaultNodeStatusModel struct {
		model *mon.Model
	}

	NodeStatusCond struct {
		Id      string
		Ids     []string
		Host    string
		Service string
		State   NodeStatus
	}
)

func NewNodeStatusModel(url, db string) NodeStatusModel {
	return &defaultNodeStatusModel{
		model: mon.MustNewModel(url, db, CollectionNodeStatus),
	}
}

func (c *NodeStatusCond) genCond() bson.M {
	filter := bson.M{}

	if c.Id != "" {
		filter["_id"] = c.Id
	} else if len(c.Ids) > 0 {
		filter["_id"] = bson.M{"$in": c.Ids}
	}

	if c.Host != "" {
		filter["host"] = c.Host
	}

	if c.Service != "" {
		filter["service"] = c.Service
	}

	if c.State != "" {
		filter["state"] = c.State
	}

	return filter
}

func (m *defaultNodeStatusModel) Insert(ctx context.Context, nodeStatus *NodeStatusRecord) error {
	now := time.Now()
	nodeStatus.CreatedAt = now
	nodeStatus.UpdatedAt = now

	_, err := m.model.InsertOne(ctx, nodeStatus)
	return err
}

func (m *defaultNodeStatusModel) Update(ctx context.Context, nodeStatus *NodeStatusRecord) error {
	nodeStatus.UpdatedAt = time.Now()

	_, err := m.model.UpdateOne(
		ctx,
		bson.M{"_id": nodeStatus.Id},
		bson.M{"$set": nodeStatus},
	)
	return err
}

func (m *defaultNodeStatusModel) Delete(ctx context.Context, id string) error {
	_, err := m.model.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (m *defaultNodeStatusModel) FindById(ctx context.Context, id string) (*NodeStatusRecord, error) {
	var nodeStatus NodeStatusRecord
	err := m.model.FindOne(ctx, &nodeStatus, bson.M{"_id": id})
	if err != nil {
		return nil, err
	}
	return &nodeStatus, nil
}

func (m *defaultNodeStatusModel) FindByHostAndService(ctx context.Context, host, service string) (*NodeStatusRecord, error) {
	var nodeStatus NodeStatusRecord
	err := m.model.FindOne(ctx, &nodeStatus, bson.M{
		"host":    host,
		"service": service,
	})
	if err != nil {
		return nil, err
	}
	return &nodeStatus, nil
}

func (m *defaultNodeStatusModel) Search(ctx context.Context, cond *NodeStatusCond) ([]*NodeStatusRecord, error) {
	var result []*NodeStatusRecord
	filter := cond.genCond()

	err := m.model.Find(ctx, &result, filter)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *defaultNodeStatusModel) Count(ctx context.Context, cond *NodeStatusCond) (int64, error) {
	count, err := m.model.CountDocuments(ctx, cond.genCond())
	return count, err
}
