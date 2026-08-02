import { Meta, StoryObj } from '@storybook/react';
import ClearFilters from './ClearFilters';

interface ClearFiltersProps {
  clearFilters: () => void;
}

export default {
  title: 'Components/Atoms/ClearFilters',
  component: ClearFilters,
} as Meta<typeof ClearFilters>;

type Story = StoryObj<ClearFiltersProps>;

export const Default: Story = {
  args: {
    clearFilters: () => {},
  },
};
