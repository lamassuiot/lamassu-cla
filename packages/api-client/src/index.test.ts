import { describe, expect, it } from 'vitest';
import { apiBasePath } from './index';

describe('api-client', () => {
  it('targets the v1 API path from the API conventions', () => {
    expect(apiBasePath).toBe('/v1');
  });
});
