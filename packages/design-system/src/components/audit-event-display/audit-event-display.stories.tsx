import type { Meta, StoryObj } from '@storybook/react-vite';
import { AuditEventDisplay } from './audit-event-display';

const meta = {
  component: AuditEventDisplay,
  args: {
    title: 'Agreement activated',
    eventType: 'agreement.activated',
    dateTime: '2026-01-15T10:02:00Z',
    timeLabel: '15 January 2026, 10:02 UTC',
    fields: [
      { term: 'Actor', description: 'System' },
      { term: 'Entity', description: 'Agreement 01HZXAMPLE0000000000000000' },
      { term: 'Request', description: 'req-example-0001' },
    ],
  },
} satisfies Meta<typeof AuditEventDisplay>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const WithHash: Story = {
  args: {
    title: 'Document version published',
    eventType: 'document_version.published',
    fields: [
      { term: 'Actor', description: 'Legal administrator' },
      {
        term: 'Document hash',
        // SHA-256 of empty input; synthetic.
        description: <code>e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855</code>,
      },
    ],
  },
};
