import { describe, expect, it } from 'vitest';
import { agreementTypes } from './index';

describe('shared-types', () => {
  it('lists the agreement types defined in the domain model', () => {
    expect(agreementTypes).toEqual(['ICLA', 'CCLA']);
  });
});
