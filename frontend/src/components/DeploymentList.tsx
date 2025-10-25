import React, { useEffect, useState } from 'react';
import { deploymentService } from '../services/deployment';
import type { Deployment } from '../types/deployment';

interface DeploymentListProps {
  onSelectDeployment?: (deployment: Deployment) => void;
  onCreateNew?: () => void;
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

  return (
    <div style={{ padding: '20px' }}>
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
                <th style={{ padding: '12px', textAlign: 'left' }}>应用名称</th>
                <th style={{ padding: '12px', textAlign: 'left' }}>包版本</th>
                <th style={{ padding: '12px', textAlign: 'left' }}>灰度策略</th>
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
                  <td style={{ padding: '12px' }}>{deployment.app_name}</td>
                  <td style={{ padding: '12px' }}>{deployment.package_version}</td>
                  <td style={{ padding: '12px' }}>{getGrayStrategyText(deployment.gray_strategy)}</td>
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
                  <td style={{ padding: '12px' }}>{deployment.release_machines?.length || 0}</td>
                  <td style={{ padding: '12px' }}>{formatTime(deployment.created_at)}</td>
                  <td style={{ padding: '12px' }}>
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
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          <div style={{ marginTop: '20px', display: 'flex', justifyContent: 'center', gap: '10px' }}>
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
              style={{
                padding: '8px 16px',
                border: '1px solid #d9d9d9',
                borderRadius: '4px',
                cursor: page === 1 ? 'not-allowed' : 'pointer',
                background: page === 1 ? '#f5f5f5' : 'white',
              }}
            >
              上一页
            </button>
            <span style={{ padding: '8px 16px' }}>
              第 {page} 页 / 共 {Math.ceil(total / pageSize)} 页
            </span>
            <button
              onClick={() => setPage((p) => p + 1)}
              disabled={page >= Math.ceil(total / pageSize)}
              style={{
                padding: '8px 16px',
                border: '1px solid #d9d9d9',
                borderRadius: '4px',
                cursor: page >= Math.ceil(total / pageSize) ? 'not-allowed' : 'pointer',
                background: page >= Math.ceil(total / pageSize) ? '#f5f5f5' : 'white',
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
