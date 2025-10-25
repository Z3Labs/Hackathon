import React, { useState, useEffect } from 'react';
import { deploymentService } from '../services/deployment';
import { appApi } from '../services/api';
import type { CreateDeploymentRequest } from '../types/deployment';

interface DeploymentFormProps {
  onSuccess?: () => void;
  onCancel?: () => void;
}

const DeploymentForm: React.FC<DeploymentFormProps> = ({ onSuccess, onCancel }) => {
  const [formData, setFormData] = useState<CreateDeploymentRequest>({
    app_name: '',
    package_version: '',
    config_path: '',
    gray_strategy: 'all',
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [apps, setApps] = useState<Array<{ id: string; name: string }>>([]);
  const [loadingApps, setLoadingApps] = useState(true);
  const [versions, setVersions] = useState<string[]>([]);
  const [loadingVersions, setLoadingVersions] = useState(false);

  useEffect(() => {
    const fetchApps = async () => {
      try {
        setLoadingApps(true);
        const response = await appApi.getAppList({ page: 1, page_size: 100 });
        setApps(response.apps || []);
      } catch (err) {
        console.error('获取应用列表失败', err);
      } finally {
        setLoadingApps(false);
      }
    };
    fetchApps();
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      await deploymentService.createDeployment(formData);
      onSuccess?.();
    } catch (err) {
      setError('创建发布记录失败');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleChange = async (field: keyof CreateDeploymentRequest, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    
    if (field === 'app_name' && value) {
      setLoadingVersions(true);
      setVersions([]);
      setFormData((prev) => ({ ...prev, package_version: '' }));
      try {
        const response = await appApi.getAppVersions(value);
        setVersions(response.versions || []);
      } catch (err) {
        console.error('获取版本列表失败', err);
        setVersions([]);
      } finally {
        setLoadingVersions(false);
      }
    }
  };

  return (
    <div style={{ padding: '20px', maxWidth: '600px', margin: '0 auto' }}>
      <h2>创建新发布</h2>
      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: '16px' }}>
          <label style={{ display: 'block', marginBottom: '8px', fontWeight: 'bold' }}>
            应用名称 <span style={{ color: '#ff4d4f' }}>*</span>
          </label>
          <select
            value={formData.app_name}
            onChange={(e) => handleChange('app_name', e.target.value)}
            required
            disabled={loadingApps}
            style={{
              width: '100%',
              padding: '8px',
              border: '1px solid #d9d9d9',
              borderRadius: '4px',
            }}
          >
            <option value="">请选择应用</option>
            {apps.map((app) => (
              <option key={app.id} value={app.name}>
                {app.name}
              </option>
            ))}
          </select>
        </div>

        <div style={{ marginBottom: '16px' }}>
          <label style={{ display: 'block', marginBottom: '8px', fontWeight: 'bold' }}>
            包版本 <span style={{ color: '#ff4d4f' }}>*</span>
          </label>
          <select
            value={formData.package_version}
            onChange={(e) => handleChange('package_version', e.target.value)}
            required
            disabled={!formData.app_name || loadingVersions || versions.length === 0}
            style={{
              width: '100%',
              padding: '8px',
              border: '1px solid #d9d9d9',
              borderRadius: '4px',
            }}
          >
            <option value="">
              {!formData.app_name ? '请先选择应用' : loadingVersions ? '加载中...' : versions.length === 0 ? '暂无版本' : '请选择版本'}
            </option>
            {versions.map((version) => (
              <option key={version} value={version}>
                {version}
              </option>
            ))}
          </select>
        </div>

        <div style={{ marginBottom: '16px' }}>
          <label style={{ display: 'block', marginBottom: '8px', fontWeight: 'bold' }}>
            配置文件路径 <span style={{ color: '#ff4d4f' }}>*</span>
          </label>
          <input
            type="text"
            value={formData.config_path}
            onChange={(e) => handleChange('config_path', e.target.value)}
            required
            placeholder="例如: /etc/app/config.yaml"
            style={{
              width: '100%',
              padding: '8px',
              border: '1px solid #d9d9d9',
              borderRadius: '4px',
            }}
          />
        </div>

        <div style={{ marginBottom: '16px' }}>
          <label style={{ display: 'block', marginBottom: '8px', fontWeight: 'bold' }}>
            灰度策略 <span style={{ color: '#ff4d4f' }}>*</span>
          </label>
          <select
            value={formData.gray_strategy}
            onChange={(e) => handleChange('gray_strategy', e.target.value as any)}
            required
            style={{
              width: '100%',
              padding: '8px',
              border: '1px solid #d9d9d9',
              borderRadius: '4px',
            }}
          >
            <option value="all">全量发布</option>
            <option value="canary">金丝雀发布</option>
            <option value="blue-green">蓝绿发布</option>
          </select>
          <div style={{ marginTop: '8px', fontSize: '12px', color: '#8c8c8c' }}>
            {formData.gray_strategy === 'canary' && '金丝雀发布:先在少量机器上部署,验证通过后逐步扩大部署范围'}
            {formData.gray_strategy === 'blue-green' && '蓝绿发布:保持两套环境,在新环境部署完成后切换,旧环境作为备份'}
            {formData.gray_strategy === 'all' && '全量发布:同时在所有机器上部署新版本'}
          </div>
        </div>

        {error && (
          <div
            style={{
              marginBottom: '16px',
              padding: '8px',
              background: '#fff1f0',
              border: '1px solid #ffa39e',
              borderRadius: '4px',
              color: '#f5222d',
            }}
          >
            {error}
          </div>
        )}

        <div style={{ display: 'flex', gap: '10px', justifyContent: 'flex-end' }}>
          {onCancel && (
            <button
              type="button"
              onClick={onCancel}
              disabled={loading}
              style={{
                padding: '8px 16px',
                border: '1px solid #d9d9d9',
                borderRadius: '4px',
                cursor: loading ? 'not-allowed' : 'pointer',
                background: 'white',
              }}
            >
              取消
            </button>
          )}
          <button
            type="submit"
            disabled={loading}
            style={{
              padding: '8px 16px',
              background: loading ? '#d9d9d9' : '#1890ff',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: loading ? 'not-allowed' : 'pointer',
            }}
          >
            {loading ? '创建中...' : '创建发布'}
          </button>
        </div>
      </form>
    </div>
  );
};

export default DeploymentForm;
