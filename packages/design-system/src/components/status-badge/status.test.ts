import { describe, expect, it } from 'vitest';
import { domainStateStatus, statusForState } from './status.ts';

// Copied from section 2.3 of specs/ux/design-tokens.md, not derived from the code under test.
const spec: Record<string, string[]> = {
  success: ['ACTIVE', 'PASSED', 'PUBLISHED'],
  warning: ['PENDING_SIGNATURE', 'VERIFYING', 'PENDING_REVIEW', 'ACTION_REQUIRED'],
  danger: ['FAILED', 'REVOKED', 'REJECTED', 'DECLINED'],
  info: ['DRAFT'],
  neutral: ['SUSPENDED', 'EXPIRED', 'VOIDED', 'DEPRECATED', 'ARCHIVED', 'SUPERSEDED', 'SKIPPED'],
};

describe('status badge mapping (AC-05-6)', () => {
  it('maps exactly the domain states in the specification', () => {
    const expected = Object.fromEntries(
      Object.entries(spec).flatMap(([status, states]) => states.map((state) => [state, status])),
    );
    expect(domainStateStatus).toEqual(expected);
  });

  it.each(Object.entries(spec).flatMap(([status, states]) => states.map((s) => [s, status])))(
    'maps %s to %s',
    (state, status) => {
      expect(statusForState(state as string)).toBe(status);
    },
  );

  it.each(['UNKNOWN', 'active', '', 'constructor', '__proto__'])(
    'renders the unknown state %j as neutral',
    (state) => {
      expect(statusForState(state)).toBe('neutral');
    },
  );
});
