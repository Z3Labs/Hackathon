export interface NodeDeployment {
  id: string;
  ip: string;
  node_deploy_status: 'pending' | 'deploying' | 'success' | 'failed' | 'skipped' | 'rolled_back';
  release_log: string;
}

export interface Deployment {
  id: string;
  app_name: string;
  status: 'pending' | 'deploying' | 'success' | 'failed' | 'rolled_back' | 'canceled';
  package_version: string;
  gray_machine_id: string;
  node_deployments: NodeDeployment[];
  created_at: number;
  updated_at: number;
}

export interface CreateDeploymentRequest {
  app_name: string;
  package_version: string;
  gray_machine_id?: string;
}

export interface UpdateDeploymentRequest {
  id: string;
  app_name: string;
  package_version: string;
  gray_machine_id?: string;
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

export interface Report {
  id: string;
  deployment_id: string;
  content: string;
  status: 'generating' | 'completed' | 'failed';
  created_at: number;
  updated_at: number;
}

export interface GetDeploymentDetailResponse {
  deployment: Deployment;
  report?: Report | null;
}

export interface CreateDeploymentResponse {
  id: string;
}

export interface UpdateDeploymentResponse {
  success: boolean;
}
