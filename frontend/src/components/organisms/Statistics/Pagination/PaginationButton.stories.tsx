import { StoryObj, Meta } from '@storybook/react';

import PaginationButton from './PaginationButton';

export default {
  title: 'Organisms/Statistics/Pagination/PaginationButton',
  component: PaginationButton,
} as Meta<typeof PaginationButton>;

export const Default: StoryObj<typeof PaginationButton> = {
  args: {
    onClick: () => {},
    disabled: false,
    arrowIconDirectionPath: 'M9 5.77469L16 12.7747L9 19.7747',
  },
};

export const Disabled: StoryObj<typeof PaginationButton> = {
  args: {
    onClick: () => {},
    disabled: true,
    arrowIconDirectionPath: 'M9 5.77469L16 12.7747L9 19.7747',
  },
};

export const Previous: StoryObj<typeof PaginationButton> = {
  args: {
    onClick: () => {},
    disabled: false,
    direction: 'previous',
    arrowIconDirectionPath: 'M15 19.7747L8 12.7747L15 5.77469',
  },
};

export const Next: StoryObj<typeof PaginationButton> = {
  args: {
    onClick: () => {},
    disabled: false,
    direction: 'next',
    arrowIconDirectionPath: 'M9 5.77469L16 12.7747L9 19.7747',
  },
};
