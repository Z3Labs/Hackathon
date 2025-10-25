import React from 'react';
import Breadcrumb, { BreadcrumbItem } from './Breadcrumb';

interface PageLayoutProps {
  breadcrumbItems: BreadcrumbItem[];
  children: React.ReactNode;
}

const PageLayout: React.FC<PageLayoutProps> = ({ breadcrumbItems, children }) => {
  return (
    <div>
      <Breadcrumb items={breadcrumbItems} />
      <div style={{ padding: '20px' }}>
        {children}
      </div>
    </div>
  );
};

export default PageLayout;
