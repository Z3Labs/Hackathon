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

type DeploymentManager struct {
	deploymentModel  model.DeploymentModel
	nodeStatusModel  model.NodeStatusModel
	applicationModel model.ApplicationModel
	executorFactory  executor.ExecutorFactoryInterface
	taskRegistry     map[string]context.CancelFunc
	taskMutex        sync.RWMutex
	alertMonitor     *AlertMonitor
}

var (
	instance *DeploymentManager
	once     sync.Once
)

func NewDeploymentManager(
	ctx context.Context,
	svc *svc.ServiceContext,
) *DeploymentManager {
	once.Do(func() {
		instance = &DeploymentManager{
			deploymentModel:  svc.DeploymentModel,
			nodeStatusModel:  svc.NodeStatusModel,
			applicationModel: svc.ApplicationModel,
			executorFactory:  executor.NewExecutorFactory(),
			taskRegistry:     make(map[string]context.CancelFunc),
		}
	})
	return instance
}

func (dm *DeploymentManager) SetAlertMonitor(monitor *AlertMonitor) {
	dm.alertMonitor = monitor
}

func GetDeploymentManager() *DeploymentManager {
	return instance
}

func (dm *DeploymentManager) registerTask(deploymentID string, cancel context.CancelFunc) {
	dm.taskMutex.Lock()
	defer dm.taskMutex.Unlock()

	if oldCancel, exists := dm.taskRegistry[deploymentID]; exists {
		oldCancel()
	}

	dm.taskRegistry[deploymentID] = cancel
}

func (dm *DeploymentManager) unregisterTask(deploymentID string) {
	dm.taskMutex.Lock()
	defer dm.taskMutex.Unlock()
	delete(dm.taskRegistry, deploymentID)
}

func (dm *DeploymentManager) cancelTask(deploymentID string) bool {
	dm.taskMutex.RLock()
	cancel, exists := dm.taskRegistry[deploymentID]
	dm.taskMutex.RUnlock()

	if exists {
		cancel()
		return true
	}
	return false
}

func (dm *DeploymentManager) ExecuteDeployment(ctx context.Context, deploymentID string) error {
	deployment, err := dm.deploymentModel.FindById(ctx, deploymentID)
	if err != nil {
		return fmt.Errorf("failed to find deployment: %w", err)
	}

	if deployment.Status != model.DeploymentStatusPending {
		return fmt.Errorf("deployment status is not pending, current status: %s", deployment.Status)
	}

	deployment.Status = model.DeploymentStatusDeploying
	if err := dm.deploymentModel.Update(ctx, deployment); err != nil {
		return fmt.Errorf("failed to update deployment status: %w", err)
	}

	if dm.alertMonitor != nil {
		app, err := dm.applicationModel.FindById(ctx, deployment.AppName)
		if err == nil && app.RollbackPolicy != nil && app.RollbackPolicy.Enabled {
			if err := dm.alertMonitor.StartMonitoring(ctx, deployment, app); err != nil {
				fmt.Printf("failed to start alert monitoring for deployment %s: %v\n", deploymentID, err)
			}
		}
	}

	taskCtx, cancel := context.WithCancel(context.Background())
	dm.registerTask(deploymentID, cancel)

	go func() {
		defer dm.unregisterTask(deploymentID)
		dm.executeNodes(taskCtx, deployment)
	}()

	return nil
}

