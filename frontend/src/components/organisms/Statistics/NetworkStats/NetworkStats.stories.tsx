import { StoryObj, Meta } from '@storybook/react';

import NetworkStats from './NetworkStats';

export default {
  title: 'Organisms/Statistics/NetworkStats',
  component: NetworkStats,
} as Meta<typeof NetworkStats>;

type Story = StoryObj<typeof NetworkStats>;

export const Default: Story = {
  args: {},
};
