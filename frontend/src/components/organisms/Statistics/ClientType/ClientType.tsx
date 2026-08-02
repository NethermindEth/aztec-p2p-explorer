import './ClientType.css';
import { useAtom } from 'jotai';
import { activeClientsAtom, isMobileAtom } from '../../../../hooks/atoms';
import { Table } from '@radix-ui/themes';
import { FilterValues } from '../../../../types';
import { useState, useEffect, useMemo } from 'react';
import StatisticsChart from '../NetworkStats/StatisticsChart'; // Import the chart

interface ClientTypeProps {
  filterValues: FilterValues;
}

type ClientPercentagesHolder = Record<string, string>;

const ClientType: React.FC<ClientTypeProps> = ({ filterValues }) => {
  const [clients] = useAtom(activeClientsAtom);
  const [isMobile] = useAtom(isMobileAtom);
  const [clientPercentages, setClientPercentages] = useState<ClientPercentagesHolder>({});
  const [clientTotals, setClientTotals] = useState<{ [key: string]: number }>({});

  // Filter and calculate clients and versions based on filters
  const filteredClients = useMemo(() => {
    const result: { [key: string]: { [version: string]: { synced: number; unsynced: number } } } = {};
    return result;
  }, [clients, filterValues]);

  // Calculate totals and percentages
  useEffect(() => {
    const totals: { [key: string]: number } = {};
    let overallTotal = 0;

    for (const client in filteredClients) {
      const versions = filteredClients[client];
      const total = Object.values(versions).reduce((acc, value) => {
        if (
          filterValues.sync.includes('All') ||
          (filterValues.sync.includes('Synced') && filterValues.sync.includes('Unsynced'))
        ) {
          return acc + value.synced + value.unsynced;
        } else if (filterValues.sync.includes('Synced')) {
          return acc + value.synced;
        } else if (filterValues.sync.includes('Unsynced')) {
          return acc + value.unsynced;
        }
        return acc;
      }, 0);
      totals[client] = total;
      overallTotal += total;
    }

    setClientTotals(totals);

    const percentages: ClientPercentagesHolder = {};
    for (const client in totals) {
      percentages[client] = overallTotal > 0 ? ((totals[client] / overallTotal) * 100).toFixed(0) : '0';
    }
    setClientPercentages(percentages);
  }, [filteredClients, filterValues]);

  const sortedClients = useMemo(() => {
    return Object.keys(clientTotals).sort((a, b) => clientTotals[b] - clientTotals[a]);
  }, [clientTotals]);

  // Prepare data for the StatisticsChart
  const chartData = useMemo(
    () => ({
      labels: sortedClients.map((client) => `${client || 'Unknown'} (${clientPercentages[client]}%)`),
      datasets: [
        {
          data: sortedClients.map((client) => clientTotals[client]),
          backgroundColor: ['#FF6384', '#36A2EB', '#FFCE56', '#4BC0C0', '#9966FF', '#FF9F40'],
          borderWidth: 0,
        },
      ],
    }),
    [sortedClients, clientTotals, clientPercentages]
  );

  const headers = [
    { className: 'table-column-header-cell-clienttype', label: '#' },
    { className: 'table-column-header-cell-clienttype', label: 'Client type' },
    { className: 'table-column-header-cell-clienttype', label: 'Amt' },
    { className: 'table-column-header-cell-clienttype', label: '%' },
  ];

  const renderTableCells = (index: number, client: string) => {
    const cells = [
      { key: 'index', content: index + 1 },
      { key: 'client', content: client || 'Unknown' },
      { key: 'amount', content: clientTotals[client] || 'Unknown' },
      {
        key: 'percentage',
        content: `${clientPercentages[client] || '-- '}%`,
      },
    ];

    return cells.map((cell) => (
      <Table.Cell key={cell.key} className="table-cell-clienttype">
        {cell.content}
      </Table.Cell>
    ));
  };

  return (
    <div className="client-type-chart-container fade-in">
      <div className="client-type-table-container">
        <Table.Root layout="auto" size="2">
          <Table.Header className="client-type-table-header">
            <Table.Row>
              {headers.map((header, index) => (
                <Table.ColumnHeaderCell key={index} className={header.className}>
                  {header.label}
                </Table.ColumnHeaderCell>
              ))}
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {sortedClients.map((client, index) => (
              <Table.Row key={client} className="table-row-clienttype">
                {renderTableCells(index, client)}
              </Table.Row>
            ))}
          </Table.Body>
        </Table.Root>
      </div>

      {/* Render the chart */}
      <StatisticsChart chartData={chartData} isMobileView={isMobile} />
    </div>
  );
};

export default ClientType;
