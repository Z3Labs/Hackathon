export interface DeploymentMachine {
  id: string;
  ip: string;
  port: number;
  release_status: 'pending' | 'deploying' | 'success' | 'failed';
  health_status: 'healthy' | 'unhealthy';
  error_status: 'normal' | 'error';
  alert_status: 'normal' | 'alert';
}

export interface Deployment {
  id: string;
  app_name: string;
  status: 'pending' | 'deploying' | 'success' | 'failed' | 'rolled_back';
  package_version: string;
  config_path: string;
  gray_strategy: 'canary' | 'blue-green' | 'all';
  release_machines: DeploymentMachine[];
  release_log: string;
  created_at: number;
  updated_at: number;
}

export interface CreateDeploymentRequest {
  app_name: string;
  package_version: string;
  config_path: string;
  gray_strategy: 'canary' | 'blue-green' | 'all';
}

export interface UpdateDeploymentRequest {
  id: string;
  app_name: string;
  package_version: string;
  config_path: string;
  gray_strategy: 'canary' | 'blue-green' | 'all';
}

export interface GetDeploymentListRequest {
  page?: number;
  page_size?: number;
  app_name?: string;
  status?: string;
}

export interface GetDeploymentListResponse {
  deployments: Deployment[];
  total: number;
  page: number;
  page_size: number;
}

export interface GetDeploymentDetailResponse {
  deployment: Deployment;
}

export interface CreateDeploymentResponse {
  id: string;
}

export interface UpdateDeploymentResponse {
  success: boolean;
}
