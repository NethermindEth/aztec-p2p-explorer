import { Table } from '@radix-ui/themes';
import Skeleton from '../../../molecules/Skeleton/Skeleton';
import { AnalyticsCountry } from '../../../../types';

type CountryTableProps = {
  stats: Array<AnalyticsCountry & { percent: string }>;
};
const CountryTable: React.FC<CountryTableProps> = ({ stats }) => {
  const columnHeaders = ['#', 'Country', 'Amt', '%'];

  return (
    <div className="country-table-container fade-in">
      <Table.Root className="table-root" layout="auto" size="2">
        <Table.Header className="table-header">
          <Table.Row className="table-row">
            {columnHeaders.map((header, index) => (
              <Table.ColumnHeaderCell key={index} className="table-column-header-cell">
                {header}
              </Table.ColumnHeaderCell>
            ))}
          </Table.Row>
        </Table.Header>
        <Table.Body className="table-body">
          {!stats.length
            ? Array(10)
                .fill(null)
                .map((_, idx) => (
                  <Table.Row className="table-row" key={idx}>
                    <Table.Cell>
                      <Skeleton height={'20px'} width={'50%'} />
                    </Table.Cell>
                    <Table.Cell>
                      <Skeleton height={'20px'} width={'100%'} />
                    </Table.Cell>
                    <Table.Cell>
                      <Skeleton height={'20px'} width={'30%'} />
                    </Table.Cell>
                    <Table.Cell>
                      <Skeleton height={'20px'} width={'30%'} />
                    </Table.Cell>
                  </Table.Row>
                ))
            : stats.map((country, index) => (
                <Table.Row className="table-row" key={country.country_code}>
                  <Table.RowHeaderCell className="table-row-header-cell">{index + 1}</Table.RowHeaderCell>
                  <Table.Cell className="region-table-cell" style={{ textTransform: 'capitalize' }}>
                    {country.country_name}
                  </Table.Cell>
                  <Table.Cell className="region-table-cell">{country.count}</Table.Cell>
                  <Table.Cell className="region-table-cell">{country.percent}</Table.Cell>
                </Table.Row>
              ))}
        </Table.Body>
      </Table.Root>
    </div>
  );
};

export default CountryTable;
