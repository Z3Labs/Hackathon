package model

const (
	// 集合名称
	CollectionApplication = "application"  // 应用
	CollectionDeployment  = "deployment"   // 发布
	CollectionMachine     = "machine"      // 机器
	CollectionReleasePlan = "release_plan" // 发布计划
	CollectionNodeStatus  = "node_status"  // 节点状态
)

type (
	HealthStatus     string // 健康状态
	ErrorStatus      string // 异常状态
	AlertStatus      string // 告警状态
	ReleaseStatus    string // 发布状态
	DeploymentStatus string // 部署状态
	GrayStrategy     string // 灰度策略
	PlanStatus       string // 发布计划状态
	StageStatus      string // 阶段状态
	NodeStatus       string // 节点状态
	PlatformType     string // 平台类型
)

const (
	HealthStatusHealthy   HealthStatus = "healthy"   // 健康
	HealthStatusUnhealthy HealthStatus = "unhealthy" // 不健康

	ErrorStatusNormal ErrorStatus = "normal" // 正常
	ErrorStatusError  ErrorStatus = "error"  // 异常

	AlertStatusNormal AlertStatus = "normal" // 正常
	AlertStatusAlert  AlertStatus = "alert"  // 告警

	ReleaseStatusPending    ReleaseStatus = "pending"     // 待发布
	ReleaseStatusDeploying  ReleaseStatus = "deploying"   // 发布中
	ReleaseStatusSuccess    ReleaseStatus = "success"     // 成功
	ReleaseStatusRolledBack ReleaseStatus = "rolled_back" // 已回滚
	ReleaseStatusFailed     ReleaseStatus = "failed"      // 失败

	DeploymentStatusPending    DeploymentStatus = "pending"     // 待发布
	DeploymentStatusDeploying  DeploymentStatus = "deploying"   // 发布中
	DeploymentStatusSuccess    DeploymentStatus = "success"     // 成功
	DeploymentStatusFailed     DeploymentStatus = "failed"      // 失败
	DeploymentStatusRolledBack DeploymentStatus = "rolled_back" // 已回滚

	GrayStrategyCanary    GrayStrategy = "canary"     // 金丝雀发布
	GrayStrategyBlueGreen GrayStrategy = "blue-green" // 蓝绿发布
	GrayStrategyAll       GrayStrategy = "all"        // 全量发布

	PlanStatusPending        PlanStatus = "pending"         // 待发布
	PlanStatusDeploying      PlanStatus = "deploying"       // 发布中
	PlanStatusPartialSuccess PlanStatus = "partial_success" // 部分成功
	PlanStatusSuccess        PlanStatus = "success"         // 成功
	PlanStatusFailed         PlanStatus = "failed"          // 失败
	PlanStatusRollingBack    PlanStatus = "rolling_back"    // 回滚中
	PlanStatusRolledBack     PlanStatus = "rolled_back"     // 已回滚
	PlanStatusCanceled       PlanStatus = "canceled"        // 已取消

	StageStatusPending   StageStatus = "pending"   // 待执行
	StageStatusDeploying StageStatus = "deploying" // 执行中
	StageStatusSuccess   StageStatus = "success"   // 成功
	StageStatusFailed    StageStatus = "failed"    // 失败

	NodeStatusPending    NodeStatus = "pending"     // 待部署
	NodeStatusDeploying  NodeStatus = "deploying"   // 部署中
	NodeStatusSuccess    NodeStatus = "success"     // 成功
	NodeStatusFailed     NodeStatus = "failed"      // 失败
	NodeStatusRolledBack NodeStatus = "rolled_back" // 已回滚

	PlatformMock     PlatformType = "mock"     // !!! 仅测试使用
	PlatformPhysical PlatformType = "physical" // 物理机
	PlatformK8s      PlatformType = "k8s"      // K8s
)
