const API_BASE = import.meta.env.VITE_API_BASE || '/api';

async function parseResponse(response) {
  const contentType = response.headers.get('content-type') || '';
  const isJSON = contentType.includes('application/json');
  const payload = isJSON ? await response.json() : null;

  if (!response.ok) {
    const message = payload?.message || `Request failed: ${response.status}`;
    throw new Error(message);
  }
  return payload;
}

export async function apiRequest(path, { method = 'GET', token, body, isForm = false } = {}) {
  const headers = {};
  if (!isForm) {
    headers['Content-Type'] = 'application/json';
  }
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: isForm ? body : body ? JSON.stringify(body) : undefined,
  });

  return parseResponse(response);
}
