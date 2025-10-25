import React, { useState } from 'react';
import { deploymentService } from '../services/deployment';
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

  const handleChange = (field: keyof CreateDeploymentRequest, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
  };

  return (
    <div style={{ padding: '20px', maxWidth: '600px', margin: '0 auto' }}>
      <h2>创建新发布</h2>
      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: '16px' }}>
          <label style={{ display: 'block', marginBottom: '8px', fontWeight: 'bold' }}>
            应用名称 *
          </label>
          <input
            type="text"
            value={formData.app_name}
            onChange={(e) => handleChange('app_name', e.target.value)}
            required
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
            包版本 *
          </label>
          <input
            type="text"
            value={formData.package_version}
            onChange={(e) => handleChange('package_version', e.target.value)}
            required
            placeholder="例如: v1.0.0"
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
            配置文件路径 *
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
            灰度策略 *
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
            {formData.gray_strategy === 'canary' && '金丝雀发布：逐步增加流量到新版本'}
            {formData.gray_strategy === 'blue-green' && '蓝绿发布：在蓝绿环境之间切换流量'}
            {formData.gray_strategy === 'all' && '全量发布：一次性将所有流量切换到新版本'}
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
