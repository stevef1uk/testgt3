(function () {
  "use strict";

  var form = document.getElementById("link-form");
  var titleInput = document.getElementById("title");
  var urlInput = document.getElementById("url");
  var linksList = document.getElementById("links");
  var error = document.getElementById("error");

  function showError(message) {
    error.textContent = message || "";
  }

  function linkItems(payload) {
    if (Array.isArray(payload)) {
      return payload;
    }
    if (payload && Array.isArray(payload.links)) {
      return payload.links;
    }
    if (payload && Array.isArray(payload.items)) {
      return payload.items;
    }
    return [];
  }

  function renderLinks(items) {
    linksList.textContent = "";

    items.forEach(function (item) {
      var li = document.createElement("li");
      var anchor = document.createElement("a");
      var title = item.title || item.name || item.url || "Untitled link";
      var href = item.url || item.href || "#";

      anchor.href = href;
      anchor.textContent = title;
      anchor.target = "_blank";
      anchor.rel = "noopener noreferrer";
      li.appendChild(anchor);

      if (item.url && item.url !== title) {
        var address = document.createElement("span");
        address.textContent = item.url;
        li.appendChild(address);
      }

      linksList.appendChild(li);
    });
  }

  function readJson(response) {
    if (!response.ok) {
      throw new Error("Request failed (" + response.status + ")");
    }
    return response.json();
  }

  function loadLinks() {
    return fetch("/api/links")
      .then(readJson)
      .then(function (payload) {
        renderLinks(linkItems(payload));
      })
      .catch(function (reason) {
        showError(reason.message || "Unable to load saved links.");
      });
  }

  form.addEventListener("submit", function (event) {
    event.preventDefault();
    showError("");

    var title = titleInput.value.trim();
    var url = urlInput.value.trim();

    if (!title || !url) {
      showError("Title and URL are required.");
      return;
    }

    fetch("/api/links", {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({ title: title, url: url })
    })
      .then(readJson)
      .then(function () {
        form.reset();
        return loadLinks();
      })
      .catch(function (reason) {
        showError(reason.message || "Unable to save link.");
      });
  });

  loadLinks();
}());