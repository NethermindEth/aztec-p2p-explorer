import { StoryObj, Meta } from '@storybook/react';
import TopNetworks from './TopNetworks';

export default {
  title: 'Organisms/Statistics/TopNetworks',
  component: TopNetworks,
} as Meta<typeof TopNetworks>;

type Story = StoryObj<typeof TopNetworks>;

const testData = {
  'ATT-INTERNET4': [
    {
      clientName: 'Google',
      clientVersion: '1.0.0',
      isSynced: true,
    },
  ],
  VERIZON: [
    {
      clientName: 'Apple',
      clientVersion: '2.0.0',
      isSynced: true,
    },
    {
      clientName: 'Microsoft',
      clientVersion: '2.1.0',
      isSynced: true,
    },
  ],
  COMCAST: [
    {
      clientName: 'Amazon',
      clientVersion: '3.0.0',
      isSynced: false,
    },
    {
      clientName: 'Netflix',
      clientVersion: '3.1.0',
      isSynced: true,
    },
  ],
};

export const Default: Story = {
  args: {
    data: {
      ...testData,
    },
  },
};

export const MissingInfo: Story = {
  args: {
    data: {
      '': [
        {
          clientName: '',
          clientVersion: '1.0.0',
          isSynced: true,
        },
      ],
      VERIZON: [
        {
          clientName: 'Apple',
          clientVersion: '',
          isSynced: true,
        },
        {
          clientName: 'Microsoft',
          clientVersion: '2.1.0',
          isSynced: true,
        },
      ],
      COMCAST: [
        {
          clientName: 'Amazon',
          clientVersion: '3.0.0',
          isSynced: false,
        },
        {
          clientName: '',
          clientVersion: '3.1.0',
          isSynced: true,
        },
      ],
    },
  },
};
