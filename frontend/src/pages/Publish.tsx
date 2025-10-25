import React, { useState } from 'react';
import DeploymentList from '../components/DeploymentList';
import DeploymentForm from '../components/DeploymentForm';
import DeploymentDetail from '../components/DeploymentDetail';
import PageLayout from '../components/PageLayout';
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

  const getBreadcrumbItems = () => {
    const items = [{ 
      label: '发布', 
      path: viewMode === 'list' ? '/publish' : undefined,
      onClick: viewMode !== 'list' ? () => {
        setViewMode('list');
        setSelectedDeployment(null);
      } : undefined
    }];
    
    if (viewMode === 'create') {
      items.push({ label: '新建发布' });
    } else if (viewMode === 'detail') {
      items.push({ label: '发布详情' });
    }
    
    return items;
  };

  return (
    <PageLayout breadcrumbItems={getBreadcrumbItems()}>
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
    </PageLayout>
  );
};

export default Publish;
