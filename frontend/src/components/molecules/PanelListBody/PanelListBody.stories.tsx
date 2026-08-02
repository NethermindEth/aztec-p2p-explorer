import { Meta, StoryObj } from '@storybook/react';
import PanelListBody from './PanelListBody';

export default {
  title: 'Components/Molecules/PanelListBody',
  component: PanelListBody,
} as Meta<typeof PanelListBody>;

type Story = StoryObj<typeof PanelListBody>;

// TODO Fill with values for each view
export const Default: Story = {
  args: {
    view: 'ListView',
  },
};

export const MapView: Story = {
  args: {
    view: 'MapView',
  },
};

export const SyncStatus: Story = {
  args: {
    view: 'SyncStatus',
  },
};

export const TopNetworks: Story = {
  args: {
    view: 'TopNetworks',
  },
};

export const TopClients: Story = {
  args: {
    view: 'TopClients',
  },
};
