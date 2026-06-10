// Import necessary dependencies
import { Create, Delete, InitSchema, Link, List } from './api.js';

// Initialize the application
InitSchema();

// Define a function to handle form submissions
async function handleSubmit(event) {
  event.preventDefault();
  const formData = new FormData(event.target);
  const data = Object.fromEntries(formData.entries());

  // Call the Create function to create a new item
  try {
    const response = await Create(data);
    console.log(response);
  } catch (error) {
    console.error(error);
  }
}

// Define a function to handle delete button clicks
async function handleDelete(event) {
  event.preventDefault();
  const id = event.target.dataset.id;

  // Call the Delete function to delete an item
  try {
    const response = await Delete(id);
    console.log(response);
  } catch (error) {
    console.error(error);
  }
}

// Define a function to handle link clicks
async function handleLinkClick(event) {
  event.preventDefault();
  const id = event.target.dataset.id;

  // Call the Link function to link an item
  try {
    const response = await Link(id);
    console.log(response);
  } catch (error) {
    console.error(error);
  }
}

// Define a function to handle list rendering
async function renderList() {
  // Call the List function to get a list of items
  try {
    const response = await List();
    console.log(response);

    // Render the list of items
    const listElement = document.getElementById('list');
    listElement.innerHTML = '';
    response.forEach((item) => {
      const listItemElement = document.createElement('li');
      listItemElement.textContent = item.name;
      listElement.appendChild(listItemElement);
    });
  } catch (error) {
    console.error(error);
  }
}

// Add event listeners to the form and buttons
document.addEventListener('DOMContentLoaded', () => {
  const formElement = document.getElementById('form');
  formElement.addEventListener('submit', handleSubmit);

  const deleteButtons = document.querySelectorAll('.delete-button');
  deleteButtons.forEach((button) => {
    button.addEventListener('click', handleDelete);
  });

  const linkButtons = document.querySelectorAll('.link-button');
  linkButtons.forEach((button) => {
    button.addEventListener('click', handleLinkClick);
  });

  renderList();
});