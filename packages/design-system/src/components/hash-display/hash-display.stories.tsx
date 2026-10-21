import type { Meta, StoryObj } from '@storybook/react-vite';
import { LiveRegionProvider } from '../live-region/live-region';
import { HashDisplay } from './hash-display';

const meta = {
  component: HashDisplay,
  decorators: [
    (Story) => (
      <LiveRegionProvider>
        <Story />
      </LiveRegionProvider>
    ),
  ],
  args: {
    label: 'Document SHA-256',
    // SHA-256 of empty input; synthetic.
    value: 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855',
    copyLabel: 'Copy',
    copiedMessage: 'Hash copied.',
    copyFailedMessage: 'Copying failed. Select the value to copy it manually.',
  },
} satisfies Meta<typeof HashDisplay>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
