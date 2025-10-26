package deployments

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Z3Labs/Hackathon/backend/internal/logic/deployments/executor"
	"github.com/Z3Labs/Hackathon/backend/internal/model"
	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeploymentManager struct {
	deploymentModel  model.DeploymentModel
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

	batchNodes := []model.NodeDeployment{}
	for i := 0; i < len(deployment.NodeDeployments); i++ {
		if deployment.NodeDeployments[i].NodeDeployStatus != model.NodeDeploymentStatusDeploying {
			continue
		}
		batchNodes = append(batchNodes, deployment.NodeDeployments[i])
	}

	if err := dm.executeBatch(ctx, deployment, batchNodes); err != nil {
		dm.deploymentModel.UpdateStatus(context.Background(), deployment.Id, model.DeploymentStatusFailed)
		return
	}
	// 全部节点发布完成，设置本次发布完成状态
	deployment, _ = dm.deploymentModel.FindById(context.Background(), deployment.Id)
	if deployment.Status != model.DeploymentStatusDeploying {
		return
	}
	succCount := 0
	for _, node := range deployment.NodeDeployments {
		if node.NodeDeployStatus == model.NodeDeploymentStatusSuccess {
			succCount++
		}
	}
	if succCount == len(deployment.NodeDeployments) {
		dm.deploymentModel.UpdateStatus(ctx, deployment.Id, model.DeploymentStatusSuccess)
		if app, err := dm.applicationModel.FindById(ctx, deployment.AppId); err == nil {
			app.PrevVersion = app.CurrentVersion
			app.CurrentVersion = deployment.PackageVersion
			dm.applicationModel.Update(ctx, app)
		}
	}
}

func (dm *DeploymentManager) executeBatch(ctx context.Context, deployment *model.Deployment, nodes []model.NodeDeployment) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(nodes))

	for i := range nodes {
		wg.Add(1)
		go func(nodeID string) {
			defer wg.Done()

			if err := dm.executeNode(ctx, deployment, nodeID); err != nil {
				errChan <- err
			}
		}(nodes[i].Id)
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

func findNodeIndex(nodes []model.NodeDeployment, nodeId string) int {
	for i := range nodes {
		if nodes[i].Id == nodeId {
			return i
		}
	}
	return -1
}
func (dm *DeploymentManager) executeNode(ctx context.Context, deployment *model.Deployment, nodeId string) error {
	deployment, err := dm.deploymentModel.FindById(context.Background(), deployment.Id)
	if err != nil {
		return err
	}
	nodeIndex := findNodeIndex(deployment.NodeDeployments, nodeId)
	if nodeIndex < 0 {
		return fmt.Errorf("node %s not found", nodeId)
	}
	node := &deployment.NodeDeployments[nodeIndex]
	logx.Infof("start executing node(%d) %s, deployment %s", nodeIndex, node.Id, deployment.Id)
	node.DeployingVersion = deployment.PackageVersion
	node.Platform = deployment.Platform
	node.UpdatedAt = time.Now()
	if node.CreatedAt.IsZero() {
		node.CreatedAt = time.Now()
	}

	if err := dm.deploymentModel.Update(context.Background(), deployment); err != nil {
		return fmt.Errorf("failed to update node status: %w", err)
	}

	executor, err := dm.executorFactory.CreateExecutor(ctx, executor.ExecutorConfig{
		Platform:    string(deployment.Platform),
		Host:        node.Id,
		IP:          node.Ip,
		Service:     deployment.AppName,
		Version:     deployment.PackageVersion,
		PrevVersion: node.PrevVersion,
		PackageURL:  deployment.Package.URL,
		MD5:         deployment.Package.MD5,
	})

	if err != nil {
		logx.Errorf("failed to create executor: %w", err)
		node.NodeDeployStatus = model.NodeDeploymentStatusFailed
		node.ReleaseLog = err.Error()
		node.UpdatedAt = time.Now()
		dm.deploymentModel.Update(context.Background(), deployment)
		return err
	}
	if err := executor.Deploy(ctx); err != nil {
		if ctx.Err() != nil {
			logx.Errorf("deployment canceled")
			node.NodeDeployStatus = model.NodeDeploymentStatusFailed
			node.ReleaseLog = "deployment canceled"
			node.UpdatedAt = time.Now()
			dm.deploymentModel.Update(context.Background(), deployment)
			return ctx.Err()
		}
		logx.Errorf("deployment failed: %w", err)
		node.NodeDeployStatus = model.NodeDeploymentStatusFailed
		node.ReleaseLog = err.Error()
		node.UpdatedAt = time.Now()

		if rollbackErr := executor.Rollback(ctx); rollbackErr != nil {
			logx.Errorf("rollback failed: %w", rollbackErr)
			node.ReleaseLog = fmt.Sprintf("deploy failed: %s, rollback failed: %s", err.Error(), rollbackErr.Error())
		} else {
			node.NodeDeployStatus = model.NodeDeploymentStatusRolledBack
		}
		node.UpdatedAt = time.Now()
		dm.deploymentModel.Update(context.Background(), deployment)
		return err
	}

	node.NodeDeployStatus = model.NodeDeploymentStatusSuccess
	node.ReleaseLog = "deployment successful"
	node.PrevVersion = node.CurrentVersion
	node.CurrentVersion = deployment.PackageVersion
	node.DeployingVersion = ""
	node.UpdatedAt = time.Now()
	logx.Infof("deployment successful: %s, node: %s, version: %s, deploying version: %s", deployment.Id, node.Id, deployment.PackageVersion, node.DeployingVersion)
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

func (dm *DeploymentManager) ContinueDeployingDeployments(ctx context.Context) error {
	statuses := []model.DeploymentStatus{
		model.DeploymentStatusDeploying,
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
				logx.Infof("continuing deployment: %s", dep.Id)
				defer dm.unregisterTask(dep.Id)
				dm.executeNodes(taskCtx, dep)
			}(deployment)
		}
	}

	return nil
}
