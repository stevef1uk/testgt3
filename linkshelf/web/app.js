// Plain JavaScript frontend for LinkShelf
// No ES module imports — uses fetch() and DOM APIs

(function () {
  'use strict';

  const API_BASE = '/api/links';

  document.addEventListener('DOMContentLoaded', function () {
    const form = document.getElementById('link-form');
    const urlInput = document.getElementById('url-input');
    const linkList = document.getElementById('link-list');
    const errorMessage = document.getElementById('error-message');

    // Fetch and render all links
    function loadLinks() {
      fetch(API_BASE)
        .then(function (response) {
          if (!response.ok) throw new Error('Failed to load links');
          return response.json();
        })
        .then(function (links) {
          renderLinks(links);
        })
        .catch(function (err) {
          showError(err.message);
        });
    }

    // Render the link list
    function renderLinks(links) {
      if (!linkList) return;
      if (links.length === 0) {
        linkList.innerHTML = '<div class="empty-state"><p>No links yet. Add one above!</p></div>';
        return;
      }

      var html = '';
      links.forEach(function (link) {
        // Parse domain from URL for display
        var domain = '';
        try {
          domain = new URL(link.url).hostname;
        } catch (e) {
          domain = link.url;
        }

        html +=
          '<li>' +
            '<div>' +
              '<a href="' + escapeHtml(link.url) + '" target="_blank" rel="noopener noreferrer">' +
                escapeHtml(link.url) +
              '</a>' +
              '<span class="domain">(' + escapeHtml(domain) + ')</span>' +
            '</div>' +
            '<button class="delete-btn" data-id="' + link.id + '">Delete</button>' +
          '</li>';
      });
      linkList.innerHTML = html;
    }

    // Add a new link
    function addLink(url) {
      fetch(API_BASE, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url: url })
      })
        .then(function (response) {
          if (!response.ok) throw new Error('Failed to add link');
          return response.json();
        })
        .then(function () {
          urlInput.value = '';
          loadLinks();
          hideError();
        })
        .catch(function (err) {
          showError(err.message);
        });
    }

    // Delete a link by ID
    function deleteLink(id) {
      fetch(API_BASE + '/' + encodeURIComponent(id), {
        method: 'DELETE'
      })
        .then(function (response) {
          if (!response.ok) throw new Error('Failed to delete link');
          loadLinks();
        })
        .catch(function (err) {
          showError(err.message);
        });
    }

    // Event delegation for delete buttons
    if (linkList) {
      linkList.addEventListener('click', function (e) {
        var target = e.target;
        if (target && target.classList.contains('delete-btn')) {
          var id = target.getAttribute('data-id');
          if (id) deleteLink(id);
        }
      });
    }

    // Form submission
    if (form) {
      form.addEventListener('submit', function (e) {
        e.preventDefault();
        var url = urlInput.value.trim();
        if (!url) {
          showError('Please enter a URL');
          return;
        }
        addLink(url);
      });
    }

    // Error display
    function showError(msg) {
      if (errorMessage) {
        errorMessage.textContent = msg;
        errorMessage.classList.add('show');
      } else {
        alert(msg);
      }
    }

    function hideError() {
      if (errorMessage) {
        errorMessage.textContent = '';
        errorMessage.classList.remove('show');
      }
    }

    // Escape HTML to prevent XSS
    function escapeHtml(str) {
      var div = document.createElement('div');
      div.appendChild(document.createTextNode(str));
      return div.innerHTML;
    }

    // Initial load
    loadLinks();
  });
})();