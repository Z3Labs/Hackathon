# 发布计划（ReleasePlan）详细设计文档

## 1. 模块定位

* **模块类型**：Plan Layer 核心模块
* **功能**：

  1. 管理发布计划生命周期（创建/执行/回滚/完成）
  2. 按 Stage 控制灰度/全量发布
  3. 按 Pacer 控制批次速度，避免大批量节点同时发布
  4. 节点状态跟踪与汇总，用于调度和回滚决策

---

## 2. 数据模型

### 2.1 ReleasePlan

```json
{
  "_id": "release_plan_001",
  "svc": "myapp",
  "target_version": "v1.2.4",
  "release_time": "2025-10-24T17:00:00Z",
  "package": {
      "url": "https://kodo.example.com/myapp/v1.2.4/myapp.tar.gz",
      "sha256": "abcdef1234567890...",
      "size": 104857600,
      "created_at": "2025-10-23T12:00:00Z"
  },
  "stages": [
    {
      "name": "gray",
      "nodes": [
        {"host":"host1","status":"pending","current_version":"v1.2.3","deploying_version":"","prev_version":"v1.2.3"},
        {"host":"host2","status":"pending","current_version":"v1.2.3","deploying_version":"","prev_version":"v1.2.3"}
      ],
      "status": "pending",
      "pacer": {"batch_size":1,"interval_seconds":30}
    },
    {
      "name": "full",
      "nodes": [
        {"host":"host3","status":"pending","current_version":"v1.2.3","deploying_version":"","prev_version":"v1.2.3"},
        {"host":"host4","status":"pending","current_version":"v1.2.3","deploying_version":"","prev_version":"v1.2.3"}
      ],
      "status": "pending",
      "pacer": {"batch_size":2,"interval_seconds":60}
    }
  ],
  "status": "pending"
}
```

### 2.2 字段说明

| 字段               | 类型       | 含义                                                                                           |
| ---------------- | -------- | -------------------------------------------------------------------------------------------- |
| `_id`            | string   | 发布计划唯一 ID                                                                                    |
| `svc`            | string   | 服务名称                                                                                         |
| `target_version` | string   | 本次发布目标版本                                                                                     |
| `release_time`   | datetime | 版本创建时间，决定全量覆盖优先级                                                                             |
| `package`        | object   | 发布包信息（URL、sha256、大小、创建时间）                                                                    |
| `stages`         | array    | 发布阶段列表，每个阶段包含节点、状态和 Pacer 配置                                                                 |
| `status`         | string   | 发布计划整体状态（pending/deploying/partial_success/success/failed/rolling_back/rolled_back/canceled） |

### 2.3 节点状态字段

| 字段                  | 含义                                                   |
| ------------------- | ---------------------------------------------------- |
| `status`            | 节点执行状态（pending/deploying/success/failed/rolled_back） |
| `current_version`   | 当前节点运行版本                                             |
| `deploying_version` | 正在发布的版本                                              |
| `prev_version`      | 上一个版本，用于回滚                                           |

---

## 3. 生命周期状态（Plan Level）

### 3.1 Plan 状态枚举

| 状态                | 含义            | 触发条件                 |
| ----------------- | ------------- | -------------------- |
| `pending`         | 发布计划已创建，但未执行  | Plan 创建完成            |
| `deploying`       | 发布计划正在执行      | Stage 首批节点开始执行       |
| `partial_success` | 部分节点成功，部分节点失败 | 节点部分失败但未触发回滚         |
| `success`         | 所有节点成功        | 所有节点状态为 success      |
| `failed`          | 节点全部失败或回滚完成   | 节点执行失败且回滚完成          |
| `rolling_back`    | 回滚中           | 收到告警或手动触发回滚          |
| `rolled_back`     | 回滚完成          | 节点状态均恢复 prev_version |
| `canceled`        | 人工取消发布        | 发布未完成被取消             |

### 3.2 Stage 状态

* 每个阶段也有状态：`pending` / `deploying` / `success` / `failed`
* 阶段状态由节点状态汇总决定

---

## 4. Golang Client 核心接口

```go
type US2Client struct {
    PlanService     *PlanManager
    RollbackService *RollbackManager
}

// 创建发布计划
func (c *US2Client) CreateReleasePlan(
    svc string,
    version string,
    pkg PackageInfo,
    stages []Stage,
) (*ReleasePlan, error)

// 触发全量发布
func (c *US2Client) TriggerFullRelease(planID string) error

// 回滚指定节点或整个计划
func (c *US2Client) RollbackRelease(planID string, nodes []string) error
```

---

## 5. 发布计划调度策略

1. **批次发布（Pacer）**

   * 每个 Stage 按 `batch_size` + `interval_seconds` 控制节点发布速度
   * 批次间节点任务可并发执行

2. **多版本发布控制**

   * 同一节点只允许单版本部署（通过 `deploying_version` 字段控制）
   * 全量覆盖以 `release_time` 决定优先级，旧版本不能覆盖新版本

3. **回滚策略**

   * 节点执行失败或告警触发回滚
   * 回滚任务独立执行，不影响其他节点
   * 回滚完成后更新节点状态为 `rolled_back`

---

## 6. 状态更新逻辑

* **Plan 状态**：汇总节点和阶段状态
* **Stage 状态**：汇总该阶段节点状态
* **Node 状态**：独立更新，Plan Scheduler 根据节点状态调度下一批次
* **MongoDB**：集中记录 Plan、Stage、Node 状态，Plan Scheduler 查询完成情况控制发布流程

---

## 7. 并发执行策略

* Node Executor 每个节点任务互不依赖，可并发执行
* Plan Layer 负责阶段调度和批次触发
* 单节点单实例，状态字段防止多版本冲突
* 回滚任务独立触发，不阻塞其他节点执行