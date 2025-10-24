package model

const (
	// 集合名称
	CollectionApplication = "application" // 应用
	CollectionDeployment  = "deployment"  // 发布
	CollectionMachine     = "machine"     // 机器
)

type (
	HealthStatus     string // 健康状态
	ErrorStatus      string // 异常状态
	AlertStatus      string // 告警状态
	ReleaseStatus    string // 发布状态
	DeploymentStatus string // 部署状态
	GrayStrategy     string // 灰度策略
)

const (
	HealthStatusHealthy   HealthStatus = "healthy"   // 健康
	HealthStatusUnhealthy HealthStatus = "unhealthy" // 不健康

	ErrorStatusNormal ErrorStatus = "normal" // 正常
	ErrorStatusError  ErrorStatus = "error"  // 异常

	AlertStatusNormal AlertStatus = "normal" // 正常
	AlertStatusAlert  AlertStatus = "alert"  // 告警

	ReleaseStatusPending   ReleaseStatus = "pending"   // 待发布
	ReleaseStatusDeploying ReleaseStatus = "deploying" // 发布中
	ReleaseStatusSuccess   ReleaseStatus = "success"   // 成功
	ReleaseStatusFailed    ReleaseStatus = "failed"    // 失败

	DeploymentStatusPending    DeploymentStatus = "pending"     // 待发布
	DeploymentStatusDeploying  DeploymentStatus = "deploying"   // 发布中
	DeploymentStatusSuccess    DeploymentStatus = "success"     // 成功
	DeploymentStatusFailed     DeploymentStatus = "failed"      // 失败
	DeploymentStatusRolledBack DeploymentStatus = "rolled_back" // 已回滚

	GrayStrategyCanary    GrayStrategy = "canary"     // 金丝雀发布
	GrayStrategyBlueGreen GrayStrategy = "blue-green" // 蓝绿发布
	GrayStrategyAll       GrayStrategy = "all"        // 全量发布
)
