import { api, onUser, restore, session, signIn, signOut } from './api.js';
import { STATUS, bookCard, confirmBox, empty, esc, myItem, pager, plural, rating, statCard, toast } from './ui.js';

const $ = (sel) => document.querySelector(sel);
const $$ = (sel) => [...document.querySelectorAll(sel)];

const catalog = { search: '', genre: 'all', sort: 'id', page: 1 };
const myList = { status: 'all', page: 1 };
const detail = { book: null, item: null };
const editor = { book: null, pendingFile: null, previewUrl: null };

function show(page) {
  $$('[data-page]').forEach((s) => { s.hidden = s.dataset.page !== page; });
  $$('[data-nav]').forEach((a) => a.classList.toggle('is-active', a.dataset.nav === page));
  window.scrollTo(0, 0);
}

function go(hash) {
  if (location.hash === hash) render();
  else location.hash = hash;
}

function fail(error) {
  toast(error.message ?? 'Что-то пошло не так', 'error');
}

/* каталог */

function openCatalog(params) {
  catalog.search = params.get('search') ?? '';
  catalog.genre = params.get('genre') ?? 'all';
  catalog.sort = params.get('sort') ?? 'id';
  catalog.page = Number(params.get('page')) || 1;

  $('#q').value = catalog.search;
  $('#sort').value = catalog.sort;
  show('/');

  loadGenres();
  loadCatalog();
}

function catalogHash() {
  const params = new URLSearchParams();
  if (catalog.search) params.set('search', catalog.search);
  if (catalog.genre !== 'all') params.set('genre', catalog.genre);
  if (catalog.sort !== 'id') params.set('sort', catalog.sort);
  if (catalog.page > 1) params.set('page', catalog.page);
  const str = params.toString();
  return str ? `#/?${str}` : '#/';
}

async function loadGenres() {
  try {
    const { genres } = await api.genres();
    $('#genre').innerHTML = '<option value="all">Все жанры</option>' +
      genres.map((g) => `<option value="${esc(g)}">${esc(g)}</option>`).join('');
    $('#genre').value = catalog.genre;
  } catch {
    /* фильтр по жанру не критичен */
  }
}

async function loadCatalog() {
  const grid = $('#catalog-grid');
  grid.innerHTML = '<p class="page-head__subtitle">Загрузка…</p>';

  try {
    const data = await api.books({ ...catalog, per_page: 12 });
    grid.innerHTML = data.items.map(bookCard).join('');
    $('#catalog-count').textContent = `${data.meta.total} ${plural(data.meta.total, 'книга', 'книги', 'книг')} в каталоге`;
    $('#catalog-pager').innerHTML = pager(data.meta, 'catalog');

    const box = $('#catalog-empty');
    box.hidden = data.items.length > 0;
    if (!data.items.length) {
      const filtered = catalog.search || catalog.genre !== 'all' || catalog.sort !== 'id';
      box.innerHTML = empty(
        '📚',
        filtered ? 'Ничего не найдено' : 'В каталоге пока нет книг',
        filtered ? 'Попробуйте изменить запрос или сбросить фильтры.' : 'Добавьте первую книгу — и она появится здесь.',
        filtered
          ? '<button class="btn btn--primary" type="button" data-reset>Сбросить фильтры</button>'
          : '<a class="btn btn--primary" href="#/books/new">Добавить книгу</a>',
      );
    }
  } catch (error) {
    grid.innerHTML = '';
    fail(error);
  }
}

/* карточка книги */

async function openBook(id) {
  show('/book');
  detail.book = null;
  detail.item = null;

  try {
    const data = await api.book(id);
    detail.book = data.book;
    renderBook();
    if (session.user) checkInList(id);
  } catch (error) {
    $('#book-media').innerHTML = '';
    $('#book-title').textContent = error.status === 404 ? 'Книга не найдена' : 'Не удалось загрузить книгу';
    $('#book-author').textContent = error.status === 404 ? 'Возможно, она была удалена из каталога.' : error.message;
    $('#book-tags').innerHTML = '';
    $('#book-facts').innerHTML = '';
    $('#book-desc').textContent = '';
  }
}

async function checkInList(id) {
  try {
    const data = await api.myBooks({ per_page: 100 });
    detail.item = data.items.find((i) => i.book_id === Number(id)) ?? null;
    renderBook();
  } catch {
    /* кнопка останется в режиме «добавить» */
  }
}

