import React, { useState } from 'react';

export interface TimeRange {
  label: string;
  minutes: number;
}

export const PRESET_RANGES: TimeRange[] = [
  { label: '最近5分钟', minutes: 5 },
  { label: '最近15分钟', minutes: 15 },
  { label: '最近30分钟', minutes: 30 },
  { label: '最近1小时', minutes: 60 },
  { label: '最近3小时', minutes: 180 },
  { label: '最近6小时', minutes: 360 },
  { label: '最近12小时', minutes: 720 },
  { label: '最近24小时', minutes: 1440 },
];

interface TimeRangeSelectorProps {
  value: number; // 当前选中的分钟数
  onChange: (minutes: number) => void;
  disabled?: boolean;
}

const TimeRangeSelector: React.FC<TimeRangeSelectorProps> = ({ value, onChange, disabled }) => {
  const [isOpen, setIsOpen] = useState(false);

  const currentLabel = PRESET_RANGES.find(r => r.minutes === value)?.label || `${value}分钟`;

  return (
    <div style={{ position: 'relative', display: 'inline-block' }}>
      <button
        onClick={() => !disabled && setIsOpen(!isOpen)}
        disabled={disabled}
        style={{
          padding: '4px 12px',
          border: 'none',
          background: 'transparent',
          cursor: disabled ? 'not-allowed' : 'pointer',
          fontSize: '13px',
          display: 'flex',
          alignItems: 'center',
          gap: '6px',
          opacity: disabled ? 0.6 : 1,
          color: '#1890ff',
        }}
      >
        <span>⏱</span>
        {currentLabel}
        <span style={{ fontSize: '10px', color: 'inherit' }}>▼</span>
      </button>

      {isOpen && !disabled && (
        <>
          {/* 点击外部关闭 */}
          <div
            style={{
              position: 'fixed',
              top: 0,
              left: 0,
              right: 0,
              bottom: 0,
              zIndex: 999,
            }}
            onClick={() => setIsOpen(false)}
          />
          {/* 下拉菜单 */}
          <div
            style={{
              position: 'absolute',
              top: '100%',
              right: 0,
              marginTop: '4px',
              background: 'white',
              border: '1px solid #d9d9d9',
              borderRadius: '4px',
              boxShadow: '0 2px 8px rgba(0,0,0,0.15)',
              zIndex: 1000,
              minWidth: '150px',
              maxHeight: '300px',
              overflow: 'auto',
            }}
            onClick={(e) => e.stopPropagation()}
          >
            {PRESET_RANGES.map((range) => (
              <div
                key={range.minutes}
                onClick={() => {
                  onChange(range.minutes);
                  setIsOpen(false);
                }}
                style={{
                  padding: '8px 12px',
                  cursor: 'pointer',
                  background: value === range.minutes ? '#f0f7ff' : 'white',
                  color: value === range.minutes ? '#1890ff' : '#333',
                  fontSize: '13px',
                }}
                onMouseEnter={(e) => {
                  if (value !== range.minutes) {
                    e.currentTarget.style.background = '#fafafa';
                  }
                }}
                onMouseLeave={(e) => {
                  if (value !== range.minutes) {
                    e.currentTarget.style.background = 'white';
                  }
                }}
              >
                {range.label}
              </div>
            ))}
          </div>
        </>
      )}
    </div>
  );
};

export default TimeRangeSelector;