func (dm *DeploymentManager) executeNodes(ctx context.Context, deployment *model.Deployment) {
	batchSize := deployment.Pacer.BatchSize
	if batchSize <= 0 {
		batchSize = 1
	}
	intervalSeconds := deployment.Pacer.IntervalSeconds

	nodes := deployment.NodeDeployments
	for i := 0; i < len(nodes); i += batchSize {
		select {
		case <-ctx.Done():
			dm.handleCancellation(deployment)
			return
		default:
		}

		current, err := dm.deploymentModel.FindById(context.Background(), deployment.Id)
		if err != nil || current.Status == model.DeploymentStatusCanceled {
			dm.handleCancellation(deployment)
			return
		}

		end := i + batchSize
		if end > len(nodes) {
			end = len(nodes)
		}

		batch := nodes[i:end]
		if err := dm.executeBatch(ctx, deployment, batch); err != nil {
			if ctx.Err() != nil {
				dm.handleCancellation(deployment)
				return
			}

			deployment.Status = model.DeploymentStatusFailed
			dm.deploymentModel.Update(context.Background(), deployment)
			return
		}

		if end < len(nodes) {
			select {
			case <-time.After(time.Duration(intervalSeconds) * time.Second):
			case <-ctx.Done():
				dm.handleCancellation(deployment)
				return
			}
		}
	}

	deployment.Status = model.DeploymentStatusSuccess
	dm.deploymentModel.Update(context.Background(), deployment)
}

func (dm *DeploymentManager) handleCancellation(deployment *model.Deployment) {
	current, _ := dm.deploymentModel.FindById(context.Background(), deployment.Id)
	if current != nil && current.Status != model.DeploymentStatusCanceled {
		current.Status = model.DeploymentStatusCanceled
		dm.deploymentModel.Update(context.Background(), current)
	}
}

