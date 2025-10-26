/**
 * PromQL 查询模板定义
 * 集中管理所有监控指标的查询语句模板
 */

/**
 * PromQL 查询模板集合
 * 每个函数支持传入单个实例、多个实例（数组）或不传（查询所有）
 */
export const PromQL = {
  // CPU相关
  cpuUsage: (instance?: string | string[]): string => {
    const filter = buildInstanceFilter(instance);
    // 如果提供了实例，使用通配符匹配（因为实例可能是 IP:PORT 格式）
    return `100 * (1 - avg(rate(node_cpu_seconds_total{mode="idle"${filter}}[5m])))`;
  },
  
  cpuIdle: (instance?: string | string[]): string => {
    const filter = buildInstanceFilter(instance);
    return `rate(node_cpu_seconds_total{mode="idle"${filter}}[5m])`;
  },
  
  // 内存相关
  memoryUsage: (instance?: string | string[]): string => {
    const filter = buildInstanceFilter(instance);
    return `100 * (1 - (node_memory_MemAvailable_bytes{${filter.slice(1)}} or (node_memory_MemFree_bytes{${filter.slice(1)}} + node_memory_Buffers_bytes{${filter.slice(1)}} + node_memory_Cached_bytes{${filter.slice(1)}})) / node_memory_MemTotal_bytes{${filter.slice(1)}})`;
  },
  
  memoryFree: (instance?: string | string[]): string => {
    const filter = buildInstanceFilter(instance);
    return `node_memory_MemFree_bytes{${filter.slice(1)}}`;
  },
  
  memoryTotal: (instance?: string | string[]): string => {
    const filter = buildInstanceFilter(instance);
    return `node_memory_MemTotal_bytes{${filter.slice(1)}}`;
  },
  
  // 网络相关
  networkReceiveRate: (instance?: string | string[]): string => {
    const filter = buildInstanceFilter(instance);
    return `rate(node_network_receive_bytes_total{${filter.slice(1)}}[5m])`;
  },
  
  networkTransmitRate: (instance?: string | string[]): string => {
    const filter = buildInstanceFilter(instance);
    return `rate(node_network_transmit_bytes_total{${filter.slice(1)}}[5m])`;
  },
  
  networkReceiveBytes: (instance?: string | string[]): string => {
    const filter = buildInstanceFilter(instance);
    return `node_network_receive_bytes_total{${filter.slice(1)}}`;
  },
  
  networkTransmitBytes: (instance?: string | string[]): string => {
    const filter = buildInstanceFilter(instance);
    return `node_network_transmit_bytes_total{${filter.slice(1)}}`;
  },
  
  // 磁盘相关
  diskUsage: (instance?: string | string[]): string => {
    const filter = buildInstanceFilter(instance);
    return `100 - ((node_filesystem_avail_bytes{${filter.slice(1)}} * 100) / node_filesystem_size_bytes{${filter.slice(1)}})`;
  },
  
  diskIoRead: (instance?: string | string[]): string => {
    const filter = buildInstanceFilter(instance);
    return `rate(node_disk_read_bytes_total{${filter.slice(1)}}[5m])`;
  },
  
  diskIoWrite: (instance?: string | string[]): string => {
    const filter = buildInstanceFilter(instance);
    return `rate(node_disk_write_bytes_total{${filter.slice(1)}}[5m])`;
  },
  
  // 负载相关
  loadAverage1m: (instance?: string | string[]): string => {
    const filter = buildInstanceFilter(instance);
    return `node_load1{${filter.slice(1)}}`;
  },
  
  loadAverage5m: (instance?: string | string[]): string => {
    const filter = buildInstanceFilter(instance);
    return `node_load5{${filter.slice(1)}}`;
  },
  
  loadAverage15m: (instance?: string | string[]): string => {
    const filter = buildInstanceFilter(instance);
    return `node_load15{${filter.slice(1)}}`;
  },
} as const;

/**
 * 构建实例过滤器
 * @param instance 单个实例、多个实例数组或undefined
 * @returns PromQL过滤器字符串，始终以逗号开头
 */
function buildInstanceFilter(instance?: string | string[]): string {
  if (!instance) {
    return '';
  }
  
  // 如果是字符串数组
  if (Array.isArray(instance)) {
    if (instance.length === 0) {
      return '';
    }
    if (instance.length === 1) {
      // 单个实例时，使用通配符匹配 hostname 标签
      return `,hostname=~".*${instance[0]}.*"`;
    }
    // 多个实例，构建正则表达式，每个实例使用通配符匹配
    const patterns = instance.map(ip => `.*${ip}.*`).join('|');
    return `,hostname=~"${patterns}"`;
  }
  
  // 单个实例字符串，使用通配符匹配 hostname 标签
  return `,hostname=~".*${instance}.*"`;
}

/**
 * 工具函数：为多个实例构建查询（已废弃，直接传数组即可）
 * @deprecated 直接使用 template(instances) 即可
 */
export function queryMultipleInstances(
  template: (instance?: string | string[]) => string,
  instances: string[]
): string {
  return template(instances);
}

/**
 * 获取指标的单位
 */
export const MetricUnits = {
  cpuUsage: '%',
  memoryUsage: '%',
  networkReceiveRate: 'bytes/s',
  networkTransmitRate: 'bytes/s',
  diskUsage: '%',
} as const;

/**
 * 获取指标的中文名称
 */
export const MetricLabels = {
  cpuUsage: 'CPU使用率',
  memoryUsage: '内存使用率',
  networkReceiveRate: '网络接收速率',
  networkTransmitRate: '网络发送速率',
  diskUsage: '磁盘使用率',
} as const;

