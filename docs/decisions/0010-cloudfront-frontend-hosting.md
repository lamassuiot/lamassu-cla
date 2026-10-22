# ADR-0010: CloudFront frontend hosting

- Status: Accepted
- Date: 2026-10-05
- Decision-log entry: A-12; related open decision D-16

## Context

The portal is a static single-page application built with Vite. It must be served over HTTPS with
security headers, without making any S3 bucket public.

## Decision

- The built portal is stored in a private, versioned, KMS-encrypted S3 bucket with S3 Block Public
  Access.
- Amazon CloudFront serves it, using **Origin Access Control** so only the distribution can read
  the bucket.
- A CloudFront response-headers policy sets the Content Security Policy, HSTS,
  `X-Content-Type-Options`, `Referrer-Policy`, and frame restrictions defined in the
  [security specification](../../specs/security/security.md#5-input-validation).
- Client-side routes fall back to `index.html`. Hashed assets are cached long-term; `index.html` is
  not cached.
- Custom domains, certificates, and WAF for the distribution are decided in D-16.

## Consequences

- Deployments upload the build and invalidate `index.html`.
- CloudFront access logs must not capture query strings that could carry personal data.
- Serving the portal and API from the same site (through a CloudFront behavior for `/v1/*`) is
  preferred so the session cookie can stay host-only; the exact routing is defined in the Phase 2
  infrastructure feature specification.
- Update 2026-10-22 (D-02): the API behavior matches `/api/*`, not `/v1/*`. The versioned API is
  under `/api/v1`, and the GitHub sign-in callback is `/api/auth/github/callback`, outside the
  versioned routes. The portal origin is `https://cla.developers.lamassu.cloud`; the domain is not
  provisioned, and D-02 does not authorize deployment.

## Alternatives considered

- **S3 static website hosting.** Rejected: requires public bucket access and has no HTTPS.
- **AWS Amplify Hosting.** Rejected: less control over headers and infrastructure in Terraform.
