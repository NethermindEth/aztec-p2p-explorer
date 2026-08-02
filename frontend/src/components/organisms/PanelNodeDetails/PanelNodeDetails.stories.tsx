import { StoryObj, Meta } from '@storybook/react';
import PanelNodeDetails from './PanelNodeDetails';

export default {
  title: 'Organisms/PanelNodeDetails',
  component: PanelNodeDetails,
} as Meta<typeof PanelNodeDetails>;

type Story = StoryObj<typeof PanelNodeDetails>;

export const Default: Story = {
  args: {
    node: {
      id: '1',
      created_at: '2023-01-01T00:00:00Z',
      last_seen: '2023-01-02T00:00:00Z',
      client: 'Corda/1.0.0',
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
  },
};
