const request = async (method, path, body) => {
  const response = await fetch(path, {
    method,
    credentials: 'same-origin',
    headers: body ? { 'Content-Type': 'application/json' } : undefined,
    body: body ? JSON.stringify(body) : undefined
  });

  const text = await response.text();
  const payload = text ? JSON.parse(text) : null;

  if (!response.ok) {
    const error = new Error(payload?.error ?? `request to ${path} failed`);
    error.status = response.status;
    throw error;
  }

  return payload;
};

export const get = (path) => request('GET', path);
export const post = (path, body) => request('POST', path, body ?? {});
export const put = (path, body) => request('PUT', path, body);
export const del = (path) => request('DELETE', path);

export const financial = (value) => Number.parseFloat(value ?? 0).toFixed(2);

export const formatDate = (value) => {
  if (!value) {
    return '';
  }
  return new Date(value).toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
};
