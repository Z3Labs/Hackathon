package deployments

import (
	"context"
	"fmt"
	"sync"

	"github.com/Z3Labs/Hackathon/backend/internal/logic/deployments/executor"
	"github.com/Z3Labs/Hackathon/backend/internal/model"
	"github.com/Z3Labs/Hackathon/backend/internal/svc"
)

type RollbackManager struct {
	deploymentModel model.DeploymentModel
	nodeStatusModel model.NodeStatusModel
	executorFactory executor.ExecutorFactoryInterface
	taskRegistry    map[string]context.CancelFunc
	taskMutex       sync.RWMutex
}

func NewRollbackManager(ctx context.Context, svcCtx *svc.ServiceContext) *RollbackManager {
	return &RollbackManager{
		deploymentModel: svcCtx.DeploymentModel,
		nodeStatusModel: svcCtx.NodeStatusModel,
		executorFactory: executor.NewExecutorFactory(),
		taskRegistry:    make(map[string]context.CancelFunc),
	}
}

func (rm *RollbackManager) registerTask(deploymentID string, cancel context.CancelFunc) {
	rm.taskMutex.Lock()
	defer rm.taskMutex.Unlock()

	if oldCancel, exists := rm.taskRegistry[deploymentID]; exists {
		oldCancel()
	}

	rm.taskRegistry[deploymentID] = cancel
}

func (rm *RollbackManager) unregisterTask(deploymentID string) {
	rm.taskMutex.Lock()
	defer rm.taskMutex.Unlock()
	delete(rm.taskRegistry, deploymentID)
}

func (rm *RollbackManager) cancelTask(deploymentID string) bool {
	rm.taskMutex.RLock()
	cancel, exists := rm.taskRegistry[deploymentID]
	rm.taskMutex.RUnlock()

	if exists {
		cancel()
		return true
	}
	return false
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

	var nodesToRollback []int
	for i := range deployment.NodeDeployments {
		node := &deployment.NodeDeployments[i]
		if len(hosts) == 0 || rm.containsHost(hosts, node.Id) {
			if node.NodeDeployStatus == model.NodeDeploymentStatusFailed || node.NodeDeployStatus == model.NodeDeploymentStatusSuccess {
				nodesToRollback = append(nodesToRollback, i)
			}
		}
	}

	if len(nodesToRollback) == 0 {
		return fmt.Errorf("no nodes found to rollback")
	}

	taskCtx, cancel := context.WithCancel(context.Background())
	rm.registerTask(deploymentID, cancel)

	go func() {
		defer rm.unregisterTask(deploymentID)
		rm.executeRollback(taskCtx, deployment, nodesToRollback)
	}()

	return nil
}

func (rm *RollbackManager) executeRollback(ctx context.Context, deployment *model.Deployment, nodeIndexes []int) {
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for _, nodeIndex := range nodeIndexes {
		select {
		case <-ctx.Done():
			wg.Wait()
			rm.handleRollbackCancellation(deployment, successCount, len(nodeIndexes))
			return
		default:
		}

		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			if err := rm.rollbackNode(ctx, deployment, idx); err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(nodeIndex)
	}

	wg.Wait()

	if successCount == len(nodeIndexes) {
		deployment.Status = model.DeploymentStatusRolledBack
	} else if successCount > 0 {
		deployment.Status = model.DeploymentStatusPartialSuccess
	} else {
		deployment.Status = model.DeploymentStatusFailed
	}

	rm.deploymentModel.Update(context.Background(), deployment)
}

func (rm *RollbackManager) handleRollbackCancellation(deployment *model.Deployment, successCount, totalCount int) {
	current, _ := rm.deploymentModel.FindById(context.Background(), deployment.Id)
	if current != nil {
		if successCount > 0 && successCount < totalCount {
			current.Status = model.DeploymentStatusPartialSuccess
		} else if successCount == 0 {
			current.Status = model.DeploymentStatusFailed
		}
		rm.deploymentModel.Update(context.Background(), current)
	}
}

