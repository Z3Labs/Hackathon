package plan

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Z3Labs/Hackathon/backend/internal/logic/deployments/executor"
	"github.com/Z3Labs/Hackathon/backend/internal/model"
	"github.com/Z3Labs/Hackathon/backend/internal/svc"
)

type PlanManager struct {
	releasePlanModel model.ReleasePlanModel
	nodeStatusModel  model.NodeStatusModel
	executorFactory  executor.ExecutorFactoryInterface
}

var (
	instance *PlanManager
	once     sync.Once
)

func NewPlanManager(
	ctx context.Context,
	svc *svc.ServiceContext,
) *PlanManager {
	once.Do(func() {
		instance = &PlanManager{
			releasePlanModel: svc.ReleasePlanModel,
			nodeStatusModel:  svc.NodeStatusModel,
			executorFactory:  executor.NewExecutorFactory(),
		}
	})
	return instance
}

func GetPlanManager() *PlanManager {
	return instance
}

func (pm *PlanManager) CreateReleasePlan(
	ctx context.Context,
	svc string,
	targetVersion string,
	pkg model.PackageInfo,
	stages []model.Stage,
) (*model.ReleasePlan, error) {
	plan := &model.ReleasePlan{
		Svc:           svc,
		TargetVersion: targetVersion,
		ReleaseTime:   time.Now(),
		Package:       pkg,
		Stages:        stages,
		Status:        model.PlanStatusPending,
	}

	if err := pm.releasePlanModel.Insert(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to insert release plan: %w", err)
	}

	return plan, nil
}

func (pm *PlanManager) ExecutePlan(ctx context.Context, planID string) error {
	plan, err := pm.releasePlanModel.FindById(ctx, planID)
	if err != nil {
		return fmt.Errorf("failed to find plan: %w", err)
	}

	if plan.Status != model.PlanStatusPending {
		return fmt.Errorf("plan status is not pending, current status: %s", plan.Status)
	}

	plan.Status = model.PlanStatusDeploying
	if err := pm.releasePlanModel.Update(ctx, plan); err != nil {
		return fmt.Errorf("failed to update plan status: %w", err)
	}

	go pm.executeStages(context.Background(), plan)

	return nil
}

func (pm *PlanManager) executeStages(ctx context.Context, plan *model.ReleasePlan) {
	for i := range plan.Stages {
		stage := &plan.Stages[i]
		stage.Status = model.StageStatusDeploying
		pm.releasePlanModel.Update(ctx, plan)

		if err := pm.executeStage(ctx, plan, stage); err != nil {
			stage.Status = model.StageStatusFailed
			plan.Status = model.PlanStatusFailed
			pm.releasePlanModel.Update(ctx, plan)
			return
		}

		stage.Status = model.StageStatusSuccess
		pm.releasePlanModel.Update(ctx, plan)
	}

	plan.Status = model.PlanStatusSuccess
	pm.releasePlanModel.Update(ctx, plan)
}

func (pm *PlanManager) executeStage(ctx context.Context, plan *model.ReleasePlan, stage *model.Stage) error {
	batchSize := stage.Pacer.BatchSize
	intervalSeconds := stage.Pacer.IntervalSeconds

	for i := 0; i < len(stage.Nodes); i += batchSize {
		end := i + batchSize
		if end > len(stage.Nodes) {
			end = len(stage.Nodes)
		}

		batch := stage.Nodes[i:end]
		if err := pm.executeBatch(ctx, plan, batch); err != nil {
			return err
		}

		if end < len(stage.Nodes) {
			time.Sleep(time.Duration(intervalSeconds) * time.Second)
		}
	}

	return nil
}

