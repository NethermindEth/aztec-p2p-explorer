import { StoryObj, Meta } from '@storybook/react';
import TableComponent from './TableComponent';

export default {
  title: 'Organisms/Statistics/TableComponent',
  component: TableComponent,
} as Meta<typeof TableComponent>;

type Story = StoryObj<typeof TableComponent>;
const testData = [
  {
    id: '2001:0db8:85a3:0000:0000:8a2e:0370:7334',
    created_at: '2023-01-01T00:00:00Z',
    last_seen: '2023-01-02T00:00:00Z',
    client: 'aztec-node/1.0.0',
    multi_addresses: [
      {
        maddr: '/ip4/35.231.95.227/tcp/7777',
        ip_info: [
          {
            ip_address: '35.231.95.227',
            port: 7777,
            as_name: 'ASOName',
            as_number: 396982,
            city_name: 'CityName',
            country_name: 'CountryName',
            country_iso: 'US',
            continent_name: 'ContinentName',
            continent_code: 'NA',
            latitude: 32.8608,
            longitude: -79.9746,
          },
        ],
      },
    ],
    protocols: ['protocol1', 'protocol2'],
    block_height: 100,
    is_synced: true,
    spec_version: '1.0.0',
  },
  {
    id: '2001:0db8:85a3:0000:0000:8a2e:0370:7335',
    created_at: '2023-01-03T00:00:00Z',
    last_seen: '2023-01-04T00:00:00Z',
    client: 'delta-node/2.0.0',
    multi_addresses: [
      {
        maddr: '/ip4/35.231.95.227/tcp/7777',
        ip_info: [
          {
            ip_address: '35.231.95.227',
            port: 7777,
            as_name: 'ASOName',
            as_number: 396982,
            city_name: 'CityName',
            country_name: 'CountryName',
            country_iso: 'US',
            continent_name: 'ContinentName',
            continent_code: 'NA',
            latitude: 32.8608,
            longitude: -79.9746,
          },
        ],
      },
    ],
    protocols: ['protocol3', 'protocol4'],
    block_height: 200,
    is_synced: false,
    spec_version: '1.0.0',
  },
];
export const Default: Story = {
  args: {
    data: testData,
    onRowClick: () => {},
    onSort: () => {},
    sortColumn: '',
    sortOrder: 'asc',
  },
};

export const MissingInfo: Story = {
  args: {
    data: testData.map((item) => ({
      ...item,
      client: '',
      country: '',
      city: '',
    })),
    onRowClick: () => {},
    onSort: () => {},
    sortColumn: '',
    sortOrder: 'asc',
  },
};