func (rm *RollbackManager) rollbackNode(ctx context.Context, deployment *model.Deployment, nodeIndex int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	deployment, err := rm.deploymentModel.FindById(context.Background(), deployment.Id)
	if err != nil {
		return err
	}

	if nodeIndex >= len(deployment.NodeDeployments) {
		return fmt.Errorf("node index out of range")
	}

	node := &deployment.NodeDeployments[nodeIndex]

	nodeStatus, err := rm.nodeStatusModel.FindByHostAndService(context.Background(), node.Id, deployment.AppName)
	if err != nil {
		return fmt.Errorf("failed to find node status: %w", err)
	}

	if nodeStatus.PrevVersion == "" {
		return fmt.Errorf("no previous version to rollback to")
	}

	executor, err := rm.executorFactory.CreateExecutor(ctx, executor.ExecutorConfig{
		Platform:    string(nodeStatus.Platform),
		Host:        node.Id,
		IP:          node.Ip,
		Service:     deployment.AppName,
		Version:     nodeStatus.CurrentVersion,
		PrevVersion: nodeStatus.PrevVersion,
		PackageURL:  deployment.Package.URL,
		SHA256:      deployment.Package.SHA256,
	})

	if err != nil {
		node.NodeDeployStatus = model.NodeDeploymentStatusFailed
		node.ReleaseLog = err.Error()
		nodeStatus.State = model.NodeStatusFailed
		nodeStatus.LastError = err.Error()
		rm.nodeStatusModel.Update(context.Background(), nodeStatus)
		rm.deploymentModel.Update(context.Background(), deployment)
		return err
	}

	if err := executor.Rollback(ctx); err != nil {
		if ctx.Err() != nil {
			node.NodeDeployStatus = model.NodeDeploymentStatusFailed
			node.ReleaseLog = "rollback canceled"
			nodeStatus.State = model.NodeStatusFailed
			nodeStatus.LastError = "rollback canceled"
			rm.nodeStatusModel.Update(context.Background(), nodeStatus)
			rm.deploymentModel.Update(context.Background(), deployment)
			return ctx.Err()
		}

		node.NodeDeployStatus = model.NodeDeploymentStatusFailed
		node.ReleaseLog = fmt.Sprintf("rollback failed: %s", err.Error())
		nodeStatus.State = model.NodeStatusFailed
		nodeStatus.LastError = node.ReleaseLog
		rm.nodeStatusModel.Update(context.Background(), nodeStatus)
		rm.deploymentModel.Update(context.Background(), deployment)
		return err
	}

	node.NodeDeployStatus = model.NodeDeploymentStatusRolledBack
	node.ReleaseLog = "rollback successful"
	nodeStatus.State = model.NodeStatusRolledBack
	nodeStatus.CurrentVersion = nodeStatus.PrevVersion
	nodeStatus.DeployingVersion = ""
	rm.nodeStatusModel.Update(context.Background(), nodeStatus)
	rm.deploymentModel.Update(context.Background(), deployment)

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

func (rm *RollbackManager) CancelRollback(ctx context.Context, deploymentID string) error {
	deployment, err := rm.deploymentModel.FindById(ctx, deploymentID)
	if err != nil {
		return fmt.Errorf("failed to find deployment: %w", err)
	}

	if deployment.Status != model.DeploymentStatusRollingBack {
		return fmt.Errorf("deployment is not in rolling back status")
	}

	deployment.Status = model.DeploymentStatusFailed
	if err := rm.deploymentModel.Update(ctx, deployment); err != nil {
		return err
	}

	rm.cancelTask(deploymentID)

	return nil
}

func (rm *RollbackManager) ContinueRollingBackDeployments(ctx context.Context) error {
	deployments, err := rm.deploymentModel.Search(ctx, &model.DeploymentCond{
		Status: string(model.DeploymentStatusRollingBack),
	})
	if err != nil {
		return fmt.Errorf("failed to search rolling back deployments: %w", err)
	}

	for _, deployment := range deployments {
		var nodesToRollback []int
		for i := range deployment.NodeDeployments {
			node := &deployment.NodeDeployments[i]
			if node.NodeDeployStatus == model.NodeDeploymentStatusFailed || node.NodeDeployStatus == model.NodeDeploymentStatusSuccess {
				nodesToRollback = append(nodesToRollback, i)
			}
		}

		if len(nodesToRollback) > 0 {
			taskCtx, cancel := context.WithCancel(context.Background())
			rm.registerTask(deployment.Id, cancel)

			go func(dep *model.Deployment, nodes []int) {
				defer rm.unregisterTask(dep.Id)
				rm.executeRollback(taskCtx, dep, nodes)
			}(deployment, nodesToRollback)
		}
	}

	return nil
}
