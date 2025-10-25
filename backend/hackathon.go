package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/Z3Labs/Hackathon/backend/internal/clients/prom"
	"github.com/Z3Labs/Hackathon/backend/internal/config"
	"github.com/Z3Labs/Hackathon/backend/internal/handler"
	"github.com/Z3Labs/Hackathon/backend/internal/logic/deployments"
	"github.com/Z3Labs/Hackathon/backend/internal/metrics"
	"github.com/Z3Labs/Hackathon/backend/internal/svc"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/hackathon-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf, rest.WithCors("*"))
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	collector := metrics.NewDeploymentCollector(ctx.DeploymentModel)
	prometheus.MustRegister(collector)
	
	server.AddRoute(rest.Route{
		Method:  "GET",
		Path:    "/deploy/metrics",
		Handler: promhttp.Handler().ServeHTTP,
	})

	deploymentManager := deployments.NewDeploymentManager(context.Background(), ctx)
	rollbackManager := deployments.NewRollbackManager(context.Background(), ctx)
	
	var alertMonitor *deployments.AlertMonitor
	if c.AI.PrometheusURL != "" {
		promClient := prom.NewVMClient(prom.NewDefaultConfig(c.AI.PrometheusURL))
		alertMonitor = deployments.NewAlertMonitor(ctx, promClient)
		deploymentManager.SetAlertMonitor(alertMonitor)
		fmt.Println("Alert monitor initialized with Prometheus URL:", c.AI.PrometheusURL)
	} else {
		fmt.Println("Alert monitor disabled: no Prometheus URL configured")
	}
	
	deploymentCron := deployments.NewDeploymentCron(deploymentManager, rollbackManager, alertMonitor)
	if err := deploymentCron.Start(); err != nil {
		panic(fmt.Sprintf("failed to start deployment cron: %v", err))
	}
	defer deploymentCron.Stop()

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
