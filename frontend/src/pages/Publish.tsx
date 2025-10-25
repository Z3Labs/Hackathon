import React, { useState } from 'react';
import DeploymentList from '../components/DeploymentList';
import DeploymentForm from '../components/DeploymentForm';
import DeploymentDetail from '../components/DeploymentDetail';
import type { Deployment } from '../types/deployment';

type ViewMode = 'list' | 'create' | 'detail';

const Publish: React.FC = () => {
  const [viewMode, setViewMode] = useState<ViewMode>('list');
  const [selectedDeployment, setSelectedDeployment] = useState<Deployment | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);

  const handleCreateNew = () => {
    setViewMode('create');
  };

  const handleSelectDeployment = (deployment: Deployment) => {
    setSelectedDeployment(deployment);
    setViewMode('detail');
  };

  const handleFormSuccess = () => {
    setViewMode('list');
    setRefreshKey((prev) => prev + 1);
  };

  const handleFormCancel = () => {
    setViewMode('list');
  };

  const handleDetailClose = () => {
    setViewMode('list');
    setSelectedDeployment(null);
  };

  return (
    <div>
      <h1 style={{ padding: '20px', borderBottom: '1px solid #f0f0f0', margin: 0 }}>发布管理</h1>
      {viewMode === 'list' && (
        <DeploymentList
          key={refreshKey}
          onSelectDeployment={handleSelectDeployment}
          onCreateNew={handleCreateNew}
        />
      )}
      {viewMode === 'create' && (
        <DeploymentForm onSuccess={handleFormSuccess} onCancel={handleFormCancel} />
      )}
      {viewMode === 'detail' && selectedDeployment && (
        <DeploymentDetail deploymentId={selectedDeployment.id} onClose={handleDetailClose} />
      )}
    </div>
  );
};

export default Publish;
