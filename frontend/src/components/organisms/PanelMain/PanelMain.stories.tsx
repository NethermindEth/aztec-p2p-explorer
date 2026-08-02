import { StoryObj, Meta } from '@storybook/react';
import PanelMain from './PanelMain';

export default {
  title: 'Organisms/PanelMain',
  component: PanelMain,
} as Meta<typeof PanelMain>;

type Story = StoryObj<typeof PanelMain>;

export const Default: Story = {
  args: {},
};
