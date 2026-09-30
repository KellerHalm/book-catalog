export const STATUS = {
  want_to_read: 'Хочу прочитать',
  reading: 'Читаю',
  finished: 'Прочитано',
};

export const esc = (value) =>
  String(value ?? '').replace(/[&<>"']/g, (c) =>
    ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));

export function plural(n, one, few, many) {
  const m10 = n % 10;
  const m100 = n % 100;
  if (m10 === 1 && m100 !== 11) return one;
  if (m10 >= 2 && m10 <= 4 && (m100 < 12 || m100 > 14)) return few;
  return many;
}

export function rating(value) {
  return value > 0 ? `★ ${Number(value).toFixed(1)}` : '—';
}

export function cover(book, extra = '') {
  return book.cover_url
    ? `<img class="cover ${extra}" src="${esc(book.cover_url)}" alt="" loading="lazy">`
    : `<div class="card__placeholder ${extra}">${esc((book.title ?? '').slice(0, 40))}</div>`;
}

export function bookCard(book) {
  const facts = [
    book.year > 0 ? book.year : null,
    book.pages > 0 ? `${book.pages} стр.` : null,
    book.rating > 0 ? rating(book.rating) : null,
  ].filter(Boolean).map((t) => `<span>${esc(t)}</span>`).join('');

  return `<article class="card">
    <a class="cover-link" href="#/books/${book.id}" tabindex="-1" aria-hidden="true">${cover(book)}</a>
    <div class="card__body">
      <h3 class="card__title"><a href="#/books/${book.id}">${esc(book.title)}</a></h3>
      <p class="card__author">${esc(book.author)}</p>
      <div class="card__facts">${facts}</div>
      <div class="card__meta"><span class="tag">${esc(book.genre)}</span></div>
    </div>
    <div class="card__actions">
      <a class="btn btn--sm btn--ghost" href="#/books/${book.id}">Подробнее</a>
      <button class="btn btn--sm" type="button" data-add="${book.id}">В список</button>
    </div>
  </article>`;
}

export function myItem(item) {
  const options = Object.entries(STATUS)
    .map(([value, label]) => `<option value="${value}"${value === item.status ? ' selected' : ''}>${label}</option>`)
    .join('');

  return `<article class="panel row row--between">
    <div class="row">
      <a href="#/books/${item.book_id}">${cover(item.book, 'thumb')}</a>
      <div>
        <h3 class="card__title"><a href="#/books/${item.book_id}">${esc(item.book.title)}</a></h3>
        <p class="card__author">${esc(item.book.author)}</p>
      </div>
    </div>
    <div class="row">
      <select class="select" data-status-for="${item.id}">${options}</select>
      <button class="btn btn--sm btn--danger" type="button" data-remove="${item.id}">Убрать</button>
    </div>
  </article>`;
}

export function statCard(value, label) {
  return `<div class="stat"><div class="stat__value">${Number(value ?? 0)}</div><div class="stat__label">${esc(label)}</div></div>`;
}

export function pager(meta, scope) {
  if (!meta || meta.total_pages < 2) return '';
  const { page, total_pages: total } = meta;
  const btn = (n, label, disabled, active) =>
    `<button class="pager__btn${active ? ' is-active' : ''}" type="button" data-scope="${scope}" data-n="${n}"${disabled ? ' disabled' : ''}>${label}</button>`;

  let out = btn(page - 1, '←', page <= 1, false);
  let prev = 0;

  for (const n of [...new Set([1, total, page - 1, page, page + 1])].filter((x) => x >= 1 && x <= total).sort((a, b) => a - b)) {
    if (prev && n - prev > 1) out += '<span class="pager__gap">…</span>';
    out += btn(n, n, false, n === page);
    prev = n;
  }

  return out + btn(page + 1, '→', page >= total, false);
}

export function empty(icon, title, text, action = '') {
  return `<div class="empty">
    <div class="empty__icon">${icon}</div>
    <h2 class="empty__title">${esc(title)}</h2>
    <p class="empty__text">${esc(text)}</p>
    ${action}
  </div>`;
}

export function toast(message, type = 'info') {
  const host = document.getElementById('toasts');
  const node = document.createElement('div');
  node.className = `toast toast--${type}`;
  node.innerHTML = '<p class="toast__text"></p><button class="toast__close" type="button" aria-label="Закрыть">×</button>';
  node.querySelector('.toast__text').textContent = message;
  host.append(node);

  const close = () => node.remove();
  node.querySelector('.toast__close').onclick = close;
  setTimeout(close, 4500);
}

export function confirmBox(text, title = 'Подтвердите действие') {
  return new Promise((resolve) => {
    const overlay = document.createElement('div');
    overlay.className = 'modal';
    overlay.innerHTML = `<div class="modal__box">
      <h2 class="modal__title"></h2>
      <p class="modal__text"></p>
      <div class="modal__actions">
        <button class="btn btn--ghost" type="button" data-no>Отмена</button>
        <button class="btn btn--danger" type="button" data-yes>Подтвердить</button>
      </div>
    </div>`;
    overlay.querySelector('.modal__title').textContent = title;
    overlay.querySelector('.modal__text').textContent = text;

    const done = (value) => {
      overlay.remove();
      document.removeEventListener('keydown', onKey);
      resolve(value);
    };
    const onKey = (e) => e.key === 'Escape' && done(false);

    overlay.querySelector('[data-no]').onclick = () => done(false);
    overlay.querySelector('[data-yes]').onclick = () => done(true);
    overlay.onclick = (e) => e.target === overlay && done(false);
    document.addEventListener('keydown', onKey);
    document.body.append(overlay);
  });
}
