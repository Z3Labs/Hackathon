import api from './api';
import type {
  CreateDeploymentRequest,
  CreateDeploymentResponse,
  UpdateDeploymentRequest,
  UpdateDeploymentResponse,
  GetDeploymentListRequest,
  GetDeploymentListResponse,
  GetDeploymentDetailResponse,
} from '../types/deployment';

export const deploymentService = {
  async createDeployment(data: CreateDeploymentRequest): Promise<CreateDeploymentResponse> {
    return api.post('/deployments', data);
  },

  async updateDeployment(data: UpdateDeploymentRequest): Promise<UpdateDeploymentResponse> {
    const { id, ...updateData } = data;
    return api.put(`/deployments/${id}`, updateData);
  },

  async getDeploymentList(params?: GetDeploymentListRequest): Promise<GetDeploymentListResponse> {
    return api.get('/deployments', { params });
  },

  async getDeploymentDetail(id: string): Promise<GetDeploymentDetailResponse> {
    return api.get(`/deployments/${id}`);
  },

  async cancelDeployment(id: string): Promise<{ success: boolean }> {
    return api.post(`/deployments/${id}/cancel`);
  },

  async rollbackDeployment(id: string): Promise<{ success: boolean }> {
    return api.post(`/deployments/${id}/rollback`);
  },

  async rollbackNodeDeployment(id: string, nodeDeploymentIds: string[]): Promise<{ success: boolean }> {
    return api.post(`/deployments/${id}/node-deployments/rollback`, { node_deployment_ids: nodeDeploymentIds });
  },

  async deployNodeDeployment(id: string, nodeDeploymentIds: string[]): Promise<{ success: boolean }> {
    return api.post(`/deployments/${id}/node-deployments/deploy`, { node_deployment_ids: nodeDeploymentIds });
  },

  async retryNodeDeployment(id: string, nodeDeploymentIds: string[]): Promise<{ success: boolean }> {
    return api.post(`/deployments/${id}/node-deployments/retry`, { node_deployment_ids: nodeDeploymentIds });
  },

  async skipNodeDeployment(id: string, nodeDeploymentIds: string[]): Promise<{ success: boolean }> {
    return api.post(`/deployments/${id}/node-deployments/skip`, { node_deployment_ids: nodeDeploymentIds });
  },

  async cancelNodeDeployment(id: string, nodeDeploymentIds: string[]): Promise<{ success: boolean }> {
    return api.post(`/deployments/${id}/node-deployments/cancel`, { node_deployment_ids: nodeDeploymentIds });
  },
};
