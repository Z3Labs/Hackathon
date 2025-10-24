package deploy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Z3Labs/Hackathon/backend/internal/clients/deploy"
	"github.com/Z3Labs/Hackathon/backend/internal/model"
)

type RollbackManager struct {
	releasePlanModel model.ReleasePlanModel
	nodeStatusModel  model.NodeStatusModel
	executorFactory  *deploy.ExecutorFactory
}

func NewRollbackManager(
	releasePlanModel model.ReleasePlanModel,
	nodeStatusModel model.NodeStatusModel,
	executorFactory *deploy.ExecutorFactory,
) *RollbackManager {
	return &RollbackManager{
		releasePlanModel: releasePlanModel,
		nodeStatusModel:  nodeStatusModel,
		executorFactory:  executorFactory,
	}
}

func (rm *RollbackManager) RollbackPlan(ctx context.Context, planID string, hosts []string) error {
	plan, err := rm.releasePlanModel.FindById(ctx, planID)
	if err != nil {
		return fmt.Errorf("failed to find plan: %w", err)
	}

	if plan.Status != model.PlanStatusFailed && plan.Status != model.PlanStatusPartialSuccess {
		return fmt.Errorf("cannot rollback plan with status: %s", plan.Status)
	}

	plan.Status = model.PlanStatusRollingBack
	if err := rm.releasePlanModel.Update(ctx, plan); err != nil {
		return fmt.Errorf("failed to update plan status: %w", err)
	}

	var nodesToRollback []model.StageNode
	for i := range plan.Stages {
		for j := range plan.Stages[i].Nodes {
			node := &plan.Stages[i].Nodes[j]
			if len(hosts) == 0 || rm.containsHost(hosts, node.Host) {
				if node.Status == model.NodeStatusFailed || node.Status == model.NodeStatusSuccess {
					nodesToRollback = append(nodesToRollback, *node)
				}
			}
		}
	}

	if len(nodesToRollback) == 0 {
		return fmt.Errorf("no nodes found to rollback")
	}

	go rm.executeRollback(context.Background(), plan, nodesToRollback)

	return nil
}

func (rm *RollbackManager) executeRollback(ctx context.Context, plan *model.ReleasePlan, nodes []model.StageNode) {
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := range nodes {
		wg.Add(1)
		go func(node *model.StageNode) {
			defer wg.Done()

			if err := rm.rollbackNode(ctx, plan, node); err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(&nodes[i])
	}

	wg.Wait()

	if successCount == len(nodes) {
		plan.Status = model.PlanStatusRolledBack
	} else if successCount > 0 {
		plan.Status = model.PlanStatusPartialSuccess
	} else {
		plan.Status = model.PlanStatusFailed
	}

	rm.releasePlanModel.Update(ctx, plan)
}

func (rm *RollbackManager) rollbackNode(ctx context.Context, plan *model.ReleasePlan, node *model.StageNode) error {
	nodeStatus, err := rm.nodeStatusModel.FindByHostAndService(ctx, node.Host, plan.Svc)
	if err != nil {
		return fmt.Errorf("failed to find node status: %w", err)
	}

	if nodeStatus.PrevVersion == "" {
		return fmt.Errorf("no previous version to rollback to")
	}

	executor, err := rm.executorFactory.CreateExecutor(ctx, deploy.ExecutorConfig{
		Platform:    string(nodeStatus.Platform),
		Host:        node.Host,
		Service:     plan.Svc,
		Version:     nodeStatus.CurrentVersion,
		PrevVersion: nodeStatus.PrevVersion,
		PackageURL:  plan.Package.URL,
		SHA256:      plan.Package.SHA256,
	})

	if err != nil {
		node.Status = model.NodeStatusFailed
		node.LastError = err.Error()
		nodeStatus.State = model.NodeStatusFailed
		nodeStatus.LastError = err.Error()
		rm.nodeStatusModel.Update(ctx, nodeStatus)
		return err
	}

	if err := executor.Rollback(ctx); err != nil {
		node.Status = model.NodeStatusFailed
		node.LastError = fmt.Sprintf("rollback failed: %s", err.Error())
		nodeStatus.State = model.NodeStatusFailed
		nodeStatus.LastError = node.LastError
		rm.nodeStatusModel.Update(ctx, nodeStatus)
		return err
	}

	node.Status = model.NodeStatusRolledBack
	node.CurrentVersion = nodeStatus.PrevVersion
	node.DeployingVersion = ""
	node.UpdatedAt = time.Now()

	nodeStatus.State = model.NodeStatusRolledBack
	nodeStatus.CurrentVersion = nodeStatus.PrevVersion
	nodeStatus.DeployingVersion = ""
	rm.nodeStatusModel.Update(ctx, nodeStatus)

	return nil
}

func (rm *RollbackManager) containsHost(hosts []string, host string) bool {
	for _, h := range hosts {
		if h == host {
			return true
		}
	}
	return false
}

func (rm *RollbackManager) GetRollbackStatus(ctx context.Context, planID string) (*model.ReleasePlan, error) {
	return rm.releasePlanModel.FindById(ctx, planID)
}
