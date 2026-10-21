import type { Meta, StoryObj } from '@storybook/react-vite';
import { Button } from '../button/button';
import { SectionLabel } from '../section-label/section-label';
import { Card } from './card';

const meta = {
  component: Card,
  args: {
    as: 'section',
    'aria-labelledby': 'card-heading',
    children: (
      <>
        <SectionLabel>Individual agreement</SectionLabel>
        <h2 id="card-heading">Individual Contributor License Agreement</h2>
        <p>Sign once to contribute to every covered repository.</p>
        <Button>Review and sign</Button>
      </>
    ),
  },
} satisfies Meta<typeof Card>;

export default meta;

export const Section: StoryObj<typeof meta> = {};
