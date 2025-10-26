import React, { useEffect, useRef, useState } from 'react';
import * as echarts from 'echarts';
import TimeRangeSelector from './TimeRangeSelector';

interface MonitorSeries {
  instance: string;
  metric: string;
  unit: string;
  labels?: Record<string, string>; // 原始标签
  data: Array<{
    timestamp: number;
    value: number;
  }>;
}

interface MonitorChartProps {
  series: MonitorSeries[];
  height?: number;
  showTimeSelector?: boolean;
  initialTimeRange?: number;
  onTimeRangeChange?: (minutes: number) => void;
  threshold?: number; // 阈值线
}

const MonitorChart: React.FC<MonitorChartProps> = ({ 
  series, 
  height = 300,
  showTimeSelector = true,
  initialTimeRange = 30,
  onTimeRangeChange,
  threshold
}) => {
  const [timeRange, setTimeRange] = useState(initialTimeRange);
  const chartRef = useRef<HTMLDivElement>(null);
  const chartInstance = useRef<echarts.ECharts | null>(null);

  useEffect(() => {
    if (!chartRef.current) return;

    console.log('[图表] 接收到的 series:', series.map(s => ({ 
      instance: s.instance, 
      unit: s.unit, 
      dataCount: s.data.length,
      sampleValues: s.data.slice(0, 3).map(p => p.value)
    })));

    // 初始化图表
    if (!chartInstance.current) {
      chartInstance.current = echarts.init(chartRef.current);
    }

    const chart = chartInstance.current;

    // 图表颜色配置
    const colors = ['#1890ff', '#52c41a', '#faad14', '#f5222d', '#722ed1', '#13c2c2', '#eb2f96'];
    
    // 准备数据
    const chartSeries: any[] = series.map((s, index) => {
      const maxValue = Math.max(...s.data.map(p => p.value));
      const minValue = Math.min(...s.data.map(p => p.value));
      console.log(`[图表] Series ${s.instance}: unit=${s.unit}, min=${minValue}, max=${maxValue}`);
      
      // 如果有 device 标签（网络指标），则显示网卡名称
      let displayName = s.instance;
      if (s.labels && s.labels.device) {
        displayName = s.labels.device;
      }
      
      const seriesConfig: any = {
        name: displayName,
        type: 'line',
        smooth: true,
        data: s.data.map((point) => [point.timestamp * 1000, point.value]),
        symbol: 'circle',
        symbolSize: 6,
        itemStyle: {
          color: colors[index % colors.length],
        },
        lineStyle: {
          color: colors[index % colors.length],
          width: 2.5,
        },
        emphasis: {
          focus: 'series',
          lineStyle: {
            width: 3,
          },
        },
      };
      
      // 如果有阈值，添加阈值线
      if (threshold !== undefined && threshold !== null) {
        seriesConfig.markLine = {
          data: [{ yAxis: threshold }],
          lineStyle: {
            color: '#ff4d4f',
            type: 'dashed',
            width: 2,
          },
          label: {
            show: true,
            position: 'end',
            formatter: `阈值: ${threshold}`,
            fontSize: 11,
            color: '#ff4d4f',
          },
        };
      }
      
      return seriesConfig;
    });

    // 配置选项
    const option: echarts.EChartsOption = {
      tooltip: {
        trigger: 'axis',
        backgroundColor: 'rgba(0, 0, 0, 0.85)',
        borderColor: 'transparent',
        textStyle: {
          color: '#fff',
          fontSize: 13,
        },
        axisPointer: {
          type: 'cross',
          label: {
            backgroundColor: '#6a7985',
            fontSize: 12,
          },
          lineStyle: {
            color: '#666',
            type: 'dashed',
          },
        },
        formatter: (params: any) => {
          if (!Array.isArray(params) || params.length === 0) {
            return '';
          }
          
          const param = params[0];
          const date = new Date(param.value[0]);
          const timeStr = date.toLocaleString('zh-CN', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
          });
          
          let result = `<div style="margin-bottom: 4px;"><strong>${timeStr}</strong></div>`;
          const unit = series[0]?.unit || '';
          
          params.forEach((p: any) => {
            const value = p.value[1];
            let displayValue = value.toFixed(2);
            result += `<div style="margin: 2px 0;">${p.marker} ${p.seriesName}: <strong>${displayValue}${unit}</strong></div>`;
          });
          
          // 如果有阈值，显示阈值
          if (threshold !== undefined && threshold !== null) {
            result += `<div style="margin-top: 8px; padding-top: 8px; border-top: 1px solid rgba(255,255,255,0.2);">`;
            result += `<div style="margin: 2px 0; color: #ffccc7;">阈值: <strong>${threshold.toFixed(2)}${unit}</strong></div>`;
            result += `</div>`;
          }
          
          return result;
        },
      },
      legend: {
        data: series.map((s) => {
          // 如果有 device 标签（网络指标），则显示网卡名称
          if (s.labels && s.labels.device) {
            return s.labels.device;
          }
          return s.instance;
        }).filter(instance => instance !== 'unknown'),
        bottom: 20,
        left: 'center',
        type: 'scroll',
        itemGap: 20,
        textStyle: {
          fontSize: 12,
          color: '#666',
        },
        icon: 'circle',
      },
      grid: {
        left: '1%',
        right: '10%',
        top: '20px',
        bottom: '20%',
        containLabel: true,
      },
      xAxis: {
        type: 'time',
        boundaryGap: [0, 0],
        axisLine: {
          show: true,
          lineStyle: {
            color: '#e0e0e0',
            width: 1,
          },
        },
        axisLabel: {
          fontSize: 11,
          color: '#999',
          formatter: (value: any) => {
            const date = new Date(value);
            const hours = date.getHours();
            const minutes = date.getMinutes();
            return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}`;
          },
          margin: 10,
          rotate: 0,
        },
        axisTick: {
          show: true,
          lineStyle: {
            color: '#e0e0e0',
            width: 1,
          },
          length: 4,
        },
        splitLine: {
          show: false,
        },
        minInterval: 300000, // 5分钟 - 自动控制标签间隔
      },
      yAxis: {
        type: 'value',
        name: series.length > 0 ? (series[0]?.unit || '') : '',
        nameLocation: 'end',
        nameGap: 20,
        nameTextStyle: {
          color: '#666',
          fontSize: 13,
          fontWeight: 500,
        },
        // 如果有阈值，确保 y 轴最大值包含阈值
        max: threshold !== undefined && threshold !== null 
          ? (() => {
              // 计算数据最大值
              let dataMax = 0;
              for (const s of series) {
                for (const point of s.data) {
                  if (point.value > dataMax) {
                    dataMax = point.value;
                  }
                }
              }
              // 取数据最大值和阈值的最大值，然后加 20% 的余量
              const maxValue = Math.max(dataMax, threshold);
              return maxValue * 1.2;
            })()
          : undefined,
        axisLine: {
          lineStyle: {
            color: '#e0e0e0',
          },
        },
        axisLabel: {
          fontSize: 12,
          color: '#666',
          margin: 12,
          formatter: (value: number) => {
            // 格式化数值，避免显示过长
            if (Math.abs(value) >= 1000000) {
              return (value / 1000000).toFixed(1) + 'M';
            } else if (Math.abs(value) >= 1000) {
              return (value / 1000).toFixed(1) + 'K';
            }
            return value.toFixed(1);
          },
        },
        splitLine: {
          show: true,
          lineStyle: {
            type: 'dashed',
            color: '#f0f0f0',
          },
        },
      },
      dataZoom: [
        {
          type: 'inside',
          start: 0,
          end: 100,
        },
        {
          type: 'slider',
          show: false,
        },
      ],
      series: chartSeries,
    };

    chart.setOption(option);

    // 处理窗口大小变化
    const handleResize = () => {
      chart?.resize();
    };
    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
    };
  }, [series, height, threshold]);

  // 判断是否有数据超过阈值
  const hasExceededThreshold = React.useMemo(() => {
    if (threshold === undefined || threshold === null) {
      return null; // 没有阈值，不显示背景色
    }
    
    // 检查所有数据点，看是否有任何一个超过阈值
    for (const s of series) {
      for (const point of s.data) {
        if (point.value > threshold) {
          return true; // 有数据超过阈值，返回红色
        }
      }
    }
    return false; // 所有数据都在阈值以下，返回绿色
  }, [series, threshold]);

  // 同步外部传入的 initialTimeRange 变化到内部状态
  useEffect(() => {
    setTimeRange(initialTimeRange);
  }, [initialTimeRange]);

  const handleTimeRangeChange = (minutes: number) => {
    setTimeRange(minutes);
    if (onTimeRangeChange) {
      onTimeRangeChange(minutes);
    }
  };

  return (
    <div>
      {showTimeSelector && (
        <div style={{ 
          marginBottom: '12px',
          display: 'flex',
          justifyContent: 'flex-start',
        }}>
          <TimeRangeSelector 
            value={timeRange} 
            onChange={handleTimeRangeChange}
          />
        </div>
      )}
      <div
        ref={chartRef}
        style={{
          width: '100%',
          height: `${height}px`,
          backgroundColor: hasExceededThreshold === true 
            ? '#fff1f0' // 红色警示背景（有数据超过阈值）
            : hasExceededThreshold === false 
            ? '#f6ffed' // 绿色背景（所有数据都在阈值以下）
            : 'transparent', // 没有阈值，不显示背景色
          borderRadius: '4px',
        }}
      />
    </div>
  );
};

export default MonitorChart;

