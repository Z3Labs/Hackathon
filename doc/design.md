## **MVP 总体架构**

```
GitHub (打 Tag) 
      ↓
GitHub Actions / CI
      ↓
Artifact（Binary + Config）上传至存储（S3 / MinIO）
      ↓
Spinnaker
 ├─ 部署到 Kubernetes（容器/服务）
 └─ 部署到物理机/VM（通过 agent / SSH 执行）
      ↓
应用实例
      ↓
VictoriaMetrics
      ↓
Alertmanager / 自定义告警服务
      ↓
多通道通知 + 自动回滚（Spinnaker rollback）
```

核心流程：

1. 打 Tag → CI 构建 → 生成 artifact → 通知 Spinnaker 创建 Release。
2. Spinnaker 根据灰度策略发布到目标环境（K8s 或物理机）。
3. Agent / Sidecar 上报指标到 VictoriaMetrics。
4. 异常告警触发多通道通知（邮件、Slack、企业微信等）。
5. 自动回滚到上一稳定版本。

---

## **用户故事与流程**

### **US1: GitHub Tag 触发 Release**

* GitHub Action Workflow：

  1. 打 tag → CI 构建二进制 + 配置文件。
  2. 上传到 artifact 存储。
  3. 调用 Spinnaker API 创建 Release 记录。
* 数据记录：

  * Release ID、tag、artifact URL、checksum、创建人、创建时间。

---

### **US2: 服务发布（灰度）**

#### **2.1 部署策略**

* **Kubernetes**：

  * Spinnaker 使用 Deployment/DaemonSet/StatefulSet 发布。
  * 支持灰度发布（Canary / Blue-Green）。
* **物理机 / VM**：

  * 部署 Agent 或通过 SSH/Ansible 拉取 artifact。
  * 按灰度策略逐批发布（例如 5% → 25% → 100%）。
* **灰度告警**：

  * Spinnaker 在每批发布前配置灰度监控规则。
  * Agent 或服务指标上报到 VictoriaMetrics。
  * VictoriaMetrics 告警规则针对当前发布版本新增。

#### **2.2 灰度告警规则示例**

* 基于请求错误率、响应延迟、CPU/内存异常。
* Pseudocode:

  ```
  if error_rate > threshold OR p95_latency > threshold:
      trigger_alert()
  ```

---

### **US3: 告警与回滚**

#### **3.1 异常告警**

* VictoriaMetrics 告警 → Alertmanager 或自定义 webhook。
* 多通道通知：

  * 邮件、Slack、企业微信、钉钉。
* 告警内容：

  * Release ID / 部署版本
  * 受影响节点列表
  * 异常指标与趋势图

#### **3.2 自动回滚**

* 回滚触发方式：

  * Alertmanager webhook 调用 Spinnaker rollback API。
* 回滚策略：

  1. 灰度阶段异常 → 仅回滚受影响批次。
  2. 全量部署异常 → 全量回滚到上一稳定 release。
* 回滚后：

  * 自动通知运维团队。
  * 生成 incident report（指标快照 + 部署日志）。

---

## **核心技术选型**

| 功能               | 技术选型 / 工具                          |
| ---------------- | ---------------------------------- |
| CI 构建 + Artifact | GitHub Actions + MinIO/S3          |
| 发布系统             | Spinnaker                          |
| 监控指标             | VictoriaMetrics                    |
| 告警               | VictoriaMetrics Alerting + Webhook |
| 灰度发布             | Spinnaker Canary / Blue-Green      |
| 回滚               | Spinnaker Rollback API             |
| 通知               | 多通道：邮件 / Slack / 企业微信 / 钉钉         |
| 部署 Agent (物理机)   | Golang/Ansible + health check      |

---

## **MVP 数据模型（最小）**

**Release**

* id, tag, version, artifact_url, checksum, created_at, created_by

**Deployment**

* id, release_id, env (prod/stage), strategy (canary/all), target_hosts, status (pending/running/success/failed/rolled_back), started_at, finished_at

**InstanceDeploymentRecord**

* host, ip, status, deployed_at, health_status, error_message

**Alert**

* release_id, deployment_id, metric, severity, timestamp, message

---

## **3 天 Hackathon 交付计划**

| 时间       | 任务                                               |
| -------- | ------------------------------------------------ |
| Day 0 上午 | 环境准备：Spinnaker、VictoriaMetrics、MinIO/S3、数据库      |
| Day 0 下午 | CI 构建 + GitHub Tag 流程 → artifact 上传 + Release 创建 |
| Day 1 上午 | 部署 Agent/SSH 脚本 + Kubernetes 部署 pipeline 配置      |
| Day 1 下午 | 灰度发布逻辑实现 + 部署控制台可视化                              |
| Day 2 上午 | 指标上报、VictoriaMetrics 告警规则创建 + 多通道通知              |
| Day 2 下午 | 自动回滚逻辑 + Spinnaker rollback 流程测试                 |
| Day 3 全天 | 全链路演示 + bug fix + hackathon demo                 |

---

## **MVP 特性总结**

* ✅ 支持 Kubernetes + 物理机/VM 多版本发布
* ✅ GitHub Tag → CI 构建 → Artifact → Spinnaker Release
* ✅ 灰度发布与灰度告警（VictoriaMetrics）
* ✅ 异常告警触发多通道通知
* ✅ 自动回滚至上一稳定版本
* ✅ 最小可交付可在 3 天内完成