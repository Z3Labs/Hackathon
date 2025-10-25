import React from 'react';
import { useNavigate } from 'react-router-dom';

export interface BreadcrumbItem {
  label: string;
  path?: string;
}

interface BreadcrumbProps {
  items: BreadcrumbItem[];
}

const Breadcrumb: React.FC<BreadcrumbProps> = ({ items }) => {
  const navigate = useNavigate();

  const handleClick = (item: BreadcrumbItem) => {
    if (item.path) {
      navigate(item.path);
    }
  };

  return (
    <div
      style={{
        padding: '12px 20px',
        borderBottom: '1px solid #f0f0f0',
        background: '#fafafa',
        display: 'flex',
        alignItems: 'center',
        fontSize: '14px',
      }}
    >
      {items.map((item, index) => (
        <React.Fragment key={index}>
          {index > 0 && (
            <span style={{ margin: '0 8px', color: '#8c8c8c' }}>{'>'}</span>
          )}
          {item.path ? (
            <span
              onClick={() => handleClick(item)}
              style={{
                cursor: 'pointer',
                color: '#1890ff',
              }}
            >
              {item.label}
            </span>
          ) : (
            <span style={{ color: '#000' }}>{item.label}</span>
          )}
        </React.Fragment>
      ))}
    </div>
  );
};

export default Breadcrumb;
