import { Meta, StoryObj } from '@storybook/react';
import FilterBar from './FilterMenuItem';

export default {
  title: 'Components/Molecules/FilterMenuItem',
  component: FilterBar,
} as Meta<typeof FilterBar>;

type Story = StoryObj<typeof FilterBar>;

export const Default: Story = {
  args: {
    option: 'All',
    selectedValues: [],
    counts: {
      All: 10000,
      'Option 1': 5000,
      'Option 2': 5000,
    },
    handleValueChange: () => {},
  },
};

export const OnMapFilter: Story = {
  args: {
    option: 'All',
    selectedValues: [],
    counts: {
      All: 10000,
      'Option 1': 5000,
      'Option 2': 5000,
    },
    handleValueChange: () => {},
  },
};

export const StatisticFilter: Story = {
  args: {
    option: 'All',
    selectedValues: [],
    counts: {
      All: 10000,
      'Option 1': 5000,
      'Option 2': 5000,
    },
    handleValueChange: () => {},
  },
};
