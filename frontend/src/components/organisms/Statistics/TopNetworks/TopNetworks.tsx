import './TopNetworks.css';
import { Table } from '@radix-ui/themes';
import { ClientInfo } from '../../../../types';
import { useEffect } from 'react';
import { useAtom } from 'jotai';
import { networkTotalsAtom, networkPercentagesAtom, sortedNetworksAtom, isMobileAtom } from '../../../../hooks/atoms';
import StatisticsChart from '../NetworkStats/StatisticsChart';

interface TopNetworksProps {
  data: { [key: string]: ClientInfo[] };
}

const TopNetworks: React.FC<TopNetworksProps> = ({ data }) => {
  const [networkTotals, setNetworkTotals] = useAtom(networkTotalsAtom);
  const [networkPercentages, setNetworkPercentages] = useAtom(networkPercentagesAtom);
  const [sortedNetworks, setSortedNetworks] = useAtom(sortedNetworksAtom);
  const [isMobile] = useAtom(isMobileAtom);

  useEffect(() => {
    // Calculate the total amount for each network
    const totals: { [key: string]: number } = {};
    let overallTotal = 0;

    for (const network in data) {
      const total = data[network].length;
      totals[network] = total;
      overallTotal += total;
    }

    setNetworkTotals(totals);

    // Calculate percentages for each network
    const percentages: { [key: string]: string } = {};
    for (const network in totals) {
      percentages[network] = ((totals[network] / overallTotal) * 100).toFixed(0);
    }

    setNetworkPercentages(percentages);

    const sortInDesecendingOrder = Object.keys(totals)
      .sort((a, b) => totals[b] - totals[a])
      .slice(0, 10);

    setSortedNetworks(sortInDesecendingOrder);
  }, [data, setNetworkTotals, setNetworkPercentages, setSortedNetworks]);

  // Renders the network table cells
  const renderNetworkCells = (network: string) => {
    type CellData = {
      key: string;
      value: string | number;
      style?: React.CSSProperties;
    };

    const cellData: CellData[] = [
      { key: 'name', value: network || 'Unknown', style: { textTransform: 'capitalize' } },
      { key: 'total', value: networkTotals[network] || 'Unknown' },
      {
        key: 'percentage',
        value: `${networkPercentages[network] || '-- '}%`,
      },
    ];

    return cellData.map((cell, index) => (
      <Table.Cell key={index} className="top-network-table-cell" style={cell.style || {}}>
        {cell.value}
      </Table.Cell>
    ));
  };

  const columnHeaders = ['#', 'Network', 'Amt', '%'];

  const chartData = {
    labels: sortedNetworks.map(
      (network) =>
        `${network.length > (isMobile ? 15 : 10) ? `${network.substring(0, isMobile ? 15 : 10)}...` : network} (${networkPercentages[network]}%)`
    ),
    datasets: [
      {
        data: sortedNetworks.map((network) => networkTotals[network]), // Raw totals
        backgroundColor: [
          '#FF6384',
          '#36A2EB',
          '#FFCE56',
          '#4BC0C0',
          '#9966FF',
          '#FF9F40',
          '#2A9D8F',
          '#36A2EB',
          '#FFCE56',
          '#4BC0C0',
        ], // Color for each network
        borderWidth: 0,
      },
    ],
  };

  return (
    <div className="network-table-container fade-in">
      <Table.Root layout="auto" size="2" className="top-network-table-root">
        <Table.Header className="top-network-table-header">
          <Table.Row>
            {columnHeaders.map((header, index) => (
              <Table.ColumnHeaderCell key={index} className="table-column-header-cell">
                {header}
              </Table.ColumnHeaderCell>
            ))}
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {sortedNetworks.map((network, index) => (
            <Table.Row className="top-network-table-row" key={network}>
              <Table.RowHeaderCell className="top-network-table-row-header-cell">{index + 1}</Table.RowHeaderCell>
              {renderNetworkCells(network)}
            </Table.Row>
          ))}
        </Table.Body>
      </Table.Root>

      <StatisticsChart chartData={chartData} isMobileView={isMobile} />
    </div>
  );
};

export default TopNetworks;
