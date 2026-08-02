import { Meta, StoryObj } from '@storybook/react';

import FilterBar from './FilterBar';
// import { Continent, Clients, NetworkItems } from '../../../types';

export default {
  title: 'Components/Molecules/FilterBar',
  component: FilterBar,
  decorators: [
    (Story) => (
      <div style={{ marginTop: '350px' }}>
        <Story />
      </div>
    ),
  ],
} as Meta<typeof FilterBar>;

type Story = StoryObj<typeof FilterBar>;

// const continents: Continent[] = [
//   {
//     continent_name: 'Europe',
//     continent_code: 'EU',
//     count: 3,
//   },
//   {
//     continent_name: 'Africa',
//     continent_code: 'AF',
//     count: 4,
//   },
//   {
//     continent_name: 'Asia',
//     continent_code: 'AS',
//     count: 5,
//   },
//   {
//     continent_name: 'Australia',
//     continent_code: 'AU',
//     count: 6,
//   },
// ];

// const clients: Clients = {
//   alpha-node: {
//     '1.0.0': { synced: 1, unsynced: 0 },
//     '1.0.1': { synced: 0, unsynced: 0 },
//   },
//   apollo: {
//     '2.0.0': { synced: 2, unsynced: 1 },
//     '2.1.0': { synced: 1, unsynced: 1 },
//   },
// };

// const networkItems: NetworkItems = {
//   network1: [
//     {
//       clientName: 'ClientA',
//       clientVersion: '1.0.0',
//       isSynced: true,
//     },
//     {
//       clientName: 'ClientB',
//       clientVersion: '1.1.0',
//       isSynced: false,
//     },
//   ],
//   network2: [
//     {
//       clientName: 'ClientC',
//       clientVersion: '2.0.0',
//       isSynced: true,
//     },
//     {
//       clientName: 'ClientD',
//       clientVersion: '2.1.0',
//       isSynced: true,
//     },
//   ],
// };

export const Default: Story = {
  args: {
    // data: { networkItems, continents, clients },
    onFilterChange: undefined,
  },
};

export const EmptyValues: Story = {
  args: {
    // data: { networkItems: {}, continents: [], clients: {} },
    onFilterChange: undefined,
  },
};
