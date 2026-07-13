// Linkshelf frontend — plain JavaScript, no ES modules
(function() {
  'use strict';

  const API_BASE = '/api/links';

  document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('link-form');
    const urlInput = document.getElementById('url-input');
    const titleInput = document.getElementById('title-input');
    const linkList = document.getElementById('link-list');

    // Fetch and render links
    function loadLinks() {
      fetch(API_BASE)
        .then(function(resp) { return resp.json(); })
        .then(function(links) {
          linkList.innerHTML = '';
          links.forEach(function(link) {
            const li = document.createElement('li');
            li.className = 'link-item';
            li.dataset.id = link.id;

            const a = document.createElement('a');
            a.href = link.url;
            a.textContent = link.title || link.url;
            a.target = '_blank';

            const deleteBtn = document.createElement('button');
            deleteBtn.className = 'delete-btn';
            deleteBtn.textContent = 'Delete';
            deleteBtn.addEventListener('click', function() {
              fetch(API_BASE + '/' + link.id, { method: 'DELETE' })
                .then(function(resp) {
                  if (resp.ok) {
                    li.remove();
                  }
                });
            });

            const li = document.createElement('li');
            li.className = 'link-item';
            li.appendChild(a);
            li.appendChild(deleteBtn);
            linkList.appendChild(li);
          });
        });
    }

    // Add link
    form.addEventListener('submit', function(e) {
      e.preventDefault();
      const url = urlInput.value.trim();
      const title = titleInput.value.trim();
      if (!url) return;

      fetch(API_BASE, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url: url, title: title || url })
      })
      .then(function(resp) {
        if (resp.ok) {
          urlInput.value = '';
          titleInput.value = '';
          loadLinks();
        }
      });
    });

    // Initial load
    loadLinks();
  });
})();