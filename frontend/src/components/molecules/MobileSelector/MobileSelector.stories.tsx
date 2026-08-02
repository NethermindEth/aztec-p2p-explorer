import { StoryObj, Meta } from '@storybook/react';
import MobileSelector from './MobileSelector';

export default {
  title: 'Components/Molecules/MobileSelector',
  component: MobileSelector,
} as Meta<typeof MobileSelector>;

type Story = StoryObj<typeof MobileSelector>;

export const Default: Story = {
  args: {
    setMobileActiveTab: () => {},
  },
};
