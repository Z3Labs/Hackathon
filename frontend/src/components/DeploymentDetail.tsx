import React, { useEffect, useState } from 'react';
import { deploymentService } from '../services/deployment';
import type { Deployment, NodeDeployment, Report } from '../types/deployment';

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
  const [countdown, setCountdown] = useState(5);
  const [report, setReport] = useState<Report | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

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

  useEffect(() => {
    const countdownTimer = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          refreshDetail();
          return 5;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(countdownTimer);
  }, [deploymentId]);

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
    return machine ? `${machine.id} (${machine.ip})` : machineId;
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
      setCountdown(5);
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
              {report.content}
            </div>
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
        <div style={{ display: 'flex', gap: '8px' }}>
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
            刷新 ({countdown}s)
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
              <th style={{ padding: '12px', textAlign: 'left' }}>机器 ID</th>
              <th style={{ padding: '12px', textAlign: 'left' }}>IP 地址</th>
              <th style={{ padding: '12px', textAlign: 'left' }}>发布状态</th>
              <th style={{ padding: '12px', textAlign: 'left' }}>发布日志</th>
            </tr>
          </thead>
          <tbody>
            {deployment.node_deployments.map((machine: NodeDeployment) => (
              <tr key={machine.id} style={{ borderBottom: '1px solid #f0f0f0' }}>
                <td style={{ padding: '12px' }}>
                  <input
                    type="checkbox"
                    checked={selectedNodeIds.includes(machine.id)}
                    onChange={(e) => handleSelectNode(machine.id, e.target.checked)}
                    disabled={!canOperate || machine.node_deploy_status === 'deploying'}
                    style={{ cursor: canOperate && machine.node_deploy_status !== 'deploying' ? 'pointer' : 'not-allowed' }}
                  />
                </td>
                <td style={{ padding: '12px' }}>{machine.id}</td>
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
              </tr>
            ))}
          </tbody>
        </table>
      ) : (
        <div style={{ padding: '20px', textAlign: 'center', color: '#8c8c8c' }}>暂无发布机器</div>
      )}

      {renderReportSection()}
    </div>
  );
};

export default DeploymentDetail;
