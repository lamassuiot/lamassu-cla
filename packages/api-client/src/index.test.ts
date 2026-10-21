import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiRequestError, apiFetch, setCsrfTokenProvider } from './fetcher';
import { getHealth } from './index';

type FetchCall = [string, RequestInit];

function respond(status: number, body?: unknown, type = 'application/json') {
  const init = { status, headers: { 'content-type': type } };
  return vi.fn(async () =>
    body === undefined ? new Response(null, init) : new Response(JSON.stringify(body), init),
  );
}

function lastCall(fetchMock: ReturnType<typeof respond>): FetchCall {
  return fetchMock.mock.calls.at(-1) as unknown as FetchCall;
}

beforeEach(() => setCsrfTokenProvider(() => undefined));
afterEach(() => vi.unstubAllGlobals());

describe('generated client (AC-02-7)', () => {
  it('calls GET /v1/health through the wrapper and returns typed data', async () => {
    const fetchMock = respond(200, { status: 'OK' });
    vi.stubGlobal('fetch', fetchMock);
    const result = await getHealth();
    expect(result.status).toBe(200);
    if (result.status === 200) expect(result.data.status).toBe('OK');
    const [url, init] = lastCall(fetchMock);
    expect(url).toBe('/v1/health');
    expect(init.method).toBe('GET');
  });

  it('returns Problem Details for error responses', async () => {
    const problem = {
      type: 'about:blank',
      title: 'Unavailable',
      status: 503,
      code: 'SERVICE_UNAVAILABLE',
      request_id: '01J9Z3K4Q8X7M2N5P6R7S8T9V0',
    };
    vi.stubGlobal('fetch', respond(503, problem, 'application/problem+json'));
    const result = await getHealth();
    expect(result.status).toBe(503);
    if (result.status === 503) expect(result.data.code).toBe('SERVICE_UNAVAILABLE');
  });
});

describe('fetch wrapper (api/conventions.md sections 4, 7, and 8)', () => {
  it('keeps the session cookie on the same origin and refuses redirects', async () => {
    const fetchMock = respond(200, {});
    vi.stubGlobal('fetch', fetchMock);
    await apiFetch('/v1/health', { credentials: 'include', redirect: 'follow' });
    const [, init] = lastCall(fetchMock);
    expect(init.credentials).toBe('same-origin');
    expect(init.redirect).toBe('error');
    expect(new Headers(init.headers).get('Accept')).toBe(
      'application/json, application/problem+json',
    );
  });

  it.each(['https://api.example.org/v1/health', '//evil.example/v1', 'v1/health'])(
    'rejects the non-same-origin path %s',
    async (url) => {
      const fetchMock = respond(200, {});
      vi.stubGlobal('fetch', fetchMock);
      await expect(apiFetch(url)).rejects.toBeInstanceOf(ApiRequestError);
      expect(fetchMock).not.toHaveBeenCalled();
    },
  );

  it('does not send a CSRF token on safe methods', async () => {
    setCsrfTokenProvider(() => 'token-1');
    const fetchMock = respond(200, {});
    vi.stubGlobal('fetch', fetchMock);
    await apiFetch('/v1/health');
    expect(new Headers(lastCall(fetchMock)[1].headers).has('X-CSRF-Token')).toBe(false);
  });

  it.each(['POST', 'put', 'PATCH', 'DELETE'])('sends X-CSRF-Token on %s', async (method) => {
    setCsrfTokenProvider(() => 'token-1');
    const fetchMock = respond(204);
    vi.stubGlobal('fetch', fetchMock);
    const result = await apiFetch<{ status: number; data: unknown }>('/v1/things', { method });
    const [, init] = lastCall(fetchMock);
    expect(init.method).toBe(method.toUpperCase());
    expect(new Headers(init.headers).get('X-CSRF-Token')).toBe('token-1');
    expect(result).toMatchObject({ status: 204, data: undefined });
  });

  it('refuses a state-changing request without a CSRF token', async () => {
    const fetchMock = respond(201, {});
    vi.stubGlobal('fetch', fetchMock);
    await expect(apiFetch('/v1/things', { method: 'POST' })).rejects.toThrow(/CSRF token/);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it('keeps caller headers such as Idempotency-Key', async () => {
    setCsrfTokenProvider(() => 'token-1');
    const fetchMock = respond(201, {});
    vi.stubGlobal('fetch', fetchMock);
    await apiFetch('/v1/things', { method: 'POST', headers: { 'Idempotency-Key': 'key-1' } });
    expect(new Headers(lastCall(fetchMock)[1].headers).get('Idempotency-Key')).toBe('key-1');
  });

  it('rejects a response that is neither JSON nor Problem Details', async () => {
    vi.stubGlobal('fetch', respond(200, '<html>', 'text/html'));
    await expect(apiFetch('/v1/health')).rejects.toThrow(/Unexpected content type/);
  });
});
