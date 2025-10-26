import api from './api';

export const monitoringService = {
  // 通用指标查询接口
  async queryMetrics(params: {
    query: string;  // PromQL查询语句
    start: string;  // 开始时间（Unix时间戳）
    end: string;    // 结束时间（Unix时间戳）
    step?: string;  // 采样间隔，默认60s
  }): Promise<{ series: any[] }> {
    return api.get('/metrics/query', { params });
  },
};

