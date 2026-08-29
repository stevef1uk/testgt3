(() => {
  const form = document.getElementById("link-form");
  const titleInput = document.getElementById("title");
  const urlInput = document.getElementById("url");
  const linksList = document.getElementById("links");
  const error = document.getElementById("error");

  function showError(message) {
    error.textContent = message || "";
  }

  function linkId(link) {
    return link.id ?? link.ID;
  }

  function linkTitle(link) {
    return link.title ?? link.Title ?? "";
  }

  function linkURL(link) {
    return link.url ?? link.URL ?? "";
  }

  function renderLinks(links) {
    linksList.replaceChildren();

    for (const link of links) {
      const item = document.createElement("li");
      const anchor = document.createElement("a");
      const remove = document.createElement("button");

      anchor.href = linkURL(link);
      anchor.textContent = linkTitle(link);
      anchor.target = "_blank";
      anchor.rel = "noopener noreferrer";

      remove.type = "button";
      remove.textContent = "Delete";
      remove.addEventListener("click", () => deleteLink(linkId(link)));

      item.append(anchor, remove);
      linksList.append(item);
    }
  }

  async function responseJSON(response) {
    if (!response.ok) {
      throw new Error(`Request failed (${response.status})`);
    }
    return response.json();
  }

  async function loadLinks() {
    try {
      const response = await fetch("/api/links");
      const links = await responseJSON(response);
      renderLinks(Array.isArray(links) ? links : []);
    } catch (err) {
      showError(err.message || "Unable to load links.");
    }
  }

  async function deleteLink(id) {
    showError("");

    try {
      const response = await fetch(`/api/links/${encodeURIComponent(id)}`, {
        method: "DELETE"
      });
      if (!response.ok) {
        throw new Error(`Request failed (${response.status})`);
      }
      await loadLinks();
    } catch (err) {
      showError(err.message || "Unable to delete link.");
    }
  }

  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    showError("");

    try {
      const response = await fetch("/api/links", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          title: titleInput.value.trim(),
          url: urlInput.value.trim()
        })
      });
      await responseJSON(response);
      form.reset();
      titleInput.focus();
      await loadLinks();
    } catch (err) {
      showError(err.message || "Unable to add link.");
    }
  });

  loadLinks();
})();