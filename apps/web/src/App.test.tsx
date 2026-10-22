import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { App } from './App';

describe('App', () => {
  it('renders the landing page inside the app shell', async () => {
    render(<App />);
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Lamassu CLA' }),
    ).toBeInTheDocument();
    expect(screen.getAllByRole('banner')[0]).toBeInTheDocument();
  });
});
