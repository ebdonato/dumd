import { GetRenderedMarkdown } from '../wailsjs/go/main/App.js';

const THEMES = ['', 'theme-dark', 'theme-sepia'];
const ZOOM_MIN = 80;
const ZOOM_MAX = 150;
const ZOOM_STEP = 10;
const ZOOM_DEFAULT = 100;

let zoomLevel = ZOOM_DEFAULT;

function cycleTheme() {
  const current = document.body.className;
  const idx = THEMES.indexOf(current);
  const next = THEMES[(idx + 1) % THEMES.length];
  document.body.className = next;
  localStorage.setItem('dumd-theme', next);
}

function restoreTheme() {
  const saved = localStorage.getItem('dumd-theme');
  if (THEMES.includes(saved)) {
    document.body.className = saved;
  }
}

function applyZoom() {
  document.body.style.fontSize = zoomLevel + '%';
  localStorage.setItem('dumd-zoom', String(zoomLevel));
}

function zoomIn() {
  if (zoomLevel < ZOOM_MAX) {
    zoomLevel += ZOOM_STEP;
    applyZoom();
  }
}

function zoomOut() {
  if (zoomLevel > ZOOM_MIN) {
    zoomLevel -= ZOOM_STEP;
    applyZoom();
  }
}

function restoreZoom() {
  const saved = localStorage.getItem('dumd-zoom');
  if (saved) {
    const val = parseInt(saved, 10);
    if (val >= ZOOM_MIN && val <= ZOOM_MAX) {
      zoomLevel = val;
      applyZoom();
    }
  }
}

document.addEventListener('DOMContentLoaded', async () => {
  restoreTheme();
  restoreZoom();

  const content = document.getElementById('content');
  try {
    const html = await GetRenderedMarkdown();
    content.innerHTML = html;
  } catch (err) {
    content.innerHTML = `<p>Error loading Markdown: ${err}</p>`;
  }
});
