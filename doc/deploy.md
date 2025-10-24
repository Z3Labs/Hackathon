明白，我们把“Layer/层”统一成 **“层”**，保持术语一致，重新整理文档。下面是更新后的综合设计文档：

---

# 发布系统综合设计文档（Plan 层 + Executor 层）

## 1. 系统概览

发布系统目标：支持物理机和未来 K8s 的多版本发布、灰度控制、告警自动回滚，基于 Golang + MongoDB 实现。

核心模块：

| 模块               | 功能                                                  |
| ---------------- | --------------------------------------------------- |
| Plan 层           | 发布计划管理、批次调度、灰度控制、Pacer 发布、全量/回滚策略                   |
| Executor 层       | 节点级发布执行，实现物理机（Ansible + systemd）和 K8s（Client API）部署 |
| MongoDB          | 保存发布计划、节点状态、版本信息                                    |
| Ansible Playbook | 物理机执行逻辑，下载包、校验、systemd 重启、回滚                        |
| K8s Client       | 将 Executor 层逻辑扩展到 K8s 部署                            |

---

## 2. 数据模型

### 2.1 发布计划（Plan）

| 字段          | 含义                                                             |
| ----------- | -------------------------------------------------------------- |
| PlanID      | 发布计划唯一标识                                                       |
| Service     | 发布服务名称                                                         |
| Version     | 发布版本                                                           |
| PackageURL  | 二进制包/镜像下载地址                                                    |
| SHA256      | 包校验值                                                           |
| ReleaseTime | 版本创建时间（用于全量覆盖排序）                                               |
| Pacer       | 批次控制配置（并发节点数、间隔时间）                                             |
| Nodes       | 目标节点列表                                                         |
| State       | `pending` / `deploying` / `success` / `failed` / `rolled_back` |
| CreatedAt   | 创建时间                                                           |
| UpdatedAt   | 更新时间                                                           |

### 2.2 节点状态（NodeStatus）

| 字段               | 含义                                                             |
| ---------------- | -------------------------------------------------------------- |
| Host             | 节点 hostname/IP 或 K8s Pod 名称                                    |
| Service          | 服务名称                                                           |
| CurrentVersion   | 当前运行版本                                                         |
| DeployingVersion | 正在部署的版本                                                        |
| PrevVersion      | 上一次版本（回滚用）                                                     |
| Platform         | `physical` / `k8s`                                             |
| State            | `pending` / `deploying` / `success` / `failed` / `rolled_back` |
| LastError        | 最近失败信息                                                         |
| UpdatedAt        | 更新时间                                                           |

---

## 3. Executor 层设计

### 3.1 Executor 接口

```go
type Executor interface {
    Deploy() error          // 部署指定版本
    Rollback() error        // 回滚到上一个版本
    Status() NodeStatus     // 查询节点状态
}
```

### 3.2 物理机实现（AnsibleExecutor）

* 下载 tar 包到指定目录 `/opt/releases/{{ svc }}/{{ version }}/`
* 文件完整性校验（SHA256）
* 渲染 systemd 配置文件 `/etc/systemd/system/{{ svc }}@{{ version }}.service`
* Reload systemd
* 重启服务
* 回滚上一个版本（PrevVersion）
* 节点状态同步 MongoDB

**核心字段与方法**：

```go
type AnsibleExecutor struct {
    Node        string
    Service     string
    Version     string
    PrevVersion string
    PackageURL  string
    SHA256      string
    Status      NodeStatus
}

func (a *AnsibleExecutor) Deploy() error
func (a *AnsibleExecutor) Rollback() error
func (a *AnsibleExecutor) Status() NodeStatus
```

### 3.3 K8s实现（K8sExecutor，未来扩展）

* 更新 Deployment/StatefulSet 镜像版本
* 等待 Pod Ready
* 回滚至 PrevVersion 镜像
* 节点状态同步 MongoDB

**核心字段与方法**：

```go
type K8sExecutor struct {
    Namespace   string
    Deployment  string
    Version     string
    PrevVersion string
    ImageURL    string
    Status      NodeStatus
}

func (k *K8sExecutor) Deploy() error
func (k *K8sExecutor) Rollback() error
func (k *K8sExecutor) Status() NodeStatus
```

### 3.4 特性

* Executor 层独立，可并发执行
* 单节点同一服务只能运行一个版本
* Plan 层调用统一接口，无需感知底层平台

---

## 4. Plan 层设计

### 4.1 功能

1. 创建发布计划（包含包信息、ReleaseTime、Pacer）
2. 支持灰度节点和全量发布
3. 批次调度控制（Pacer 配置）
4. 收集节点状态并汇总计划状态
5. 自动回滚失败节点
6. 支持多版本同时发布，避免相互覆盖（按 ReleaseTime 决定全量覆盖顺序）

### 4.2 流程

```
创建发布计划
      │
      ▼
批次调度（Pacer 控制）
      │
      ▼
为每个节点创建 NodeExecutor（physical/k8s）
      │
      ▼
Executor.Deploy()
      │
      ├─ 成功 → 更新节点状态 success
      ├─ 失败 → Executor.Rollback() → 更新节点状态 rolled_back/failed
      ▼
汇总节点状态 → 更新计划状态（MongoDB）
```

### 4.3 计划状态

| 状态          | 含义            |
| ----------- | ------------- |
| pending     | 计划创建但未开始执行    |
| deploying   | 部分节点正在执行      |
| success     | 所有节点成功部署      |
| failed      | 有节点部署失败且未回滚成功 |
| rolled_back | 已回滚到上一个版本     |

---

## 5. 并发与批次控制

* Executor 层独立执行节点任务
* Plan 层根据 Pacer 配置控制批次数量和间隔
* 防止节点之间多版本冲突：单节点只允许一个 DeployingVersion
* 节点状态 MongoDB 同步用于计划汇总

---

## 6. 错误处理与回滚策略

| 场景               | 处理方式                             |
| ---------------- | -------------------------------- |
| 下载/校验失败          | 节点状态标记 `failed`，阻止执行             |
| systemd/K8s 部署失败 | 自动回滚到 PrevVersion                |
| 回滚失败             | 节点状态 `failed`，告警通知               |
| 多版本冲突            | 根据 ReleaseTime 决定覆盖顺序，阻止旧版本覆盖新版本 |

---

## 7. MongoDB 状态同步

* 节点状态写入 MongoDB，Plan 层聚合
* 节点状态字段：Host、Service、CurrentVersion、DeployingVersion、PrevVersion、State、LastError
* 发布计划聚合节点状态生成 Stage / Plan 状态

---

## 8. 总结

* **Plan 层**：发布计划、灰度控制、批次调度、汇总状态、回滚策略
* **Executor 层**：节点级执行，抽象接口，物理机和 K8s 两种实现
* **MongoDB**：统一状态存储
* **特点**：

  * 多版本同时发布
  * 单节点单服务版本限制
  * 批次控制避免节点拥塞
  * 状态同步和回滚机制保证发布安全
