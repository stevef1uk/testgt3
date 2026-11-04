(() => {
  const linksList = document.getElementById('links');
  const form = document.querySelector('form');
  const inputs = document.querySelectorAll('input');
  const addButton = document.querySelector('button');

  if (!linksList || inputs.length < 2) {
    return;
  }

  const titleInput = inputs[0];
  const urlInput = inputs[1];

  async function loadLinks() {
    const response = await fetch('/api/items');
    if (!response.ok) {
      throw new Error('Unable to load links');
    }

    const links = await response.json();
    renderLinks(Array.isArray(links) ? links : []);
  }

  function renderLinks(links) {
    linksList.replaceChildren();

    links.forEach((link) => {
      const item = document.createElement('li');
      const anchor = document.createElement('a');
      const removeButton = document.createElement('button');

      anchor.href = link.url;
      anchor.textContent = link.title;
      anchor.target = '_blank';
      anchor.rel = 'noreferrer';
      removeButton.type = 'button';
      removeButton.textContent = 'Delete';
      removeButton.addEventListener('click', async () => {
        const response = await fetch(`/api/items/${encodeURIComponent(link.id)}`, {
          method: 'DELETE'
        });
        if (!response.ok) {
          throw new Error('Unable to delete link');
        }
        await loadLinks();
      });

      item.append(anchor, removeButton);
      linksList.append(item);
    });
  }

  async function addLink(event) {
    if (event) {
      event.preventDefault();
    }

    const title = titleInput.value.trim();
    const url = urlInput.value.trim();
    if (!title || !url) {
      return;
    }

    const response = await fetch('/api/items', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ title, url })
    });
    if (!response.ok) {
      throw new Error('Unable to add link');
    }

    titleInput.value = '';
    urlInput.value = '';
    await loadLinks();
  }

  if (form) {
    form.addEventListener('submit', addLink);
  } else if (addButton) {
    addButton.addEventListener('click', addLink);
  }

  loadLinks().catch((error) => {
    linksList.textContent = error.message;
  });
})();