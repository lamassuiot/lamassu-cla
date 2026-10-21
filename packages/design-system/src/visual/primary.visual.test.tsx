import '../styles/tokens.css';
import '../styles/base.css';
import { cleanup, render } from '@testing-library/react';
import type { ReactElement } from 'react';
import { afterEach, expect, test } from 'vitest';
import { page } from 'vitest/browser';
import { Button } from '../components/button/button';
import { Card } from '../components/card/card';
import { Dialog } from '../components/dialog/dialog';
import { FormField } from '../components/form-field/form-field';
import { Input } from '../components/input/input';
import { StatusBadge } from '../components/status-badge/status-badge';
import { Table } from '../components/table/table';
import { applyTheme } from '../theme';
import styles from './visual.module.css';

// F05-T10 and F05-T11: the primary set in both themes, at a wide window, at 200% zoom of a 1280 px
// window (640 px), and at the 320 px reflow width. Baselines exist only for the pinned image.
afterEach(cleanup);

const rows = [
  { id: '1', name: 'Individual CLA', version: '2.0', state: 'PUBLISHED' },
  { id: '2', name: 'Corporate CLA', version: '2.0', state: 'DRAFT' },
];

const fixtures: Record<string, () => ReactElement> = {
  button: () => (
    <div className={styles.row}>
      <Button>Sign</Button>
      <Button variant="secondary">Cancel</Button>
      <Button variant="ghost">Details</Button>
      <Button variant="danger">Revoke</Button>
      <Button disabled>Unavailable</Button>
    </div>
  ),
  'form-field': () => (
    <>
      <FormField label="Organization name" hint="As registered.">
        {(control) => <Input {...control} defaultValue="Example Org" />}
      </FormField>
      <FormField label="Email" error="Enter a valid email address.">
        {(control) => <Input {...control} type="email" defaultValue="not-an-email" />}
      </FormField>
    </>
  ),
  card: () => (
    <Card>
      <p>Your individual agreement covers your contributions.</p>
    </Card>
  ),
  'status-badge': () => (
    <div className={styles.row}>
      <StatusBadge state="ACTIVE" label="Active" />
      <StatusBadge state="PENDING_SIGNATURE" label="Pending signature" />
      <StatusBadge state="REVOKED" label="Revoked" />
      <StatusBadge state="DRAFT" label="Draft" />
      <StatusBadge state="EXPIRED" label="Expired" />
    </div>
  ),
  table: () => (
    <Table
      caption="Agreement versions"
      rows={rows}
      rowKey={(row) => row.id}
      columns={[
        { key: 'name', header: 'Agreement', cell: (row) => row.name, rowHeader: true },
        { key: 'version', header: 'Version', cell: (row) => row.version },
        {
          key: 'state',
          header: 'Status',
          cell: (row) => <StatusBadge state={row.state} label={row.state.toLowerCase()} />,
        },
      ]}
    />
  ),
  dialog: () => (
    <Dialog
      open
      onClose={() => undefined}
      title="Publish version 2.0?"
      description="Published versions cannot be changed."
      actions={
        <>
          <Button variant="secondary">Cancel</Button>
          <Button>Publish</Button>
        </>
      }
    />
  ),
};

for (const theme of ['light', 'dark'] as const) {
  for (const width of [1280, 640, 320]) {
    for (const [name, fixture] of Object.entries(fixtures)) {
      test(`${name} ${theme} ${width}`, async () => {
        applyTheme(theme);
        await page.viewport(width, 900);
        render(
          <div className={styles.fixture} data-testid="fixture">
            {fixture()}
          </div>,
        );
        const target = name === 'dialog' ? page.getByRole('dialog') : page.getByTestId('fixture');
        await expect.element(target).toMatchScreenshot(`${name}-${theme}-${width}`);
      });
    }
  }
}
