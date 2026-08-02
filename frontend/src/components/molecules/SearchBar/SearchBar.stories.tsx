import { StoryObj, Meta } from '@storybook/react';
import SearchBar from './SearchBar';

export default {
  title: 'Components/Molecules/SearchBar',
  component: SearchBar,
} as Meta<typeof SearchBar>;

type Story = StoryObj<typeof SearchBar>;

export const Default: Story = {
  args: {
    onChange: () => {},
  },
};
