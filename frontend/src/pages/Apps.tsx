import React, { useState, useEffect } from 'react'
import { appApi, machineApi } from '../services/api'
import { Application, Machine, CreateAppReq, GetAppListResp, GetAppDetailResp, GetMachineListResp, PrometheusAlert } from '../types'
import { useApiRequest } from '../hooks/useApiRequest'
import { Toaster } from 'react-hot-toast'
import PageLayout from '../components/PageLayout'
import './Apps.css'

const Apps: React.FC = () => {
  const [apps, setApps] = useState<Application[]>([])
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showEditModal, setShowEditModal] = useState(false)
  const [showDetailModal, setShowDetailModal] = useState(false)
  const [showMachineListModal, setShowMachineListModal] = useState(false)
  const [isMachineEditMode, setIsMachineEditMode] = useState(false)
  const [selectedApp, setSelectedApp] = useState<Application | null>(null)
  const [availableMachines, setAvailableMachines] = useState<Machine[]>([])
  const [selectedMachineIds, setSelectedMachineIds] = useState<string[]>([])
  const [searchName, setSearchName] = useState('')
  const [activeTab, setActiveTab] = useState<'basic' | 'red' | 'rollback'>('basic')
  const [pagination, setPagination] = useState({
    page: 1,
    pageSize: 10,
    total: 0
  })
  
  // 使用统一的API请求Hook
  const { request, loading } = useApiRequest()

  // 表单数据
  const [formData, setFormData] = useState<CreateAppReq>({
    name: '',
    repo: '',
    deploy_path: '',
    config_path: '',
    start_cmd: '',
    stop_cmd: '',
    red_metrics_config: {
      enabled: false,
      rate_metric: undefined,
      error_metric: undefined,
      duration_metric: undefined,
      health_threshold: undefined
    },
    rollback_policy: {
      enabled: false,
      alert_rules: [],
      auto_rollback: false,
      notify_channel: ''
    }
  })


  // 获取应用列表
  const fetchApps = async () => {
    const result = await request(
      () => appApi.getAppList({
        page: pagination.page,
        page_size: pagination.pageSize,
        name: searchName || undefined
      }) as unknown as Promise<GetAppListResp>,
      {
        errorMessage: '获取应用列表失败'
      }
    )
    
    if (result) {
      setApps(result.apps || [])
      setPagination(prev => ({
        ...prev,
        total: result.total || 0
      }))
    }
  }

  useEffect(() => {
    fetchApps()
  }, [pagination.page, pagination.pageSize, searchName])

  // ESC键关闭弹窗
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        if (showCreateModal) {
          setShowCreateModal(false)
        } else if (showEditModal) {
          setShowEditModal(false)
        } else if (showDetailModal) {
          setShowDetailModal(false)
        } else if (showMachineListModal) {
          setShowMachineListModal(false)
        }
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => {
      document.removeEventListener('keydown', handleKeyDown)
    }
  }, [showCreateModal, showEditModal, showDetailModal, showMachineListModal])

  // 创建应用
  const handleCreateApp = async () => {
    const result = await request(
      () => appApi.createApp(formData),
      {
        successMessage: '应用创建成功',
        errorMessage: '创建应用失败',
        showSuccessToast: true
      }
    )
    
    if (result) {
      setShowCreateModal(false)
      resetForm()
      fetchApps()
    }
  }

  // 更新应用
  const handleUpdateApp = async () => {
    if (!selectedApp) return
    
    const result = await request(
      () => appApi.updateApp(selectedApp.id, {
        ...formData,
        id: selectedApp.id
      }),
      {
        successMessage: '应用更新成功',
        errorMessage: '更新应用失败',
        showSuccessToast: true
      }
    )
    
    if (result) {
      setShowEditModal(false)
      resetForm()
      fetchApps()
    }
  }

  // 重置表单
  const resetForm = () => {
    setFormData({
      name: '',
      repo: '',
      deploy_path: '',
      config_path: '',
      start_cmd: '',
      stop_cmd: '',
      red_metrics_config: {
        enabled: false,
        rate_metric: undefined,
        error_metric: undefined,
        duration_metric: undefined,
        health_threshold: undefined
      },
      rollback_policy: {
        enabled: false,
        alert_rules: [],
        auto_rollback: false,
        notify_channel: ''
      }
    })
    setSelectedApp(null)
    setActiveTab('basic')
  }

  // 打开编辑模态框
  const openEditModal = (app: Application) => {
    setSelectedApp(app)
    setFormData({
      name: app.name,
      repo: app.repo || '',
      deploy_path: app.deploy_path,
      config_path: app.config_path || '',
      start_cmd: app.start_cmd,
      stop_cmd: app.stop_cmd,
      red_metrics_config: app.red_metrics_config || {
        enabled: false,
        rate_metric: undefined,
        error_metric: undefined,
        duration_metric: undefined,
        health_threshold: undefined
      },
      rollback_policy: app.rollback_policy || {
        enabled: false,
        alert_rules: [],
        auto_rollback: false,
        notify_channel: ''
      }
    })
    setActiveTab('basic')
    setShowEditModal(true)
  }

  // 打开详情模态框
  const openDetailModal = async (appId: string) => {
    const result = await request(
      () => appApi.getAppDetail(appId) as unknown as Promise<GetAppDetailResp>,
      {
        errorMessage: '获取应用详情失败'
      }
    )
    
    if (result) {
      setSelectedApp(result.application)
      setShowDetailModal(true)
    }
  }

  // 打开机器列表模态框
  const openMachineListModal = async (appId: string) => {
    const result = await request(
      () => appApi.getAppDetail(appId) as unknown as Promise<GetAppDetailResp>,
      {
        errorMessage: '获取应用机器列表失败'
      }
    )
    
    if (result) {
      setSelectedApp(result.application)
      setIsMachineEditMode(false)
      setShowMachineListModal(true)
    }
  }

  // 进入编辑模式
  const enterMachineEditMode = async () => {
    // 获取所有可用机器
    const result = await request(
      () => machineApi.getMachineList({ page: 1, page_size: 1000 }) as unknown as Promise<GetMachineListResp>,
      {
        errorMessage: '获取机器列表失败'
      }
    )
    
    if (result) {
      setAvailableMachines(result.machines || [])
      // 设置当前已选中的机器ID
      setSelectedMachineIds(selectedApp?.machines?.map(m => m.id) || [])
      setIsMachineEditMode(true)
    }
  }

  // 保存机器关联
  const saveMachineAssociations = async () => {
    if (!selectedApp) return
    
    const result = await request(
      () => appApi.updateApp(selectedApp.id, {
        id: selectedApp.id,
        name: selectedApp.name,
        repo: selectedApp.repo,
        deploy_path: selectedApp.deploy_path,
        config_path: selectedApp.config_path,
        start_cmd: selectedApp.start_cmd,
        stop_cmd: selectedApp.stop_cmd,
        machine_ids: selectedMachineIds,
        red_metrics_config: selectedApp.red_metrics_config,
        rollback_policy: selectedApp.rollback_policy
      }),
      {
        successMessage: '机器关联更新成功',
        errorMessage: '更新机器关联失败',
        showSuccessToast: true
      }
    )
    
    if (result) {
      setIsMachineEditMode(false)
      // 重新获取应用详情
      const updatedApp = await request(
        () => appApi.getAppDetail(selectedApp.id) as unknown as Promise<GetAppDetailResp>,
        {
          errorMessage: '获取应用详情失败'
        }
      )
      if (updatedApp) {
        setSelectedApp(updatedApp.application)
      }
      fetchApps()
    }
  }

  // 切换机器选择
  const toggleMachineSelection = (machineId: string) => {
    setSelectedMachineIds(prev => {
      if (prev.includes(machineId)) {
        return prev.filter(id => id !== machineId)
      } else {
        return [...prev, machineId]
      }
    })
  }

  // 获取状态颜色
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'healthy':
      case 'normal':
        return '#52c41a'
      case 'unhealthy':
      case 'error':
        return '#ff4d4f'
      case 'alert':
        return '#faad14'
      default:
        return '#d9d9d9'
    }
  }

  // 获取状态文本
  const getStatusText = (status: string) => {
    switch (status) {
      case 'healthy':
        return '健康'
      case 'unhealthy':
        return '不健康'
      case 'normal':
        return '正常'
      case 'error':
        return '异常'
      case 'alert':
        return '告警'
      default:
        return status
    }
  }

  return (
    <PageLayout breadcrumbItems={[{ label: '应用管理', path: '/apps' }]}>
      <Toaster 
        position="top-right"
        toastOptions={{
          duration: 4000,
          style: {
            background: 'var(--bg-primary)',
            color: 'var(--text-primary)',
            border: '1px solid var(--border-color)',
          },
        }}
      />
      <div className="apps-header">
        <div className="apps-actions">
          <div className="search-box">
            <input
              type="text"
              placeholder="搜索应用名称"
              value={searchName}
              onChange={(e) => setSearchName(e.target.value)}
            />
          </div>
          <button 
            className="btn btn-primary"
            onClick={() => setShowCreateModal(true)}
          >
            创建应用
          </button>
        </div>
      </div>

      <div className="apps-content">
        {loading ? (
          <div className="loading">加载中...</div>
        ) : (
          <div className="apps-table">
            <table>
              <thead>
                <tr>
                  <th>应用名称</th>
                  <th>版本</th>
                  <th>部署路径</th>
                  <th>配置路径</th>
                  <th>机器状态</th>
                  <th>创建时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {apps.map((app) => (
                  <tr key={app.id}>
                    <td>{app.name}</td>
                      <td>{app.currentVersion}</td>
                      <td>{app.deploy_path}</td>
                      <td>{app.config_path || '-'}</td>
                      <td>
                      <div className="status-info">
                        <span className="status-item">
                          总数: {app.machine_count}
                        </span>
                        <span className="status-item" style={{ color: getStatusColor('healthy') }}>
                          健康: {app.health_count}
                        </span>
                        <span className="status-item" style={{ color: getStatusColor('error') }}>
                          异常: {app.error_count}
                        </span>
                        <span className="status-item" style={{ color: getStatusColor('alert') }}>
                          告警: {app.alert_count}
                        </span>
                      </div>
                    </td>
                    <td>{new Date(app.created_at * 1000).toLocaleString()}</td>
                    <td>
                      <div className="action-buttons">
                        <button 
                          className="btn btn-sm btn-info"
                          onClick={() => openDetailModal(app.id)}
                        >
                          详情
                        </button>
                        <button 
                          className="btn btn-sm btn-success"
                          onClick={() => openMachineListModal(app.id)}
                        >
                          机器列表
                        </button>
                        <button 
                          className="btn btn-sm btn-warning"
                          onClick={() => openEditModal(app)}
                        >
                          编辑
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* 分页 */}
        <div className="pagination">
          <button 
            disabled={pagination.page === 1}
            onClick={() => setPagination(prev => ({ ...prev, page: prev.page - 1 }))}
          >
            上一页
          </button>
          <span>
            第 {pagination.page} 页，共 {Math.ceil(pagination.total / pagination.pageSize)} 页
          </span>
          <button 
            disabled={pagination.page >= Math.ceil(pagination.total / pagination.pageSize)}
            onClick={() => setPagination(prev => ({ ...prev, page: prev.page + 1 }))}
          >
            下一页
          </button>
        </div>
      </div>

      {/* 创建应用模态框 */}
      {showCreateModal && (
        <div className="modal-overlay">
          <div className="modal">
            <div className="modal-header">
              <h3>创建应用</h3>
              <button onClick={() => setShowCreateModal(false)}>×</button>
            </div>
            <div className="modal-body">
              <div className="form-group">
                <label>应用名称</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
                />
              </div>
              <div className="form-group">
                <label>仓库地址（可选）</label>
                <input
                  type="text"
                  value={formData.repo}
                  onChange={(e) => setFormData(prev => ({ ...prev, repo: e.target.value }))}
                  placeholder="例如: https://github.com/owner/repo"
                />
              </div>
              <div className="form-group">
                <label>部署路径</label>
                <input
                  type="text"
                  value={formData.deploy_path}
                  onChange={(e) => setFormData(prev => ({ ...prev, deploy_path: e.target.value }))}
                />
              </div>
              <div className="form-group">
                <label>配置文件路径（可选）</label>
                <input
                  type="text"
                  value={formData.config_path}
                  onChange={(e) => setFormData(prev => ({ ...prev, config_path: e.target.value }))}
                  placeholder="例如: /etc/app/config.yaml"
                />
              </div>
              <div className="form-group">
                <label>启动命令</label>
                <input
                  type="text"
                  value={formData.start_cmd}
                  onChange={(e) => setFormData(prev => ({ ...prev, start_cmd: e.target.value }))}
                />
              </div>
              <div className="form-group">
                <label>停止命令</label>
                <input
                  type="text"
                  value={formData.stop_cmd}
                  onChange={(e) => setFormData(prev => ({ ...prev, stop_cmd: e.target.value }))}
                />
              </div>
            </div>
            <div className="modal-footer">
              <button onClick={() => setShowCreateModal(false)}>取消</button>
              <button className="btn-primary" onClick={handleCreateApp}>创建</button>
            </div>
          </div>
        </div>
      )}

      {/* 编辑应用模态框 */}
      {showEditModal && (
        <div className="modal-overlay">
          <div className="modal modal-large">
            <div className="modal-header">
              <h3>编辑应用</h3>
              <button onClick={() => setShowEditModal(false)}>×</button>
            </div>
            
            {/* 标签页导航 */}
            <div style={{ borderBottom: '1px solid var(--border-color)', padding: '0 20px' }}>
              <div style={{ display: 'flex', gap: '20px' }}>
                <button
                  onClick={() => setActiveTab('basic')}
                  style={{
                    padding: '10px 20px',
                    border: 'none',
                    background: 'transparent',
                    cursor: 'pointer',
                    borderBottom: activeTab === 'basic' ? '2px solid #1890ff' : 'none',
                    color: activeTab === 'basic' ? '#1890ff' : 'inherit'
                  }}
                >
                  基本信息
                </button>
                <button
                  onClick={() => setActiveTab('red')}
                  style={{
                    padding: '10px 20px',
                    border: 'none',
                    background: 'transparent',
                    cursor: 'pointer',
                    borderBottom: activeTab === 'red' ? '2px solid #1890ff' : 'none',
                    color: activeTab === 'red' ? '#1890ff' : 'inherit'
                  }}
                >
                  RED 指标配置
                </button>
                <button
                  onClick={() => setActiveTab('rollback')}
                  style={{
                    padding: '10px 20px',
                    border: 'none',
                    background: 'transparent',
                    cursor: 'pointer',
                    borderBottom: activeTab === 'rollback' ? '2px solid #1890ff' : 'none',
                    color: activeTab === 'rollback' ? '#1890ff' : 'inherit'
                  }}
                >
                  回滚策略
                </button>
              </div>
            </div>

            <div className="modal-body" style={{ maxHeight: '60vh', overflowY: 'auto' }}>
              {/* 基本信息标签页 */}
              {activeTab === 'basic' && (
                <>
                  <div className="form-group">
                    <label>应用名称</label>
                    <input
                      type="text"
                      value={formData.name}
                      onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
                    />
                  </div>
                  <div className="form-group">
                    <label>仓库地址（可选）</label>
                    <input
                      type="text"
                      value={formData.repo}
                      onChange={(e) => setFormData(prev => ({ ...prev, repo: e.target.value }))}
                      placeholder="例如: https://github.com/owner/repo"
                    />
                  </div>
                  <div className="form-group">
                    <label>部署路径</label>
                    <input
                      type="text"
                      value={formData.deploy_path}
                      onChange={(e) => setFormData(prev => ({ ...prev, deploy_path: e.target.value }))}
                    />
                  </div>
                  <div className="form-group">
                    <label>配置文件路径（可选）</label>
                    <input
                      type="text"
                      value={formData.config_path}
                      onChange={(e) => setFormData(prev => ({ ...prev, config_path: e.target.value }))}
                      placeholder="例如: /etc/app/config.yaml"
                    />
                  </div>
                  <div className="form-group">
                    <label>启动命令</label>
                    <input
                      type="text"
                      value={formData.start_cmd}
                      onChange={(e) => setFormData(prev => ({ ...prev, start_cmd: e.target.value }))}
                    />
                  </div>
                  <div className="form-group">
                    <label>停止命令</label>
                    <input
                      type="text"
                      value={formData.stop_cmd}
                      onChange={(e) => setFormData(prev => ({ ...prev, stop_cmd: e.target.value }))}
                    />
                  </div>
                </>
              )}

              {/* RED 指标配置标签页 */}
              {activeTab === 'red' && (
                <>
                  <div className="form-group">
                    <label style={{ display: 'flex', alignItems: 'center', gap: '10px', width: 'fit-content' }}>
                      <input
                        type="checkbox"
                        style={{ width: '16px', height: '16px' }}
                        checked={formData.red_metrics_config?.enabled || false}
                        onChange={(e) => setFormData(prev => ({
                          ...prev,
                          red_metrics_config: {
                            ...prev.red_metrics_config!,
                            enabled: e.target.checked
                          }
                        }))}
                      />
                      启用 RED 指标监控
                    </label>
                  </div>

                  {formData.red_metrics_config?.enabled && (
                    <>
                      {/* Rate 指标配置 */}
                      <div className="detail-section" style={{ 
                        marginTop: '24px', 
                        padding: '20px',
                        backgroundColor: '#f8f9fa',
                        borderRadius: '8px',
                        border: '1px solid #e9ecef'
                      }}>
                        <h4 style={{ marginBottom: '20px', fontSize: '16px', fontWeight: '600' }}>Rate (请求速率) 指标</h4>
                        <div className="form-group" style={{ marginBottom: '16px' }}>
                          <label style={{ 
                            display: 'block', 
                            marginBottom: '8px', 
                            fontSize: '14px', 
                            fontWeight: '500',
                            color: '#495057'
                          }}>指标名称</label>
                          <input
                            type="text"
                            style={{ width: '100%' }}
                            value={formData.red_metrics_config?.rate_metric?.metric_name || ''}
                            onChange={(e) => setFormData(prev => ({
                              ...prev,
                              red_metrics_config: {
                                ...prev.red_metrics_config!,
                                rate_metric: {
                                  ...prev.red_metrics_config?.rate_metric,
                                  metric_name: e.target.value,
                                  promql: prev.red_metrics_config?.rate_metric?.promql || '',
                                  labels: prev.red_metrics_config?.rate_metric?.labels || {},
                                  description: prev.red_metrics_config?.rate_metric?.description || ''
                                }
                              }
                            }))}
                            placeholder="例如: http_requests_total"
                          />
                        </div>
                        <div className="form-group" style={{ marginBottom: '16px' }}>
                          <label style={{ 
                            display: 'block', 
                            marginBottom: '8px', 
                            fontSize: '14px', 
                            fontWeight: '500',
                            color: '#495057'
                          }}>PromQL 查询语句</label>
                          <textarea
                            style={{ width: '100%' }}
                            value={formData.red_metrics_config?.rate_metric?.promql || ''}
                            onChange={(e) => setFormData(prev => ({
                              ...prev,
                              red_metrics_config: {
                                ...prev.red_metrics_config!,
                                rate_metric: {
                                  ...prev.red_metrics_config?.rate_metric!,
                                  promql: e.target.value
                                }
                              }
                            }))}
                            placeholder="例如: rate(http_requests_total[5m])"
                            rows={6}
                          />
                        </div>
                        <div className="form-group" style={{ marginBottom: '0' }}>
                          <label style={{ 
                            display: 'block', 
                            marginBottom: '8px', 
                            fontSize: '14px', 
                            fontWeight: '500',
                            color: '#495057'
                          }}>描述</label>
                          <input
                            type="text"
                            style={{ width: '100%' }}
                            value={formData.red_metrics_config?.rate_metric?.description || ''}
                            onChange={(e) => setFormData(prev => ({
                              ...prev,
                              red_metrics_config: {
                                ...prev.red_metrics_config!,
                                rate_metric: {
                                  ...prev.red_metrics_config?.rate_metric!,
                                  description: e.target.value
                                }
                              }
                            }))}
                            placeholder="指标描述"
                          />
                        </div>
                      </div>

                      {/* Error 指标配置 */}
                      <div className="detail-section" style={{ 
                        marginTop: '24px', 
                        padding: '20px',
                        backgroundColor: '#f8f9fa',
                        borderRadius: '8px',
                        border: '1px solid #e9ecef'
                      }}>
                        <h4 style={{ marginBottom: '20px', fontSize: '16px', fontWeight: '600' }}>Error (错误率) 指标</h4>
                        <div className="form-group" style={{ marginBottom: '16px' }}>
                          <label style={{ 
                            display: 'block', 
                            marginBottom: '8px', 
                            fontSize: '14px', 
                            fontWeight: '500',
                            color: '#495057'
                          }}>指标名称</label>
                          <input
                            type="text"
                            style={{ width: '100%' }}
                            value={formData.red_metrics_config?.error_metric?.metric_name || ''}
                            onChange={(e) => setFormData(prev => ({
                              ...prev,
                              red_metrics_config: {
                                ...prev.red_metrics_config!,
                                error_metric: {
                                  ...prev.red_metrics_config?.error_metric,
                                  metric_name: e.target.value,
                                  promql: prev.red_metrics_config?.error_metric?.promql || '',
                                  labels: prev.red_metrics_config?.error_metric?.labels || {},
                                  description: prev.red_metrics_config?.error_metric?.description || ''
                                }
                              }
                            }))}
                            placeholder="例如: http_requests_errors_total"
                          />
                        </div>
                        <div className="form-group" style={{ marginBottom: '16px' }}>
                          <label style={{ 
                            display: 'block', 
                            marginBottom: '8px', 
                            fontSize: '14px', 
                            fontWeight: '500',
                            color: '#495057'
                          }}>PromQL 查询语句</label>
                          <textarea
                            style={{ width: '100%' }}
                            value={formData.red_metrics_config?.error_metric?.promql || ''}
                            onChange={(e) => setFormData(prev => ({
                              ...prev,
                              red_metrics_config: {
                                ...prev.red_metrics_config!,
                                error_metric: {
                                  ...prev.red_metrics_config?.error_metric!,
                                  promql: e.target.value
                                }
                              }
                            }))}
                            placeholder="例如: rate(http_requests_errors_total[5m])"
                            rows={6}
                          />
                        </div>
                        <div className="form-group" style={{ marginBottom: '0' }}>
                          <label style={{ 
                            display: 'block', 
                            marginBottom: '8px', 
                            fontSize: '14px', 
                            fontWeight: '500',
                            color: '#495057'
                          }}>描述</label>
                          <input
                            type="text"
                            style={{ width: '100%' }}
                            value={formData.red_metrics_config?.error_metric?.description || ''}
                            onChange={(e) => setFormData(prev => ({
                              ...prev,
                              red_metrics_config: {
                                ...prev.red_metrics_config!,
                                error_metric: {
                                  ...prev.red_metrics_config?.error_metric!,
                                  description: e.target.value
                                }
                              }
                            }))}
                            placeholder="指标描述"
                          />
                        </div>
                      </div>

                      {/* Duration 指标配置 */}
                      <div className="detail-section" style={{ 
                        marginTop: '24px', 
                        padding: '20px',
                        backgroundColor: '#f8f9fa',
                        borderRadius: '8px',
                        border: '1px solid #e9ecef'
                      }}>
                        <h4 style={{ marginBottom: '20px', fontSize: '16px', fontWeight: '600' }}>Duration (响应时长) 指标</h4>
                        <div className="form-group" style={{ marginBottom: '16px' }}>
                          <label style={{ 
                            display: 'block', 
                            marginBottom: '8px', 
                            fontSize: '14px', 
                            fontWeight: '500',
                            color: '#495057'
                          }}>指标名称</label>
                          <input
                            type="text"
                            style={{ width: '100%' }}
                            value={formData.red_metrics_config?.duration_metric?.metric_name || ''}
                            onChange={(e) => setFormData(prev => ({
                              ...prev,
                              red_metrics_config: {
                                ...prev.red_metrics_config!,
                                duration_metric: {
                                  ...prev.red_metrics_config?.duration_metric,
                                  metric_name: e.target.value,
                                  promql: prev.red_metrics_config?.duration_metric?.promql || '',
                                  labels: prev.red_metrics_config?.duration_metric?.labels || {},
                                  description: prev.red_metrics_config?.duration_metric?.description || ''
                                }
                              }
                            }))}
                            placeholder="例如: http_request_duration_seconds"
                          />
                        </div>
                        <div className="form-group" style={{ marginBottom: '16px' }}>
                          <label style={{ 
                            display: 'block', 
                            marginBottom: '8px', 
                            fontSize: '14px', 
                            fontWeight: '500',
                            color: '#495057'
                          }}>PromQL 查询语句</label>
                          <textarea
                            style={{ width: '100%' }}
                            value={formData.red_metrics_config?.duration_metric?.promql || ''}
                            onChange={(e) => setFormData(prev => ({
                              ...prev,
                              red_metrics_config: {
                                ...prev.red_metrics_config!,
                                duration_metric: {
                                  ...prev.red_metrics_config?.duration_metric!,
                                  promql: e.target.value
                                }
                              }
                            }))}
                            placeholder="例如: histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))"
                            rows={6}
                          />
                        </div>
                        <div className="form-group" style={{ marginBottom: '0' }}>
                          <label style={{ 
                            display: 'block', 
                            marginBottom: '8px', 
                            fontSize: '14px', 
                            fontWeight: '500',
                            color: '#495057'
                          }}>描述</label>
                          <input
                            type="text"
                            style={{ width: '100%' }}
                            value={formData.red_metrics_config?.duration_metric?.description || ''}
                            onChange={(e) => setFormData(prev => ({
                              ...prev,
                              red_metrics_config: {
                                ...prev.red_metrics_config!,
                                duration_metric: {
                                  ...prev.red_metrics_config?.duration_metric!,
                                  description: e.target.value
                                }
                              }
                            }))}
                            placeholder="指标描述"
                          />
                        </div>
                      </div>

                      {/* 健康度阈值配置 */}
                      <div className="detail-section" style={{ 
                        marginTop: '24px', 
                        padding: '20px',
                        backgroundColor: '#f8f9fa',
                        borderRadius: '8px',
                        border: '1px solid #e9ecef'
                      }}>
                        <h4 style={{ marginBottom: '20px', fontSize: '16px', fontWeight: '600' }}>健康度阈值</h4>
                        <div className="form-group" style={{ marginBottom: '16px' }}>
                          <label style={{ 
                            display: 'block', 
                            marginBottom: '8px', 
                            fontSize: '14px', 
                            fontWeight: '500',
                            color: '#495057'
                          }}>最低请求速率 (rate_min)</label>
                          <input
                            type="number"
                            step="0.01"
                            style={{ width: '100%' }}
                            value={formData.red_metrics_config?.health_threshold?.rate_min || ''}
                            onChange={(e) => setFormData(prev => ({
                              ...prev,
                              red_metrics_config: {
                                ...prev.red_metrics_config!,
                                health_threshold: {
                                  ...prev.red_metrics_config?.health_threshold,
                                  rate_min: parseFloat(e.target.value) || 0,
                                  error_rate_max: prev.red_metrics_config?.health_threshold?.error_rate_max || 0,
                                  duration_p95_max: prev.red_metrics_config?.health_threshold?.duration_p95_max || 0
                                }
                              }
                            }))}
                            placeholder="例如: 10"
                          />
                        </div>
                        <div className="form-group" style={{ marginBottom: '16px' }}>
                          <label style={{ 
                            display: 'block', 
                            marginBottom: '8px', 
                            fontSize: '14px', 
                            fontWeight: '500',
                            color: '#495057'
                          }}>最大错误率 (error_rate_max)</label>
                          <input
                            type="number"
                            step="0.001"
                            style={{ width: '100%' }}
                            value={formData.red_metrics_config?.health_threshold?.error_rate_max || ''}
                            onChange={(e) => setFormData(prev => ({
                              ...prev,
                              red_metrics_config: {
                                ...prev.red_metrics_config!,
                                health_threshold: {
                                  ...prev.red_metrics_config?.health_threshold!,
                                  error_rate_max: parseFloat(e.target.value) || 0
                                }
                              }
                            }))}
                            placeholder="例如: 0.05 (5%)"
                          />
                        </div>
                        <div className="form-group" style={{ marginBottom: '0' }}>
                          <label style={{ 
                            display: 'block', 
                            marginBottom: '8px', 
                            fontSize: '14px', 
                            fontWeight: '500',
                            color: '#495057'
                          }}>P95 响应时长上限 (毫秒)</label>
                          <input
                            type="number"
                            step="0.01"
                            style={{ width: '100%' }}
                            value={formData.red_metrics_config?.health_threshold?.duration_p95_max || ''}
                            onChange={(e) => setFormData(prev => ({
                              ...prev,
                              red_metrics_config: {
                                ...prev.red_metrics_config!,
                                health_threshold: {
                                  ...prev.red_metrics_config?.health_threshold!,
                                  duration_p95_max: parseFloat(e.target.value) || 0
                                }
                              }
                            }))}
                            placeholder="例如: 0.5"
                          />
                        </div>
                      </div>
                    </>
                  )}
                </>
              )}

              {/* 回滚策略标签页 */}
              {activeTab === 'rollback' && (
                <>
                  <div className="form-group">
                    <label style={{ display: 'flex', alignItems: 'center', gap: '10px', width: 'fit-content' }}>
                      <input
                        type="checkbox"
                        style={{ width: '16px', height: '16px' }}
                        checked={formData.rollback_policy?.enabled || false}
                        onChange={(e) => setFormData(prev => ({
                          ...prev,
                          rollback_policy: {
                            ...prev.rollback_policy!,
                            enabled: e.target.checked
                          }
                        }))}
                      />
                      启用回滚策略
                    </label>
                  </div>

                  {formData.rollback_policy?.enabled && (
                    <>
                      <div className="form-group">
                        <label style={{ display: 'flex', alignItems: 'center', gap: '10px', width: 'fit-content' }}>
                          <input
                            type="checkbox"
                            style={{ width: '16px', height: '16px' }}
                            checked={formData.rollback_policy?.auto_rollback || false}
                            onChange={(e) => setFormData(prev => ({
                              ...prev,
                              rollback_policy: {
                                ...prev.rollback_policy!,
                                auto_rollback: e.target.checked
                              }
                            }))}
                          />
                          自动执行回滚
                        </label>
                      </div>

                      <div className="form-group">
                        <label>通知渠道</label>
                        <input
                          type="text"
                          value={formData.rollback_policy?.notify_channel || ''}
                          onChange={(e) => setFormData(prev => ({
                            ...prev,
                            rollback_policy: {
                              ...prev.rollback_policy!,
                              notify_channel: e.target.value
                            }
                          }))}
                          placeholder="例如: webhook_url 或 email"
                        />
                      </div>

                      <div className="detail-section" style={{ marginTop: '20px' }}>
                        <h4>Prometheus 告警规则</h4>
                        <p style={{ color: '#666', fontSize: '14px', marginBottom: '10px' }}>
                          配置触发回滚的 Prometheus 告警规则
                        </p>
                        
                        {(formData.rollback_policy?.alert_rules || []).map((rule, index) => (
                          <div key={index} style={{ 
                            padding: '15px', 
                            border: '1px solid var(--border-color)', 
                            borderRadius: '4px',
                            marginBottom: '15px'
                          }}>
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '10px' }}>
                              <h5 style={{ margin: 0 }}>告警规则 #{index + 1}</h5>
                              <button 
                                onClick={() => {
                                  const newRules = [...(formData.rollback_policy?.alert_rules || [])]
                                  newRules.splice(index, 1)
                                  setFormData(prev => ({
                                    ...prev,
                                    rollback_policy: {
                                      ...prev.rollback_policy!,
                                      alert_rules: newRules
                                    }
                                  }))
                                }}
                                style={{ 
                                  padding: '5px 10px',
                                  background: '#ff4d4f',
                                  color: 'white',
                                  border: 'none',
                                  borderRadius: '4px',
                                  cursor: 'pointer'
                                }}
                              >
                                删除
                              </button>
                            </div>
                            
                            <div className="form-group">
                              <label>告警名称</label>
                              <input
                                type="text"
                                value={rule.name}
                                onChange={(e) => {
                                  const newRules = [...(formData.rollback_policy?.alert_rules || [])]
                                  newRules[index] = { ...newRules[index], name: e.target.value }
                                  setFormData(prev => ({
                                    ...prev,
                                    rollback_policy: {
                                      ...prev.rollback_policy!,
                                      alert_rules: newRules
                                    }
                                  }))
                                }}
                                placeholder="例如: HighErrorRate"
                              />
                            </div>
                            
                            <div className="form-group">
                              <label>PromQL 表达式</label>
                              <textarea
                                value={rule.alert_expr}
                                onChange={(e) => {
                                  const newRules = [...(formData.rollback_policy?.alert_rules || [])]
                                  newRules[index] = { ...newRules[index], alert_expr: e.target.value }
                                  setFormData(prev => ({
                                    ...prev,
                                    rollback_policy: {
                                      ...prev.rollback_policy!,
                                      alert_rules: newRules
                                    }
                                  }))
                                }}
                                placeholder="例如: rate(http_requests_errors_total[5m]) > 0.1"
                                rows={5}
                              />
                            </div>
                            
                            <div className="form-group">
                              <label>持续时长</label>
                              <input
                                type="text"
                                value={rule.duration}
                                onChange={(e) => {
                                  const newRules = [...(formData.rollback_policy?.alert_rules || [])]
                                  newRules[index] = { ...newRules[index], duration: e.target.value }
                                  setFormData(prev => ({
                                    ...prev,
                                    rollback_policy: {
                                      ...prev.rollback_policy!,
                                      alert_rules: newRules
                                    }
                                  }))
                                }}
                                placeholder="例如: 5m"
                              />
                            </div>
                            
                            <div className="form-group">
                              <label>告警级别</label>
                              <select
                                value={rule.severity}
                                onChange={(e) => {
                                  const newRules = [...(formData.rollback_policy?.alert_rules || [])]
                                  newRules[index] = { ...newRules[index], severity: e.target.value }
                                  setFormData(prev => ({
                                    ...prev,
                                    rollback_policy: {
                                      ...prev.rollback_policy!,
                                      alert_rules: newRules
                                    }
                                  }))
                                }}
                              >
                                <option value="critical">Critical</option>
                                <option value="warning">Warning</option>
                                <option value="info">Info</option>
                              </select>
                            </div>
                          </div>
                        ))}
                        
                        <button
                          onClick={() => {
                            const newRule: PrometheusAlert = {
                              name: '',
                              alert_expr: '',
                              duration: '5m',
                              severity: 'critical',
                              labels: {},
                              annotations: {}
                            }
                            setFormData(prev => ({
                              ...prev,
                              rollback_policy: {
                                ...prev.rollback_policy!,
                                alert_rules: [...(prev.rollback_policy?.alert_rules || []), newRule]
                              }
                            }))
                          }}
                          style={{
                            padding: '10px 20px',
                            background: '#1890ff',
                            color: 'white',
                            border: 'none',
                            borderRadius: '4px',
                            cursor: 'pointer'
                          }}
                        >
                          + 添加告警规则
                        </button>
                      </div>
                    </>
                  )}
                </>
              )}
            </div>
            
            <div className="modal-footer">
              <button onClick={() => setShowEditModal(false)}>取消</button>
              <button className="btn-primary" onClick={handleUpdateApp}>保存</button>
            </div>
          </div>
        </div>
      )}

      {/* 机器列表模态框 */}
      {showMachineListModal && selectedApp && (
        <div className="modal-overlay">
          <div className="modal modal-large">
            <div className="modal-header">
              <h3>{selectedApp.name} - 机器列表</h3>
              <button onClick={() => setShowMachineListModal(false)}>×</button>
            </div>
            <div className="modal-body">
              {!isMachineEditMode ? (
                <>
                  <div className="detail-section">
                    <h4>机器状态统计</h4>
                    <div className="status-stats">
                      <div className="stat-item">
                        <span className="stat-label">总机器数:</span>
                        <span className="stat-value">{selectedApp.machine_count}</span>
                      </div>
                      <div className="stat-item">
                        <span className="stat-label">健康机器:</span>
                        <span className="stat-value" style={{ color: getStatusColor('healthy') }}>
                          {selectedApp.health_count}
                        </span>
                      </div>
                      <div className="stat-item">
                        <span className="stat-label">异常机器:</span>
                        <span className="stat-value" style={{ color: getStatusColor('error') }}>
                          {selectedApp.error_count}
                        </span>
                      </div>
                      <div className="stat-item">
                        <span className="stat-label">告警机器:</span>
                        <span className="stat-value" style={{ color: getStatusColor('alert') }}>
                          {selectedApp.alert_count}
                        </span>
                      </div>
                    </div>
                  </div>

                  <div className="detail-section">
                    <h4>机器列表</h4>
                    {selectedApp.machines && selectedApp.machines.length > 0 ? (
                      <div className="machines-table">
                        <table>
                          <thead>
                            <tr>
                              <th>机器名称</th>
                              <th>IP地址</th>
                              <th>端口</th>
                              <th>描述</th>
                              <th>健康状态</th>
                              <th>异常状态</th>
                              <th>告警状态</th>
                            </tr>
                          </thead>
                          <tbody>
                            {selectedApp.machines.map((machine) => (
                              <tr key={machine.id}>
                                <td>{machine.name}</td>
                                <td>{machine.ip}</td>
                                <td>{machine.port}</td>
                                <td>{machine.description}</td>
                                <td>
                                  <span style={{ color: getStatusColor(machine.health_status) }}>
                                    {getStatusText(machine.health_status)}
                                  </span>
                                </td>
                                <td>
                                  <span style={{ color: getStatusColor(machine.error_status) }}>
                                    {getStatusText(machine.error_status)}
                                  </span>
                                </td>
                                <td>
                                  <span style={{ color: getStatusColor(machine.alert_status) }}>
                                    {getStatusText(machine.alert_status)}
                                  </span>
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    ) : (
                      <div className="empty-state">暂无关联机器</div>
                    )}
                  </div>
                </>
              ) : (
                <div className="detail-section">
                  <h4>选择关联机器</h4>
                  <div className="machine-selection">
                    {availableMachines.map((machine) => (
                      <div key={machine.id} className="machine-checkbox-item">
                        <label>
                          <input
                            type="checkbox"
                            checked={selectedMachineIds.includes(machine.id)}
                            onChange={() => toggleMachineSelection(machine.id)}
                          />
                          <span className="machine-info">
                            <strong>{machine.name}</strong> - {machine.ip}:{machine.port}
                            {machine.description && <span className="machine-desc"> ({machine.description})</span>}
                          </span>
                        </label>
                      </div>
                    ))}
                  </div>
                  <div style={{ marginTop: '10px', color: '#666' }}>
                    已选择 {selectedMachineIds.length} 台机器
                  </div>
                </div>
              )}
            </div>
            <div className="modal-footer">
              {!isMachineEditMode ? (
                <>
                  <button className="btn-primary" onClick={enterMachineEditMode}>编辑机器关联</button>
                  <button onClick={() => setShowMachineListModal(false)}>关闭</button>
                </>
              ) : (
                <>
                  <button className="btn-primary" onClick={saveMachineAssociations}>保存</button>
                  <button onClick={() => setIsMachineEditMode(false)}>取消</button>
                </>
              )}
            </div>
          </div>
        </div>
      )}

      {/* 应用详情模态框 */}
      {showDetailModal && selectedApp && (
        <div className="modal-overlay">
          <div className="modal modal-large">
            <div className="modal-header">
              <h3>应用详情 - {selectedApp.name}</h3>
              <button onClick={() => setShowDetailModal(false)}>×</button>
            </div>
            <div className="modal-body">
              <div className="detail-section">
                <h4>基本信息</h4>
                <div className="detail-grid">
                  <div className="detail-item">
                    <label>应用ID:</label>
                    <span>{selectedApp.id}</span>
                  </div>
                  <div className="detail-item">
                    <label>应用名称:</label>
                    <span>{selectedApp.name}</span>
                  </div>
                  <div className="detail-item">
                    <label>仓库地址:</label>
                    <span>{selectedApp.repo || '-'}</span>
                  </div>
                  <div className="detail-item">
                    <label>版本:</label>
                    <span>{selectedApp.currentVersion}</span>
                  </div>
                  <div className="detail-item">
                    <label>部署路径:</label>
                    <span>{selectedApp.deploy_path}</span>
                  </div>
                  <div className="detail-item">
                    <label>配置文件路径:</label>
                    <span>{selectedApp.config_path || '-'}</span>
                  </div>
                  <div className="detail-item">
                    <label>启动命令:</label>
                    <span>{selectedApp.start_cmd}</span>
                  </div>
                  <div className="detail-item">
                    <label>停止命令:</label>
                    <span>{selectedApp.stop_cmd}</span>
                  </div>
                  <div className="detail-item">
                    <label>创建时间:</label>
                    <span>{new Date(selectedApp.created_at * 1000).toLocaleString()}</span>
                  </div>
                  <div className="detail-item">
                    <label>更新时间:</label>
                    <span>{new Date(selectedApp.updated_at * 1000).toLocaleString()}</span>
                  </div>
                </div>
              </div>

              <div className="detail-section">
                <h4>机器状态统计</h4>
                <div className="status-stats">
                  <div className="stat-item">
                    <span className="stat-label">总机器数:</span>
                    <span className="stat-value">{selectedApp.machine_count}</span>
                  </div>
                  <div className="stat-item">
                    <span className="stat-label">健康机器:</span>
                    <span className="stat-value" style={{ color: getStatusColor('healthy') }}>
                      {selectedApp.health_count}
                    </span>
                  </div>
                  <div className="stat-item">
                    <span className="stat-label">异常机器:</span>
                    <span className="stat-value" style={{ color: getStatusColor('error') }}>
                      {selectedApp.error_count}
                    </span>
                  </div>
                  <div className="stat-item">
                    <span className="stat-label">告警机器:</span>
                    <span className="stat-value" style={{ color: getStatusColor('alert') }}>
                      {selectedApp.alert_count}
                    </span>
                  </div>
                </div>
              </div>

              <div className="detail-section">
                <h4>机器列表</h4>
                <div className="machines-table">
                  <table>
                    <thead>
                      <tr>
                        <th>机器ID</th>
                        <th>IP地址</th>
                        <th>端口</th>
                        <th>健康状态</th>
                        <th>异常状态</th>
                        <th>告警状态</th>
                      </tr>
                    </thead>
                    <tbody>
                      {(selectedApp.machines || []).map((machine) => (
                        <tr key={machine.id}>
                          <td>{machine.id}</td>
                          <td>{machine.ip}</td>
                          <td>{machine.port}</td>
                          <td>
                            <span style={{ color: getStatusColor(machine.health_status) }}>
                              {getStatusText(machine.health_status)}
                            </span>
                          </td>
                          <td>
                            <span style={{ color: getStatusColor(machine.error_status) }}>
                              {getStatusText(machine.error_status)}
                            </span>
                          </td>
                          <td>
                            <span style={{ color: getStatusColor(machine.alert_status) }}>
                              {getStatusText(machine.alert_status)}
                            </span>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>

              {/* RED 指标配置 */}
              {selectedApp.red_metrics_config && selectedApp.red_metrics_config.enabled && (
                <div className="detail-section">
                  <h4>RED 指标配置</h4>
                  <div className="detail-grid">
                    <div className="detail-item">
                      <label>监控状态:</label>
                      <span style={{ color: '#52c41a' }}>已启用</span>
                    </div>
                    
                    {selectedApp.red_metrics_config.rate_metric && (
                      <>
                        <div className="detail-item" style={{ gridColumn: '1 / -1' }}>
                          <label>Rate 指标:</label>
                          <span>{selectedApp.red_metrics_config.rate_metric.metric_name || '未配置'}</span>
                        </div>
                        {selectedApp.red_metrics_config.rate_metric.promql && (
                          <div className="detail-item" style={{ gridColumn: '1 / -1' }}>
                            <label>PromQL:</label>
                            <code style={{ background: '#f5f5f5', padding: '5px', borderRadius: '4px', display: 'block' }}>
                              {selectedApp.red_metrics_config.rate_metric.promql}
                            </code>
                          </div>
                        )}
                      </>
                    )}
                    
                    {selectedApp.red_metrics_config.error_metric && (
                      <>
                        <div className="detail-item" style={{ gridColumn: '1 / -1' }}>
                          <label>Error 指标:</label>
                          <span>{selectedApp.red_metrics_config.error_metric.metric_name || '未配置'}</span>
                        </div>
                        {selectedApp.red_metrics_config.error_metric.promql && (
                          <div className="detail-item" style={{ gridColumn: '1 / -1' }}>
                            <label>PromQL:</label>
                            <code style={{ background: '#f5f5f5', padding: '5px', borderRadius: '4px', display: 'block' }}>
                              {selectedApp.red_metrics_config.error_metric.promql}
                            </code>
                          </div>
                        )}
                      </>
                    )}
                    
                    {selectedApp.red_metrics_config.duration_metric && (
                      <>
                        <div className="detail-item" style={{ gridColumn: '1 / -1' }}>
                          <label>Duration 指标:</label>
                          <span>{selectedApp.red_metrics_config.duration_metric.metric_name || '未配置'}</span>
                        </div>
                        {selectedApp.red_metrics_config.duration_metric.promql && (
                          <div className="detail-item" style={{ gridColumn: '1 / -1' }}>
                            <label>PromQL:</label>
                            <code style={{ background: '#f5f5f5', padding: '5px', borderRadius: '4px', display: 'block' }}>
                              {selectedApp.red_metrics_config.duration_metric.promql}
                            </code>
                          </div>
                        )}
                      </>
                    )}
                    
                    {selectedApp.red_metrics_config.health_threshold && (
                      <>
                        <div className="detail-item" style={{ gridColumn: '1 / -1', marginTop: '10px' }}>
                          <label style={{ fontWeight: 'bold' }}>健康度阈值:</label>
                        </div>
                        <div className="detail-item">
                          <label>最低请求速率:</label>
                          <span>{selectedApp.red_metrics_config.health_threshold.rate_min}</span>
                        </div>
                        <div className="detail-item">
                          <label>最大错误率:</label>
                          <span>{selectedApp.red_metrics_config.health_threshold.error_rate_max}</span>
                        </div>
                        <div className="detail-item">
                          <label>P95 响应时长上限:</label>
                          <span>{selectedApp.red_metrics_config.health_threshold.duration_p95_max}ms</span>
                        </div>
                      </>
                    )}
                  </div>
                </div>
              )}

              {/* 回滚策略配置 */}
              {selectedApp.rollback_policy && selectedApp.rollback_policy.enabled && (
                <div className="detail-section">
                  <h4>回滚策略</h4>
                  <div className="detail-grid">
                    <div className="detail-item">
                      <label>策略状态:</label>
                      <span style={{ color: '#52c41a' }}>已启用</span>
                    </div>
                    <div className="detail-item">
                      <label>自动回滚:</label>
                      <span style={{ color: selectedApp.rollback_policy.auto_rollback ? '#52c41a' : '#999' }}>
                        {selectedApp.rollback_policy.auto_rollback ? '是' : '否'}
                      </span>
                    </div>
                    <div className="detail-item">
                      <label>通知渠道:</label>
                      <span>{selectedApp.rollback_policy.notify_channel || '未配置'}</span>
                    </div>
                    
                    {selectedApp.rollback_policy.alert_rules && selectedApp.rollback_policy.alert_rules.length > 0 && (
                      <div className="detail-item" style={{ gridColumn: '1 / -1', marginTop: '10px' }}>
                        <label style={{ fontWeight: 'bold' }}>告警规则 ({selectedApp.rollback_policy.alert_rules.length} 条):</label>
                        <div style={{ marginTop: '10px' }}>
                          {selectedApp.rollback_policy.alert_rules.map((rule, index) => (
                            <div key={index} style={{ 
                              padding: '10px', 
                              background: '#f5f5f5', 
                              borderRadius: '4px', 
                              marginBottom: '10px' 
                            }}>
                              <div><strong>{rule.name}</strong> - {rule.severity}</div>
                              <div style={{ marginTop: '5px', fontSize: '12px' }}>
                                <code>{rule.alert_expr}</code>
                              </div>
                              <div style={{ marginTop: '5px', fontSize: '12px', color: '#666' }}>
                                持续时长: {rule.duration}
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>
            <div className="modal-footer">
              <button onClick={() => setShowDetailModal(false)}>关闭</button>
            </div>
          </div>
        </div>
      )}
    </PageLayout>
  )
}

export default Apps
