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
  const [actionLoading, setActionLoading] = useState<string | null>(null);

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
    try {
      const response = await deploymentService.getDeploymentDetail(deploymentId);
      setDeployment(response.deployment);
    } catch (err) {
      console.error('刷新详情失败:', err);
    }
  };

  const handleDeploy = async (nodeId: string) => {
    setActionLoading(nodeId);
    try {
      await deploymentService.deployNodeDeployment(deploymentId, [nodeId]);
      await refreshDetail();
      alert('发布操作成功');
    } catch (err) {
      console.error('发布失败:', err);
      alert('发布操作失败');
    } finally {
      setActionLoading(null);
    }
  };

  const handleRetry = async (nodeId: string) => {
    setActionLoading(nodeId);
    try {
      await deploymentService.retryNodeDeployment(deploymentId, [nodeId]);
      await refreshDetail();
      alert('重试操作成功');
    } catch (err) {
      console.error('重试失败:', err);
      alert('重试操作失败');
    } finally {
      setActionLoading(null);
    }
  };

  const handleSkip = async (nodeId: string) => {
    if (!confirm('确定要跳过该设备吗?')) return;
    setActionLoading(nodeId);
    try {
      await deploymentService.skipNodeDeployment(deploymentId, [nodeId]);
      await refreshDetail();
      alert('跳过操作成功');
    } catch (err) {
      console.error('跳过失败:', err);
      alert('跳过操作失败');
    } finally {
      setActionLoading(null);
    }
  };

  const handleRollback = async (nodeId: string) => {
    if (!confirm('确定要回滚该设备吗?')) return;
    setActionLoading(nodeId);
    try {
      await deploymentService.rollbackNodeDeployment(deploymentId, [nodeId]);
      await refreshDetail();
      alert('回滚操作成功');
    } catch (err) {
      console.error('回滚失败:', err);
      alert('回滚操作失败');
    } finally {
      setActionLoading(null);
    }
  };

  const getNodeActions = (node: NodeDeployment) => {
    const actions = [];
    const cannotOperate = ['canceled', 'rolled_back'].includes(deployment?.status || '');

    if (cannotOperate) return [];

    if (node.node_deploy_status === 'pending') {
      actions.push(
        { label: '发布', onClick: () => handleDeploy(node.id), color: '#1890ff' },
        { label: '跳过', onClick: () => handleSkip(node.id), color: '#8c8c8c' }
      );
    } else if (node.node_deploy_status === 'failed') {
      actions.push(
        { label: '重试', onClick: () => handleRetry(node.id), color: '#faad14' },
        { label: '跳过', onClick: () => handleSkip(node.id), color: '#8c8c8c' }
      );
    } else if (node.node_deploy_status === 'success') {
      actions.push(
        { label: '回滚', onClick: () => handleRollback(node.id), color: '#722ed1' }
      );
    }

    return actions;
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

      <h3 style={{ marginBottom: '16px' }}>发布机器列表</h3>
      {deployment.node_deployments && deployment.node_deployments.length > 0 ? (
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ background: '#fafafa', borderBottom: '1px solid #f0f0f0' }}>
              <th style={{ padding: '12px', textAlign: 'left' }}>机器 ID</th>
              <th style={{ padding: '12px', textAlign: 'left' }}>IP 地址</th>
              <th style={{ padding: '12px', textAlign: 'left' }}>发布状态</th>
              <th style={{ padding: '12px', textAlign: 'left' }}>发布日志</th>
              <th style={{ padding: '12px', textAlign: 'left' }}>操作</th>
            </tr>
          </thead>
          <tbody>
            {deployment.node_deployments.map((machine: NodeDeployment) => {
              const actions = getNodeActions(machine);
              return (
                <tr key={machine.id} style={{ borderBottom: '1px solid #f0f0f0' }}>
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
                  <td style={{ padding: '12px' }}>
                    {actions.length > 0 ? (
                      <div style={{ display: 'flex', gap: '8px' }}>
                        {actions.map((action, index) => (
                          <button
                            key={index}
                            onClick={action.onClick}
                            disabled={actionLoading === machine.id}
                            style={{
                              padding: '4px 12px',
                              border: 'none',
                              borderRadius: '4px',
                              background: actionLoading === machine.id ? '#d9d9d9' : action.color,
                              color: 'white',
                              cursor: actionLoading === machine.id ? 'not-allowed' : 'pointer',
                              fontSize: '12px',
                            }}
                          >
                            {actionLoading === machine.id ? '处理中...' : action.label}
                          </button>
                        ))}
                      </div>
                    ) : (
                      <span style={{ color: '#8c8c8c' }}>-</span>
                    )}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      ) : (
        <div style={{ padding: '20px', textAlign: 'center', color: '#8c8c8c' }}>暂无发布机器</div>
      )}
    </div>
  );
};

export default DeploymentDetail;
