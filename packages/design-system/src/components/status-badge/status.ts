// Domain-state to status-color map (design tokens, section 2.3; F-05 AC-05-6).
export const statuses = ['success', 'warning', 'danger', 'info', 'neutral'] as const;
export type Status = (typeof statuses)[number];

export const domainStateStatus = {
  ACTIVE: 'success',
  PASSED: 'success',
  PUBLISHED: 'success',
  PENDING_SIGNATURE: 'warning',
  VERIFYING: 'warning',
  PENDING_REVIEW: 'warning',
  ACTION_REQUIRED: 'warning',
  FAILED: 'danger',
  REVOKED: 'danger',
  REJECTED: 'danger',
  DECLINED: 'danger',
  DRAFT: 'info',
  SUSPENDED: 'neutral',
  EXPIRED: 'neutral',
  VOIDED: 'neutral',
  DEPRECATED: 'neutral',
  ARCHIVED: 'neutral',
  SUPERSEDED: 'neutral',
  SKIPPED: 'neutral',
} as const satisfies Record<string, Status>;

export type DomainState = keyof typeof domainStateStatus;

export function statusForState(state: string): Status {
  return Object.hasOwn(domainStateStatus, state)
    ? domainStateStatus[state as DomainState]
    : 'neutral';
}
