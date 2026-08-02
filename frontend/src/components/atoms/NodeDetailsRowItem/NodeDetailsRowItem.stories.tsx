import { Meta, StoryObj } from '@storybook/react';
import NodeDetailsRowItem from './RowItem';

export default {
  title: 'Components/Atoms/NodeDetailsRowItem',
  component: NodeDetailsRowItem,
} as Meta<typeof NodeDetailsRowItem>;

type Story = StoryObj<typeof NodeDetailsRowItem>;

export const Default: Story = {
  args: {
    title: 'City',
    content: 'Berlin',
  },
};

export const EmptyValue: Story = {
  args: {
    title: 'City',
    content: '--',
  },
};
