import type { Meta, StoryObj } from '@storybook/react-vite';
import { AgreementStatusDisplay } from './agreement-status-display';

const meta = {
  component: AgreementStatusDisplay,
  args: {
    title: 'Individual CLA',
    state: 'ACTIVE',
    stateLabel: 'Active',
    signatureLevel: { term: 'Signature level', description: 'Electronic signature' },
    details: [
      { term: 'Version', description: '2.0 (English)' },
      { term: 'Signed', description: '15 January 2026' },
    ],
  },
} satisfies Meta<typeof AgreementStatusDisplay>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Active: Story = {};
export const PendingSignature: Story = {
  args: { state: 'PENDING_SIGNATURE', stateLabel: 'Pending signature', details: [] },
};
export const Revoked: Story = {
  args: {
    state: 'REVOKED',
    stateLabel: 'Revoked',
    details: [{ term: 'Ended', description: '1 March 2026' }],
  },
};
