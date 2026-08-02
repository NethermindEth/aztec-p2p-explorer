import { StoryObj, Meta } from '@storybook/react';
import Pagination from './Pagination';

export default {
  title: 'Organisms/Statistics/Pagination',
  component: Pagination,
} as Meta<typeof Pagination>;

type Story = StoryObj<typeof Pagination>;

export const Default: Story = {
  args: {
    currentPage: 1,
    totalPages: 10,
    onFirst: () => {},
    onPrevious: () => {},
    onNext: () => {},
    onLast: () => {},
  },
};
