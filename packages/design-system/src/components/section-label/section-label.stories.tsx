import type { Meta, StoryObj } from '@storybook/react-vite';
import { SectionLabel } from './section-label';

const meta = {
  component: SectionLabel,
  args: { children: 'Contributor agreements' },
} satisfies Meta<typeof SectionLabel>;

export default meta;

export const Default: StoryObj<typeof meta> = {};
