import { useEffect, useMemo, useRef, useState } from 'react';
import { Line } from 'react-chartjs-2';
import { Chart, ChartData, ChartArea, ChartOptions, registerables, Plugin } from 'chart.js';
import { useAtom } from 'jotai';
import { historyAtom } from '../../../../hooks/atoms';
import { useAnalyticsPeerHistory } from '../../../../api/analytics';
import { useNavigate } from 'react-router-dom';
import useHashSearchQuery from '../../../../hooks/useHashSearchQuery';
import LoadingSpinner from '../../../molecules/LoadingSpinner/LoadingSpinner';
import './NodeHistory.css';

Chart.register(...registerables);

const NodeHistory: React.FC = () => {
  const search = useHashSearchQuery();
  const navigate = useNavigate();
  const chartRef = useRef<Chart<'line'> | null>(null);
  const [range, setRange] = useState<'Week' | 'Month' | 'All'>(
    (search.get('range') as 'Week' | 'Month' | 'All') || 'All'
  );

  const getStartEndDates = (range: 'Week' | 'Month' | 'All') => {
    const end = new Date();
    let start = new Date();

    if (range === 'Week') {
      start.setDate(end.getDate() - 7);
    } else if (range === 'Month') {
      start.setDate(end.getDate() - 30);
    } else {
      start = new Date('2024-01-01');
    }

    return { start, end };
  };

  const { start, end } = useMemo(() => getStartEndDates(range), [range]);
  const { data } = useAnalyticsPeerHistory({ start, end });
  const [history, setHistory] = useAtom(historyAtom);

  useEffect(() => {
    if (data?.history) setHistory(data.history);
  }, [data, setHistory]);

  const formatDate = (dateString: string): string => {
    const date = new Date(dateString);
    return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
  };

  const labels = history.map((entry) => formatDate(entry.date)) || [];
  const values = history.map((entry) => entry.peer_count) || [];

  const getGradient = (ctx: CanvasRenderingContext2D, chartArea: ChartArea) => {
    const { top, bottom } = chartArea;
    const gradient = ctx.createLinearGradient(0, top, 0, bottom);
    // Teal/green gradient to match app styling
    gradient.addColorStop(0, 'rgba(1, 200, 164, 0.25)');
    gradient.addColorStop(0.35, 'rgba(1, 200, 164, 0.18)');
    gradient.addColorStop(1, 'rgba(1, 200, 164, 0.00)');
    return gradient;
  };

  const dataConfig: ChartData<'line'> = {
    labels,
    datasets: [
      {
        label: 'Node count over time',
        data: values,
        fill: true,
        borderColor: 'rgba(1, 200, 164, 1)',
        borderWidth: 2,
        tension: 0.3,
        pointRadius: 2,
        pointHoverRadius: 3,
        pointBackgroundColor: 'rgba(1, 200, 164, 1)',
        pointBorderColor: 'rgba(1, 200, 164, 1)',
        backgroundColor: (context) => {
          const chart = context.chart;
          const { ctx, chartArea } = chart;

          if (!chartArea) {
            return 'rgba(62, 99, 221, 0.00)';
          }

          return getGradient(ctx, chartArea);
        },
      },
    ],
  };

  const options: ChartOptions<'line'> = {
    maintainAspectRatio: false,
    responsive: true,
    animation: false,
    scales: {
      x: {
        grid: { display: false },
        ticks: {
          color: 'rgba(241, 247, 254, 0.71)',
        },
      },
      y: {
        beginAtZero: true,
        grid: {
          color: 'rgba(255, 255, 255, 0.10)',
        },
        ticks: {
          color: 'rgba(241, 247, 254, 0.71)',
        },
      },
    },
    plugins: {
      legend: {
        display: false,
      },
      tooltip: {
        mode: 'index',
        intersect: false,
        backgroundColor: 'rgba(5, 7, 10, 0.9)',
        borderColor: 'rgba(255, 255, 255, 0.1)',
        borderWidth: 1,
        titleColor: '#fff',
        bodyColor: '#fff',
      },
    },
    interaction: {
      mode: 'nearest',
      axis: 'x',
      intersect: false,
    },
  };

  // Draw a dashed vertical guide line at the hovered x position
  const hoverGuidePlugin: Plugin<'line'> = {
    id: 'hoverGuide',
    afterDatasetsDraw(chart) {
      const {
        ctx,
        tooltip,
        chartArea: { top, bottom },
      } = chart as unknown as Chart<'line'> & { tooltip: any };
      const active = tooltip?.getActiveElements?.();
      if (!active || active.length === 0) return;
      const x = active[0].element?.x;
      if (typeof x !== 'number') return;
      ctx.save();
      ctx.beginPath();
      ctx.setLineDash([4, 4]);
      ctx.moveTo(x, top);
      ctx.lineTo(x, bottom);
      ctx.lineWidth = 1;
      ctx.strokeStyle = 'rgba(255, 255, 255, 0.6)';
      ctx.stroke();
      ctx.restore();
    },
  };

  const handleRangeChange = (newRange: 'Week' | 'Month' | 'All') => {
    setRange(newRange);
    navigate(`/explore#history?range=${newRange}`);
  };

  const timeRanges: ('Week' | 'Month' | 'All')[] = ['Week', 'Month', 'All'];

  return (
    <div className="node-history">
      <div className="node-history-top-bar">
        <h3 className="node-history-title">Maximum node count per day</h3>
        <div className="time-range">
          {timeRanges.map((r) => (
            <button key={r} onClick={() => handleRangeChange(r)} className={range === r ? 'selected' : ''}>
              {r}
            </button>
          ))}
        </div>
      </div>
      <div className="chart">
        {!history || history.length < 1 ? (
          <LoadingSpinner />
        ) : (
          <Line ref={chartRef} data={dataConfig} options={options} plugins={[hoverGuidePlugin]} />
        )}
      </div>
    </div>
  );
};

export default NodeHistory;
