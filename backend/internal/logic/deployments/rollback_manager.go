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

type RollbackManager struct {
	deploymentModel  model.DeploymentModel
	applicationModel model.ApplicationModel
	executorFactory  executor.ExecutorFactoryInterface
	taskRegistry     map[string]context.CancelFunc
	taskMutex        sync.RWMutex
}

func NewRollbackManager(ctx context.Context, svcCtx *svc.ServiceContext) *RollbackManager {
	return &RollbackManager{
		deploymentModel:  svcCtx.DeploymentModel,
		applicationModel: svcCtx.ApplicationModel,
		executorFactory:  executor.NewExecutorFactory(),
		taskRegistry:     make(map[string]context.CancelFunc),
	}
}

func (rm *RollbackManager) executeRollback(ctx context.Context, deployment *model.Deployment, nodes []string, prevVersion string) int {
	if len(nodes) == 0 {
		return 0
	}
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex
	for _, id := range nodes {
		logx.Infof("start rolling back deployment:%s, node:%s", deployment.Id, id)
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			err := rm.rollbackNode(ctx, deployment, id, prevVersion)
			if err != nil {
				logx.Errorf("rolling back deployment:%s, node:%s, failed, err = %s ", deployment.Id, id, err)
			} else {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(id)
	}

	wg.Wait()
	return successCount
}

func (rm *RollbackManager) rollbackNode(ctx context.Context, deployment *model.Deployment, id string, preVersion string) error {

	nodeIndex := findNodeIndex(deployment.NodeDeployments, id)
	if nodeIndex < 0 {
		return fmt.Errorf("node index out of range")
	}
	if preVersion == "" {
		logx.Info("bad version, version = %s", preVersion)
		return fmt.Errorf("invalid prev version")
	}
	node := &deployment.NodeDeployments[nodeIndex]
	executor, err := rm.executorFactory.CreateExecutor(ctx, executor.ExecutorConfig{
		Platform:    string(node.Platform),
		Host:        node.Id,
		IP:          node.Ip,
		Service:     deployment.AppName,
		Version:     node.CurrentVersion,
		PrevVersion: preVersion,
		PackageURL:  deployment.Package.URL,
		MD5:         deployment.Package.MD5,
	})

	if err != nil {
		node.NodeDeployStatus = model.NodeDeploymentStatusFailed
		node.ReleaseLog = err.Error()
		node.UpdatedAt = time.Now()
		rm.deploymentModel.Update(context.Background(), deployment)
		return err
	}
	logx.Infof("%s@%s start rolling_back", deployment.AppName, node.Id)
	if err := executor.Rollback(ctx); err != nil {
		if ctx.Err() != nil {
			node.NodeDeployStatus = model.NodeDeploymentStatusFailed
			node.ReleaseLog = "rollback canceled"
			node.UpdatedAt = time.Now()
			rm.deploymentModel.Update(context.Background(), deployment)
			return ctx.Err()
		}

		node.NodeDeployStatus = model.NodeDeploymentStatusFailed
		node.ReleaseLog = fmt.Sprintf("rollback failed: %s", err.Error())
		node.UpdatedAt = time.Now()
		rm.deploymentModel.Update(context.Background(), deployment)
		return err
	}

	node.NodeDeployStatus = model.NodeDeploymentStatusRolledBack
	node.ReleaseLog = "rollback successful"
	node.CurrentVersion = node.PrevVersion
	node.DeployingVersion = ""
	node.UpdatedAt = time.Now()

	deployment, err = rm.deploymentModel.FindById(context.Background(), deployment.Id)
	if err != nil {
		return err
	}
	// 避免影响其他节点
	for i := range deployment.NodeDeployments {
		if deployment.NodeDeployments[i].Id == node.Id {
			deployment.NodeDeployments[i] = *node
		}
	}
	rm.deploymentModel.Update(context.Background(), deployment)

	return nil
}

func (rm *RollbackManager) ContinueRollingBackDeployments(ctx context.Context) error {

	// 发布中任务，单节点回滚
	deployments, err := rm.deploymentModel.Search(ctx, &model.DeploymentCond{Status: string(model.DeploymentStatusDeploying)})
	if err != nil {
		return fmt.Errorf("failed to search deployments with status %s: %w", model.DeploymentStatusDeploying, err)
	}

	for _, deployment := range deployments {
		var nodesToRollback []string
		for _, node := range deployment.NodeDeployments {
			if node.NodeDeployStatus == model.NodeDeploymentStatusRollingBack {
				nodesToRollback = append(nodesToRollback, node.Id)
			}
		}
		app, err := rm.applicationModel.FindById(ctx, deployment.AppId)
		if err != nil {
			logx.Errorf("find app failed, errr = %s", err)
		}
		if len(nodesToRollback) > 0 {
			rm.executeRollback(ctx, deployment, nodesToRollback, app.CurrentVersion)
		}
	}
	// 整个发布回滚
	deployments, err = rm.deploymentModel.Search(ctx, &model.DeploymentCond{Status: string(model.DeploymentStatusRollingBack)})
	if err != nil {
		return fmt.Errorf("failed to search deployments with status %s: %w", model.DeploymentStatusRollingBack, err)
	}

	for _, deployment := range deployments {
		var nodesToRollback []string
		for _, node := range deployment.NodeDeployments {
			if node.NodeDeployStatus == model.NodeDeploymentStatusSuccess {
				nodesToRollback = append(nodesToRollback, node.Id)
			}
		}
		app, err := rm.applicationModel.FindById(ctx, deployment.AppId)
		if err != nil {
			logx.Errorf("find app failed, errr = %s", err)
		}
		if len(nodesToRollback) > 0 {
			succCount := rm.executeRollback(ctx, deployment, nodesToRollback, app.PrevVersion)
			// 发布单级别回滚需要更新发布单整体状态
			if succCount == len(nodesToRollback) {
				deployment.Status = model.DeploymentStatusRolledBack
				app.CurrentVersion = app.PrevVersion
				rm.applicationModel.Update(ctx, app)
			} else {
				deployment.Status = model.DeploymentStatusFailed
			}
			rm.deploymentModel.UpdateStatus(context.Background(), deployment.Id, deployment.Status)
		}

	}
	return nil
}
