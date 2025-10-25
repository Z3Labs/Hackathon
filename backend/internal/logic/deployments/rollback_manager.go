package deployments

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Z3Labs/Hackathon/backend/internal/logic/deployments/executor"
	"github.com/Z3Labs/Hackathon/backend/internal/model"
	"github.com/Z3Labs/Hackathon/backend/internal/svc"
)

type RollbackManager struct {
	deploymentModel model.DeploymentModel
	nodeStatusModel model.NodeStatusModel
	executorFactory executor.ExecutorFactoryInterface
}

func NewRollbackManager(ctx context.Context, svcCtx *svc.ServiceContext) *RollbackManager {
	return &RollbackManager{
		deploymentModel: svcCtx.DeploymentModel,
		nodeStatusModel: svcCtx.NodeStatusModel,
		executorFactory: executor.NewExecutorFactory(),
	}
}

func (rm *RollbackManager) RollbackDeployment(ctx context.Context, deploymentID string, hosts []string) error {
	deployment, err := rm.deploymentModel.FindById(ctx, deploymentID)
	if err != nil {
		return fmt.Errorf("failed to find deployment: %w", err)
	}

	if deployment.Status != model.DeploymentStatusFailed && deployment.Status != model.DeploymentStatusPartialSuccess {
		return fmt.Errorf("cannot rollback deployment with status: %s", deployment.Status)
	}

	deployment.Status = model.DeploymentStatusRollingBack
	if err := rm.deploymentModel.Update(ctx, deployment); err != nil {
		return fmt.Errorf("failed to update deployment status: %w", err)
	}

	var nodesToRollback []model.StageNode
	for i := range deployment.Stages {
		for j := range deployment.Stages[i].Nodes {
			node := &deployment.Stages[i].Nodes[j]
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

	go rm.executeRollback(context.Background(), deployment, nodesToRollback)

	return nil
}

func (rm *RollbackManager) executeRollback(ctx context.Context, deployment *model.Deployment, nodes []model.StageNode) {
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := range nodes {
		wg.Add(1)
		go func(node *model.StageNode) {
			defer wg.Done()

			if err := rm.rollbackNode(ctx, deployment, node); err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(&nodes[i])
	}

	wg.Wait()

	if successCount == len(nodes) {
		deployment.Status = model.DeploymentStatusRolledBack
	} else if successCount > 0 {
		deployment.Status = model.DeploymentStatusPartialSuccess
	} else {
		deployment.Status = model.DeploymentStatusFailed
	}

	rm.deploymentModel.Update(ctx, deployment)
}

func (rm *RollbackManager) rollbackNode(ctx context.Context, deployment *model.Deployment, node *model.StageNode) error {
	nodeStatus, err := rm.nodeStatusModel.FindByHostAndService(ctx, node.Host, deployment.AppName)
	if err != nil {
		return fmt.Errorf("failed to find node status: %w", err)
	}

	if nodeStatus.PrevVersion == "" {
		return fmt.Errorf("no previous version to rollback to")
	}

	executor, err := rm.executorFactory.CreateExecutor(ctx, executor.ExecutorConfig{
		Platform:    string(nodeStatus.Platform),
		Host:        node.Host,
		IP:          node.IP,
		Service:     deployment.AppName,
		Version:     nodeStatus.CurrentVersion,
		PrevVersion: nodeStatus.PrevVersion,
		PackageURL:  deployment.Package.URL,
		SHA256:      deployment.Package.SHA256,
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

func (rm *RollbackManager) GetRollbackStatus(ctx context.Context, deploymentID string) (*model.Deployment, error) {
	return rm.deploymentModel.FindById(ctx, deploymentID)
}

func (rm *RollbackManager) ContinueRollingBackDeployments(ctx context.Context) error {
	deployments, err := rm.deploymentModel.Search(ctx, &model.DeploymentCond{
		Status: string(model.DeploymentStatusRollingBack),
	})
	if err != nil {
		return fmt.Errorf("failed to search rolling back deployments: %w", err)
	}

	for _, deployment := range deployments {
		var nodesToRollback []model.StageNode
		for i := range deployment.Stages {
			for j := range deployment.Stages[i].Nodes {
				node := &deployment.Stages[i].Nodes[j]
				if node.Status == model.NodeStatusFailed || node.Status == model.NodeStatusSuccess {
					nodesToRollback = append(nodesToRollback, *node)
				}
			}
		}

		if len(nodesToRollback) > 0 {
			go rm.executeRollback(context.Background(), deployment, nodesToRollback)
		}
	}

	return nil
}
