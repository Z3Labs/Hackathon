package plan

import (
	"context"
	"testing"
	"time"

	"github.com/Z3Labs/Hackathon/backend/internal/logic/deployments/executor"
	"github.com/Z3Labs/Hackathon/backend/internal/model"
)

func TestRollbackManager_RollbackPlan_Success(t *testing.T) {
	ctx := context.Background()
	svc := setupTestDB(t)
	mockExecutorFactory := executor.NewMockExecutorFactory()
	ctx = context.Background()
	rm := NewRollbackManager(ctx, svc, mockExecutorFactory)

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
				{
					Host:           "host1",
					Status:         model.NodeStatusFailed,
					CurrentVersion: "v1.0.0",
					PrevVersion:    "v0.9.0",
				},
			},
			Pacer: model.PacerConfig{BatchSize: 1, IntervalSeconds: 0},
		},
	}

	plan := &model.ReleasePlan{
		Svc:           "test-service",
		TargetVersion: "v1.0.0",
		Package:       pkg,
		Stages:        stages,
		Status:        model.PlanStatusFailed,
	}
	svc.ReleasePlanModel.Insert(ctx, plan)
	defer cleanupTestData(t, ctx, svc.ReleasePlanModel, plan.Id)

	nodeStatus := &model.NodeDeployStatusRecord{
		Host:           "host1",
		Service:        "test-service",
		CurrentVersion: "v1.0.0",
		PrevVersion:    "v0.9.0",
		Platform:       model.PlatformPhysical,
		State:          model.NodeStatusFailed,
	}
	svc.NodeStatusModel.Insert(ctx, nodeStatus)

	err := rm.RollbackPlan(ctx, plan.Id, nil)
	if err != nil {
		t.Fatalf("RollbackPlan failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	updatedPlan, err := rm.GetRollbackStatus(ctx, plan.Id)
	if err != nil {
		t.Fatalf("GetRollbackStatus failed: %v", err)
	}

	if updatedPlan.Status != model.PlanStatusRolledBack {
		t.Errorf("expected status=rolledback, got %s", updatedPlan.Status)
	}
}

func TestRollbackManager_RollbackPlan_WithSpecificHosts(t *testing.T) {
	ctx := context.Background()
	svc := setupTestDB(t)
	mockExecutorFactory := executor.NewMockExecutorFactory()

	rm := NewRollbackManager(ctx, svc, mockExecutorFactory)

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
				{
					Host:           "host1",
					Status:         model.NodeStatusFailed,
					CurrentVersion: "v1.0.0",
					PrevVersion:    "v0.9.0",
				},
				{
					Host:           "host2",
					Status:         model.NodeStatusFailed,
					CurrentVersion: "v1.0.0",
					PrevVersion:    "v0.9.0",
				},
			},
			Pacer: model.PacerConfig{BatchSize: 2, IntervalSeconds: 0},
		},
	}

	plan := &model.ReleasePlan{
		Svc:           "test-service",
		TargetVersion: "v1.0.0",
		Package:       pkg,
		Stages:        stages,
		Status:        model.PlanStatusFailed,
	}
	svc.ReleasePlanModel.Insert(ctx, plan)
	defer cleanupTestData(t, ctx, svc.ReleasePlanModel, plan.Id)

	svc.NodeStatusModel.Insert(ctx, &model.NodeDeployStatusRecord{
		Host:           "host1",
		Service:        "test-service",
		CurrentVersion: "v1.0.0",
		PrevVersion:    "v0.9.0",
		Platform:       model.PlatformPhysical,
		State:          model.NodeStatusFailed,
	})

	svc.NodeStatusModel.Insert(ctx, &model.NodeDeployStatusRecord{
		Host:           "host2",
		Service:        "test-service",
		CurrentVersion: "v1.0.0",
		PrevVersion:    "v0.9.0",
		Platform:       model.PlatformPhysical,
		State:          model.NodeStatusFailed,
	})

	err := rm.RollbackPlan(ctx, plan.Id, []string{"host1"})
	if err != nil {
		t.Fatalf("RollbackPlan failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	updatedPlan, err := rm.GetRollbackStatus(ctx, plan.Id)
	if err != nil {
		t.Fatalf("GetRollbackStatus failed: %v", err)
	}

	if updatedPlan.Status != model.PlanStatusRolledBack {
		t.Errorf("expected status=rolledback, got %s", updatedPlan.Status)
	}
}

func TestRollbackManager_RollbackPlan_InvalidStatus(t *testing.T) {
	ctx := context.Background()
	svc := setupTestDB(t)
	mockExecutorFactory := executor.NewMockExecutorFactory()

	rm := NewRollbackManager(context.Background(), svc, mockExecutorFactory)

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

	plan := &model.ReleasePlan{
		Svc:           "test-service",
		TargetVersion: "v1.0.0",
		Package:       pkg,
		Stages:        stages,
		Status:        model.PlanStatusPending,
	}
	svc.ReleasePlanModel.Insert(ctx, plan)
	defer cleanupTestData(t, ctx, svc.ReleasePlanModel, plan.Id)

	err := rm.RollbackPlan(ctx, plan.Id, nil)
	if err == nil {
		t.Fatal("expected error when rolling back plan with invalid status")
	}
}

func TestRollbackManager_RollbackPlan_NoPreviousVersion(t *testing.T) {
	ctx := context.Background()
	svc := setupTestDB(t)
	mockExecutorFactory := executor.NewMockExecutorFactory()

	rm := NewRollbackManager(ctx, svc, mockExecutorFactory)

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
				{
					Host:           "host1",
					Status:         model.NodeStatusFailed,
					CurrentVersion: "v1.0.0",
					PrevVersion:    "",
				},
			},
			Pacer: model.PacerConfig{BatchSize: 1, IntervalSeconds: 0},
		},
	}

	plan := &model.ReleasePlan{
		Svc:           "test-service",
		TargetVersion: "v1.0.0",
		Package:       pkg,
		Stages:        stages,
		Status:        model.PlanStatusFailed,
	}
	svc.ReleasePlanModel.Insert(ctx, plan)
	defer cleanupTestData(t, ctx, svc.ReleasePlanModel, plan.Id)

	nodeStatus := &model.NodeDeployStatusRecord{
		Host:           "host1",
		Service:        "test-service",
		CurrentVersion: "v1.0.0",
		PrevVersion:    "",
		Platform:       model.PlatformPhysical,
		State:          model.NodeStatusFailed,
	}
	svc.NodeStatusModel.Insert(ctx, nodeStatus)

	err := rm.RollbackPlan(ctx, plan.Id, nil)
	if err != nil {
		t.Fatalf("RollbackPlan should not fail: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	updatedPlan, err := rm.GetRollbackStatus(ctx, plan.Id)
	if err != nil {
		t.Fatalf("GetRollbackStatus failed: %v", err)
	}

	if updatedPlan.Status != model.PlanStatusRolledBack {
		t.Errorf("expected status=rolledback, got %s", updatedPlan.Status)
	}
}

func TestRollbackManager_RollbackPlan_NoNodesToRollback(t *testing.T) {
	ctx := context.Background()
	svc := setupTestDB(t)
	mockExecutorFactory := executor.NewMockExecutorFactory()

	rm := NewRollbackManager(ctx, svc, mockExecutorFactory)

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

	plan := &model.ReleasePlan{
		Svc:           "test-service",
		TargetVersion: "v1.0.0",
		Package:       pkg,
		Stages:        stages,
		Status:        model.PlanStatusFailed,
	}
	svc.ReleasePlanModel.Insert(ctx, plan)
	defer cleanupTestData(t, ctx, svc.ReleasePlanModel, plan.Id)

	err := rm.RollbackPlan(ctx, plan.Id, nil)
	if err == nil {
		t.Fatal("expected error when no nodes to rollback")
	}
}

func TestRollbackManager_ContainsHost(t *testing.T) {
	rm := &RollbackManager{}

	hosts := []string{"host1", "host2", "host3"}

	if !rm.containsHost(hosts, "host1") {
		t.Error("expected containsHost to return true for host1")
	}

	if rm.containsHost(hosts, "host4") {
		t.Error("expected containsHost to return false for host4")
	}
}
