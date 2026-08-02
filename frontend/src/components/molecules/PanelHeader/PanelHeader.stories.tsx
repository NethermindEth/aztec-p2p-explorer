import { StoryObj, Meta } from '@storybook/react';
import { SWRConfig } from 'swr';
import HeaderSection from './PanelHeader';
import { Spinner } from '@radix-ui/themes';

export default {
  title: 'Components/Molecules/PanelHeader',
  component: HeaderSection,
} as Meta<typeof HeaderSection>;

type Story = StoryObj<typeof HeaderSection>;

const MockedPeersProvider: React.FC<{ children: React.ReactNode; isLoading?: boolean; error?: any }> = ({
  children,
  isLoading,
  error,
}) => {
  if (error) return <div>Error: {error.message}</div>;
  return isLoading ? <Spinner /> : <>{children}</>;
};

export const Default: Story = {
  decorators: [
    (StoryComponent) => (
      <SWRConfig value={{ provider: () => new Map() }}>
        <MockedPeersProvider>
          <StoryComponent />
        </MockedPeersProvider>
      </SWRConfig>
    ),
  ],
};

export const Loading: Story = {
  decorators: [
    (StoryComponent) => (
      <SWRConfig value={{ provider: () => new Map() }}>
        <MockedPeersProvider isLoading={true}>
          <StoryComponent />
        </MockedPeersProvider>
      </SWRConfig>
    ),
  ],
};

// TODO fix this. In case of error the values should be --
export const Error: Story = {
  decorators: [
    (StoryComponent) => (
      <SWRConfig value={{ provider: () => new Map() }}>
        <MockedPeersProvider error={'Failed to fetch'}>
          <StoryComponent />
        </MockedPeersProvider>
      </SWRConfig>
    ),
  ],
};
