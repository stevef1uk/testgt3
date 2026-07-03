// Linkshelf SPA – plain JavaScript, no ES modules
(function() {
  'use strict';

  const API = '/api/links';

  document.addEventListener('DOMContentLoaded', init);

  async function init() {
    await fetchLinks();
    document.getElementById('link-form').addEventListener('submit', addLink);
  }

  async function fetchLinks() {
    const list = document.getElementById('link-list');
    list.innerHTML = '';
    try {
      const res = await fetch(API);
      if (!res.ok) throw new Error('Network response was not ok');
      const links = await res.json();
      links.forEach(link => {
        const li = createLinkElement(link);
        list.appendChild(li);
      });
    } catch (err) {
      console.error('Failed to fetch links:', err);
    }
  }

  async function addLink(e) {
    e.preventDefault();
    const input = document.getElementById('url-input');
    const url = input.value.trim();
    if (!url) return;
    try {
      const res = await fetch(API, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url })
      });
      if (!res.ok) throw new Error('Failed to add link');
      const link = await res.json();
      const list = document.getElementById('link-list');
      const li = createLinkElement(link);
      list.appendChild(li);
      input.value = '';
    } catch (err) {
      console.error('Failed to add link:', err);
    }
  }

  async function deleteLink(id, li) {
    try {
      const res = await fetch(`${API}/${id}`, { method: 'DELETE' });
      if (!res.ok) throw new Error('Failed to delete link');
      li.remove();
    } catch (err) {
      console.error('Failed to delete link:', err);
    }
  }

  function createLinkElement(link) {
    const li = document.createElement('li');
    li.className = 'link-item';
    li.dataset.id = link.id;
    const a = document.createElement('a');
    a.href = link.url;
    a.textContent = link.title || link.url;
    a.target = '_blank';
    const delBtn = document.createElement('button');
    delBtn.className = 'delete-btn';
    delBtn.textContent = 'Delete';
    delBtn.addEventListener('click', () => deleteLink(link.id, li));
    li.appendChild(a);
    li.appendChild(delBtn);
    return li;
  }
})();