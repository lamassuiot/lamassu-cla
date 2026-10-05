export const agreementTypes = ['ICLA', 'CCLA'] as const;

export type AgreementType = (typeof agreementTypes)[number];
