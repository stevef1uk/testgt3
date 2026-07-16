document.addEventListener('DOMContentLoaded', () => {
  const listEl = document.getElementById('link-list');
  const formEl = document.getElementById('new-link-form');
  const urlInput = document.getElementById('url-input');
  const descInput = document.getElementById('desc-input');

  function fetchLinks() {
    fetch('/api/links')
      .then(r => r.json())
      .then(data => {
        listEl.innerHTML = '';
        data.forEach(link => {
          const li = document.createElement('li');

          const a = document.createElement('a');
          a.href = link.url;
          a.textContent = link.description || link.url;
          a.target = '_blank';
          li.appendChild(a);

          const del = document.createElement('button');
          del.textContent = 'Delete';
          del.dataset.id = link.id;
          del.addEventListener('click', () => deleteLink(link.id));
          li.appendChild(del);

          listEl.appendChild(li);
        });
      })
      .catch(err => console.error('Failed to fetch links:', err));
  }

  function addLink(e) {
    e.preventDefault();
    const payload = {
      url: urlInput.value,
      title: descInput.value,
    };
    fetch('/api/links', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
      .then(r => {
        if (!r.ok) throw new Error('Network response was not ok');
        return r.json();
      })
      .then(() => {
        urlInput.value = '';
        descInput.value = '';
        fetchLinks();
      })
      .catch(err => console.error('Failed to add link:', err));
  }

  function deleteLink(id) {
    fetch(`/api/links/${id}`, { method: 'DELETE' })
      .then(r => {
        if (!r.ok) throw new Error('Delete failed');
        fetchLinks();
      })
      .catch(err => console.error('Failed to delete link:', err));
  }

  if (formEl) {
    formEl.addEventListener('submit', addLink);
  }

  fetchLinks();
});