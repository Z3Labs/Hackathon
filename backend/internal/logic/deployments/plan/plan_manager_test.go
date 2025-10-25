package plan

import (
	"context"
	"testing"
	"time"

	"github.com/Z3Labs/Hackathon/backend/internal/config"
	"github.com/Z3Labs/Hackathon/backend/internal/model"
	"github.com/Z3Labs/Hackathon/backend/internal/svc"
)

const (
	testMongoURL = "mongodb://127.0.0.1:27017"
	testDB       = "hackathon_test"
)

func setupTestDB(t *testing.T) *svc.ServiceContext {
	cfg := config.Config{
		Mongo: config.MongoDBConfig{
			URL:      testMongoURL,
			Database: testDB,
		},
	}
	return svc.NewServiceContext(cfg)
}

func cleanupTestData(t *testing.T, ctx context.Context, releasePlanModel model.ReleasePlanModel, planID string) {
	if planID != "" {
		releasePlanModel.Delete(ctx, planID)
	}
}

func TestPlanManager_CreateReleasePlan(t *testing.T) {
	ctx := context.Background()
	svcCtx := setupTestDB(t)

	pm := NewPlanManager(ctx, svcCtx)

	pkg := model.PackageInfo{
		URL:       "http://example.com/package.tar.gz",
		SHA256:    "abc123",
		Size:      1024,
		CreatedAt: time.Now(),
	}

	stages := []model.Stage{
		{
			Name: "stage1",
			Nodes: []model.StageNode{
				{Host: "host1", Status: model.NodeStatusPending},
			},
			Pacer: model.PacerConfig{BatchSize: 1, IntervalSeconds: 0},
		},
	}

	plan, err := pm.CreateReleasePlan(ctx, "test-service", "v1.0.0", pkg, stages)
	if err != nil {
		t.Fatalf("CreateReleasePlan failed: %v", err)
	}
	defer cleanupTestData(t, ctx, svcCtx.ReleasePlanModel, plan.Id)

	if plan.Svc != "test-service" {
		t.Errorf("expected svc=test-service, got %s", plan.Svc)
	}

	if plan.TargetVersion != "v1.0.0" {
		t.Errorf("expected version=v1.0.0, got %s", plan.TargetVersion)
	}

	if plan.Status != model.PlanStatusPending {
		t.Errorf("expected status=pending, got %s", plan.Status)
	}
}

