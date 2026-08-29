(function () {
  "use strict";

  var form = document.querySelector("form");
  var titleInput = document.getElementById("title");
  var urlInput = document.getElementById("url");
  var list = document.getElementById("links");
  var errorEl = document.getElementById("error");

  function setError(msg) {
    if (errorEl) {
      errorEl.textContent = msg || "";
    }
  }

  function escapeText(s) {
    return String(s == null ? "" : s);
  }

  function renderLinks(items) {
    while (list.firstChild) {
      list.removeChild(list.firstChild);
    }
    if (!items || items.length === 0) {
      var empty = document.createElement("li");
      empty.className = "empty";
      empty.textContent = "No links yet.";
      list.appendChild(empty);
      return;
    }
    for (var i = 0; i < items.length; i++) {
      list.appendChild(renderItem(items[i]));
    }
  }

  function renderItem(link) {
    var li = document.createElement("li");

    var a = document.createElement("a");
    a.href = link.url;
    a.textContent = link.title || link.url;
    a.target = "_blank";
    a.rel = "noopener noreferrer";

    var btn = document.createElement("button");
    btn.type = "button";
    btn.className = "delete";
    btn.textContent = "Delete";
    btn.setAttribute("data-id", link.id);
    btn.addEventListener("click", function (e) {
      var id = e.currentTarget.getAttribute("data-id");
      deleteLink(id);
    });

    li.appendChild(a);
    li.appendChild(btn);
    return li;
  }

  function loadLinks() {
    setError("");
    return fetch("/links", { method: "GET", headers: { "Accept": "application/json" } })
      .then(function (res) {
        if (!res.ok) {
          throw new Error("Failed to load links (status " + res.status + ")");
        }
        return res.json();
      })
      .then(function (data) {
        renderLinks(data);
      })
      .catch(function (err) {
        setError("Could not load links: " + err.message);
        renderLinks([]);
      });
  }

  function createLink(title, url) {
    setError("");
    return fetch("/links", {
      method: "POST",
      headers: { "Content-Type": "application/json", "Accept": "application/json" },
      body: JSON.stringify({ title: title, url: url })
    })
      .then(function (res) {
        if (!res.ok) {
          return res.text().then(function (t) {
            throw new Error(t || ("Create failed (status " + res.status + ")"));
          });
        }
        return res.json();
      })
      .then(function () {
        titleInput.value = "";
        urlInput.value = "";
        return loadLinks();
      })
      .catch(function (err) {
        setError("Could not add link: " + err.message);
      });
  }

  function deleteLink(id) {
    setError("");
    return fetch("/links/" + encodeURIComponent(id), {
      method: "DELETE",
      headers: { "Accept": "application/json" }
    })
      .then(function (res) {
        if (!res.ok) {
          throw new Error("Delete failed (status " + res.status + ")");
        }
        return loadLinks();
      })
      .catch(function (err) {
        setError("Could not delete link: " + err.message);
      });
  }

  function onSubmit(e) {
    e.preventDefault();
    var title = escapeText(titleInput.value).trim();
    var url = escapeText(urlInput.value).trim();
    if (!title || !url) {
      setError("Both title and URL are required.");
      return;
    }
    if (!/^https?:\/\//i.test(url)) {
      setError("URL must start with http:// or https://");
      return;
    }
    createLink(title, url);
  }

  if (form) {
    form.addEventListener("submit", onSubmit);
  }

  if (list) {
    loadLinks();
  }
})();