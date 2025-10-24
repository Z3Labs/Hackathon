package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
)

type (
	Deployment struct {
		Id          string    `bson:"_id,omitempty" json:"id,omitempty"`
		// TOTO: Add your fields here
		CreatedTime time.Time `bson:"createdTime"   json:"createdTime"`
		UpdatedTime time.Time `bson:"updatedTime"   json:"updatedTime"`
	}

	DeploymentModel interface {
		Insert(ctx context.Context, deployment *Deployment) error
		Update(ctx context.Context, deployment *Deployment) error
		Delete(ctx context.Context, id string) error
		FindById(ctx context.Context, id string) (*Deployment, error)
		Search(ctx context.Context, cond *DeploymentCond) ([]*Deployment, error)
		Count(ctx context.Context, cond *DeploymentCond) (int64, error)
	}

	defaultDeploymentModel struct {
		model *mon.Model
	}

	DeploymentCond struct {
		Id  string
		Ids []string
	}
)

func NewDeploymentModel(url, db string) DeploymentModel {
	return &defaultDeploymentModel{
		model: mon.MustNewModel(url, db, "Deployment"),
	}
}

func (c *DeploymentCond) genCond() bson.M {
	filter := bson.M{}

	if c.Id != "" {
		filter["_id"] = c.Id
	} else if len(c.Ids) > 0 {
		filter["_id"] = bson.M{"$in": c.Ids}
	}

	return filter
}

func (m *defaultDeploymentModel) Insert(ctx context.Context, deployment *Deployment) error {
	deployment.CreatedTime = time.Now()
	deployment.UpdatedTime = time.Now()
	
	_, err := m.model.InsertOne(ctx, deployment)
	return err
}

func (m *defaultDeploymentModel) Update(ctx context.Context, deployment *Deployment) error {
	deployment.UpdatedTime = time.Now()
	
	_, err := m.model.UpdateOne(
		ctx,
		bson.M{"_id": deployment.Id},
		bson.M{"$set": deployment},
	)
	return err
}

func (m *defaultDeploymentModel) Delete(ctx context.Context, id string) error {
	_, err := m.model.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (m *defaultDeploymentModel) FindById(ctx context.Context, id string) (*Deployment, error) {
	var deployment Deployment
	err := m.model.FindOne(ctx, &deployment, bson.M{"_id": id})
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

func (m *defaultDeploymentModel) Search(ctx context.Context, cond *DeploymentCond) ([]*Deployment, error) {
	var result []*Deployment
	filter := cond.genCond()

	err := m.model.Find(ctx, &result, filter)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *defaultDeploymentModel) Count(ctx context.Context, cond *DeploymentCond) (int64, error) {
	count, err := m.model.CountDocuments(ctx, cond.genCond())
	return count, err
} 