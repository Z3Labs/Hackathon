package main

import (
	"context"
	"flag"
	"fmt"

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
		Path:    "/metrics",
		Handler: promhttp.Handler().ServeHTTP,
	})

	deploymentManager := deployments.NewDeploymentManager(context.Background(), ctx)
	rollbackManager := deployments.NewRollbackManager(context.Background(), ctx)
	
	deploymentCron := deployments.NewDeploymentCron(deploymentManager, rollbackManager)
	if err := deploymentCron.Start(); err != nil {
		panic(fmt.Sprintf("failed to start deployment cron: %v", err))
	}
	defer deploymentCron.Stop()

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
