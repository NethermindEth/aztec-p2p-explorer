import { Meta, StoryObj } from '@storybook/react';
import PanelBlocks from './PanelBlocks';
import tpsIcon from '../../../assets/icons/tps.svg';
import txsIcon from '../../../assets/icons/txs.svg';
import blockIcon from '../../../assets/icons/block.svg';

export default {
  title: 'Components/Atoms/PanelBlocks',
  component: PanelBlocks,
} as Meta<typeof PanelBlocks>;

type Story = StoryObj<typeof PanelBlocks>;

export const Default: Story = {
  args: {
    label: 'TPS',
    svg: tpsIcon,
    value: '1234.56',
  },
};

export const Blocks: Story = {
  args: {
    label: 'Blocks',
    svg: blockIcon,
    value: '123456',
  },
};

export const Transactions: Story = {
  args: {
    label: 'Transactions',
    svg: txsIcon,
    value: '789012',
  },
};

// for any of the information tx, blocks etc
export const EmptyValue: Story = {
  args: {
    label: 'Transactions',
    svg: txsIcon,
    value: undefined,
  },
};
