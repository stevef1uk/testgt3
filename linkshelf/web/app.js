/*
 * Implementation of the client‑side JavaScript for the Linkshelf SPA.
 * It follows the contract defined in architecture.md / SPEC.md:
 *
 *   • GET    /api/links           → returns []Link
 *   • POST   /api/links           → creates a Link, returns the created object
 *   • DELETE /api/links/{id}      → deletes a Link, returns 204 on success
 *
 * The HTML (index.html) provides:
 *   – <div id="output"></div>   for transient messages
 *   – <ul   id="link-list"></ul> for the list of links
 *   – <form>…</form>            with a <input type="url"> field
 *
 * This script:
 *   1. Loads the link list on page load.
 *   2. Submits new links via the form.
 *   3. Renders each link with a Delete button.
 *   4. Handles delete actions and refreshes the list.
 *
 * All fetch calls include error handling that surfaces server‑provided
 * error messages when available.
 */
document.addEventListener('DOMContentLoaded', () => {
  const outputEl = document.getElementById('output');
  const listEl   = document.getElementById('link-list'); // Changed from 'list' to 'link-list'
  const formEl   = document.getElementById('add-link-form'); // Changed from querySelector('form') to getElementById
  const urlInput = document.getElementById('link-input'); // Changed from querySelector('input[type="url"]') to getElementById

  const showMessage = (msg, isError = false) => {
    outputEl.textContent = msg;
    outputEl.style.color = isError ? 'red' : 'green';
    setTimeout(() => { outputEl.textContent = ''; }, 3000);
  };

  const fetchLinks = async () => {
    try {
      const resp = await fetch('/api/links');
      if (!resp.ok) {
        let err = `HTTP ${resp.status}`;
        try { const data = await resp.json(); err = data.error || err; } catch (_) {}
        throw new Error(err);
      }
      const links = await resp.json();
      renderLinks(links);
    } catch (e) {
      console.error(e);
      showMessage(`Failed to load links: ${e.message}`, true);
    }
  };

  const renderLinks = (links) => {
    listEl.innerHTML = '';
    if (!Array.isArray(links) || links.length === 0) {
      const li = document.createElement('li');
      li.textContent = 'No links yet.';
      listEl.appendChild(li);
      return;
    }
    links.forEach(link => {
      const li = document.createElement('li');
      li.innerHTML = `
        <a href="${link.url}" target="_blank">${link.url}</a>
        <button data-id="${link.id}" style="margin-left:8px;">Delete</button>
      `;
      listEl.appendChild(li);
    });
  };

  formEl.addEventListener('submit', async (e) => {
    e.preventDefault();
    const url = urlInput.value.trim();
    if (!url) { showMessage('URL cannot be empty.', true); return; }
    try {
      const resp = await fetch('/api/links', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url })
      });
      if (!resp.ok) {
        const data = await resp.json();
        throw new Error(data.error || `HTTP ${resp.status}`);
      }
      // The created link object is returned by the API, but not directly used here.
      // We refresh the list to show the newly added link.
      await resp.json();
      showMessage('Link created successfully!');
      urlInput.value = ''; // Clear the input field
      fetchLinks(); // Refresh the list of links
    } catch (e) {
      console.error(e);
      showMessage(`Failed to create link: ${e.message}`, true);
    }
  });

  listEl.addEventListener('click', async (e) => {
    if (e.target.tagName !== 'BUTTON') return;
    const id = e.target.dataset.id;
    if (!id) return; // Ensure the button has a data-id attribute
    try {
      const resp = await fetch(`/api/links/${encodeURIComponent(id)}`, { method: 'DELETE' });
      // A successful DELETE operation might return 204 No Content or 200 OK.
      // We check for !resp.ok to catch errors and specifically check for 204.
      if (!resp.ok && resp.status !== 204) {
        const data = await resp.json();
        throw new Error(data.error || `HTTP ${resp.status}`);
      }
      showMessage('Link deleted successfully!');
      fetchLinks(); // Refresh the list after deletion
    } catch (err) {
      console.error(err);
      showMessage(`Failed to delete link: ${err.message}`, true);
    }
  });

  // Load initial list when the DOM is ready
  fetchLinks();
});