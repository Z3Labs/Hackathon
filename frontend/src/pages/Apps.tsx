import React, { useState, useEffect } from 'react'
import { appApi, machineApi } from '../services/api'
import { Application, Machine, CreateAppReq, GetAppListResp, GetAppDetailResp, GetMachineListResp } from '../types'
import { useApiRequest } from '../hooks/useApiRequest'
import { Toaster } from 'react-hot-toast'
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
    deploy_path: '',
    start_cmd: '',
    stop_cmd: ''
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
        id: selectedApp.id,
        ...formData
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
      deploy_path: '',
      start_cmd: '',
      stop_cmd: ''
    })
    setSelectedApp(null)
  }

  // 打开编辑模态框
  const openEditModal = (app: Application) => {
    setSelectedApp(app)
    setFormData({
      name: app.name,
      deploy_path: app.deploy_path,
      start_cmd: app.start_cmd,
      stop_cmd: app.stop_cmd
    })
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
        deploy_path: selectedApp.deploy_path,
        start_cmd: selectedApp.start_cmd,
        stop_cmd: selectedApp.stop_cmd,
        machine_ids: selectedMachineIds
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
    <div className="apps-container">
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
        <h1>应用管理</h1>
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
                <label>部署路径</label>
                <input
                  type="text"
                  value={formData.deploy_path}
                  onChange={(e) => setFormData(prev => ({ ...prev, deploy_path: e.target.value }))}
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
          <div className="modal">
            <div className="modal-header">
              <h3>编辑应用</h3>
              <button onClick={() => setShowEditModal(false)}>×</button>
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
                <label>部署路径</label>
                <input
                  type="text"
                  value={formData.deploy_path}
                  onChange={(e) => setFormData(prev => ({ ...prev, deploy_path: e.target.value }))}
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
                    <label>版本:</label>
                    <span>{selectedApp.currentVersion}</span>
                  </div>
                  <div className="detail-item">
                    <label>部署路径:</label>
                    <span>{selectedApp.deploy_path}</span>
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
            </div>
            <div className="modal-footer">
              <button onClick={() => setShowDetailModal(false)}>关闭</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default Apps
