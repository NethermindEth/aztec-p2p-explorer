import React from 'react';
import type { Preview, StoryFn } from '@storybook/react';
import '@radix-ui/themes/styles.css';
import { Theme } from '@radix-ui/themes';
import '../src/globalStyles.css';
const CustomTheme = (Story: StoryFn, context: any) => {
  return (
    <Theme className="theme" appearance="dark">
      <Story {...context} />
    </Theme>
  );
};

const preview: Preview = {
  decorators: [CustomTheme],
  parameters: {
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
  },
};

export default preview;
