/* Client-side JavaScript for the Linkshelf web UI.
   Uses fetch() to interact with the REST API at /api/links.
   No ES module imports — plain JavaScript for browser compatibility. */

(function () {
    "use strict";

    const API_URL = "/api/links";

    document.addEventListener("DOMContentLoaded", function () {
        const form = document.getElementById("linkForm");
        const urlInput = document.getElementById("urlInput");
        const titleInput = document.getElementById("titleInput");
        const linksList = document.getElementById("linksList");

        if (!form || !urlInput || !titleInput || !linksList) {
            console.error("Required DOM elements not found");
            return;
        }

        // Load existing links on page load
        loadLinks(linksList);

        // Handle form submission to create a new link
        form.addEventListener("submit", function (e) {
            e.preventDefault();
            const url = urlInput.value.trim();
            const title = titleInput.value.trim();
            if (!url || !title) {
                alert("URL and title are required");
                return;
            }
            createLink(url, title, urlInput, titleInput, linksList);
        });
    });

    function loadLinks(linksList) {
        fetch(API_URL)
            .then(function (resp) {
                if (!resp.ok) {
                    throw new Error("Failed to fetch links: " + resp.statusText);
                }
                return resp.json();
            })
            .then(function (links) {
                renderLinks(links, linksList);
            })
            .catch(function (err) {
                console.error("Error loading links:", err);
            });
    }

    function createLink(url, title, urlInput, titleInput, linksList) {
        fetch(API_URL, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ url: url, title: title })
        })
            .then(function (resp) {
                if (!resp.ok) {
                    throw new Error("Failed to create link: " + resp.statusText);
                }
                return resp.json();
            })
            .then(function () {
                urlInput.value = "";
                titleInput.value = "";
                // Reload the list to show the new link
                loadLinks(linksList);
            })
            .catch(function (err) {
                console.error("Error creating link:", err);
                alert("Failed to create link. See console for details.");
            });
    }

    function deleteLink(id, linksList) {
        fetch(API_URL + "/" + id, {
            method: "DELETE"
        })
            .then(function (resp) {
                if (!resp.ok) {
                    throw new Error("Failed to delete link: " + resp.statusText);
                }
                // Reload the list
                loadLinks(linksList);
            })
            .catch(function (err) {
                console.error("Error deleting link:", err);
                alert("Failed to delete link. See console for details.");
            });
    }

    function renderLinks(links, linksList) {
        linksList.innerHTML = "";

        if (!links || links.length === 0) {
            var emptyMsg = document.createElement("li");
            emptyMsg.textContent = "No links yet. Add one above!";
            linksList.appendChild(emptyMsg);
            return;
        }

        links.forEach(function (link) {
            var li = document.createElement("li");

            var titleSpan = document.createElement("span");
            titleSpan.className = "link-title";
            titleSpan.textContent = link.title || "Untitled";

            var urlAnchor = document.createElement("a");
            urlAnchor.className = "link-url";
            urlAnchor.href = link.url;
            urlAnchor.target = "_blank";
            urlAnchor.rel = "noopener noreferrer";
            urlAnchor.textContent = link.url;

            var deleteBtn = document.createElement("button");
            deleteBtn.className = "delete-btn";
            deleteBtn.textContent = "Delete";
            deleteBtn.addEventListener("click", function () {
                deleteLink(link.id, linksList);
            });

            li.appendChild(titleSpan);
            li.appendChild(urlAnchor);
            li.appendChild(deleteBtn);
            linksList.appendChild(li);
        });
    }
})();