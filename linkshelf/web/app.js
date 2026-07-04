(function() {
  var form = document.getElementById('addLinkForm');
  var list = document.getElementById('linkList');

  function loadLinks() {
    fetch('/api/links')
      .then(function(res) { return res.json(); })
      .then(function(links) {
        list.innerHTML = '';
        links.forEach(function(link) {
          var li = document.createElement('li');
          li.innerHTML = '<a href="' + escapeHtml(link.url) + '" target="_blank">' + escapeHtml(link.title) + '</a> ' +
            '<button data-id="' + link.id + '">Delete</button>';
          li.querySelector('button').addEventListener('click', function() {
            deleteLink(link.id);
          });
          list.appendChild(li);
        });
      });
  }

  function escapeHtml(str) {
    var div = document.createElement('div');
    div.appendChild(document.createTextNode(str));
    return div.innerHTML;
  }

  function deleteLink(id) {
    fetch('/api/links/' + id, { method: 'DELETE' })
      .then(function(res) {
        if (res.ok) loadLinks();
      });
  }

  form.addEventListener('submit', function(e) {
    e.preventDefault();
    var url = document.getElementById('url').value;
    var title = document.getElementById('title').value;
    fetch('/api/links', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ url: url, title: title })
    }).then(function(res) {
      if (res.ok) {
        form.reset();
        loadLinks();
      }
    });
  });

  loadLinks();
})();