func TestPlanManager_ExecutePlan_Success(t *testing.T) {
	ctx := context.Background()
	svcCtx := setupTestDB(t)

	pm := NewPlanManager(ctx, svcCtx)

	pkg := model.PackageInfo{
		URL:       "http://example.com/package.tar.gz",
		SHA256:    "abc123",
		Size:      1024,
		CreatedAt: time.Now(),
	}

	stages := []model.Stage{
		{
			Name: "stage1",
			Nodes: []model.StageNode{
				{Host: "host1", Status: model.NodeStatusPending, CurrentVersion: "v0.9.0"},
			},
			Pacer: model.PacerConfig{BatchSize: 1, IntervalSeconds: 0},
		},
	}

	plan, err := pm.CreateReleasePlan(ctx, "test-service", "v1.0.0", pkg, stages)
	if err != nil {
		t.Fatalf("CreateReleasePlan failed: %v", err)
	}
	defer cleanupTestData(t, ctx, svcCtx.ReleasePlanModel, plan.Id)

	err = pm.ExecutePlan(ctx, plan.Id)
	if err != nil {
		t.Fatalf("ExecutePlan failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	updatedPlan, err := pm.GetPlanStatus(ctx, plan.Id)
	if err != nil {
		t.Fatalf("GetPlanStatus failed: %v", err)
	}

	if updatedPlan.Status != model.PlanStatusSuccess {
		t.Errorf("expected status=success, got %s", updatedPlan.Status)
	}
}

func TestPlanManager_ExecutePlan_WithMultipleStages(t *testing.T) {
	ctx := context.Background()
	svcCtx := setupTestDB(t)

	pm := NewPlanManager(ctx, svcCtx)

	pkg := model.PackageInfo{
		URL:       "http://example.com/package.tar.gz",
		SHA256:    "abc123",
		Size:      1024,
		CreatedAt: time.Now(),
	}

	stages := []model.Stage{
		{
			Name: "stage1",
			Nodes: []model.StageNode{
				{Host: "host1", Status: model.NodeStatusPending, CurrentVersion: "v0.9.0"},
				{Host: "host2", Status: model.NodeStatusPending, CurrentVersion: "v0.9.0"},
			},
			Pacer: model.PacerConfig{BatchSize: 2, IntervalSeconds: 0},
		},
		{
			Name: "stage2",
			Nodes: []model.StageNode{
				{Host: "host3", Status: model.NodeStatusPending, CurrentVersion: "v0.9.0"},
			},
			Pacer: model.PacerConfig{BatchSize: 1, IntervalSeconds: 0},
		},
	}

	plan, err := pm.CreateReleasePlan(ctx, "test-service", "v1.0.0", pkg, stages)
	if err != nil {
		t.Fatalf("CreateReleasePlan failed: %v", err)
	}
	defer cleanupTestData(t, ctx, svcCtx.ReleasePlanModel, plan.Id)

	err = pm.ExecutePlan(ctx, plan.Id)
	if err != nil {
		t.Fatalf("ExecutePlan failed: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	updatedPlan, err := pm.GetPlanStatus(ctx, plan.Id)
	if err != nil {
		t.Fatalf("GetPlanStatus failed: %v", err)
	}

	if updatedPlan.Status != model.PlanStatusSuccess {
		t.Errorf("expected status=success, got %s", updatedPlan.Status)
	}

	for i, stage := range updatedPlan.Stages {
		if stage.Status != model.StageStatusSuccess {
			t.Errorf("stage %d: expected status=success, got %s", i, stage.Status)
		}
	}
}

func TestPlanManager_ExecutePlan_WithBatching(t *testing.T) {
	ctx := context.Background()
	svcCtx := setupTestDB(t)

	pm := NewPlanManager(ctx, svcCtx)

	pkg := model.PackageInfo{
		URL:       "http://example.com/package.tar.gz",
		SHA256:    "abc123",
		Size:      1024,
		CreatedAt: time.Now(),
	}

	stages := []model.Stage{
		{
			Name: "stage1",
			Nodes: []model.StageNode{
				{Host: "host1", Status: model.NodeStatusPending, CurrentVersion: "v0.9.0"},
				{Host: "host2", Status: model.NodeStatusPending, CurrentVersion: "v0.9.0"},
				{Host: "host3", Status: model.NodeStatusPending, CurrentVersion: "v0.9.0"},
			},
			Pacer: model.PacerConfig{BatchSize: 2, IntervalSeconds: 0},
		},
	}

	plan, err := pm.CreateReleasePlan(ctx, "test-service", "v1.0.0", pkg, stages)
	if err != nil {
		t.Fatalf("CreateReleasePlan failed: %v", err)
	}
	defer cleanupTestData(t, ctx, svcCtx.ReleasePlanModel, plan.Id)

	err = pm.ExecutePlan(ctx, plan.Id)
	if err != nil {
		t.Fatalf("ExecutePlan failed: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	updatedPlan, err := pm.GetPlanStatus(ctx, plan.Id)
	if err != nil {
		t.Fatalf("GetPlanStatus failed: %v", err)
	}

	if updatedPlan.Status != model.PlanStatusSuccess {
		t.Errorf("expected status=success, got %s", updatedPlan.Status)
	}
}

func TestPlanManager_ExecutePlan_InvalidStatus(t *testing.T) {
	ctx := context.Background()
	svcCtx := setupTestDB(t)

	pm := NewPlanManager(ctx, svcCtx)

	pkg := model.PackageInfo{
		URL:       "http://example.com/package.tar.gz",
		SHA256:    "abc123",
		Size:      1024,
		CreatedAt: time.Now(),
	}

	stages := []model.Stage{
		{
			Name: "stage1",
			Nodes: []model.StageNode{
				{Host: "host1", Status: model.NodeStatusPending},
			},
			Pacer: model.PacerConfig{BatchSize: 1, IntervalSeconds: 0},
		},
	}

	plan, err := pm.CreateReleasePlan(ctx, "test-service", "v1.0.0", pkg, stages)
	if err != nil {
		t.Fatalf("CreateReleasePlan failed: %v", err)
	}
	defer cleanupTestData(t, ctx, svcCtx.ReleasePlanModel, plan.Id)

	plan.Status = model.PlanStatusSuccess
	svcCtx.ReleasePlanModel.Update(ctx, plan)

	err = pm.ExecutePlan(ctx, plan.Id)
	if err == nil {
		t.Fatal("expected error when executing plan with invalid status")
	}
}

func TestPlanManager_CancelPlan(t *testing.T) {
	ctx := context.Background()
	svcCtx := setupTestDB(t)

	pm := NewPlanManager(ctx, svcCtx)

	pkg := model.PackageInfo{
		URL:       "http://example.com/package.tar.gz",
		SHA256:    "abc123",
		Size:      1024,
		CreatedAt: time.Now(),
	}

	stages := []model.Stage{
		{
			Name: "stage1",
			Nodes: []model.StageNode{
				{Host: "host1", Status: model.NodeStatusPending},
			},
			Pacer: model.PacerConfig{BatchSize: 1, IntervalSeconds: 0},
		},
	}

	plan, err := pm.CreateReleasePlan(ctx, "test-service", "v1.0.0", pkg, stages)
	if err != nil {
		t.Fatalf("CreateReleasePlan failed: %v", err)
	}
	defer cleanupTestData(t, ctx, svcCtx.ReleasePlanModel, plan.Id)

	err = pm.CancelPlan(ctx, plan.Id)
	if err != nil {
		t.Fatalf("CancelPlan failed: %v", err)
	}

	updatedPlan, err := pm.GetPlanStatus(ctx, plan.Id)
	if err != nil {
		t.Fatalf("GetPlanStatus failed: %v", err)
	}

	if updatedPlan.Status != model.PlanStatusCanceled {
		t.Errorf("expected status=canceled, got %s", updatedPlan.Status)
	}
}

func TestPlanManager_CancelPlan_InvalidStatus(t *testing.T) {
	ctx := context.Background()
	svcCtx := setupTestDB(t)

	pm := NewPlanManager(ctx, svcCtx)

	pkg := model.PackageInfo{
		URL:       "http://example.com/package.tar.gz",
		SHA256:    "abc123",
		Size:      1024,
		CreatedAt: time.Now(),
	}

	stages := []model.Stage{
		{
			Name: "stage1",
			Nodes: []model.StageNode{
				{Host: "host1", Status: model.NodeStatusPending},
			},
			Pacer: model.PacerConfig{BatchSize: 1, IntervalSeconds: 0},
		},
	}

	plan, err := pm.CreateReleasePlan(ctx, "test-service", "v1.0.0", pkg, stages)
	if err != nil {
		t.Fatalf("CreateReleasePlan failed: %v", err)
	}
	defer cleanupTestData(t, ctx, svcCtx.ReleasePlanModel, plan.Id)

	plan.Status = model.PlanStatusSuccess
	svcCtx.ReleasePlanModel.Update(ctx, plan)

	err = pm.CancelPlan(ctx, plan.Id)
	if err == nil {
		t.Fatal("expected error when canceling plan with invalid status")
	}
}
