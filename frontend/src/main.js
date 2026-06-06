import { GetRenderedMarkdown } from '../wailsjs/go/main/App.js';

document.addEventListener('DOMContentLoaded', async () => {
  const content = document.getElementById('content');
  try {
    const html = await GetRenderedMarkdown();
    content.innerHTML = html;
  } catch (err) {
    content.innerHTML = `<p>Error loading Markdown: ${err}</p>`;
  }
});
