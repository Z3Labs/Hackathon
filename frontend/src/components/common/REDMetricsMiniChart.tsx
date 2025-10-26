import React, { useEffect, useRef } from 'react';
import * as echarts from 'echarts';

interface REDMetricsMiniChartProps {
  rate?: Array<{ timestamp: number; value: number }>;
  error?: Array<{ timestamp: number; value: number }>;
  duration?: Array<{ timestamp: number; value: number }>;
  rateThreshold?: number;
  errorThreshold?: number;
  durationThreshold?: number;
  metricName: 'Rate' | 'Error' | 'Duration';
}

const REDMetricsMiniChart: React.FC<REDMetricsMiniChartProps> = ({
  rate,
  error,
  duration,
  rateThreshold,
  errorThreshold,
  durationThreshold,
  metricName,
}) => {
  const chartRef = useRef<HTMLDivElement>(null);
  const chartInstance = useRef<echarts.ECharts | null>(null);

  useEffect(() => {
    if (!chartRef.current) return;

    if (!chartInstance.current) {
      chartInstance.current = echarts.init(chartRef.current);
    }

    // 选择对应的指标数据和阈值
    let timeSeriesData: Array<{ timestamp: number; value: number }> | undefined;
    let threshold: number | undefined;
    let backgroundColor = '#f6ffed'; // 绿色背景

    if (metricName === 'Rate') {
      timeSeriesData = rate;
      threshold = rateThreshold;
    } else if (metricName === 'Error') {
      timeSeriesData = error;
      threshold = errorThreshold;
    } else if (metricName === 'Duration') {
      timeSeriesData = duration;
      threshold = durationThreshold;
    }

    // 如果没有数据，清空图表
    if (!timeSeriesData || timeSeriesData.length === 0) {
      chartInstance.current.clear();
      return;
    }

    // 使用真实的时间序列数据
    const data: [number, number][] = [];
    timeSeriesData.forEach(d => {
      data.push([d.timestamp, d.value]);
    });

    // 判断背景色：检查所有数据点，只要有任何一点超过阈值就显示红色
    // 所有指标统一逻辑：值越高越坏
    if (threshold !== undefined) {
      const hasExceededThreshold = timeSeriesData.some(d => d.value > threshold);
      if (hasExceededThreshold) {
        backgroundColor = '#fff1f0'; // 红色背景
      }
    }

    const option: echarts.EChartsOption = {
      backgroundColor: backgroundColor,
      tooltip: {
        trigger: 'axis',
        backgroundColor: 'rgba(0, 0, 0, 0.85)',
        textStyle: {
          color: '#fff',
          fontSize: 11,
        },
        formatter: (params: any) => {
          if (!Array.isArray(params) || params.length === 0) return '';
          const param = params[0];
          const time = new Date(param.value[0]);
          const timeStr = time.toLocaleString('zh-CN', {
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
          });
          let result = `${timeStr}<br/>${param.marker} ${param.seriesName}: ${param.value[1].toFixed(2)}`;
          if (threshold !== undefined) {
            result += `<br/>阈值: ${threshold.toFixed(2)}`;
          }
          return result;
        },
      },
      grid: {
        left: '15%',
        right: '5%',
        top: '10%',
        bottom: '25%',
        containLabel: false,
      },
      xAxis: {
        type: 'time',
        show: false,
        boundaryGap: [0, 0],
      },
      yAxis: {
        type: 'value',
        show: false,
        scale: false,
      },
      series: [
        {
          name: metricName,
          type: 'line',
          smooth: true,
          data: data,
          symbol: 'none',
          lineStyle: {
            color: '#1890ff',
            width: 1.5,
          },
          areaStyle: {
            color: {
              type: 'linear',
              x: 0,
              y: 0,
              x2: 0,
              y2: 1,
              colorStops: [
                { offset: 0, color: 'rgba(24, 144, 255, 0.3)' },
                { offset: 1, color: 'rgba(24, 144, 255, 0)' },
              ],
            },
          },
          emphasis: {
            focus: 'series',
          },
          markLine: threshold !== undefined
            ? {
                data: [{ yAxis: threshold }],
                lineStyle: {
                  color: '#ff4d4f',
                  type: 'dashed' as const,
                  width: 1.5,
                },
                label: {
                  show: false,
                },
                symbol: 'none',
                silent: true,
              }
            : undefined,
        },
      ],
      animation: false,
    };

    chartInstance.current.setOption(option);

    // 处理窗口大小变化
    const handleResize = () => {
      chartInstance.current?.resize();
    };
    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
    };
  }, [rate, error, duration, rateThreshold, errorThreshold, durationThreshold, metricName]);

  // 检查是否有数据
  let timeSeriesData: Array<{ timestamp: number; value: number }> | undefined;
  if (metricName === 'Rate') {
    timeSeriesData = rate;
  } else if (metricName === 'Error') {
    timeSeriesData = error;
  } else if (metricName === 'Duration') {
    timeSeriesData = duration;
  }

  // 如果没有数据，显示提示
  if (!timeSeriesData || timeSeriesData.length === 0) {
    return (
      <div
        style={{
          width: '120px',
          height: '40px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          background: '#fafafa',
          borderRadius: '2px',
          fontSize: '11px',
          color: '#999',
        }}
      >
        RED 无数据
      </div>
    );
  }

  return (
    <div
      ref={chartRef}
      style={{
        width: '120px',
        height: '40px',
      }}
    />
  );
};

export default REDMetricsMiniChart;

