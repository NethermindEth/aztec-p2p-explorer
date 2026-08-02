import { StoryObj, Meta } from '@storybook/react';
import PanelSection from './PanelSection';

export default {
  title: 'Components/Molecules/PanelSection',
  component: PanelSection,
} as Meta<typeof PanelSection>;

type Story = StoryObj<typeof PanelSection>;

export const Default: Story = {
  args: {
    title: 'Sync status',
    items: [
      {
        label: 'Synced',
        value: '211',
        percentage: '100%',
      },
      {
        label: 'Syncing',
        value: '2310',
        percentage: '0%',
      },
    ],
  },
};
