// linkshelf/web/app.js — frontend logic (plain JS, no bundler)

(function () {
  'use strict';

  const form = document.getElementById('link-form');
  const urlInput = document.getElementById('url-input');
  const linkList = document.getElementById('link-list');
  const statusEl = document.getElementById('status');

  if (!form || !urlInput || !linkList) {
    console.error('Required DOM elements missing');
    return;
  }

  async function loadLinks() {
    try {
      const res = await fetch('/api/links');
      if (!res.ok) throw new Error('Failed to fetch links');
      const links = await res.json();
      renderLinks(links);
    } catch (err) {
      showStatus('Error loading links: ' + err.message);
    }
  }

  function renderLinks(links) {
    linkList.innerHTML = '';
    links.forEach(function (link) {
      const li = document.createElement('li');
      const a = document.createElement('a');
      a.href = link.url;
      a.textContent = link.url;
      a.target = '_blank';
      a.rel = 'noopener noreferrer';

      const deleteBtn = document.createElement('button');
      deleteBtn.textContent = 'Delete';
      deleteBtn.dataset.id = link.id;
      deleteBtn.addEventListener('click', function () {
        deleteLink(link.id);
      });

      li.appendChild(a);
      li.appendChild(deleteBtn);
      linkList.appendChild(li);
    });
  }

  function showStatus(msg) {
    if (statusEl) {
      statusEl.textContent = msg;
    }
  }

  async function addLink(url) {
    try {
      const res = await fetch('/api/links', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url: url }),
      });
      if (!res.ok) throw new Error('Failed to create link');
      const created = await res.json();
      showStatus('Link added');
      urlInput.value = '';
      loadLinks();
    } catch (err) {
      showStatus('Error: ' + err.message);
    }
  }

  async function deleteLink(id) {
    try {
      const res = await fetch('/api/links/' + id, {
        method: 'DELETE',
      });
      if (!res.ok) throw new Error('Failed to delete link');
      showStatus('Link deleted');
      loadLinks();
    } catch (err) {
      showStatus('Error: ' + err.message);
    }
  }

  form.addEventListener('submit', function (e) {
    e.preventDefault();
    const url = urlInput.value.trim();
    if (url) {
      addLink(url);
    }
  });

  // Initial load
  loadLinks();
})();