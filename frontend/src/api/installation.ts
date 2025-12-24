import { apiClient } from './client';
import type { InstallationCheckResponse, InstallationRequest, Installation } from '../types';

export const installationService = {
  check: () => 
    apiClient.get<InstallationCheckResponse>('/api/v1/installation/check'),
  
  install: (data: InstallationRequest) => 
    apiClient.post<Installation>('/api/v1/installation/install', data),
};
