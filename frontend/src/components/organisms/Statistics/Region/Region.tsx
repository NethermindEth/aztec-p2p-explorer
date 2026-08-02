import CountryTable from './CountryTable';
import { useAtom, useAtomValue } from 'jotai';
import { countriesStatsAtom, isMobileAtom } from '../../../../hooks/atoms';
import StatisticsChart from '../NetworkStats/StatisticsChart';
import './Region.css';

const Region = () => {
  const [isMobile] = useAtom(isMobileAtom);
  const contriesStats = useAtomValue(countriesStatsAtom);
  const top10Countries = contriesStats?.slice(0, 10);

  const chartData = {
    labels: top10Countries?.map((country) => `${country.country_name} (${country.percent})`),
    datasets: [
      {
        data: top10Countries?.map((country) => country.count) || [], // Raw totals for each country
        backgroundColor: [
          '#233572',
          '#3e63dd',
          '#36A2EB',
          '#4BC0C0',
          '#9966FF',
          '#FF9F40',
          '#2A9D8F',
          '#36A2EB',
          '#FFCE56',
          '#4BC0C0',
        ], // Random colors
        borderWidth: 0,
      },
    ],
  };

  return (
    <div className="region-table-chart-container">
      <CountryTable stats={top10Countries || []} />

      <StatisticsChart chartData={chartData} isMobileView={isMobile} />
    </div>
  );
};

export default Region;
