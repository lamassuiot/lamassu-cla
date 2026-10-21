import type { Meta, StoryObj } from '@storybook/react-vite';
import { Button } from '../button/button';
import { LiveRegionProvider, useAnnounce } from './live-region';

function Demo({ message }: { message: string }) {
  const announce = useAnnounce();
  return (
    <>
      <Button onClick={() => announce(message)}>Announce politely</Button>
      <Button variant="secondary" onClick={() => announce(message, 'assertive')}>
        Announce urgently
      </Button>
    </>
  );
}

const meta = {
  component: Demo,
  decorators: [
    (Story) => (
      <LiveRegionProvider>
        <Story />
      </LiveRegionProvider>
    ),
  ],
  args: { message: 'Agreement status updated.' },
} satisfies Meta<typeof Demo>;

export default meta;

export const Default: StoryObj<typeof meta> = {};
