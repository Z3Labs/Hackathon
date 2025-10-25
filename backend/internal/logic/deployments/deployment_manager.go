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
	deploymentModel model.DeploymentModel
	nodeStatusModel model.NodeStatusModel
	executorFactory executor.ExecutorFactoryInterface
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
			deploymentModel: svc.DeploymentModel,
			nodeStatusModel: svc.NodeStatusModel,
			executorFactory: executor.NewExecutorFactory(),
		}
	})
	return instance
}

func GetDeploymentManager() *DeploymentManager {
	return instance
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

	go dm.executeStages(context.Background(), deployment)

	return nil
}

func (dm *DeploymentManager) executeStages(ctx context.Context, deployment *model.Deployment) {
	for i := range deployment.Stages {
		stage := &deployment.Stages[i]
		stage.Status = model.StageStatusDeploying
		dm.deploymentModel.Update(ctx, deployment)

		if err := dm.executeStage(ctx, deployment, stage); err != nil {
			stage.Status = model.StageStatusFailed
			deployment.Status = model.DeploymentStatusFailed
			dm.deploymentModel.Update(ctx, deployment)
			return
		}

		stage.Status = model.StageStatusSuccess
		dm.deploymentModel.Update(ctx, deployment)
	}

	deployment.Status = model.DeploymentStatusSuccess
	dm.deploymentModel.Update(ctx, deployment)
}

func (dm *DeploymentManager) executeStage(ctx context.Context, deployment *model.Deployment, stage *model.Stage) error {
	batchSize := stage.Pacer.BatchSize
	intervalSeconds := stage.Pacer.IntervalSeconds

	for i := 0; i < len(stage.Nodes); i += batchSize {
		end := i + batchSize
		if end > len(stage.Nodes) {
			end = len(stage.Nodes)
		}

		batch := stage.Nodes[i:end]
		if err := dm.executeBatch(ctx, deployment, batch); err != nil {
			return err
		}

		if end < len(stage.Nodes) {
			time.Sleep(time.Duration(intervalSeconds) * time.Second)
		}
	}

	return nil
}

func (dm *DeploymentManager) executeBatch(ctx context.Context, deployment *model.Deployment, nodes []model.StageNode) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(nodes))

	for i := range nodes {
		wg.Add(1)
		go func(node *model.StageNode) {
			defer wg.Done()

			if err := dm.executeNode(ctx, deployment, node); err != nil {
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

func (dm *DeploymentManager) executeNode(ctx context.Context, deployment *model.Deployment, node *model.StageNode) error {
	node.Status = model.NodeStatusDeploying
	node.DeployingVersion = deployment.PackageVersion
	node.UpdatedAt = time.Now()

	nodeStatus := &model.NodeDeployStatusRecord{
		Host:             node.Host,
		Service:          deployment.AppName,
		CurrentVersion:   node.CurrentVersion,
		DeployingVersion: deployment.PackageVersion,
		PrevVersion:      node.PrevVersion,
		Platform:         deployment.Platform,
		State:            model.NodeStatusDeploying,
	}

	existing, _ := dm.nodeStatusModel.FindByHostAndService(ctx, node.Host, deployment.AppName)
	if existing != nil {
		nodeStatus.Id = existing.Id
		dm.nodeStatusModel.Update(ctx, nodeStatus)
	} else {
		dm.nodeStatusModel.Insert(ctx, nodeStatus)
	}

	executor, err := dm.executorFactory.CreateExecutor(ctx, executor.ExecutorConfig{
		Platform:    string(deployment.Platform),
		Host:        node.Host,
		IP:          node.IP,
		Service:     deployment.AppName,
		Version:     deployment.PackageVersion,
		PrevVersion: node.PrevVersion,
		PackageURL:  deployment.Package.URL,
		SHA256:      deployment.Package.SHA256,
	})

	if err != nil {
		node.Status = model.NodeStatusFailed
		node.LastError = err.Error()
		nodeStatus.State = model.NodeStatusFailed
		nodeStatus.LastError = err.Error()
		dm.nodeStatusModel.Update(ctx, nodeStatus)
		return err
	}

	if err := executor.Deploy(ctx); err != nil {
		node.Status = model.NodeStatusFailed
		node.LastError = err.Error()
		nodeStatus.State = model.NodeStatusFailed
		nodeStatus.LastError = err.Error()
		dm.nodeStatusModel.Update(ctx, nodeStatus)

		if rollbackErr := executor.Rollback(ctx); rollbackErr != nil {
			node.LastError = fmt.Sprintf("deploy failed: %s, rollback failed: %s", err.Error(), rollbackErr.Error())
			nodeStatus.LastError = node.LastError
		} else {
			node.Status = model.NodeStatusRolledBack
			nodeStatus.State = model.NodeStatusRolledBack
		}
		dm.nodeStatusModel.Update(ctx, nodeStatus)
		return err
	}

	node.Status = model.NodeStatusSuccess
	node.CurrentVersion = deployment.PackageVersion
	node.PrevVersion = node.CurrentVersion
	node.DeployingVersion = ""
	node.UpdatedAt = time.Now()

	nodeStatus.State = model.NodeStatusSuccess
	nodeStatus.CurrentVersion = deployment.PackageVersion
	nodeStatus.PrevVersion = nodeStatus.CurrentVersion
	nodeStatus.DeployingVersion = ""
	dm.nodeStatusModel.Update(ctx, nodeStatus)

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
	return dm.deploymentModel.Update(ctx, deployment)
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
			go dm.executeStages(context.Background(), deployment)
		}
	}

	return nil
}
