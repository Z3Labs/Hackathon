import React, { useEffect, useState } from 'react';
import { deploymentService } from '../services/deployment';
import type { Deployment, NodeDeployment } from '../types/deployment';

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

  useEffect(() => {
    const fetchDetail = async () => {
      setLoading(true);
      setError(null);
      try {
        const response = await deploymentService.getDeploymentDetail(deploymentId);
        setDeployment(response.deployment);
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

  const getGrayStrategyText = (strategy: string) => {
    const strategyMap: Record<string, string> = {
      canary: '金丝雀发布',
      'blue-green': '蓝绿发布',
      all: '全量发布',
    };
    return strategyMap[strategy] || strategy;
  };

  const formatTime = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleString('zh-CN');
  };

  const refreshDetail = async () => {
    setLoading(true);
    try {
      const response = await deploymentService.getDeploymentDetail(deploymentId);
      setDeployment(response.deployment);
      setSelectedNodeIds([]);
      setCountdown(5);
    } catch (err) {
      console.error('刷新详情失败:', err);
      alert('刷新详情失败');
    } finally {
      setLoading(false);
    }
  };

  const handleSelectAll = (checked: boolean) => {
    if (checked && deployment?.node_deployments) {
      setSelectedNodeIds(deployment.node_deployments.map(node => node.id));
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
    if (!confirm(`确定要发布选中的 ${selectedNodeIds.length} 个设备吗？`)) return;
    
    setActionLoading(true);
    try {
      await deploymentService.deployNodeDeployment(deploymentId, selectedNodeIds);
      await refreshDetail();
      alert('批量发布操作成功');
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
    if (!confirm(`确定要重试选中的 ${selectedNodeIds.length} 个设备吗？`)) return;
    
    setActionLoading(true);
    try {
      await deploymentService.retryNodeDeployment(deploymentId, selectedNodeIds);
      await refreshDetail();
      alert('批量重试操作成功');
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
    if (!confirm(`确定要跳过选中的 ${selectedNodeIds.length} 个设备吗？`)) return;
    
    setActionLoading(true);
    try {
      await deploymentService.skipNodeDeployment(deploymentId, selectedNodeIds);
      await refreshDetail();
      alert('批量跳过操作成功');
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
    if (!confirm(`确定要回滚选中的 ${selectedNodeIds.length} 个设备吗？`)) return;
    
    setActionLoading(true);
    try {
      await deploymentService.rollbackNodeDeployment(deploymentId, selectedNodeIds);
      await refreshDetail();
      alert('批量回滚操作成功');
    } catch (err) {
      console.error('批量回滚失败:', err);
      alert('批量回滚操作失败');
    } finally {
      setActionLoading(false);
    }
  };

  const handleBatchCancel = async () => {
    if (selectedNodeIds.length === 0) {
      alert('请先选择要取消的设备');
      return;
    }
    if (!confirm(`确定要取消选中的 ${selectedNodeIds.length} 个设备吗？`)) return;
    
    setActionLoading(true);
    try {
      await deploymentService.cancelNodeDeployment(deploymentId, selectedNodeIds);
      await refreshDetail();
      alert('批量取消操作成功');
    } catch (err) {
      console.error('批量取消失败:', err);
      alert('批量取消操作失败');
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
    <div style={{ padding: '20px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px' }}>
        <h2>发布详情</h2>
        {onClose && (
          <button
            onClick={onClose}
            style={{
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

      <div style={{ background: '#fafafa', padding: '16px', borderRadius: '4px', marginBottom: '20px' }}>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
          <div>
            <div style={{ color: '#8c8c8c', marginBottom: '4px' }}>应用名称</div>
            <div style={{ fontWeight: 'bold' }}>{deployment.app_name}</div>
          </div>
          <div>
            <div style={{ color: '#8c8c8c', marginBottom: '4px' }}>包版本</div>
            <div style={{ fontWeight: 'bold' }}>{deployment.package_version}</div>
          </div>
          <div>
            <div style={{ color: '#8c8c8c', marginBottom: '4px' }}>灰度策略</div>
            <div>{getGrayStrategyText(deployment.gray_strategy)}</div>
          </div>
          <div>
            <div style={{ color: '#8c8c8c', marginBottom: '4px' }}>发布状态</div>
            <div>
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
            </div>
          </div>
          <div>
            <div style={{ color: '#8c8c8c', marginBottom: '4px' }}>配置文件路径</div>
            <div>{deployment.config_path}</div>
          </div>
          <div>
            <div style={{ color: '#8c8c8c', marginBottom: '4px' }}>机器数量</div>
            <div>{deployment.node_deployments?.length || 0}</div>
          </div>
          <div>
            <div style={{ color: '#8c8c8c', marginBottom: '4px' }}>创建时间</div>
            <div>{formatTime(deployment.created_at)}</div>
          </div>
          <div>
            <div style={{ color: '#8c8c8c', marginBottom: '4px' }}>更新时间</div>
            <div>{formatTime(deployment.updated_at)}</div>
          </div>
        </div>
      </div>

      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
        <h3 style={{ margin: 0 }}>发布机器列表</h3>
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
                批量发布
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
                批量重试
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
                批量跳过
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
                批量回滚
              </button>
              <button
                onClick={handleBatchCancel}
                disabled={actionLoading || selectedNodeIds.length === 0}
                style={{
                  padding: '6px 16px',
                  border: 'none',
                  borderRadius: '4px',
                  background: actionLoading || selectedNodeIds.length === 0 ? '#d9d9d9' : '#f5222d',
                  color: 'white',
                  cursor: actionLoading || selectedNodeIds.length === 0 ? 'not-allowed' : 'pointer',
                  fontSize: '14px',
                }}
              >
                批量取消
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
                  checked={selectedNodeIds.length === deployment.node_deployments.length}
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
                    disabled={!canOperate}
                    style={{ cursor: canOperate ? 'pointer' : 'not-allowed' }}
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
    </div>
  );
};

export default DeploymentDetail;
