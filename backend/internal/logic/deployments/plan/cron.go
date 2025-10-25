package plan

import (
	"context"
	"fmt"

	"github.com/robfig/cron/v3"
)

type PlanCron struct {
	cron            *cron.Cron
	planManager     *PlanManager
	rollbackManager *RollbackManager
}

func NewPlanCron(planManager *PlanManager, rollbackManager *RollbackManager) *PlanCron {
	return &PlanCron{
		cron:            cron.New(),
		planManager:     planManager,
		rollbackManager: rollbackManager,
	}
}

func (pc *PlanCron) Start() error {
	_, err := pc.cron.AddFunc("@every 1m", func() {
		ctx := context.Background()
		
		if err := pc.planManager.ProcessPendingPlans(ctx); err != nil {
			fmt.Printf("process pending plans error: %v\n", err)
		}
		
		if err := pc.planManager.ContinueDeployingPlans(ctx); err != nil {
			fmt.Printf("continue deploying plans error: %v\n", err)
		}
		
		if err := pc.rollbackManager.ContinueRollingBackPlans(ctx); err != nil {
			fmt.Printf("continue rolling back plans error: %v\n", err)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	pc.cron.Start()
	fmt.Println("Plan cron job started, will process plans every minute")
	return nil
}

func (pc *PlanCron) Stop() {
	pc.cron.Stop()
	fmt.Println("Plan cron job stopped")
}