function renderBook() {
  const book = detail.book;
  if (!book) return;

  const authed = Boolean(session.user);
  const inList = Boolean(detail.item);

  $('#book-media').innerHTML = book.cover_url
    ? `<img src="${esc(book.cover_url)}" alt="Обложка: ${esc(book.title)}">`
    : `<div class="card__placeholder">${esc(book.title.slice(0, 40))}</div>`;
  $('#book-title').textContent = book.title;
  $('#book-author').textContent = book.author;
  $('#book-tags').innerHTML =
    `<span class="tag tag--accent">${esc(book.genre)}</span>` +
    (inList ? `<span class="tag tag--success">${STATUS[detail.item.status]}</span>` : '');
  $('#book-facts').innerHTML = [
    ['Год', book.year > 0 ? book.year : '—'],
    ['Страниц', book.pages > 0 ? book.pages : '—'],
    ['Рейтинг', rating(book.rating)],
  ].map(([k, v]) => `<div><div class="book__fact-label">${k}</div><div class="book__fact-value">${esc(v)}</div></div>`).join('');
  $('#book-desc').textContent = book.description || 'Описание пока не добавлено.';
  $('#book-desc').classList.toggle('book__description--empty', !book.description);

  const toggle = $('#book-toggle');
  toggle.hidden = !authed;
  toggle.textContent = inList ? 'Убрать из списка' : 'Добавить в список';
  toggle.className = inList ? 'btn btn--ghost' : 'btn btn--primary';

  $('#book-status-row').hidden = !inList;
  if (inList) $('#book-status').value = detail.item.status;

  $('#book-edit').hidden = !authed;
  $('#book-edit').href = `#/books/${book.id}/edit`;
  $('#book-delete').hidden = !authed;
}

/* форма книги */

function showCover(url, canDelete = true) {
  $('#cover-box').hidden = false;
  $('#cover-hint').hidden = true;
  $('#cover-img').hidden = !url;
  $('#cover-del').hidden = !url || !canDelete;
  if (url) {
    $('#cover-img').src = url;
  } else {
    $('#cover-img').removeAttribute('src');
  }
}

function clearPendingCover() {
  if (editor.previewUrl) {
    URL.revokeObjectURL(editor.previewUrl);
  }
  editor.previewUrl = null;
  editor.pendingFile = null;
}

async function openForm(id) {
  show('/form');
  editor.book = null;
  clearPendingCover();
  $('#cover-file').value = '';
  $('#form-error').hidden = true;
  $('#genre-list').innerHTML = '';
  showCover('', false);
  for (const field of ['f-title', 'f-author', 'f-genre', 'f-year', 'f-pages', 'f-rating', 'f-desc']) {
    $('#' + field).value = '';
  }

  api.genres()
    .then(({ genres }) => {
      $('#genre-list').innerHTML = genres.map((g) => `<option value="${esc(g)}">`).join('');
    })
    .catch(() => {});

  if (!id) {
    $('#form-title').textContent = 'Новая книга';
    $('#form-submit').textContent = 'Добавить в каталог';
    return;
  }

  try {
    const { book } = await api.book(id);
    editor.book = book;
    $('#form-title').textContent = `Редактирование: ${book.title}`;
    $('#form-submit').textContent = 'Сохранить';
    $('#f-title').value = book.title;
    $('#f-author').value = book.author;
    $('#f-genre').value = book.genre;
    $('#f-year').value = book.year;
    $('#f-pages').value = book.pages;
    $('#f-rating').value = book.rating;
    $('#f-desc').value = book.description;
    showCover(book.cover_url, true);
  } catch (error) {
    fail(error);
  }
}

async function submitForm(event) {
  event.preventDefault();
  const id = location.hash.match(/\/books\/(\d+)\/edit/)?.[1];
  const payload = {
    title: $('#f-title').value.trim(),
    author: $('#f-author').value.trim(),
    genre: $('#f-genre').value.trim(),
    year: Number($('#f-year').value) || 0,
    pages: Number($('#f-pages').value) || 0,
    rating: Number($('#f-rating').value) || 0,
    description: $('#f-desc').value.trim(),
  };

  const error = $('#form-error');
  if (!payload.title || !payload.author || !payload.genre) {
    error.textContent = 'Заполните название, автора и жанр.';
    error.hidden = false;
    return;
  }

  const submit = $('#form-submit');
  const label = submit.textContent;
  submit.disabled = true;
  submit.textContent = id ? 'Сохраняем…' : 'Добавляем…';

  try {
    const res = id ? await api.updateBook(id, payload) : await api.createBook(payload);
    let book = res.book;

    if (editor.pendingFile) {
      try {
        book = (await api.uploadCover(book.id, editor.pendingFile)).book;
        toast('Книга и обложка добавлены в каталог', 'success');
      } catch (coverError) {
        toast(`Книга добавлена, но обложка не загрузилась: ${coverError.message}`, 'warn');
      }
      clearPendingCover();
    } else {
      toast(id ? 'Изменения сохранены' : 'Книга добавлена в каталог', 'success');
    }

    go(`#/books/${book.id}`);
  } catch (err) {
    error.textContent = err.message;
    error.hidden = false;
    submit.disabled = false;
    submit.textContent = label;
  }
}

