const TOKEN_KEY = 'bc_token';

export const session = { user: null };

const listeners = new Set();

export function onUser(fn) {
  listeners.add(fn);
  fn(session.user);
}

function emit(user) {
  session.user = user;
  for (const fn of listeners) fn(user);
}

function query(params) {
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value !== '' && value !== null && value !== undefined) search.set(key, String(value));
  }
  const str = search.toString();
  return str ? `?${str}` : '';
}

async function call(path, { method = 'GET', body, form } = {}) {
  const headers = {};
  const token = localStorage.getItem(TOKEN_KEY);
  if (token) headers.Authorization = `Bearer ${token}`;

  let payload;
  if (form) {
    payload = form;
  } else if (body) {
    headers['Content-Type'] = 'application/json';
    payload = JSON.stringify(body);
  }

  let res;
  try {
    res = await fetch(path, { method, headers, body: payload });
  } catch {
    throw new Error('Сервер недоступен. Проверьте соединение.');
  }

  if (res.status === 204) return null;
  const data = await res.json().catch(() => null);

  if (!res.ok) {
    if (res.status === 401) {
      localStorage.removeItem(TOKEN_KEY);
      emit(null);
    }
    const error = new Error(data?.error?.message ?? 'Не удалось выполнить запрос');
    error.status = res.status;
    throw error;
  }
  return data;
}

export const api = {
  books: (params) => call('/api/books' + query(params)),
  genres: () => call('/api/books/genres'),
  book: (id) => call(`/api/books/${id}`),
  createBook: (book) => call('/api/books', { method: 'POST', body: book }),
  updateBook: (id, book) => call(`/api/books/${id}`, { method: 'PUT', body: book }),
  deleteBook: (id) => call(`/api/books/${id}`, { method: 'DELETE' }),
  uploadCover: (id, file) => {
    const form = new FormData();
    form.append('cover', file);
    return call(`/api/books/${id}/cover`, { method: 'POST', form });
  },
  deleteCover: (id) => call(`/api/books/${id}/cover`, { method: 'DELETE' }),

  myBooks: (params) => call('/api/my/books' + query(params)),
  addToList: (bookId) => call(`/api/my/books/${bookId}`, { method: 'POST' }),
  setStatus: (id, status) => call(`/api/my/books/${id}`, { method: 'PATCH', body: { status } }),
  removeItem: (id) => call(`/api/my/books/${id}`, { method: 'DELETE' }),

  login: (login, password) => call('/api/auth/login', { method: 'POST', body: { login, password } }),
  register: (data) => call('/api/auth/register', { method: 'POST', body: data }),
  me: () => call('/api/auth/me'),
};

export async function signIn(kind, data) {
  const res = kind === 'login' ? await api.login(data.login, data.password) : await api.register(data);
  localStorage.setItem(TOKEN_KEY, res.token);
  emit(res.user ?? null);
}

export async function restore() {
  if (!localStorage.getItem(TOKEN_KEY)) return emit(null);
  try {
    const data = await api.me();
    emit(data.user ?? null);
  } catch {
    localStorage.removeItem(TOKEN_KEY);
    emit(null);
  }
}

export function signOut() {
  localStorage.removeItem(TOKEN_KEY);
  emit(null);
}
