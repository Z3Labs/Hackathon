import React, { useEffect, useRef, useState } from 'react';
import * as echarts from 'echarts';
import TimeRangeSelector from './TimeRangeSelector';

interface MonitorSeries {
  instance: string;
  metric: string;
  unit: string;
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
}

const MonitorChart: React.FC<MonitorChartProps> = ({ 
  series, 
  height = 300,
  showTimeSelector = true,
  initialTimeRange = 30,
  onTimeRangeChange 
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
      
      return {
        name: s.instance,
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
          
          return result;
        },
      },
      legend: {
        data: series.map((s) => s.instance),
        bottom: 15,
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
        left: '3%',
        right: '3%',
        top: '45px',
        bottom: '18%',
        containLabel: false,
      },
      xAxis: {
        type: 'time',
        boundaryGap: [0, 0],
        axisLine: {
          show: true,
          lineStyle: {
            color: '#d9d9d9',
            width: 1,
          },
        },
        axisLabel: {
          fontSize: 12,
          color: '#8c8c8c',
          formatter: (value: any) => {
            const date = new Date(value);
            const hours = date.getHours();
            const minutes = date.getMinutes();
            return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}`;
          },
          margin: 8,
        },
        axisTick: {
          show: true,
          lineStyle: {
            color: '#d9d9d9',
          },
        },
        splitLine: {
          show: false,
        },
      },
      yAxis: {
        type: 'value',
        name: series.length > 0 ? (series[0]?.unit || '') : '',
        nameLocation: 'end',
        nameGap: 15,
        nameTextStyle: {
          color: '#666',
          fontSize: 13,
          fontWeight: 500,
        },
        axisLine: {
          lineStyle: {
            color: '#e0e0e0',
          },
        },
        axisLabel: {
          fontSize: 12,
          color: '#666',
          margin: 8,
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
          show: true,
          right: '6%',
          height: 20,
          borderColor: '#e0e0e0',
          dataBackground: {
            areaStyle: {
              color: '#f0f0f0',
            },
          },
          selectedDataBackground: {
            areaStyle: {
              color: '#e6f7ff',
            },
          },
          handleStyle: {
            color: '#1890ff',
          },
          textStyle: {
            color: '#666',
            fontSize: 11,
          },
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
  }, [series, height]);

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
        }}
      />
    </div>
  );
};

export default MonitorChart;