func (dm *DeploymentManager) executeBatch(ctx context.Context, deployment *model.Deployment, nodes []model.NodeDeployment) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(nodes))

	for i := range nodes {
		wg.Add(1)
		go func(nodeIndex int) {
			defer wg.Done()

			if err := dm.executeNode(ctx, deployment, nodeIndex); err != nil {
				errChan <- err
			}
		}(i)
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

func (dm *DeploymentManager) executeNode(ctx context.Context, deployment *model.Deployment, nodeIndex int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	deployment, err := dm.deploymentModel.FindById(context.Background(), deployment.Id)
	if err != nil {
		return err
	}

	if nodeIndex >= len(deployment.NodeDeployments) {
		return fmt.Errorf("node index out of range")
	}

	node := &deployment.NodeDeployments[nodeIndex]
	node.NodeDeployStatus = model.NodeDeploymentStatusDeploying

	if err := dm.deploymentModel.Update(context.Background(), deployment); err != nil {
		return fmt.Errorf("failed to update node status: %w", err)
	}

	nodeStatus := &model.NodeDeployStatusRecord{
		Host:             node.Id,
		Service:          deployment.AppName,
		CurrentVersion:   "",
		DeployingVersion: deployment.PackageVersion,
		PrevVersion:      "",
		Platform:         deployment.Platform,
		State:            model.NodeStatusDeploying,
	}

	existing, _ := dm.nodeStatusModel.FindByHostAndService(context.Background(), node.Id, deployment.AppName)
	if existing != nil {
		nodeStatus.Id = existing.Id
		nodeStatus.CurrentVersion = existing.CurrentVersion
		nodeStatus.PrevVersion = existing.PrevVersion
		dm.nodeStatusModel.Update(context.Background(), nodeStatus)
	} else {
		dm.nodeStatusModel.Insert(context.Background(), nodeStatus)
	}

	executor, err := dm.executorFactory.CreateExecutor(ctx, executor.ExecutorConfig{
		Platform:    string(deployment.Platform),
		Host:        node.Id,
		IP:          node.Ip,
		Service:     deployment.AppName,
		Version:     deployment.PackageVersion,
		PrevVersion: nodeStatus.PrevVersion,
		PackageURL:  deployment.Package.URL,
		MD5:         deployment.Package.MD5,
	})

	if err != nil {
		node.NodeDeployStatus = model.NodeDeploymentStatusFailed
		node.ReleaseLog = err.Error()
		nodeStatus.State = model.NodeStatusFailed
		nodeStatus.LastError = err.Error()
		dm.nodeStatusModel.Update(context.Background(), nodeStatus)
		dm.deploymentModel.Update(context.Background(), deployment)
		return err
	}

	if err := executor.Deploy(ctx); err != nil {
		if ctx.Err() != nil {
			node.NodeDeployStatus = model.NodeDeploymentStatusFailed
			node.ReleaseLog = "deployment canceled"
			nodeStatus.State = model.NodeStatusFailed
			nodeStatus.LastError = "deployment canceled"
			dm.nodeStatusModel.Update(context.Background(), nodeStatus)
			dm.deploymentModel.Update(context.Background(), deployment)
			return ctx.Err()
		}

		node.NodeDeployStatus = model.NodeDeploymentStatusFailed
		node.ReleaseLog = err.Error()
		nodeStatus.State = model.NodeStatusFailed
		nodeStatus.LastError = err.Error()
		dm.nodeStatusModel.Update(context.Background(), nodeStatus)

		if rollbackErr := executor.Rollback(ctx); rollbackErr != nil {
			node.ReleaseLog = fmt.Sprintf("deploy failed: %s, rollback failed: %s", err.Error(), rollbackErr.Error())
			nodeStatus.LastError = node.ReleaseLog
		} else {
			node.NodeDeployStatus = model.NodeDeploymentStatusRolledBack
			nodeStatus.State = model.NodeStatusRolledBack
		}
		dm.nodeStatusModel.Update(ctx, nodeStatus)
		dm.deploymentModel.Update(context.Background(), deployment)
		return err
	}

	node.NodeDeployStatus = model.NodeDeploymentStatusSuccess
	node.ReleaseLog = "deployment successful"
	nodeStatus.State = model.NodeStatusSuccess
	nodeStatus.CurrentVersion = deployment.PackageVersion
	nodeStatus.PrevVersion = nodeStatus.CurrentVersion
	nodeStatus.DeployingVersion = ""
	dm.nodeStatusModel.Update(ctx, nodeStatus)
	dm.deploymentModel.Update(context.Background(), deployment)

	return nil
}

func (dm *DeploymentManager) GetDeploymentStatus(ctx context.Context, deploymentID string) (*model.Deployment, error) {
	return dm.deploymentModel.FindById(ctx, deploymentID)
}

func (dm *DeploymentManager) CancelDeployment(ctx context.Context, deploymentID string) error {
	deployment, err := dm.deploymentModel.FindById(ctx, deploymentID)
	if err != nil {
		return fmt.Errorf("failed to find deployment: %w", err)
	}

	if deployment.Status != model.DeploymentStatusPending && deployment.Status != model.DeploymentStatusDeploying {
		return fmt.Errorf("cannot cancel deployment with status: %s", deployment.Status)
	}

	deployment.Status = model.DeploymentStatusCanceled
	if err := dm.deploymentModel.Update(ctx, deployment); err != nil {
		return err
	}

	dm.cancelTask(deploymentID)

	return nil
}

func (dm *DeploymentManager) ProcessPendingDeployments(ctx context.Context) error {
	deployments, err := dm.deploymentModel.Search(ctx, &model.DeploymentCond{
		Status: string(model.DeploymentStatusPending),
	})
	if err != nil {
		return fmt.Errorf("failed to search pending deployments: %w", err)
	}

	for _, deployment := range deployments {
		if err := dm.ExecuteDeployment(ctx, deployment.Id); err != nil {
			fmt.Printf("failed to execute deployment %s: %v\n", deployment.Id, err)
		}
	}

	return nil
}

func (dm *DeploymentManager) ContinueDeployingDeployments(ctx context.Context) error {
	statuses := []model.DeploymentStatus{
		model.DeploymentStatusDeploying,
		model.DeploymentStatusPartialSuccess,
	}

	for _, status := range statuses {
		deployments, err := dm.deploymentModel.Search(ctx, &model.DeploymentCond{
			Status: string(status),
		})
		if err != nil {
			return fmt.Errorf("failed to search deployments with status %s: %w", status, err)
		}

		for _, deployment := range deployments {
			taskCtx, cancel := context.WithCancel(context.Background())
			dm.registerTask(deployment.Id, cancel)

			go func(dep *model.Deployment) {
				defer dm.unregisterTask(dep.Id)
				dm.executeNodes(taskCtx, dep)
			}(deployment)
		}
	}

	return nil
}
