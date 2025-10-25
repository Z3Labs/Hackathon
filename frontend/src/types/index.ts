// 机器信息
export interface Machine {
  id: string
  name: string
  ip: string
  port: number
  username: string
  password: string
  description: string
  health_status: string // healthy-健康, unhealthy-不健康
  error_status: string  // normal-正常, error-异常
  alert_status: string  // normal-正常, alert-告警
  created_at: number
  updated_at: number
}

// 应用信息
export interface Application {
  id: string
  name: string
  deploy_path: string
  start_cmd: string
  stop_cmd: string
  currentVersion: string
  machine_count: number
  health_count: number
  error_count: number
  alert_count: number
  machines: Machine[]
  created_at: number
  updated_at: number
}

// 发布机器信息
export interface DeploymentMachine {
  id: string
  ip: string
  port: number
  release_status: string // pending-待发布, deploying-发布中, success-成功, failed-失败
  health_status: string   // healthy-健康, unhealthy-不健康
  error_status: string   // normal-正常, error-异常
  alert_status: string   // normal-正常, alert-告警
}

// 发布记录信息
export interface Deployment {
  id: string
  app_name: string
  status: string // pending-待发布, deploying-发布中, success-成功, failed-失败, rolled_back-已回滚
  package_version: string
  config_path: string
  gray_strategy: string // canary-金丝雀发布, blue-green-蓝绿发布, all-全量发布
  release_machines: DeploymentMachine[]
  release_log: string
  created_at: number
  updated_at: number
}

// API请求响应类型
export interface CreateAppReq {
  name: string
  deploy_path: string
  start_cmd: string
  stop_cmd: string
}

export interface CreateAppResp {
  id: string
}

export interface UpdateAppReq {
  id: string
  name: string
  deploy_path: string
  start_cmd: string
  stop_cmd: string
  machine_ids?: string[]
}

export interface UpdateAppResp {
  success: boolean
}

export interface GetAppListReq {
  page?: number
  page_size?: number
  name?: string
}

export interface GetAppListResp {
  apps: Application[]
  total: number
  page: number
  page_size: number
}

export interface GetAppDetailReq {
  id: string
}

export interface GetAppDetailResp {
  application: Application
}

// 机器相关请求响应类型
export interface CreateMachineReq {
  name: string
  ip: string
  port: number
  username: string
  password: string
  description: string
}

export interface CreateMachineResp {
  id: string
}

export interface UpdateMachineReq {
  id: string
  name: string
  ip: string
  port: number
  username: string
  password: string
  description: string
}

export interface UpdateMachineResp {
  success: boolean
}

export interface GetMachineListReq {
  page?: number
  page_size?: number
  name?: string
  ip?: string
  health_status?: string
  error_status?: string
  alert_status?: string
}

export interface GetMachineListResp {
  machines: Machine[]
  total: number
  page: number
  page_size: number
}

export interface GetMachineDetailReq {
  id: string
}

export interface GetMachineDetailResp {
  machine: Machine
}

export interface DeleteMachineReq {
  id: string
}

export interface DeleteMachineResp {
  success: boolean
}

export interface TestMachineConnectionReq {
  id: string
}

export interface TestMachineConnectionResp {
  success: boolean
  message: string
}
