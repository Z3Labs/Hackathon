# Node Executor 接口设计

## 1. Executor 接口

```go
package executor

type Executor interface {
    // Deploy 部署指定版本服务
    Deploy() error

    // Rollback 回滚到上一个版本
    Rollback() error

    // Status 获取当前节点状态
    Status() NodeStatus
}
```

### NodeStatus 结构体

```go
type NodeStatus struct {
    Host           string // 节点或 Pod 名称
    Service        string
    CurrentVersion string
    DeployingVersion string
    PrevVersion    string
    Platform       string // physical | k8s
    State          string // pending | deploying | success | failed | rolled_back
    LastError      string
}
```

---

## 2. 物理机实现（AnsibleExecutor）

```go
package executor

type AnsibleExecutor struct {
    Node       string
    Service    string
    Version    string
    PrevVersion string
    PackageURL string
    SHA256     string
    Status     NodeStatus
}

// Deploy 使用 ansible playbook 部署物理机
func (a *AnsibleExecutor) Deploy() error {
    // 调用 ansible-playbook 执行下载、校验、systemd 重启
    // 更新 a.Status
}

// Rollback 回滚物理机服务
func (a *AnsibleExecutor) Rollback() error {
    // 调用 ansible-playbook 回滚至 PrevVersion
    // 更新 a.Status
}

func (a *AnsibleExecutor) Status() NodeStatus {
    return a.Status
}
```

### 特点

* 单节点只能运行一个版本，使用 systemd 服务名加版本控制
* 包下载、SHA256 校验、systemd 重启由 playbook 统一管理
* 状态同步 MongoDB

---

## 3. K8s 实现（K8sExecutor）

```go
package executor

type K8sExecutor struct {
    Namespace   string
    Deployment  string
    Version     string
    PrevVersion string
    ImageURL    string
    Status      NodeStatus
}

// Deploy 更新 Deployment 镜像
func (k *K8sExecutor) Deploy() error {
    // 使用 client-go 更新 Deployment 镜像
    // 等待 Pod Ready
    // 更新 k.Status
}

// Rollback 回滚 Deployment 至 PrevVersion 镜像
func (k *K8sExecutor) Rollback() error {
    // 使用 client-go 回滚 Deployment
    // 更新 k.Status
}

func (k *K8sExecutor) Status() NodeStatus {
    return k.Status
}
```

### 特点

* 支持 Deployment/StatefulSet 镜像更新
* 可以与 Plan Layer 批次发布、Pacer 控制并发发布保持一致
* 节点状态同步 MongoDB
* 与物理机接口完全一致，Plan Layer 调用无需感知底层平台

---

## 4. Plan Layer 调用示例

```go
var exec executor.Executor

if platform == "physical" {
    exec = &executor.AnsibleExecutor{
        Node: node,
        Service: svc,
        Version: version,
        PrevVersion: prevVersion,
        PackageURL: pkgURL,
        SHA256: sha256,
    }
} else if platform == "k8s" {
    exec = &executor.K8sExecutor{
        Namespace: ns,
        Deployment: svc,
        Version: version,
        PrevVersion: prevVersion,
        ImageURL: imageURL,
    }
}

// 部署
err := exec.Deploy()
if err != nil {
    // 自动回滚
    exec.Rollback()
}

// 获取状态
status := exec.Status()
```

---

## 5. 优点

1. **接口抽象**：Plan Layer 调用统一接口，无需关心物理机或 K8s
2. **可扩展**：新增平台（Docker Swarm、Nomad 等）只需实现 `Executor` 接口
3. **统一状态管理**：MongoDB 统一记录节点状态，支持批次调度和回滚逻辑
4. **平台独立**：AnsibleExecutor 与 K8sExecutor 可以独立开发、测试