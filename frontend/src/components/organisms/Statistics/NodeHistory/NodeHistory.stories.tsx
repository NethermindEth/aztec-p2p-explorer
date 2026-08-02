import { StoryObj, Meta } from '@storybook/react';
import NodeHistory from './NodeHistory';

export default {
  title: 'Organisms/Statistics/NodeHistory',
  component: NodeHistory,
} as Meta<typeof NodeHistory>;

type Story = StoryObj<typeof NodeHistory>;

export const Default: Story = {
  args: {},
};
