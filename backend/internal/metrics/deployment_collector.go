package metrics

import (
	"context"
	"sync"

	"github.com/Z3Labs/Hackathon/backend/internal/model"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeploymentCollector struct {
	deploymentModel model.DeploymentModel
	
	deployingInfo   *prometheus.GaugeVec
	rollingBackInfo *prometheus.GaugeVec
	
	mu sync.Mutex
}

func NewDeploymentCollector(deploymentModel model.DeploymentModel) *DeploymentCollector {
	return &DeploymentCollector{
		deploymentModel: deploymentModel,
		deployingInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "deploying_info",
				Help: "Information about deployments currently in progress",
			},
			[]string{"host", "app", "version"},
		),
		rollingBackInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "rolling_back_info",
				Help: "Information about deployments currently rolling back",
			},
			[]string{"host", "app", "version"},
		),
	}
}

func (c *DeploymentCollector) Describe(ch chan<- *prometheus.Desc) {
	c.deployingInfo.Describe(ch)
	c.rollingBackInfo.Describe(ch)
}

func (c *DeploymentCollector) Collect(ch chan<- prometheus.Metric) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.deployingInfo.Reset()
	c.rollingBackInfo.Reset()

	ctx := context.Background()

	deployingDeployments, err := c.deploymentModel.Search(ctx, &model.DeploymentCond{
		Status: string(model.DeploymentStatusDeploying),
	})
	if err != nil {
		logx.Errorf("Failed to fetch deploying deployments: %v", err)
	} else {
		for _, deployment := range deployingDeployments {
			for _, node := range deployment.NodeDeployments {
				c.deployingInfo.WithLabelValues(
					node.Id,
					deployment.AppName,
					deployment.PackageVersion,
				).Set(1)
			}
		}
	}

	rollingBackDeployments, err := c.deploymentModel.Search(ctx, &model.DeploymentCond{
		Status: string(model.DeploymentStatusRollingBack),
	})
	if err != nil {
		logx.Errorf("Failed to fetch rolling back deployments: %v", err)
	} else {
		for _, deployment := range rollingBackDeployments {
			for _, node := range deployment.NodeDeployments {
				c.rollingBackInfo.WithLabelValues(
					node.Id,
					deployment.AppName,
					deployment.PackageVersion,
				).Set(1)
			}
		}
	}

	c.deployingInfo.Collect(ch)
	c.rollingBackInfo.Collect(ch)
}
