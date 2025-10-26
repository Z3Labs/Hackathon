import axios from 'axios'

// 创建axios实例
const api = axios.create({
  baseURL: '/api/v1', // 使用相对路径，通过nginx代理
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    // 可以在这里添加token等认证信息
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器 - 简化版本，只处理基本的数据提取
api.interceptors.response.use(
  (response) => {
    // 如果响应有data字段，返回data，否则返回整个响应
    return response.data?.data || response.data
  },
  (error) => {
    // 提取更友好的错误信息
    let errorMessage = '网络请求失败'
    
    if (error.response) {
      // 服务器返回了错误状态码
      errorMessage = error.response.data?.message || `服务器错误: ${error.response.status}`
    } else if (error.request) {
      // 请求已发出但没有收到响应
      errorMessage = '无法连接到服务器，请检查网络连接'
    } else {
      // 其他错误
      errorMessage = error.message || '请求配置错误'
    }
    
    // 创建一个包含友好错误信息的错误对象
    const friendlyError = new Error(errorMessage)
    friendlyError.name = error.name
    friendlyError.stack = error.stack
    
    return Promise.reject(friendlyError)
  }
)

// 应用相关接口
export const appApi = {
  // 创建应用
  createApp: (data: {
    name: string
    repo?: string
    deploy_path: string
    config_path?: string
    start_cmd: string
    stop_cmd: string
    rollback_policy?: any
    red_metrics_config?: any
  }) => api.post('/apps', data),

  // 更新应用
  updateApp: (id: string, data: {
    id: string
    name: string
    repo?: string
    deploy_path: string
    config_path?: string
    start_cmd: string
    stop_cmd: string
    machine_ids?: string[]
    rollback_policy?: any
    red_metrics_config?: any
  }) => api.put(`/apps/${id}`, data),

  // 获取应用列表
  getAppList: (params?: {
    page?: number
    page_size?: number
    name?: string
  }) => api.get('/apps', { params }),

  // 获取应用详情
  getAppDetail: (id: string) => api.get(`/apps/${id}`),

  // 获取应用版本列表
  getAppVersions: (appName: string) => api.get('/apps/versions', { params: { app_name: appName } }),
}

// 发布记录相关接口
export const deploymentApi = {
  // 创建发布记录
  createDeployment: (data: {
    app_name: string
    package_version: string
    gray_machine_id?: string
  }) => api.post('/deployments', data),

  // 更新发布记录
  updateDeployment: (id: string, data: {
    app_name: string
    package_version: string
    gray_machine_id?: string
  }) => api.put(`/deployments/${id}`, data),

  // 获取发布记录列表
  getDeploymentList: (params?: {
    page?: number
    page_size?: number
    app_name?: string
    status?: string
  }) => api.get('/deployments', { params }),

  // 获取发布记录详情
  getDeploymentDetail: (id: string) => api.get(`/deployments/${id}`),
}

// 机器相关接口
export const machineApi = {
  // 创建机器
  createMachine: (data: {
    name: string
    ip: string
    port: number
    username: string
    password: string
    description: string
  }) => api.post('/machines', data),

  // 更新机器
  updateMachine: (id: string, data: {
    name: string
    ip: string
    port: number
    username: string
    password: string
    description: string
  }) => api.put(`/machines/${id}`, data),

  // 获取机器列表
  getMachineList: (params?: {
    page?: number
    page_size?: number
    name?: string
    ip?: string
    health_status?: string
    error_status?: string
    alert_status?: string
  }) => api.get('/machines', { params }),

  // 获取机器详情
  getMachineDetail: (id: string) => api.get(`/machines/${id}`),

  // 删除机器
  deleteMachine: (id: string) => api.delete(`/machines/${id}`),

  // 测试机器连接
  testMachineConnection: (id: string) => api.post(`/machines/${id}/test`),

  // 获取机器hostname
  getMachineHostname: (data: {
    ip: string
    port: number
    username: string
    password: string
  }) => api.post('/machines/hostname', data),
}

export default api