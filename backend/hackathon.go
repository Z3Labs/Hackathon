package main

import (
	"flag"
	"fmt"

	"github.com/Z3Labs/Hackathon/backend/internal/config"
	"github.com/Z3Labs/Hackathon/backend/internal/handler"
	"github.com/Z3Labs/Hackathon/backend/internal/logic/deployments/executor"
	"github.com/Z3Labs/Hackathon/backend/internal/logic/deployments/plan"
	"github.com/Z3Labs/Hackathon/backend/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/hackathon-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	executorFactory := executor.NewExecutorFactory()
	planManager := plan.NewPlanManager(ctx.ReleasePlanModel, ctx.NodeStatusModel, executorFactory)
	
	planCron := plan.NewPlanCron(planManager)
	if err := planCron.Start(); err != nil {
		panic(fmt.Sprintf("failed to start plan cron: %v", err))
	}
	defer planCron.Stop()

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
