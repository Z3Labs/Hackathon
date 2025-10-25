package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	ReleasePlan struct {
		Id            string       `bson:"_id,omitempty"   json:"id,omitempty"`
		Svc           string       `bson:"svc"             json:"svc"`
		TargetVersion string       `bson:"targetVersion"   json:"target_version"`
		Platform      PlatformType `bson:"platform"        json:"platform"`
		ReleaseTime   time.Time    `bson:"releaseTime"     json:"release_time"`
		Package       PackageInfo  `bson:"package"         json:"package"`
		Stages        []Stage      `bson:"stages"          json:"stages"`
		Status        PlanStatus   `bson:"status"          json:"status"`
		CreatedAt     time.Time    `bson:"createdAt"       json:"created_at"`
		UpdatedAt     time.Time    `bson:"updatedAt"       json:"updated_at"`
	}

	PackageInfo struct {
		URL       string    `bson:"url"       json:"url"`
		SHA256    string    `bson:"sha256"    json:"sha256"`
		Size      int64     `bson:"size"      json:"size"`
		CreatedAt time.Time `bson:"createdAt" json:"created_at"`
	}

	Stage struct {
		Name   string      `bson:"name"   json:"name"`
		Nodes  []StageNode `bson:"nodes"  json:"nodes"`
		Status StageStatus `bson:"status" json:"status"`
		Pacer  PacerConfig `bson:"pacer"  json:"pacer"`
	}

	StageNode struct {
		Host             string     `bson:"host"             json:"host"`
		IP               string     `bson:"ip"               json:"ip"`
		Status           NodeStatus `bson:"status"           json:"status"`
		CurrentVersion   string     `bson:"currentVersion"   json:"current_version"`
		DeployingVersion string     `bson:"deployingVersion" json:"deploying_version"`
		PrevVersion      string     `bson:"prevVersion"      json:"prev_version"`
		LastError        string     `bson:"lastError"        json:"last_error"`
		UpdatedAt        time.Time  `bson:"updatedAt"        json:"updated_at"`
	}

	PacerConfig struct {
		BatchSize       int `bson:"batchSize"       json:"batch_size"`
		IntervalSeconds int `bson:"intervalSeconds" json:"interval_seconds"`
	}

	ReleasePlanModel interface {
		Insert(ctx context.Context, plan *ReleasePlan) error
		Update(ctx context.Context, plan *ReleasePlan) error
		Delete(ctx context.Context, id string) error
		FindById(ctx context.Context, id string) (*ReleasePlan, error)
		Search(ctx context.Context, cond *ReleasePlanCond) ([]*ReleasePlan, error)
		Count(ctx context.Context, cond *ReleasePlanCond) (int64, error)
	}

	defaultReleasePlanModel struct {
		model *mon.Model
	}

	ReleasePlanCond struct {
		Id     string
		Ids    []string
		Svc    string
		Status PlanStatus
	}
)

func NewReleasePlanModel(url, db string) ReleasePlanModel {
	return &defaultReleasePlanModel{
		model: mon.MustNewModel(url, db, CollectionReleasePlan),
	}
}

func (c *ReleasePlanCond) genCond() bson.M {
	filter := bson.M{}

	if c.Id != "" {
		filter["_id"] = c.Id
	} else if len(c.Ids) > 0 {
		filter["_id"] = bson.M{"$in": c.Ids}
	}

	if c.Svc != "" {
		filter["svc"] = c.Svc
	}

	if c.Status != "" {
		filter["status"] = c.Status
	}

	return filter
}

func (m *defaultReleasePlanModel) Insert(ctx context.Context, plan *ReleasePlan) error {
	now := time.Now()
	plan.CreatedAt = now
	plan.UpdatedAt = now
	if plan.Id == "" {
		plan.Id = primitive.NewObjectID().String()
	}
	_, err := m.model.InsertOne(ctx, plan)
	return err
}

func (m *defaultReleasePlanModel) Update(ctx context.Context, plan *ReleasePlan) error {
	plan.UpdatedAt = time.Now()

	_, err := m.model.UpdateOne(
		ctx,
		bson.M{"_id": plan.Id},
		bson.M{"$set": plan},
	)
	return err
}

func (m *defaultReleasePlanModel) Delete(ctx context.Context, id string) error {
	_, err := m.model.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (m *defaultReleasePlanModel) FindById(ctx context.Context, id string) (*ReleasePlan, error) {
	var plan ReleasePlan
	err := m.model.FindOne(ctx, &plan, bson.M{"_id": id})
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (m *defaultReleasePlanModel) Search(ctx context.Context, cond *ReleasePlanCond) ([]*ReleasePlan, error) {
	var result []*ReleasePlan
	filter := cond.genCond()

	err := m.model.Find(ctx, &result, filter)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *defaultReleasePlanModel) Count(ctx context.Context, cond *ReleasePlanCond) (int64, error) {
	count, err := m.model.CountDocuments(ctx, cond.genCond())
	return count, err
}
