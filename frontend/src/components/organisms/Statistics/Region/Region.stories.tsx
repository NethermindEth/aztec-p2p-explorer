import { StoryObj, Meta } from '@storybook/react';
import Region from './Region';

export default {
  title: 'Organisms/Statistics/Region',
  component: Region,
} as Meta<typeof Region>;

type Story = StoryObj<typeof Region>;

export const Default: Story = {
  args: {
    data: {
      countryTotals: { Germany: 50, France: 30, Spain: 20 },
      sortedCountries: ['Germany', 'France', 'Spain'],
      countryPercentages: { Germany: '50', France: '30', Spain: '20' },
    },
  },
};

export const MissingInfo: Story = {
  args: {
    data: {
      countryTotals: { Germany: 50, France: 30, Spain: 20 },
      sortedCountries: ['Germany', '', 'Spain'],
      countryPercentages: { Germany: '', France: '30', Spain: '20' },
    },
  },
};
