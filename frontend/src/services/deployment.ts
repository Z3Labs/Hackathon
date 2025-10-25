import axios from 'axios';
import type {
  CreateDeploymentRequest,
  CreateDeploymentResponse,
  UpdateDeploymentRequest,
  UpdateDeploymentResponse,
  GetDeploymentListRequest,
  GetDeploymentListResponse,
  GetDeploymentDetailResponse,
} from '../types/deployment';

const API_BASE_URL = '/api/v1';

export const deploymentService = {
  async createDeployment(data: CreateDeploymentRequest): Promise<CreateDeploymentResponse> {
    const response = await axios.post<CreateDeploymentResponse>(
      `${API_BASE_URL}/deployments`,
      data
    );
    return response.data;
  },

  async updateDeployment(data: UpdateDeploymentRequest): Promise<UpdateDeploymentResponse> {
    const { id, ...updateData } = data;
    const response = await axios.put<UpdateDeploymentResponse>(
      `${API_BASE_URL}/deployments/${id}`,
      updateData
    );
    return response.data;
  },

  async getDeploymentList(params?: GetDeploymentListRequest): Promise<GetDeploymentListResponse> {
    const response = await axios.get<GetDeploymentListResponse>(
      `${API_BASE_URL}/deployments`,
      { params }
    );
    return response.data;
  },

  async getDeploymentDetail(id: string): Promise<GetDeploymentDetailResponse> {
    const response = await axios.get<GetDeploymentDetailResponse>(
      `${API_BASE_URL}/deployments/${id}`
    );
    return response.data;
  },
};
