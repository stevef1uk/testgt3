"use strict";

const form = document.getElementById("link-form");
const titleInput = document.getElementById("title");
const urlInput = document.getElementById("url");
const linksList = document.getElementById("links");
const statusMessage = document.getElementById("status");
const linksURL = "/api/links";

function renderLinks(links) {
  linksList.replaceChildren();
  if (!Array.isArray(links) || links.length === 0) {
    const empty = document.createElement("li");
    empty.textContent = "No links saved yet.";
    linksList.appendChild(empty);
    return;
  }

  links.forEach((link) => {
    const item = document.createElement("li");
    const anchor = document.createElement("a");
    anchor.href = link.url;
    anchor.textContent = link.title || link.url;
    anchor.target = "_blank";
    anchor.rel = "noopener noreferrer";
    const remove = document.createElement("button");
    remove.type = "button";
    remove.textContent = "Delete";
    remove.addEventListener("click", () => deleteLink(link.id));
    item.append(anchor, remove);
    linksList.appendChild(item);
  });
}

async function loadLinks() {
  const response = await fetch(linksURL);
  if (!response.ok) throw new Error("Could not load links");
  renderLinks(await response.json());
}

async function deleteLink(id) {
  const response = await fetch(`${linksURL}/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
  if (!response.ok) throw new Error("Could not delete link");
  await loadLinks();
}

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  statusMessage.textContent = "";
  const response = await fetch(linksURL, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ title: titleInput.value.trim(), url: urlInput.value.trim() }),
  });
  if (!response.ok) {
    statusMessage.textContent = "Unable to save link.";
    return;
  }
  form.reset();
  await loadLinks();
});

loadLinks().catch((error) => {
  console.error(error);
  statusMessage.textContent = "Unable to load links.";
});

async function loadLinks() 

async function deleteLink(id) 

form.addEventListener("submit", async (event) => 

loadLinks().catch((error) => {
  console.error(error);
});