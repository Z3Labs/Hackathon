package plan

import (
	"context"
	"fmt"

	"github.com/robfig/cron/v3"
)

type PlanCron struct {
	cron        *cron.Cron
	planManager *PlanManager
}

func NewPlanCron(planManager *PlanManager) *PlanCron {
	return &PlanCron{
		cron:        cron.New(),
		planManager: planManager,
	}
}

func (pc *PlanCron) Start() error {
	_, err := pc.cron.AddFunc("@every 1m", func() {
		ctx := context.Background()
		if err := pc.planManager.ProcessPendingPlans(ctx); err != nil {
			fmt.Printf("cron job error: %v\n", err)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	pc.cron.Start()
	fmt.Println("Plan cron job started, will check pending plans every minute")
	return nil
}

func (pc *PlanCron) Stop() {
	pc.cron.Stop()
	fmt.Println("Plan cron job stopped")
}