/* мои книги */

async function openMy(params) {
  myList.status = params.get('status') ?? 'all';
  myList.page = Number(params.get('page')) || 1;
  show('/my');
  await loadMy();
}

function myHash() {
  const params = new URLSearchParams();
  if (myList.status !== 'all') params.set('status', myList.status);
  if (myList.page > 1) params.set('page', myList.page);
  const str = params.toString();
  return str ? `#/my?${str}` : '#/my';
}

async function loadMy() {
  const tabs = ['all', ...Object.keys(STATUS)];
  $('#my-tabs').innerHTML = tabs.map((value) => {
    const label = value === 'all' ? 'Все' : STATUS[value];
    const active = value === myList.status;
    return `<button class="btn btn--sm ${active ? 'btn--primary' : 'btn--ghost'}" type="button" data-tab="${value}">${label}</button>`;
  }).join('');

  try {
    const data = await api.myBooks({ ...myList, per_page: 12 });
    $('#my-stats').innerHTML = [
      statCard(data.stats.total, 'Всего книг'),
      statCard(data.stats.want_to_read, STATUS.want_to_read),
      statCard(data.stats.reading, STATUS.reading),
      statCard(data.stats.finished, STATUS.finished),
    ].join('');
    $('#my-pager').innerHTML = pager(data.meta, 'my');

    const list = $('#my-list');
    list.innerHTML = data.items.length
      ? data.items.map(myItem).join('')
      : empty(
          '📖',
          myList.status === 'all' ? 'Список пока пуст' : `Нет книг со статусом «${STATUS[myList.status]}»`,
          myList.status === 'all'
            ? 'Добавляйте книги из каталога и отмечайте прочитанное.'
            : 'Смените фильтр, чтобы увидеть остальные книги.',
          myList.status === 'all' ? '<a class="btn btn--primary" href="#/">Перейти в каталог</a>' : '',
        );
  } catch (error) {
    fail(error);
  }
}

/* авторизация */

function openAuth(kind) {
  show(`/${kind}`);
}

async function submitAuth(event, kind) {
  event.preventDefault();
  const box = $(`#${kind}-error`);
  box.hidden = true;

  const data = kind === 'login'
    ? { login: $('#l-login').value.trim(), password: $('#l-pass').value }
    : { username: $('#r-user').value.trim(), email: $('#r-mail').value.trim(), password: $('#r-pass').value };

  try {
    await signIn(kind, data);
    toast(kind === 'login' ? 'Вы вошли в аккаунт' : 'Аккаунт создан', 'success');
    go('#/');
  } catch (error) {
    box.textContent = error.message;
    box.hidden = false;
  }
}

function renderHeader(user) {
  $$('[data-auth]').forEach((node) => { node.hidden = !user; });
  $('#auth-actions').innerHTML = user
    ? `<span class="topbar__user">${esc(user.username)}</span>
       <button class="btn btn--ghost btn--sm" type="button" data-signout>Выйти</button>`
    : `<a class="btn btn--ghost btn--sm" href="#/login">Войти</a>
       <a class="btn btn--primary btn--sm" href="#/register">Регистрация</a>`;
}

/* роутер */

async function render() {
  const raw = location.hash.slice(1) || '/';
  const [path, query] = raw.split('?');
  const params = new URLSearchParams(query ?? '');
  const authed = Boolean(session.user);

  const edit = path.match(/^\/books\/(\d+)\/edit$/);
  const one = path.match(/^\/books\/(\d+)$/);

  if (path === '/') return openCatalog(params);
  if (path === '/my') return authed ? openMy(params) : go('#/login');
  if (path === '/books/new') return authed ? openForm(null) : go('#/login');
  if (edit) return authed ? openForm(edit[1]) : go('#/login');
  if (one) return openBook(one[1]);
  if (path === '/login' || path === '/register') return authed ? go('#/my') : openAuth(path.slice(1));

  go('#/');
}

/* события */

