import React, { useEffect, useState } from 'react';
import { deploymentService } from '../services/deployment';
import type { Deployment } from '../types/deployment';
import { monitoringService } from '../services/monitoring';
import { appApi } from '../services/api';
import REDMetricsMiniChart from './common/REDMetricsMiniChart';
import type { REDMetrics } from '../types';

interface DeploymentListProps {
  onSelectDeployment?: (deployment: Deployment) => void;
  onCreateNew?: () => void;
}

interface REDMetricsData {
  rate?: Array<{ timestamp: number; value: number }>;
  error?: Array<{ timestamp: number; value: number }>;
  duration?: Array<{ timestamp: number; value: number }>;
  rateLatest?: number;
  errorLatest?: number;
  durationLatest?: number;
}

const DeploymentList: React.FC<DeploymentListProps> = ({ onSelectDeployment, onCreateNew }) => {
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [pageSize] = useState(10);
  const [appNameFilter, setAppNameFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [redMetricsData, setRedMetricsData] = useState<Record<string, REDMetricsData>>({});
  const [redMetricsConfigs, setRedMetricsConfigs] = useState<Record<string, REDMetrics>>({});

  const fetchDeployments = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await deploymentService.getDeploymentList({
        page,
        page_size: pageSize,
        app_name: appNameFilter || undefined,
        status: statusFilter || undefined,
      });
      setDeployments(response.deployments || []);
      setTotal(response.total);
    } catch (err) {
      setError('获取发布记录失败');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  // 获取单个部署的RED指标数据
  const fetchDeploymentREDMetrics = async (deployment: Deployment) => {
    try {
      // 获取应用配置
      const appResult = await appApi.getAppList({ name: deployment.app_name, page: 1, page_size: 1 }) as any;
      if (!appResult || !appResult.apps || appResult.apps.length === 0) return;
      
      const app = appResult.apps[0];
      if (!app.red_metrics_config || !app.red_metrics_config.enabled) return;
      
      const redMetrics = app.red_metrics_config;
      setRedMetricsConfigs(prev => ({ ...prev, [deployment.id]: redMetrics }));
      
      // 获取最新30分钟的RED指标数据
      const now = Math.floor(Date.now() / 1000);
      const start = now - 30 * 60; // 最近30分钟
      const data: REDMetricsData = {};
      
      // 查询Rate
      if (redMetrics.rate_metric?.promql) {
        try {
          const response = await monitoringService.queryMetrics({
            query: redMetrics.rate_metric.promql.replace(/\{\{hostname\}\}/g, '".*"'),
            start: start.toString(),
            end: now.toString(),
            step: '60s',
          });
          // 保存完整的时间序列数据
          if (response.series && response.series.length > 0) {
            const series = response.series[0];
            if (series.data && series.data.length > 0) {
              data.rate = series.data.map((d: { timestamp: number; value: number }) => ({
                timestamp: d.timestamp * 1000, // 转为毫秒
                value: d.value,
              }));
              const validValues = series.data.filter((d: { timestamp: number; value: number }) => d.value !== null && d.value !== undefined);
              if (validValues.length > 0) {
                data.rateLatest = validValues[validValues.length - 1].value;
              }
            }
          }
        } catch (err) {
          console.error('查询Rate指标失败:', err);
        }
      }
      
      // 查询Error
      if (redMetrics.error_metric?.promql) {
        try {
          const response = await monitoringService.queryMetrics({
            query: redMetrics.error_metric.promql.replace(/\{\{hostname\}\}/g, '".*"'),
            start: start.toString(),
            end: now.toString(),
            step: '60s',
          });
          if (response.series && response.series.length > 0) {
            const series = response.series[0];
            if (series.data && series.data.length > 0) {
              data.error = series.data.map((d: { timestamp: number; value: number }) => ({
                timestamp: d.timestamp * 1000, // 转为毫秒
                value: d.value,
              }));
              const validValues = series.data.filter((d: { timestamp: number; value: number }) => d.value !== null && d.value !== undefined);
              if (validValues.length > 0) {
                data.errorLatest = validValues[validValues.length - 1].value;
              }
            }
          }
        } catch (err) {
          console.error('查询Error指标失败:', err);
        }
      }
      
      // 查询Duration
      if (redMetrics.duration_metric?.promql) {
        try {
          const response = await monitoringService.queryMetrics({
            query: redMetrics.duration_metric.promql.replace(/\{\{hostname\}\}/g, '".*"'),
            start: start.toString(),
            end: now.toString(),
            step: '60s',
          });
          if (response.series && response.series.length > 0) {
            const series = response.series[0];
            if (series.data && series.data.length > 0) {
              data.duration = series.data.map((d: { timestamp: number; value: number }) => ({
                timestamp: d.timestamp * 1000, // 转为毫秒
                value: d.value,
              }));
              const validValues = series.data.filter((d: { timestamp: number; value: number }) => d.value !== null && d.value !== undefined);
              if (validValues.length > 0) {
                data.durationLatest = validValues[validValues.length - 1].value;
              }
            }
          }
        } catch (err) {
          console.error('查询Duration指标失败:', err);
        }
      }
      
      setRedMetricsData(prev => ({ ...prev, [deployment.id]: data }));
    } catch (err) {
      console.error('获取RED指标失败:', err);
    }
  };

  useEffect(() => {
    const fetchAllREDMetrics = async () => {
      for (const deployment of deployments) {
        await fetchDeploymentREDMetrics(deployment);
      }
    };
    
    fetchAllREDMetrics();
  }, [deployments]);

  useEffect(() => {
    fetchDeployments();
  }, [page, appNameFilter, statusFilter]);

  const getStatusText = (status: string) => {
    const statusMap: Record<string, string> = {
      pending: '待发布',
      deploying: '发布中',
      success: '成功',
      failed: '失败',
      rolled_back: '已回滚',
      canceled: '已取消',
      skipped: '已跳过',
    };
    return statusMap[status] || status;
  };

  const getStatusColor = (status: string) => {
    const colorMap: Record<string, string> = {
      pending: '#faad14',
      deploying: '#1890ff',
      success: '#52c41a',
      failed: '#f5222d',
      rolled_back: '#722ed1',
      canceled: '#8c8c8c',
      skipped: '#d9d9d9',
    };
    return colorMap[status] || '#d9d9d9';
  };

  const getGrayMachineInfo = (deployment: Deployment) => {
    if (!deployment.gray_machine_id) return '未设置';
    const machine = deployment.node_deployments?.find(m => m.id === deployment.gray_machine_id);
    return machine ? `${machine.id}` : deployment.gray_machine_id;
  };

  const formatTime = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleString('zh-CN');
  };

  const handleCancel = async (e: React.MouseEvent, deployment: Deployment) => {
    e.stopPropagation();
    if (!confirm('确定要取消这个发布吗？')) return;
    
    try {
      await deploymentService.cancelDeployment(deployment.id);
      alert('取消成功');
      fetchDeployments();
    } catch (err) {
      alert('取消失败');
      console.error(err);
    }
  };

  const handleRollback = async (e: React.MouseEvent, deployment: Deployment) => {
    e.stopPropagation();
    
    const hasDeployingMachine = deployment.node_deployments?.some(
      (m) => m.node_deploy_status === 'deploying'
    );
    
    if (hasDeployingMachine) {
      alert('存在发布中的设备，无法回滚');
      return;
    }
    
    if (!confirm('确定要回滚这个发布吗？')) return;
    
    try {
      await deploymentService.rollbackDeployment(deployment.id);
      alert('回滚成功');
      fetchDeployments();
    } catch (err) {
      alert('回滚失败');
      console.error(err);
    }
  };

  return (
    <div>
      <div style={{ marginBottom: '20px', display: 'flex', gap: '10px', alignItems: 'center' }}>
        <input
          type="text"
          placeholder="应用名称"
          value={appNameFilter}
          onChange={(e) => setAppNameFilter(e.target.value)}
          style={{ padding: '8px', border: '1px solid #d9d9d9', borderRadius: '4px' }}
        />
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          style={{ padding: '8px', border: '1px solid #d9d9d9', borderRadius: '4px' }}
        >
          <option value="">全部状态</option>
          <option value="pending">待发布</option>
          <option value="deploying">发布中</option>
          <option value="success">成功</option>
          <option value="failed">失败</option>
          <option value="rolled_back">已回滚</option>
          <option value="canceled">已取消</option>
        </select>
        <button
          onClick={fetchDeployments}
          style={{
            padding: '8px 16px',
            background: '#1890ff',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
          }}
        >
          刷新
        </button>
        {onCreateNew && (
          <button
            onClick={onCreateNew}
            style={{
              padding: '8px 16px',
              background: '#52c41a',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
          >
            新建发布
          </button>
        )}
      </div>

      {loading && <div>加载中...</div>}
      {error && <div style={{ color: '#f5222d' }}>{error}</div>}

      {!loading && !error && (
        <>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ background: '#fafafa', borderBottom: '1px solid #f0f0f0' }}>
                <th style={{ padding: '12px', textAlign: 'left', maxWidth: '200px' }}>应用名称</th>
                <th style={{ padding: '12px', textAlign: 'left' }}>包版本</th>
                <th style={{ padding: '12px', textAlign: 'left' }}>灰度设备</th>
                <th style={{ padding: '12px', textAlign: 'left' }}>状态</th>
                <th style={{ padding: '12px', textAlign: 'left' }}>机器数量</th>
                <th style={{ padding: '12px', textAlign: 'left' }}>创建时间</th>
                <th style={{ padding: '12px', textAlign: 'left' }}>操作</th>
              </tr>
            </thead>
            <tbody>
              {deployments.map((deployment) => (
                <tr
                  key={deployment.id}
                  style={{ borderBottom: '1px solid #f0f0f0', cursor: 'pointer' }}
                  onClick={() => onSelectDeployment?.(deployment)}
                >
                  <td style={{ padding: '12px', maxWidth: '200px' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px', minWidth: 0 }}>
                      <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', flexShrink: 1 }}>{deployment.app_name}</span>
                      {redMetricsData[deployment.id] && redMetricsConfigs[deployment.id] && (
                        <>
                          {redMetricsConfigs[deployment.id].rate_metric && (
                            <div>
                              <REDMetricsMiniChart
                                rate={redMetricsData[deployment.id].rate}
                                rateThreshold={redMetricsConfigs[deployment.id].health_threshold?.rate_min}
                                metricName="Rate"
                              />
                            </div>
                          )}
                          {redMetricsConfigs[deployment.id].error_metric && (
                            <div>
                              <REDMetricsMiniChart
                                error={redMetricsData[deployment.id].error}
                                errorThreshold={redMetricsConfigs[deployment.id].health_threshold?.error_rate_max}
                                metricName="Error"
                              />
                            </div>
                          )}
                          {redMetricsConfigs[deployment.id].duration_metric && (
                            <div>
                              <REDMetricsMiniChart
                                duration={redMetricsData[deployment.id].duration}
                                durationThreshold={redMetricsConfigs[deployment.id].health_threshold?.duration_p99_max}
                                metricName="Duration"
                              />
                            </div>
                          )}
                        </>
                      )}
                    </div>
                  </td>
                  <td style={{ padding: '12px' }}>{deployment.package_version}</td>
                  <td style={{ padding: '12px' }}>{getGrayMachineInfo(deployment)}</td>
                  <td style={{ padding: '12px' }}>
                    <span
                      style={{
                        padding: '4px 8px',
                        borderRadius: '4px',
                        background: getStatusColor(deployment.status),
                        color: 'white',
                        fontSize: '12px',
                      }}
                    >
                      {getStatusText(deployment.status)}
                    </span>
                  </td>
                  <td style={{ padding: '12px' }}>{deployment.node_deployments?.length || 0}</td>
                  <td style={{ padding: '12px' }}>{formatTime(deployment.created_at)}</td>
                  <td style={{ padding: '12px' }}>
                    <div style={{ display: 'flex', gap: '8px' }}>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          onSelectDeployment?.(deployment);
                        }}
                        style={{
                          padding: '4px 12px',
                          background: '#1890ff',
                          color: 'white',
                          border: 'none',
                          borderRadius: '4px',
                          cursor: 'pointer',
                        }}
                      >
                        查看详情
                      </button>
                      {(deployment.status === 'pending' || deployment.status === 'deploying') && (
                        <button
                          onClick={(e) => handleCancel(e, deployment)}
                          style={{
                            padding: '4px 12px',
                            background: '#ff4d4f',
                            color: 'white',
                            border: 'none',
                            borderRadius: '4px',
                            cursor: 'pointer',
                          }}
                        >
                          取消
                        </button>
                      )}
                      {deployment.status === 'deploying' && (
                        <button
                          onClick={(e) => handleRollback(e, deployment)}
                          style={{
                            padding: '4px 12px',
                            background: '#faad14',
                            color: 'white',
                            border: 'none',
                            borderRadius: '4px',
                            cursor: 'pointer',
                          }}
                        >
                          回滚
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          <div style={{ marginTop: '20px', display: 'flex', justifyContent: 'flex-end', gap: '8px' }}>
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
              style={{
                padding: '4px 8px',
                border: '1px solid #d9d9d9',
                borderRadius: '3px',
                cursor: page === 1 ? 'not-allowed' : 'pointer',
                background: page === 1 ? '#f5f5f5' : 'white',
                fontSize: '12px',
              }}
            >
              上一页
            </button>
            <span style={{ padding: '4px 8px', fontSize: '12px' }}>
              第 {page} 页 / 共 {Math.ceil(total / pageSize)} 页
            </span>
            <button
              onClick={() => setPage((p) => p + 1)}
              disabled={page >= Math.ceil(total / pageSize)}
              style={{
                padding: '4px 8px',
                border: '1px solid #d9d9d9',
                borderRadius: '3px',
                cursor: page >= Math.ceil(total / pageSize) ? 'not-allowed' : 'pointer',
                background: page >= Math.ceil(total / pageSize) ? '#f5f5f5' : 'white',
                fontSize: '12px',
              }}
            >
              下一页
            </button>
          </div>
        </>
      )}
    </div>
  );
};

export default DeploymentList;
