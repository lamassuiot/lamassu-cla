import { describe, expect, it } from 'vitest';
import { tokenPrefix } from './index';

describe('design-system', () => {
  it('exposes the CSS custom-property prefix from the token specification', () => {
    expect(tokenPrefix).toBe('--cla-');
  });
});
