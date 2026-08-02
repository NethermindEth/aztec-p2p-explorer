import { Meta, StoryObj } from '@storybook/react';
import NodeExplorer from './NodeExplorer';

export default {
  title: 'Components/Atoms/NodeExplorer',
  component: NodeExplorer,
} as Meta<typeof NodeExplorer>;

type Story = StoryObj<typeof NodeExplorer>;

export const Default: Story = {
  args: {
    onShowNetworkStats: () => {},
  },
};