func (pm *PlanManager) executeBatch(ctx context.Context, plan *model.ReleasePlan, nodes []model.StageNode) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(nodes))

	for i := range nodes {
		wg.Add(1)
		go func(node *model.StageNode) {
			defer wg.Done()

			if err := pm.executeNode(ctx, plan, node); err != nil {
				errChan <- err
			}
		}(&nodes[i])
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func (pm *PlanManager) executeNode(ctx context.Context, plan *model.ReleasePlan, node *model.StageNode) error {
	node.Status = model.NodeStatusDeploying
	node.DeployingVersion = plan.TargetVersion
	node.UpdatedAt = time.Now()

	nodeStatus := &model.NodeDeployStatusRecord{
		Host:             node.Host,
		Service:          plan.Svc,
		CurrentVersion:   node.CurrentVersion,
		DeployingVersion: plan.TargetVersion,
		PrevVersion:      node.PrevVersion,
		Platform:         model.PlatformPhysical,
		State:            model.NodeStatusDeploying,
	}

	existing, _ := pm.nodeStatusModel.FindByHostAndService(ctx, node.Host, plan.Svc)
	if existing != nil {
		nodeStatus.Id = existing.Id
		pm.nodeStatusModel.Update(ctx, nodeStatus)
	} else {
		pm.nodeStatusModel.Insert(ctx, nodeStatus)
	}

	executor, err := pm.executorFactory.CreateExecutor(ctx, executor.ExecutorConfig{
		Platform:    string(model.PlatformPhysical),
		Host:        node.Host,
		IP:          node.IP,
		Service:     plan.Svc,
		Version:     plan.TargetVersion,
		PrevVersion: node.PrevVersion,
		PackageURL:  plan.Package.URL,
		SHA256:      plan.Package.SHA256,
	})

	if err != nil {
		node.Status = model.NodeStatusFailed
		node.LastError = err.Error()
		nodeStatus.State = model.NodeStatusFailed
		nodeStatus.LastError = err.Error()
		pm.nodeStatusModel.Update(ctx, nodeStatus)
		return err
	}

	if err := executor.Deploy(ctx); err != nil {
		node.Status = model.NodeStatusFailed
		node.LastError = err.Error()
		nodeStatus.State = model.NodeStatusFailed
		nodeStatus.LastError = err.Error()
		pm.nodeStatusModel.Update(ctx, nodeStatus)

		if rollbackErr := executor.Rollback(ctx); rollbackErr != nil {
			node.LastError = fmt.Sprintf("deploy failed: %s, rollback failed: %s", err.Error(), rollbackErr.Error())
			nodeStatus.LastError = node.LastError
		} else {
			node.Status = model.NodeStatusRolledBack
			nodeStatus.State = model.NodeStatusRolledBack
		}
		pm.nodeStatusModel.Update(ctx, nodeStatus)
		return err
	}

	node.Status = model.NodeStatusSuccess
	node.CurrentVersion = plan.TargetVersion
	node.PrevVersion = node.CurrentVersion
	node.DeployingVersion = ""
	node.UpdatedAt = time.Now()

	nodeStatus.State = model.NodeStatusSuccess
	nodeStatus.CurrentVersion = plan.TargetVersion
	nodeStatus.PrevVersion = nodeStatus.CurrentVersion
	nodeStatus.DeployingVersion = ""
	pm.nodeStatusModel.Update(ctx, nodeStatus)

	return nil
}

func (pm *PlanManager) GetPlanStatus(ctx context.Context, planID string) (*model.ReleasePlan, error) {
	return pm.releasePlanModel.FindById(ctx, planID)
}

func (pm *PlanManager) CancelPlan(ctx context.Context, planID string) error {
	plan, err := pm.releasePlanModel.FindById(ctx, planID)
	if err != nil {
		return fmt.Errorf("failed to find plan: %w", err)
	}

	if plan.Status != model.PlanStatusPending && plan.Status != model.PlanStatusDeploying {
		return fmt.Errorf("cannot cancel plan with status: %s", plan.Status)
	}

	plan.Status = model.PlanStatusCanceled
	return pm.releasePlanModel.Update(ctx, plan)
}

func (pm *PlanManager) ProcessPendingPlans(ctx context.Context) error {
	plans, err := pm.releasePlanModel.Search(ctx, &model.ReleasePlanCond{
		Status: model.PlanStatusPending,
	})
	if err != nil {
		return fmt.Errorf("failed to search pending plans: %w", err)
	}

	for _, plan := range plans {
		if err := pm.ExecutePlan(ctx, plan.Id); err != nil {
			fmt.Printf("failed to execute plan %s: %v\n", plan.Id, err)
		}
	}

	return nil
}

func (pm *PlanManager) ContinueDeployingPlans(ctx context.Context) error {
	statuses := []model.PlanStatus{
		model.PlanStatusDeploying,
		model.PlanStatusPartialSuccess,
	}

	for _, status := range statuses {
		plans, err := pm.releasePlanModel.Search(ctx, &model.ReleasePlanCond{
			Status: status,
		})
		if err != nil {
			return fmt.Errorf("failed to search plans with status %s: %w", status, err)
		}

		for _, plan := range plans {
			go pm.executeStages(context.Background(), plan)
		}
	}

	return nil
}
