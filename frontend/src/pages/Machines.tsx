import React, { useState, useEffect } from 'react'
import { machineApi } from '../services/api'
import { Machine, CreateMachineReq, GetMachineListResp, GetMachineDetailResp } from '../types'
import { useApiRequest } from '../hooks/useApiRequest'
import { Toaster } from 'react-hot-toast'
import Breadcrumb from '../components/Breadcrumb'
import './Apps.css'

const Machines: React.FC = () => {
  const [machines, setMachines] = useState<Machine[]>([])
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showEditModal, setShowEditModal] = useState(false)
  const [showDetailModal, setShowDetailModal] = useState(false)
  const [selectedMachine, setSelectedMachine] = useState<Machine | null>(null)
  const [searchName, setSearchName] = useState('')
  const [searchIp, setSearchIp] = useState('')
  const [pagination, setPagination] = useState({
    page: 1,
    pageSize: 10,
    total: 0
  })
  
  const { request, loading } = useApiRequest()

  const [formData, setFormData] = useState<CreateMachineReq>({
    name: '',
    ip: '',
    port: 22,
    username: '',
    password: '',
    description: ''
  })

  const [testStatus, setTestStatus] = useState<'idle' | 'testing' | 'success' | 'failed'>('idle')
  const [testMessage, setTestMessage] = useState('')

  const fetchMachines = async () => {
    const result = await request(
      () => machineApi.getMachineList({
        page: pagination.page,
        page_size: pagination.pageSize,
        name: searchName || undefined,
        ip: searchIp || undefined
      }) as unknown as Promise<GetMachineListResp>,
      {
        errorMessage: '获取机器列表失败'
      }
    )
    
    if (result) {
      setMachines(result.machines || [])
      setPagination(prev => ({
        ...prev,
        total: result.total || 0
      }))
    }
  }

  useEffect(() => {
    fetchMachines()
  }, [pagination.page, pagination.pageSize, searchName, searchIp])

  const handleTestConnectionInModal = async () => {
    if (!formData.name || !formData.ip || !formData.username || !formData.password) {
      alert('请填写完整的机器信息')
      return
    }

    setTestStatus('testing')
    setTestMessage('')

    try {
      const tempMachine = await machineApi.createMachine({ ...formData })
      const testResult = await machineApi.testMachineConnection(tempMachine.id) as unknown as { success: boolean; message: string }
      
      if (testResult.success) {
        setTestStatus('success')
        setTestMessage('连接测试成功')
      } else {
        setTestStatus('failed')
        setTestMessage(testResult.message || '连接测试失败')
        await machineApi.deleteMachine(tempMachine.id)
      }
    } catch (err) {
      setTestStatus('failed')
      setTestMessage('连接测试失败')
    }
  }

  const handleCreateMachine = async () => {
    if (testStatus !== 'success') {
      alert('请先测试连接并且测试通过后才能创建机器')
      return
    }

    setShowCreateModal(false)
    resetForm()
    setTestStatus('idle')
    setTestMessage('')
    fetchMachines()
  }

  const handleUpdateMachine = async () => {
    if (!selectedMachine) return
    
    const result = await request(
      () => machineApi.updateMachine(selectedMachine.id, formData as any),
      {
        successMessage: '机器更新成功',
        errorMessage: '更新机器失败',
        showSuccessToast: true
      }
    )
    
    if (result) {
      setShowEditModal(false)
      resetForm()
      fetchMachines()
    }
  }

  const handleDeleteMachine = async (id: string) => {
    if (!confirm('确定要删除这台机器吗？')) return
    
    const result = await request(
      () => machineApi.deleteMachine(id),
      {
        successMessage: '机器删除成功',
        errorMessage: '删除机器失败',
        showSuccessToast: true
      }
    )
    
    if (result) {
      fetchMachines()
    }
  }

  const handleTestConnectionInDetail = async () => {
    if (!selectedMachine) return
    
    const result = await request(
      () => machineApi.testMachineConnection(selectedMachine.id) as unknown as Promise<{ success: boolean; message: string }>,
      {
        successMessage: '连接测试成功',
        errorMessage: '连接测试失败',
        showSuccessToast: true
      }
    )
    
    if (result) {
      alert(result.message)
    }
  }

  const resetForm = () => {
    setFormData({
      name: '',
      ip: '',
      port: 22,
      username: '',
      password: '',
      description: ''
    })
    setSelectedMachine(null)
    setTestStatus('idle')
    setTestMessage('')
  }

  const openEditModal = (machine: Machine) => {
    setSelectedMachine(machine)
    setFormData({
      name: machine.name,
      ip: machine.ip,
      port: machine.port,
      username: machine.username,
      password: machine.password,
      description: machine.description
    })
    setShowEditModal(true)
  }

  const openDetailModal = async (machineId: string) => {
    const result = await request(
      () => machineApi.getMachineDetail(machineId) as unknown as Promise<GetMachineDetailResp>,
      {
        errorMessage: '获取机器详情失败'
      }
    )
    
    if (result) {
      setSelectedMachine(result.machine)
      setShowDetailModal(true)
    }
  }

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
      <Breadcrumb items={[{ label: '机器管理' }]} />
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
        <h1>机器管理</h1>
        <div className="apps-actions">
          <div className="search-box">
            <input
              type="text"
              placeholder="搜索机器名称"
              value={searchName}
              onChange={(e) => setSearchName(e.target.value)}
            />
            <input
              type="text"
              placeholder="搜索IP地址"
              value={searchIp}
              onChange={(e) => setSearchIp(e.target.value)}
              style={{ marginLeft: '10px' }}
            />
          </div>
          <button 
            className="btn btn-primary"
            onClick={() => setShowCreateModal(true)}
          >
            添加机器
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
                  <th>机器名称</th>
                  <th>IP地址</th>
                  <th>端口</th>
                  <th>用户名</th>
                  <th>描述</th>
                  <th>健康状态</th>
                  <th>异常状态</th>
                  <th>告警状态</th>
                  <th>创建时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {machines.map((machine) => (
                  <tr key={machine.id}>
                    <td>{machine.name}</td>
                    <td>{machine.ip}</td>
                    <td>{machine.port}</td>
                    <td>{machine.username}</td>
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
                    <td>{new Date(machine.created_at * 1000).toLocaleString()}</td>
                    <td>
                      <div className="action-buttons">
                        <button 
                          className="btn btn-sm btn-info"
                          onClick={() => openDetailModal(machine.id)}
                        >
                          详情
                        </button>
                        <button 
                          className="btn btn-sm btn-warning"
                          onClick={() => openEditModal(machine)}
                        >
                          编辑
                        </button>
                        <button 
                          className="btn btn-sm btn-danger"
                          onClick={() => handleDeleteMachine(machine.id)}
                        >
                          删除
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

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

      {showCreateModal && (
        <div className="modal-overlay">
          <div className="modal">
            <div className="modal-header">
              <h3>添加机器</h3>
              <button onClick={() => setShowCreateModal(false)}>×</button>
            </div>
            <div className="modal-body">
              <div className="form-group">
                <label>机器名称 <span style={{ color: 'red' }}>*</span></label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
                />
              </div>
              <div className="form-group">
                <label>IP地址 <span style={{ color: 'red' }}>*</span></label>
                <input
                  type="text"
                  value={formData.ip}
                  onChange={(e) => setFormData(prev => ({ ...prev, ip: e.target.value }))}
                />
              </div>
              <div className="form-group">
                <label>SSH端口 <span style={{ color: 'red' }}>*</span></label>
                <input
                  type="number"
                  value={formData.port}
                  onChange={(e) => setFormData(prev => ({ ...prev, port: parseInt(e.target.value) }))}
                />
              </div>
              <div className="form-group">
                <label>SSH用户名 <span style={{ color: 'red' }}>*</span></label>
                <input
                  type="text"
                  value={formData.username}
                  onChange={(e) => setFormData(prev => ({ ...prev, username: e.target.value }))}
                />
              </div>
              <div className="form-group">
                <label>SSH密码 <span style={{ color: 'red' }}>*</span></label>
                <input
                  type="password"
                  value={formData.password}
                  onChange={(e) => setFormData(prev => ({ ...prev, password: e.target.value }))}
                />
              </div>
              <div className="form-group">
                <label>机器描述</label>
                <input
                  type="text"
                  value={formData.description}
                  onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                />
              </div>
              <div className="form-group">
                <button 
                  className="btn btn-success" 
                  onClick={handleTestConnectionInModal}
                  disabled={testStatus === 'testing'}
                  style={{ width: '100%' }}
                >
                  {testStatus === 'testing' ? '测试中...' : '测试连接'}
                </button>
                {testStatus === 'success' && (
                  <div style={{ color: '#52c41a', marginTop: '8px' }}>{testMessage}</div>
                )}
                {testStatus === 'failed' && (
                  <div style={{ color: '#ff4d4f', marginTop: '8px' }}>{testMessage}</div>
                )}
              </div>
            </div>
            <div className="modal-footer">
              <button onClick={() => { setShowCreateModal(false); resetForm(); }}>取消</button>
              <button className="btn-primary" onClick={handleCreateMachine}>创建</button>
            </div>
          </div>
        </div>
      )}

      {showEditModal && (
        <div className="modal-overlay">
          <div className="modal">
            <div className="modal-header">
              <h3>编辑机器</h3>
              <button onClick={() => setShowEditModal(false)}>×</button>
            </div>
            <div className="modal-body">
              <div className="form-group">
                <label>机器名称 <span style={{ color: 'red' }}>*</span></label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
                />
              </div>
              <div className="form-group">
                <label>IP地址 <span style={{ color: 'red' }}>*</span></label>
                <input
                  type="text"
                  value={formData.ip}
                  onChange={(e) => setFormData(prev => ({ ...prev, ip: e.target.value }))}
                />
              </div>
              <div className="form-group">
                <label>SSH端口 <span style={{ color: 'red' }}>*</span></label>
                <input
                  type="number"
                  value={formData.port}
                  onChange={(e) => setFormData(prev => ({ ...prev, port: parseInt(e.target.value) }))}
                />
              </div>
              <div className="form-group">
                <label>SSH用户名 <span style={{ color: 'red' }}>*</span></label>
                <input
                  type="text"
                  value={formData.username}
                  onChange={(e) => setFormData(prev => ({ ...prev, username: e.target.value }))}
                />
              </div>
              <div className="form-group">
                <label>SSH密码 <span style={{ color: 'red' }}>*</span></label>
                <input
                  type="password"
                  value={formData.password}
                  onChange={(e) => setFormData(prev => ({ ...prev, password: e.target.value }))}
                />
              </div>
              <div className="form-group">
                <label>机器描述</label>
                <input
                  type="text"
                  value={formData.description}
                  onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                />
              </div>
            </div>
            <div className="modal-footer">
              <button onClick={() => setShowEditModal(false)}>取消</button>
              <button className="btn-primary" onClick={handleUpdateMachine}>保存</button>
            </div>
          </div>
        </div>
      )}

      {showDetailModal && selectedMachine && (
        <div className="modal-overlay">
          <div className="modal modal-large">
            <div className="modal-header">
              <h3>机器详情 - {selectedMachine.name}</h3>
              <button onClick={() => setShowDetailModal(false)}>×</button>
            </div>
            <div className="modal-body">
              <div className="detail-section">
                <h4>基本信息</h4>
                <div className="detail-grid">
                  <div className="detail-item">
                    <label>机器ID:</label>
                    <span>{selectedMachine.id}</span>
                  </div>
                  <div className="detail-item">
                    <label>机器名称:</label>
                    <span>{selectedMachine.name}</span>
                  </div>
                  <div className="detail-item">
                    <label>IP地址:</label>
                    <span>{selectedMachine.ip}</span>
                  </div>
                  <div className="detail-item">
                    <label>SSH端口:</label>
                    <span>{selectedMachine.port}</span>
                  </div>
                  <div className="detail-item">
                    <label>SSH用户名:</label>
                    <span>{selectedMachine.username}</span>
                  </div>
                  <div className="detail-item">
                    <label>描述:</label>
                    <span>{selectedMachine.description}</span>
                  </div>
                  <div className="detail-item">
                    <label>创建时间:</label>
                    <span>{new Date(selectedMachine.created_at * 1000).toLocaleString()}</span>
                  </div>
                  <div className="detail-item">
                    <label>更新时间:</label>
                    <span>{new Date(selectedMachine.updated_at * 1000).toLocaleString()}</span>
                  </div>
                </div>
              </div>

              <div className="detail-section">
                <h4>状态信息</h4>
                <div className="status-stats">
                  <div className="stat-item">
                    <span className="stat-label">健康状态:</span>
                    <span className="stat-value" style={{ color: getStatusColor(selectedMachine.health_status) }}>
                      {getStatusText(selectedMachine.health_status)}
                    </span>
                  </div>
                  <div className="stat-item">
                    <span className="stat-label">异常状态:</span>
                    <span className="stat-value" style={{ color: getStatusColor(selectedMachine.error_status) }}>
                      {getStatusText(selectedMachine.error_status)}
                    </span>
                  </div>
                  <div className="stat-item">
                    <span className="stat-label">告警状态:</span>
                    <span className="stat-value" style={{ color: getStatusColor(selectedMachine.alert_status) }}>
                      {getStatusText(selectedMachine.alert_status)}
                    </span>
                  </div>
                </div>
              </div>
            </div>
            <div className="modal-footer">
              <button className="btn btn-success" onClick={handleTestConnectionInDetail}>测试连接</button>
              <button onClick={() => setShowDetailModal(false)}>关闭</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default Machines
