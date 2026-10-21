import type { Meta, StoryObj } from '@storybook/react-vite';
import { Timeline, TimelineItem } from './timeline-item';

const meta = {
  component: TimelineItem,
  decorators: [
    (Story) => (
      <Timeline label="Agreement history">
        <Story />
        <TimelineItem
          title="Signing started"
          dateTime="2026-01-15T09:30:00Z"
          timeLabel="15 January 2026, 09:30 UTC"
        />
      </Timeline>
    ),
  ],
  args: {
    title: 'Agreement activated',
    dateTime: '2026-01-15T10:02:00Z',
    timeLabel: '15 January 2026, 10:02 UTC',
    status: 'success',
  },
} satisfies Meta<typeof TimelineItem>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Success: Story = {};
export const WithDescription: Story = {
  args: {
    title: 'Verification failed',
    status: 'danger',
    description: 'The delivered signature level is below the required level.',
  },
};
