import React, { useEffect, useState } from 'react';
import { deploymentService } from '../services/deployment';
import type { Deployment, DeploymentMachine } from '../types/deployment';

interface DeploymentDetailProps {
  deploymentId: string;
  onClose?: () => void;
}

const DeploymentDetail: React.FC<DeploymentDetailProps> = ({ deploymentId, onClose }) => {
  const [deployment, setDeployment] = useState<Deployment | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

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
      normal: '正常',
      error: '异常',
      healthy: '健康',
      unhealthy: '不健康',
      alert: '告警',
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
      normal: '#52c41a',
      error: '#f5222d',
      healthy: '#52c41a',
      unhealthy: '#f5222d',
      alert: '#faad14',
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
            <div>{deployment.release_machines?.length || 0}</div>
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
      {deployment.release_machines && deployment.release_machines.length > 0 ? (
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ background: '#fafafa', borderBottom: '1px solid #f0f0f0' }}>
              <th style={{ padding: '12px', textAlign: 'left' }}>机器 ID</th>
              <th style={{ padding: '12px', textAlign: 'left' }}>IP 地址</th>
              <th style={{ padding: '12px', textAlign: 'left' }}>端口</th>
              <th style={{ padding: '12px', textAlign: 'left' }}>发布状态</th>
              <th style={{ padding: '12px', textAlign: 'left' }}>健康状态</th>
              <th style={{ padding: '12px', textAlign: 'left' }}>异常状态</th>
              <th style={{ padding: '12px', textAlign: 'left' }}>告警状态</th>
            </tr>
          </thead>
          <tbody>
            {deployment.release_machines.map((machine: DeploymentMachine) => (
              <tr key={machine.id} style={{ borderBottom: '1px solid #f0f0f0' }}>
                <td style={{ padding: '12px' }}>{machine.id}</td>
                <td style={{ padding: '12px' }}>{machine.ip}</td>
                <td style={{ padding: '12px' }}>{machine.port}</td>
                <td style={{ padding: '12px' }}>
                  <span
                    style={{
                      padding: '4px 8px',
                      borderRadius: '4px',
                      background: getStatusColor(machine.release_status),
                      color: 'white',
                      fontSize: '12px',
                    }}
                  >
                    {getStatusText(machine.release_status)}
                  </span>
                </td>
                <td style={{ padding: '12px' }}>
                  <span
                    style={{
                      padding: '4px 8px',
                      borderRadius: '4px',
                      background: getStatusColor(machine.health_status),
                      color: 'white',
                      fontSize: '12px',
                    }}
                  >
                    {getStatusText(machine.health_status)}
                  </span>
                </td>
                <td style={{ padding: '12px' }}>
                  <span
                    style={{
                      padding: '4px 8px',
                      borderRadius: '4px',
                      background: getStatusColor(machine.error_status),
                      color: 'white',
                      fontSize: '12px',
                    }}
                  >
                    {getStatusText(machine.error_status)}
                  </span>
                </td>
                <td style={{ padding: '12px' }}>
                  <span
                    style={{
                      padding: '4px 8px',
                      borderRadius: '4px',
                      background: getStatusColor(machine.alert_status),
                      color: 'white',
                      fontSize: '12px',
                    }}
                  >
                    {getStatusText(machine.alert_status)}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      ) : (
        <div style={{ padding: '20px', textAlign: 'center', color: '#8c8c8c' }}>暂无发布机器</div>
      )}

      {deployment.release_log && (
        <>
          <h3 style={{ marginTop: '20px', marginBottom: '16px' }}>发布日志</h3>
          <div
            style={{
              background: '#000',
              color: '#0f0',
              padding: '16px',
              borderRadius: '4px',
              fontFamily: 'monospace',
              fontSize: '12px',
              whiteSpace: 'pre-wrap',
              maxHeight: '400px',
              overflow: 'auto',
            }}
          >
            {deployment.release_log}
          </div>
        </>
      )}
    </div>
  );
};

export default DeploymentDetail;
