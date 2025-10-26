# PromQL 测试查询

为了排查问题，建议在 VictoriaMetrics UI 中测试以下查询：

## 1. 检查所有实例
```
node_cpu_seconds_total
```
这应该返回所有有 CPU 数据的实例

## 2. 查看实例标签值
```
{__name__="node_cpu_seconds_total"}
```
查看实例标签的实际格式

## 3. 不带实例过滤的查询
```
100 * (1 - avg(rate(node_cpu_seconds_total{mode="idle"}[5m])))
```

## 4. 可能需要的格式
如果实例格式是 `IP:PORT`（比如 `150.158.152.112:9100`），那么查询应该是：
```
100 * (1 - avg(rate(node_cpu_seconds_total{mode="idle",instance="150.158.152.112:9100"}[5m])))
```

## 5. 或者实例可能需要通配符匹配
```
100 * (1 - avg(rate(node_cpu_seconds_total{mode="idle",instance=~".*150\\.158\\.152\\.112.*"}[5m])))
```

