import React, { useEffect, useState, useCallback } from 'react';
import { deploymentService } from '../services/deployment';
import { monitoringService } from '../services/monitoring';
import { PromQL } from '../utils/promql';
import type { Deployment, NodeDeployment, Report, ReportData } from '../types/deployment';
import MonitorChart from './common/MonitorChart';

interface DeploymentDetailProps {
  deploymentId: string;
  onClose?: () => void;
}

const DeploymentDetail: React.FC<DeploymentDetailProps> = ({ deploymentId, onClose }) => {
  const [deployment, setDeployment] = useState<Deployment | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState(false);
  const [selectedNodeIds, setSelectedNodeIds] = useState<string[]>([]);
  const [countdown, setCountdown] = useState(30);
  const [report, setReport] = useState<Report | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  
  // 监控数据相关状态
  const [monitorMetric, setMonitorMetric] = useState<string>('cpu');
  const [monitorTimeRange, setMonitorTimeRange] = useState<number>(30);
  
  // 每台机器的监控展开状态
  const [expandedMonitorMachine, setExpandedMonitorMachine] = useState<string | null>(null);
  const [machineMonitorData, setMachineMonitorData] = useState<Record<string, any>>({});
  
  // 自动刷新控制
  const [autoRefresh, setAutoRefresh] = useState<boolean>(true);
  
  // 诊断报告的 promQL 查询结果
  const [reportPromQLResults, setReportPromQLResults] = useState<Record<string, any[]>>({});

  // 获取单台机器的监控数据
  const fetchMachineMonitorData = useCallback(async (machineName: string, minutes: number = 30, metric?: string) => {
    if (!deployment) return;
    
    // 使用传入的 metric 或当前的 monitorMetric
    const currentMetric = metric || monitorMetric;
    
    try {
      let promQL;
      let metricName;
      let unit;
      
      // 根据机器名称进行筛选（hostname 标签值就是机器名称）
      switch (currentMetric) {
        case 'cpu':
          promQL = PromQL.cpuUsage(machineName);
          metricName = 'CPU使用率';
          unit = '%';
          break;
        case 'memory':
          promQL = PromQL.memoryUsage(machineName);
          metricName = '内存使用率';
          unit = '%';
          break;
        case 'network':
          promQL = PromQL.networkReceiveRate(machineName);
          metricName = '网络接收速率';
          unit = 'bytes/s';
          break;
        default:
          promQL = PromQL.cpuUsage(machineName);
          metricName = 'CPU使用率';
          unit = '%';
      }
      const now = Math.floor(Date.now() / 1000);
      const start = now - minutes * 60;
      
      const response = await monitoringService.queryMetrics({
        query: promQL,
        start: start.toString(),
        end: now.toString(),
        step: '60s',
      });
      
      console.log(`[监控] 机器: ${machineName}, 指标类型: ${currentMetric}, 单位: ${unit}, 数据点数: ${response.series.length}`);
      
      // 为每个 series 添加 metric 名称和单位
      const enrichedSeries = response.series.map(s => ({
        ...s,
        metric: metricName,
        unit: unit,
      }));
      
      console.log(`[监控] 增强后的 series:`, enrichedSeries.map(s => ({ instance: s.instance, unit: s.unit })));
      
      setMachineMonitorData((prev) => ({
        ...prev,
        [machineName]: enrichedSeries || [],
      }));
    } catch (err) {
      console.error('获取机器监控数据失败:', err);
    }
  }, [deployment, monitorMetric, monitorTimeRange]);

  useEffect(() => {
    const fetchDetail = async () => {
      setLoading(true);
      setError(null);
      try {
        const response = await deploymentService.getDeploymentDetail(deploymentId);
        setDeployment(response.deployment);
        setReport(response.report ?? null);
      } catch (err) {
        setError('获取发布详情失败');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    fetchDetail();
  }, [deploymentId]);

  // 解析诊断报告内容
  const parseReportContent = (report: Report): ReportData => {
    try {
      // 尝试解析 content 为 JSON
      const parsed = JSON.parse(report.content);
      if (parsed.promQL && parsed.content) {
        return {
          promQL: parsed.promQL,
          content: parsed.content,
        };
      }
    } catch (e) {
      // 如果解析失败，说明 content 是纯文本
    }
    
    // 如果没有 promQL 字段，使用 report 对象的 promQL
    return {
      promQL: report.promQL,
      content: report.content,
    };
  };

  // 查询报告的 promQL
  const fetchReportPromQLResults = useCallback(async (promQLArray: string[]) => {
    const results: Record<string, any[]> = {};
    
    for (const query of promQLArray) {
      try {
        const now = Math.floor(Date.now() / 1000);
        const start = now - 30 * 60; // 最近 30 分钟
        
        const response = await monitoringService.queryMetrics({
          query: query,
          start: start.toString(),
          end: now.toString(),
          step: '60s',
        });
        
        results[query] = response.series || [];
      } catch (err) {
        console.error(`查询 promQL 失败: ${query}`, err);
        results[query] = [];
      }
    }
    
    setReportPromQLResults(results);
  }, []);

  useEffect(() => {
    if (!autoRefresh) return;
    
    const countdownTimer = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          refreshDetail();
          // 如果监控图表是展开的，也刷新监控数据
          if (expandedMonitorMachine) {
            // 使用最新的 monitorMetric 和 monitorTimeRange
            fetchMachineMonitorData(expandedMonitorMachine, monitorTimeRange, monitorMetric);
          }
          return 30;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(countdownTimer);
  }, [deploymentId, autoRefresh, expandedMonitorMachine, fetchMachineMonitorData, monitorMetric, monitorTimeRange]);

  // 当报告加载时，如果有 promQL，自动查询
  useEffect(() => {
    if (report && report.status === 'completed') {
      const reportData = parseReportContent(report);
      if (reportData.promQL && reportData.promQL.length > 0) {
        fetchReportPromQLResults(reportData.promQL);
      }
    }
  }, [report, fetchReportPromQLResults]);

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

  const getGrayMachineInfo = (machineId: string) => {
    if (!machineId || !deployment?.node_deployments) return '未设置';
    const machine = deployment.node_deployments.find(m => m.id === machineId);
    return machine ? `${machine.name} (${machine.ip})` : machineId;
  };

  const formatTime = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleString('zh-CN');
  };

  const showSuccessMessage = (message: string) => {
    setSuccessMessage(message);
    setTimeout(() => {
      setSuccessMessage(null);
    }, 3000);
  };

  const refreshDetail = async () => {
    setLoading(true);
    try {
      const response = await deploymentService.getDeploymentDetail(deploymentId);
      setDeployment(response.deployment);
      setReport(response.report ?? null);
      setCountdown(30);
      
      // 如果监控图表是展开的，也刷新监控数据
      if (expandedMonitorMachine) {
        fetchMachineMonitorData(expandedMonitorMachine, monitorTimeRange, monitorMetric);
      }
    } catch (err) {
      console.error('刷新详情失败:', err);
      alert('刷新详情失败');
    } finally {
      setLoading(false);
    }
  };

  const renderReportSection = () => {
    if (!report) {
      return (
        <div>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', margin: '12px 0' }}>
            <h3 style={{ margin: 0, fontSize: '16px' }}>诊断报告</h3>
            <span style={{ color: '#8c8c8c', fontSize: 12 }}>暂无报告</span>
          </div>
          <div style={{ background: '#fff', border: '1px solid #f0f0f0', borderRadius: 6, padding: 12 }}>
            <div style={{ color: '#8c8c8c', fontSize: '13px' }}>当发布触发异常或完成分析后将自动生成诊断报告。</div>
          </div>
        </div>
      );
    }

    const statusColor: Record<Report['status'], string> = {
      generating: '#1890ff',
      completed: '#52c41a',
      failed: '#f5222d',
    };
    const statusText: Record<Report['status'], string> = {
      generating: '报告生成中...',
      completed: '报告生成完成',
      failed: '报告生成失败',
    };

    const reportData = parseReportContent(report);

    return (
      <div>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', margin: '12px 0' }}>
          <h3 style={{ margin: 0, fontSize: '16px' }}>诊断报告</h3>
          <span
            style={{
              padding: '3px 6px',
              borderRadius: 3,
              background: statusColor[report.status],
              color: '#fff',
              fontSize: 11,
            }}
          >
            {statusText[report.status]}
          </span>
        </div>

        <div style={{ background: '#fff', border: '1px solid #f0f0f0', borderRadius: 6, padding: 12 }}>
          {report.status === 'generating' && (
            <div style={{ color: '#8c8c8c', display: 'flex', alignItems: 'center', gap: 6, fontSize: '13px' }}>
              <span className="spin" style={{ width: 14, height: 14, border: '2px solid #1890ff', borderTopColor: 'transparent', borderRadius: '50%', display: 'inline-block', animation: 'spin 1s linear infinite' }} />
              报告生成中，请稍候...
            </div>
          )}

          {report.status === 'failed' && (
            <div style={{ color: '#f5222d', fontSize: '13px' }}>
              生成失败，请稍后重试或刷新页面。
            </div>
          )}

          {report.status === 'completed' && (
            <>
              {/* 如果有 promQL，显示查询结果 */}
              {reportData.promQL && reportData.promQL.length > 0 && (
                <div style={{ marginBottom: 16 }}>
                  <div style={{ marginBottom: 12, fontWeight: 500, fontSize: '14px', color: '#262626' }}>
                    监控数据指标
                  </div>
                  {reportData.promQL.map((query, index) => {
                    const results = reportPromQLResults[query];
                    return (
                      <div key={index} style={{ marginBottom: 16, border: '1px solid #e8e8e8', borderRadius: 4 }}>
                        <div style={{ padding: '8px 12px', background: '#fafafa', borderBottom: '1px solid #e8e8e8' }}>
                          <div style={{ fontSize: '12px', color: '#8c8c8c', marginBottom: 4 }}>PromQL 查询</div>
                          <div style={{ fontSize: '13px', fontFamily: 'monospace', color: '#1890ff', wordBreak: 'break-all' }}>{query}</div>
                        </div>
                        <div style={{ padding: 12 }}>
                          {results && results.length > 0 ? (
                            <MonitorChart 
                              series={results} 
                              height={300} 
                              initialTimeRange={30}
                              showTimeSelector={false}
                            />
                          ) : (
                            <div style={{ padding: '20px', textAlign: 'center', color: '#8c8c8c', fontSize: '13px' }}>
                              加载中...
                            </div>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
              
              {/* 显示报告内容 */}
              <div style={{
                background: '#fafafa',
                border: '1px solid #f0f0f0',
                borderRadius: 4,
                padding: 8,
                whiteSpace: 'pre-wrap',
                lineHeight: 1.5,
                color: '#262626',
                fontSize: '13px',
              }}>
                {reportData.content}
              </div>
            </>
          )}

          <div style={{ marginTop: 8, color: '#8c8c8c', fontSize: 11 }}>
            更新时间：{new Date((report.updated_at || report.created_at) * 1000).toLocaleString('zh-CN')}
          </div>
        </div>
      </div>
    );
  };

  const handleSelectAll = (checked: boolean) => {
    if (checked && deployment?.node_deployments) {
      const selectableNodeIds = deployment.node_deployments
        .filter(node => node.node_deploy_status !== 'deploying')
        .map(node => node.id);
      setSelectedNodeIds(selectableNodeIds);
    } else {
      setSelectedNodeIds([]);
    }
  };

  const handleSelectNode = (nodeId: string, checked: boolean) => {
    if (checked) {
      setSelectedNodeIds([...selectedNodeIds, nodeId]);
    } else {
      setSelectedNodeIds(selectedNodeIds.filter(id => id !== nodeId));
    }
  };

  const handleBatchDeploy = async () => {
    if (selectedNodeIds.length === 0) {
      alert('请先选择要发布的设备');
      return;
    }
    
    setActionLoading(true);
    try {
      await deploymentService.deployNodeDeployment(deploymentId, selectedNodeIds);
      await refreshDetail();
      // 清除选择状态，因为设备已进入发布中状态
      setSelectedNodeIds([]);
      showSuccessMessage(`成功发布 ${selectedNodeIds.length} 个设备`);
    } catch (err) {
      console.error('批量发布失败:', err);
      alert('批量发布操作失败');
    } finally {
      setActionLoading(false);
    }
  };

  const handleBatchRetry = async () => {
    if (selectedNodeIds.length === 0) {
      alert('请先选择要重试的设备');
      return;
    }
    
    setActionLoading(true);
    try {
      await deploymentService.retryNodeDeployment(deploymentId, selectedNodeIds);
      await refreshDetail();
      // 清除选择状态
      setSelectedNodeIds([]);
      showSuccessMessage(`成功重试 ${selectedNodeIds.length} 个设备`);
    } catch (err) {
      console.error('批量重试失败:', err);
      alert('批量重试操作失败');
    } finally {
      setActionLoading(false);
    }
  };

  const handleBatchSkip = async () => {
    if (selectedNodeIds.length === 0) {
      alert('请先选择要跳过的设备');
      return;
    }
    
    setActionLoading(true);
    try {
      await deploymentService.skipNodeDeployment(deploymentId, selectedNodeIds);
      await refreshDetail();
      // 清除选择状态
      setSelectedNodeIds([]);
      showSuccessMessage(`成功跳过 ${selectedNodeIds.length} 个设备`);
    } catch (err) {
      console.error('批量跳过失败:', err);
      alert('批量跳过操作失败');
    } finally {
      setActionLoading(false);
    }
  };

  const handleBatchRollback = async () => {
    if (selectedNodeIds.length === 0) {
      alert('请先选择要回滚的设备');
      return;
    }
    
    setActionLoading(true);
    try {
      await deploymentService.rollbackNodeDeployment(deploymentId, selectedNodeIds);
      await refreshDetail();
      // 清除选择状态
      setSelectedNodeIds([]);
      showSuccessMessage(`成功回滚 ${selectedNodeIds.length} 个设备`);
    } catch (err) {
      console.error('批量回滚失败:', err);
      alert('批量回滚操作失败');
    } finally {
      setActionLoading(false);
    }
  };

  const canOperate = !['canceled', 'rolled_back'].includes(deployment?.status || '');

  // 切换机器监控展开/收起
  const toggleMachineMonitor = (machineName: string) => {
    if (expandedMonitorMachine === machineName) {
      setExpandedMonitorMachine(null);
    } else {
      setExpandedMonitorMachine(machineName);
      // 加载该机器的监控数据，使用当前的 monitorTimeRange
      fetchMachineMonitorData(machineName, monitorTimeRange, monitorMetric);
    }
  };

  // 渲染单台机器的监控图表
  const renderMachineMonitorChart = (machineName: string) => {
    const data = machineMonitorData[machineName] || [];
    
    return (
      <div style={{ 
        background: '#ffffff', 
        border: '1px solid #e6f0ff',
        borderRadius: '8px',
        padding: '16px',
        boxShadow: '0 2px 8px rgba(0,0,0,0.04)'
      }}>
        <div style={{ display: 'flex', gap: '8px', marginBottom: '16px' }}>
          {['cpu', 'memory', 'network'].map((metric) => (
            <button
              key={metric}
              onClick={async () => {
                // 先清空旧数据，避免显示错误
                setMachineMonitorData((prev) => ({
                  ...prev,
                  [machineName]: [],
                }));
                // 更新指标类型
                setMonitorMetric(metric);
                // 立即使用新的 metric 加载数据，使用当前的 monitorTimeRange
                fetchMachineMonitorData(machineName, monitorTimeRange, metric);
              }}
              style={{
                padding: '6px 16px',
                border: `1px solid ${monitorMetric === metric ? '#1890ff' : '#d9d9d9'}`,
                borderRadius: '6px',
                background: monitorMetric === metric ? '#1890ff' : '#ffffff',
                color: monitorMetric === metric ? 'white' : '#666',
                cursor: 'pointer',
                fontSize: '13px',
                fontWeight: monitorMetric === metric ? 500 : 400,
                transition: 'all 0.2s',
              }}
            >
              {metric === 'cpu' ? 'CPU' : metric === 'memory' ? '内存' : '网络'}
            </button>
          ))}
        </div>
        
        {data.length > 0 ? (
          <MonitorChart 
            series={data} 
            height={400} 
            initialTimeRange={monitorTimeRange}
            onTimeRangeChange={(minutes) => {
              // 更新时间范围状态并重新加载数据
              setMonitorTimeRange(minutes);
              fetchMachineMonitorData(machineName, minutes, monitorMetric);
            }}
          />
        ) : (
          <div style={{ padding: '60px', textAlign: 'center', color: '#8c8c8c', fontSize: '14px' }}>
            <div style={{ fontSize: '32px', marginBottom: '12px' }}>📊</div>
            <div>加载监控数据中...</div>
          </div>
        )}
      </div>
    );
  };

  if (loading) {
    return <div style={{ padding: '20px' }}>加载中...</div>;
  }

  if (error) {
    return (
      <div style={{ padding: '20px' }}>
        <div style={{ color: '#f5222d' }}>{error}</div>
        {onClose && (
          <button
            onClick={onClose}
            style={{
              marginTop: '16px',
              padding: '8px 16px',
              border: '1px solid #d9d9d9',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
          >
            返回
          </button>
        )}
      </div>
    );
  }

  if (!deployment) {
    return null;
  }

  return (
    <div style={{ padding: '16px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
        <h2 style={{ margin: 0, fontSize: '18px' }}>发布详情</h2>
        {onClose && (
          <button
            onClick={onClose}
            style={{
              padding: '6px 12px',
              border: '1px solid #d9d9d9',
              borderRadius: '4px',
              cursor: 'pointer',
              fontSize: '14px',
            }}
          >
            返回
          </button>
        )}
      </div>

      {successMessage && (
        <div style={{
          position: 'fixed',
          top: '20px',
          right: '20px',
          background: '#f6ffed',
          border: '1px solid #b7eb8f',
          borderRadius: '6px',
          padding: '12px 16px',
          color: '#52c41a',
          fontSize: '14px',
          display: 'flex',
          alignItems: 'center',
          gap: '8px',
          boxShadow: '0 4px 12px rgba(0, 0, 0, 0.15)',
          zIndex: 1000,
          minWidth: '200px',
        }}>
          <span style={{ fontSize: '16px' }}>✓</span>
          {successMessage}
        </div>
      )}

      <div style={{ background: '#fafafa', padding: '12px', borderRadius: '4px', marginBottom: '16px' }}>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr 1fr', gap: '12px' }}>
          <div>
            <div style={{ color: '#8c8c8c', marginBottom: '2px', fontSize: '12px' }}>应用名称</div>
            <div style={{ fontWeight: 'bold', fontSize: '14px' }}>{deployment.app_name}</div>
          </div>
          <div>
            <div style={{ color: '#8c8c8c', marginBottom: '2px', fontSize: '12px' }}>包版本</div>
            <div style={{ fontWeight: 'bold', fontSize: '14px' }}>{deployment.package_version}</div>
          </div>
          <div>
            <div style={{ color: '#8c8c8c', marginBottom: '2px', fontSize: '12px' }}>灰度设备</div>
            <div style={{ fontSize: '14px' }}>{getGrayMachineInfo(deployment.gray_machine_id)}</div>
          </div>
          <div>
            <div style={{ color: '#8c8c8c', marginBottom: '2px', fontSize: '12px' }}>发布状态</div>
            <div>
              <span
                style={{
                  padding: '2px 6px',
                  borderRadius: '3px',
                  background: getStatusColor(deployment.status),
                  color: 'white',
                  fontSize: '11px',
                }}
              >
                {getStatusText(deployment.status)}
              </span>
            </div>
          </div>
          <div>
            <div style={{ color: '#8c8c8c', marginBottom: '2px', fontSize: '12px' }}>机器数量</div>
            <div style={{ fontSize: '14px' }}>{deployment.node_deployments?.length || 0}</div>
          </div>
          <div>
            <div style={{ color: '#8c8c8c', marginBottom: '2px', fontSize: '12px' }}>创建时间</div>
            <div style={{ fontSize: '14px' }}>{formatTime(deployment.created_at)}</div>
          </div>
          <div>
            <div style={{ color: '#8c8c8c', marginBottom: '2px', fontSize: '12px' }}>更新时间</div>
            <div style={{ fontSize: '14px' }}>{formatTime(deployment.updated_at)}</div>
          </div>
        </div>
      </div>

      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', margin: '12px 0' }}>
        <h3 style={{ margin: 0, fontSize: '16px' }}>发布机器列表</h3>
        <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
          <a
            onClick={(e) => {
              e.preventDefault();
              setAutoRefresh(!autoRefresh);
            }}
            href="#"
            style={{
              color: autoRefresh ? '#1890ff' : '#666',
              cursor: 'pointer',
              fontSize: '13px',
              textDecoration: 'none',
              display: 'inline-block',
              lineHeight: '32px',
            }}
          >
            {autoRefresh ? (
              <span>✓ 自动刷新</span>
            ) : (
              <span>○ 自动刷新</span>
            )}
          </a>
          <button
            onClick={refreshDetail}
            disabled={actionLoading}
            style={{
              padding: '6px 16px',
              border: '1px solid #d9d9d9',
              borderRadius: '4px',
              background: 'white',
              cursor: actionLoading ? 'not-allowed' : 'pointer',
              fontSize: '14px',
            }}
          >
            {autoRefresh ? `刷新 (${countdown}s)` : '刷新'}
          </button>
          {canOperate && (
            <>
              <button
                onClick={handleBatchDeploy}
                disabled={actionLoading || selectedNodeIds.length === 0}
                style={{
                  padding: '6px 16px',
                  border: 'none',
                  borderRadius: '4px',
                  background: actionLoading || selectedNodeIds.length === 0 ? '#d9d9d9' : '#1890ff',
                  color: 'white',
                  cursor: actionLoading || selectedNodeIds.length === 0 ? 'not-allowed' : 'pointer',
                  fontSize: '14px',
                }}
              >
                发布
              </button>
              <button
                onClick={handleBatchRetry}
                disabled={actionLoading || selectedNodeIds.length === 0}
                style={{
                  padding: '6px 16px',
                  border: 'none',
                  borderRadius: '4px',
                  background: actionLoading || selectedNodeIds.length === 0 ? '#d9d9d9' : '#faad14',
                  color: 'white',
                  cursor: actionLoading || selectedNodeIds.length === 0 ? 'not-allowed' : 'pointer',
                  fontSize: '14px',
                }}
              >
                重试
              </button>
              <button
                onClick={handleBatchSkip}
                disabled={actionLoading || selectedNodeIds.length === 0}
                style={{
                  padding: '6px 16px',
                  border: 'none',
                  borderRadius: '4px',
                  background: actionLoading || selectedNodeIds.length === 0 ? '#d9d9d9' : '#8c8c8c',
                  color: 'white',
                  cursor: actionLoading || selectedNodeIds.length === 0 ? 'not-allowed' : 'pointer',
                  fontSize: '14px',
                }}
              >
                跳过
              </button>
              <button
                onClick={handleBatchRollback}
                disabled={actionLoading || selectedNodeIds.length === 0}
                style={{
                  padding: '6px 16px',
                  border: 'none',
                  borderRadius: '4px',
                  background: actionLoading || selectedNodeIds.length === 0 ? '#d9d9d9' : '#722ed1',
                  color: 'white',
                  cursor: actionLoading || selectedNodeIds.length === 0 ? 'not-allowed' : 'pointer',
                  fontSize: '14px',
                }}
              >
                回滚
              </button>
            </>
          )}
        </div>
      </div>
      {deployment.node_deployments && deployment.node_deployments.length > 0 ? (
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ background: '#fafafa', borderBottom: '1px solid #f0f0f0' }}>
              <th style={{ padding: '12px', textAlign: 'left', width: '50px' }}>
                <input
                  type="checkbox"
                  checked={
                    deployment.node_deployments.filter(n => n.node_deploy_status !== 'deploying').length > 0 &&
                    selectedNodeIds.length === deployment.node_deployments.filter(n => n.node_deploy_status !== 'deploying').length
                  }
                  onChange={(e) => handleSelectAll(e.target.checked)}
                  disabled={!canOperate}
                  style={{ cursor: canOperate ? 'pointer' : 'not-allowed' }}
                />
              </th>
              <th style={{ padding: '12px', textAlign: 'left' }}>机器名称</th>
              <th style={{ padding: '12px', textAlign: 'left' }}>IP 地址</th>
              <th style={{ padding: '12px', textAlign: 'left' }}>发布状态</th>
              <th style={{ padding: '12px', textAlign: 'left' }}>发布日志</th>
              <th style={{ padding: '12px', textAlign: 'left', width: '140px' }}>操作</th>
            </tr>
          </thead>
          <tbody>
            {deployment.node_deployments.map((machine: NodeDeployment) => (
              <React.Fragment key={machine.id}>
                <tr style={{ borderBottom: '1px solid #f0f0f0' }}>
                <td style={{ padding: '12px' }}>
                  <input
                    type="checkbox"
                    checked={selectedNodeIds.includes(machine.id)}
                    onChange={(e) => handleSelectNode(machine.id, e.target.checked)}
                    disabled={!canOperate || machine.node_deploy_status === 'deploying'}
                    style={{ cursor: canOperate && machine.node_deploy_status !== 'deploying' ? 'pointer' : 'not-allowed' }}
                  />
                </td>
                <td style={{ padding: '12px' }}>{machine.name}</td>
                <td style={{ padding: '12px' }}>{machine.ip}</td>
                <td style={{ padding: '12px' }}>
                  <span
                    style={{
                      padding: '4px 8px',
                      borderRadius: '4px',
                      background: getStatusColor(machine.node_deploy_status),
                      color: 'white',
                      fontSize: '12px',
                    }}
                  >
                    {getStatusText(machine.node_deploy_status)}
                  </span>
                </td>
                <td style={{ padding: '12px', maxWidth: '300px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                  {machine.release_log || '-'}
                </td>
                <td style={{ padding: '12px' }}>
                  <button
                    onClick={() => toggleMachineMonitor(machine.name)}
                    style={{
                      padding: '4px 12px',
                      border: 'none',
                      borderRadius: '4px',
                      background: '#1890ff',
                      color: 'white',
                      cursor: 'pointer',
                      fontSize: '12px',
                      display: 'flex',
                      alignItems: 'center',
                      gap: '4px',
                    }}
                  >
                    {expandedMonitorMachine === machine.name ? (
                      <>收起 <span style={{ fontSize: '10px' }}>▽</span></>
                    ) : (
                      <>指标监控 <span style={{ fontSize: '10px' }}>▶</span></>
                    )}
                  </button>
                </td>
                </tr>
                {expandedMonitorMachine === machine.name && (
                  <tr>
                    <td colSpan={6} style={{ padding: '12px', background: '#fafafa' }}>
                      {renderMachineMonitorChart(machine.name)}
                    </td>
                  </tr>
                )}
              </React.Fragment>
            ))}
          </tbody>
        </table>
      ) : (
        <div style={{ padding: '20px', textAlign: 'center', color: '#8c8c8c' }}>暂无发布机器</div>
      )}

      <div style={{ marginTop: '24px', borderTop: '1px solid #f0f0f0', paddingTop: '16px' }}>
        {renderReportSection()}
      </div>
      <style>{`
        @keyframes spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
      `}</style>
    </div>
  );
};

export default DeploymentDetail;