function bind() {
  let timer;
  $('#q').addEventListener('input', (e) => {
    const value = e.target.value;
    clearTimeout(timer);
    timer = setTimeout(() => {
      catalog.search = value.trim();
      catalog.page = 1;
      history.replaceState(null, '', catalogHash());
      loadCatalog();
    }, 320);
  });

  $('#genre').addEventListener('change', (e) => {
    catalog.genre = e.target.value;
    catalog.page = 1;
    history.replaceState(null, '', catalogHash());
    loadCatalog();
  });

  $('#sort').addEventListener('change', (e) => {
    catalog.sort = e.target.value;
    catalog.page = 1;
    history.replaceState(null, '', catalogHash());
    loadCatalog();
  });

  $('#book-form').addEventListener('submit', submitForm);

  $('#cover-file').addEventListener('change', async (e) => {
    const input = e.target;
    const file = input.files?.[0];
    if (!file) return;

    if (!editor.book) {
      clearPendingCover();
      editor.pendingFile = file;
      editor.previewUrl = URL.createObjectURL(file);
      showCover(editor.previewUrl, true);
      toast('Обложка прикрепится к книге после сохранения', 'info');
      return;
    }

    input.disabled = true;
    try {
      const { book } = await api.uploadCover(editor.book.id, file);
      editor.book = book;
      showCover(book.cover_url, true);
      toast('Обложка загружена', 'success');
    } catch (error) {
      fail(error);
    } finally {
      input.disabled = false;
      input.value = '';
    }
  });

  $('#cover-del').addEventListener('click', async () => {
    if (!editor.book) {
      clearPendingCover();
      $('#cover-file').value = '';
      showCover('', false);
      toast('Обложка убрана', 'info');
      return;
    }
    if (!await confirmBox('Файл обложки будет удалён с диска.', 'Удалить обложку?')) return;
    try {
      await api.deleteCover(editor.book.id);
      showCover('');
      toast('Обложка удалена', 'info');
    } catch (error) {
      fail(error);
    }
  });

  $('#book-toggle').addEventListener('click', async () => {
    try {
      if (detail.item) {
        await api.removeItem(detail.item.id);
        detail.item = null;
        toast('Книга убрана из списка', 'info');
      } else {
        const { item } = await api.addToList(detail.book.id);
        detail.item = item;
        toast(`«${detail.book.title}» добавлена в ваш список`, 'success');
      }
      renderBook();
    } catch (error) {
      fail(error);
    }
  });

  $('#book-status').addEventListener('change', async (e) => {
    try {
      const { item } = await api.setStatus(detail.item.id, e.target.value);
      detail.item = item;
      renderBook();
      toast(`Статус изменён: ${STATUS[item.status]}`, 'success');
    } catch (error) {
      fail(error);
    }
  });

  $('#book-delete').addEventListener('click', async () => {
    if (!await confirmBox(`«${detail.book.title}» будет удалена из каталога. Действие необратимо.`, 'Удалить книгу?')) return;
    try {
      await api.deleteBook(detail.book.id);
      toast('Книга удалена', 'info');
      go('#/');
    } catch (error) {
      fail(error);
    }
  });

  $('#login-form').addEventListener('submit', (e) => submitAuth(e, 'login'));
  $('#register-form').addEventListener('submit', (e) => submitAuth(e, 'register'));

  document.addEventListener('click', async (e) => {
    const target = e.target;

    if (target.closest('[data-signout]')) {
      signOut();
      toast('Вы вышли из аккаунта', 'info');
      return go('#/');
    }

    const page = target.closest('[data-scope]');
    if (page && !page.disabled) {
      const n = Number(page.dataset.n);
      if (page.dataset.scope === 'catalog') {
        catalog.page = n;
        history.replaceState(null, '', catalogHash());
        loadCatalog();
      } else {
        myList.page = n;
        history.replaceState(null, '', myHash());
        loadMy();
      }
      return;
    }

    const add = target.closest('[data-add]');
    if (add) {
      try {
        await api.addToList(add.dataset.add);
        add.disabled = true;
        add.textContent = 'В списке';
        toast('Книга добавлена в ваш список', 'success');
      } catch (error) {
        fail(error);
      }
      return;
    }

    const tab = target.closest('[data-tab]');
    if (tab) {
      myList.status = tab.dataset.tab;
      myList.page = 1;
      history.replaceState(null, '', myHash());
      loadMy();
      return;
    }

    const remove = target.closest('[data-remove]');
    if (remove && await confirmBox('Книга останется в каталоге.', 'Убрать из списка?')) {
      try {
        await api.removeItem(remove.dataset.remove);
        toast('Книга убрана из списка', 'info');
        loadMy();
      } catch (error) {
        fail(error);
      }
      return;
    }

    if (target.closest('[data-reset]')) {
      catalog.search = '';
      catalog.genre = 'all';
      catalog.sort = 'id';
      catalog.page = 1;
      $('#q').value = '';
      $('#sort').value = 'id';
      history.replaceState(null, '', '#/');
      loadGenres();
      loadCatalog();
    }
  });

  document.addEventListener('change', async (e) => {
    const select = e.target.closest('[data-status-for]');
    if (!select) return;
    select.disabled = true;
    try {
      await api.setStatus(select.dataset.statusFor, select.value);
      toast(`Статус изменён: ${STATUS[select.value]}`, 'success');
      loadMy();
    } catch (error) {
      fail(error);
      select.disabled = false;
    }
  });
}

onUser(renderHeader);
bind();
addEventListener('hashchange', render);
await restore();
render();
