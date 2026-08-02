import React from 'react';
import { Pie } from 'react-chartjs-2';
import { ChartData, ChartOptions } from 'chart.js/auto';
import './StatisticsChart.css';
import Skeleton from '../../../molecules/Skeleton/Skeleton';

interface StatisticsChartProps {
  chartData: ChartData<'pie'>;
  isMobileView: boolean;
}

const StatisticsChart: React.FC<StatisticsChartProps> = ({ chartData, isMobileView }) => {
  const chartStyleOptions: ChartOptions<'pie'> = {
    responsive: true,
    maintainAspectRatio: false,
    animation: false,
    plugins: {
      legend: {
        display: false,
      },
    },
    radius: isMobileView ? '100%' : '100%',
  };

  const renderCustomLegend = (data: ChartData<'pie'>) => {
    const legendData = data.labels?.map((label, index) => {
      const value = data.datasets[0].data[index] as number;
      let backgroundColor = '';
      if (Array.isArray(data.datasets[0].backgroundColor)) {
        backgroundColor = data.datasets[0].backgroundColor[index] as string;
      }
      return { label: label as string, value, backgroundColor };
    });

    return (
      <div className={`legend-container ${isMobileView ? 'mobile-legend' : 'desktop-legend'}`}>
        {!chartData.labels?.length
          ? Array(10)
              .fill(null)
              .map((_, idx) => (
                <div key={idx} className={`legend-item ${isMobileView ? 'legend-item-mobile' : 'legend-item-desktop'}`}>
                  <Skeleton width={'20px'} height={'20px'} mr={'10px'} />
                  <Skeleton width={'80%'} height={'20px'} />
                </div>
              ))
          : legendData?.map((item, idx) => (
              <div key={idx} className={`legend-item ${isMobileView ? 'legend-item-mobile' : 'legend-item-desktop'}`}>
                <div className="legend-color-box" style={{ backgroundColor: item.backgroundColor }} />
                <span>{item.label}</span>
              </div>
            ))}
      </div>
    );
  };

  return (
    <div className="chart-container">
      <div className={`chart-wrapper ${isMobileView ? 'chart-mobile' : 'chart-desktop'}`}>
        {!chartData?.labels?.length ? (
          <Skeleton width={'262px'} height={'262px'} radius={'100%'} mx={'auto'} />
        ) : (
          <Pie data={chartData} options={chartStyleOptions} />
        )}
      </div>
      {renderCustomLegend(chartData)}
    </div>
  );
};

export default StatisticsChart;
