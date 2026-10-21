import type { Meta, StoryObj } from '@storybook/react-vite';
import { StatusBadge } from '../status-badge/status-badge';
import { Table, type TableColumn } from './table';

interface Row {
  id: string;
  name: string;
  version: string;
  state: string;
  signers: number;
}

const columns: TableColumn<Row>[] = [
  { key: 'name', header: 'Agreement', cell: (row) => row.name, rowHeader: true },
  { key: 'version', header: 'Version', cell: (row) => row.version },
  {
    key: 'state',
    header: 'Status',
    cell: (row) => <StatusBadge state={row.state} label={row.state.toLowerCase()} />,
  },
  { key: 'signers', header: 'Signers', cell: (row) => row.signers, numeric: true },
];

const rows: Row[] = [
  { id: '1', name: 'Individual CLA', version: '2.0', state: 'PUBLISHED', signers: 128 },
  { id: '2', name: 'Corporate CLA', version: '2.0', state: 'PUBLISHED', signers: 12 },
  { id: '3', name: 'Individual CLA', version: '3.0', state: 'DRAFT', signers: 0 },
];

const meta = {
  component: Table<Row>,
  args: { caption: 'Agreement versions', columns, rows, rowKey: (row) => row.id },
} satisfies Meta<typeof Table<Row>>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Empty: Story = { args: { rows: [] } };
