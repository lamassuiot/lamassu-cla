import type { Meta, StoryObj } from '@storybook/react-vite';
import { Button } from '../button/button';
import { PageHeader } from './page-header';

const meta = {
  component: PageHeader,
  args: {
    title: 'Agreements',
    description: 'Individual and corporate agreements that cover your contributions.',
  },
} satisfies Meta<typeof PageHeader>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const WithActions: Story = { args: { actions: <Button>Sign the ICLA</Button> } };
export const TitleOnly: Story = {
  args: { title: 'Audit log' },
  render: ({ title }) => <PageHeader title={title} />,
};
