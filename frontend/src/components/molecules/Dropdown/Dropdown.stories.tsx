import { Meta, StoryObj } from '@storybook/react';
import Dropdown from './Dropdown';

export default {
  title: 'Components/Molecules/Dropdown',
  component: Dropdown,
  decorators: [
    (Story) => (
      <div style={{ marginTop: '350px' }}>
        <Story />
      </div>
    ),
  ],
} as Meta<typeof Dropdown>;

type Story = StoryObj<typeof Dropdown>;

export const Default: Story = {
  args: {
    label: 'Country',
    value: ['All'],
    options: ['All', 'Germany', 'Poland', 'France', 'Spain', 'Italy', 'UK', 'USA'],
    counts: {
      All: 10000,
      'Option 1': 5000,
      'Option 2': 5000,
    },
    onFilterChange: (value) => {
      console.log(value);
    },
  },
};

export const Open: Story = {
  args: {
    label: 'Country',
    value: ['All'],
    options: ['All', 'Germany', 'Poland', 'France', 'Spain', 'Italy', 'UK', 'USA'],
    counts: {
      All: 10000,
      'Option 1': 5000,
      'Option 2': 5000,
    },
    onFilterChange: (value) => {
      console.log(value);
    },
    isOpen: true,
  },
};

export const EmptyValues: Story = {
  args: {
    label: 'Country',
    value: [],
    options: [],
    counts: {},
    onFilterChange: (value) => {
      console.log(value);
    },
  },
};
