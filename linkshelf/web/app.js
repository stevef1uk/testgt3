"use strict";

const form = document.getElementById("link-form");
const urlInput = document.getElementById("url");
const linksList = document.getElementById("links");
const linksURL = "/api/links";

function renderLinks(links) 

async function loadLinks() 

async function deleteLink(id) 

form.addEventListener("submit", async (event) => 

loadLinks().catch((error) => {
  console.error(error);
});