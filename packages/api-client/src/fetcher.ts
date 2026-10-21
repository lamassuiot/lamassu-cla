// Fetch wrapper for the generated client (decision A-22, api/conventions.md sections 4, 7, and 8).
// Requests stay on the portal's origin so the `__Host-` session cookie is never sent elsewhere.

const unsafeMethods = new Set(['POST', 'PUT', 'PATCH', 'DELETE']);

export type CsrfTokenProvider = () => string | undefined;

let csrfTokenProvider: CsrfTokenProvider = () => undefined;

// How the portal obtains the session CSRF token depends on the session model (D-06).
export function setCsrfTokenProvider(provider: CsrfTokenProvider): void {
  csrfTokenProvider = provider;
}

export class ApiRequestError extends Error {
  override name = 'ApiRequestError';
}

function sameOriginPath(url: string): string {
  if (!url.startsWith('/') || url.startsWith('//')) {
    throw new ApiRequestError(`API requests must use a same-origin path, got "${url}".`);
  }
  return url;
}

async function readBody(response: Response): Promise<unknown> {
  if ([204, 205, 304].includes(response.status)) return undefined;
  const text = await response.text();
  if (text === '') return undefined;
  const type = response.headers.get('content-type') ?? '';
  if (type.includes('application/json') || type.includes('application/problem+json')) {
    return JSON.parse(text);
  }
  throw new ApiRequestError(`Unexpected content type "${type}" (status ${response.status}).`);
}

// Returns the `{ data, status, headers }` shape that the Orval `fetch` client declares.
export async function apiFetch<T>(url: string, init: RequestInit = {}): Promise<T> {
  const method = (init.method ?? 'GET').toUpperCase();
  const headers = new Headers(init.headers);
  headers.set('Accept', 'application/json, application/problem+json');
  if (unsafeMethods.has(method)) {
    const token = csrfTokenProvider();
    if (!token) {
      throw new ApiRequestError(`${method} ${url} needs a CSRF token, and none is available.`);
    }
    headers.set('X-CSRF-Token', token);
  }
  const response = await fetch(sameOriginPath(url), {
    ...init,
    method,
    headers,
    credentials: 'same-origin',
    redirect: 'error',
  });
  const data = await readBody(response);
  return { data, status: response.status, headers: response.headers } as T;
}
