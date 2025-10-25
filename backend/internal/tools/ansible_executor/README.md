# Ansible Executor 测试工具

用于测试和调试 Ansible Executor 的命令行工具。

## 功能

- **部署测试**: 执行服务部署操作
- **回滚测试**: 执行版本回滚操作
- **状态查看**: 查询当前部署状态

## 编译

```bash
cd backend/internal/tools/ansible_executor
go build -o ansible-executor-tool
```

## 使用方法

### 1. 部署操作

```bash
./ansible-executor-tool \
  -action=deploy \
  -host=192.168.1.100 \
  -ip=192.168.1.100 \
  -service=my-service \
  -version=v1.2.0 \
  -prev-version=v1.1.0 \
  -package-url=http://example.com/packages/my-service-v1.2.0.tar.gz \
  -sha256=abc123def456... \
  -playbook=/path/to/deploy.yml
```

### 2. 回滚操作

```bash
./ansible-executor-tool \
  -action=rollback \
  -host=192.168.1.100 \
  -service=my-service \
  -version=v1.2.0 \
  -prev-version=v1.1.0 \
  -package-url=http://example.com/packages/my-service-v1.1.0.tar.gz \
  -sha256=xyz789abc123...
```

### 3. 查看状态

```bash
./ansible-executor-tool \
  -action=status \
  -host=192.168.1.100 \
  -service=my-service
```

## 参数说明

| 参数 | 说明 | 是否必填 | 默认值 |
|------|------|----------|--------|
| `-action` | 操作类型: deploy, rollback, status | 否 | deploy |
| `-host` | 目标主机地址 | 是 | - |
| `-ip` | 目标机器 IP (用于 ansible -i 参数) | 否 | - |
| `-service` | 服务名称 | 是 | - |
| `-version` | 部署版本 | deploy/rollback时必填 | - |
| `-prev-version` | 上一个版本 | rollback时必填 | - |
| `-package-url` | 软件包 URL | deploy/rollback时必填 | - |
| `-sha256` | 软件包 SHA256 校验和 | deploy/rollback时必填 | - |
| `-playbook` | Ansible playbook 文件路径 | 否 | /workspace/backend/playbooks/deploy.yml |
| `-timeout` | 执行超时时间（秒） | 否 | 600 |

## 环境要求

1. 已安装 Ansible
2. 目标主机已配置 SSH 免密登录
3. Ansible playbook 文件存在且可访问

## 示例输出

### 部署成功

```
=== Ansible Executor 测试工具 ===

操作类型: deploy
目标主机: 192.168.1.100
服务名称: my-service
部署版本: v1.2.0
上一版本: v1.1.0
软件包URL: http://example.com/packages/my-service-v1.2.0.tar.gz
SHA256: abc123def456...
Playbook: /workspace/backend/playbooks/deploy.yml

================================

开始执行部署...
[Ansible输出...]

✅ 操作成功 (耗时: 45.2s)

当前状态:
  主机: 192.168.1.100
  服务: my-service
  当前版本: v1.2.0
  部署中版本: 
  上一版本: v1.2.0
  平台: physical
  状态: success
  更新时间: 2024-01-20 10:30:45
```

### 部署失败

```
开始执行部署...
[Ansible输出...]

❌ 操作失败 (耗时: 12.5s)
错误详情: failed to execute ansible-playbook: exit status 1
```

## 调试技巧

1. **查看详细日志**: Ansible 命令使用了 `-v` 选项，会输出详细的执行信息
2. **测试连接**: 在运行工具前，可以先用 `ansible` 命令测试主机连接
3. **验证 playbook**: 使用 `ansible-playbook --syntax-check` 验证 playbook 语法
4. **超时设置**: 对于大型部署，可以适当增加 `-timeout` 值

## 故障排查

### 常见问题

1. **连接失败**: 检查 SSH 配置和网络连接
2. **权限错误**: 确保目标主机上有足够的权限执行部署操作
3. **Playbook 不存在**: 检查 `-playbook` 参数指定的路径是否正确
4. **超时**: 增加 `-timeout` 参数值或检查网络速度
