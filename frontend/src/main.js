import { GetRenderedMarkdown, CloseApp } from '../wailsjs/go/main/App.js';

const THEMES = ['', 'theme-dark', 'theme-sepia'];
const ZOOM_MIN = 80;
const ZOOM_MAX = 150;
const ZOOM_STEP = 10;
const ZOOM_DEFAULT = 100;

let zoomLevel = ZOOM_DEFAULT;
let popoverVisible = false;

function isPopoverOpen() {
  return popoverVisible;
}

function updateZoomLabel() {
  const label = document.getElementById('zoom-label');
  if (label) label.textContent = zoomLevel + '%';
}

function updateThemeButtons() {
  document.querySelectorAll('.theme-btn').forEach(btn => {
    btn.classList.toggle('active', btn.dataset.theme === document.body.className);
  });
}

function cycleTheme() {
  const current = document.body.className;
  const idx = THEMES.indexOf(current);
  const next = THEMES[(idx + 1) % THEMES.length];
  document.body.className = next;
  localStorage.setItem('dumd-theme', next);
  updateThemeButtons();
}

function setTheme(themeName) {
  document.body.className = themeName;
  localStorage.setItem('dumd-theme', themeName);
  updateThemeButtons();
}

function restoreTheme() {
  const saved = localStorage.getItem('dumd-theme');
  if (THEMES.includes(saved)) {
    document.body.className = saved;
  }
  updateThemeButtons();
}

function openPopover() {
  popoverVisible = true;
  document.getElementById('settings-popover').classList.add('open');
  updateThemeButtons();
  updateZoomLabel();
}

function closePopover() {
  popoverVisible = false;
  document.getElementById('settings-popover').classList.remove('open');
}

function applyZoom() {
  document.body.style.fontSize = zoomLevel + '%';
  localStorage.setItem('dumd-zoom', String(zoomLevel));
  updateZoomLabel();
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

window.addEventListener('keydown', (e) => {
  if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') return;

  switch (e.key) {
    case 'Escape':
      if (isPopoverOpen()) {
        closePopover();
      } else {
        CloseApp();
      }
      break;
    case ' ':
      e.preventDefault();
      const dir = e.shiftKey ? -0.8 : 0.8;
      window.scrollBy({ top: window.innerHeight * dir, behavior: 'smooth' });
      break;
    case 'j':
      window.scrollBy({ top: 50, behavior: 'auto' });
      break;
    case 'k':
      window.scrollBy({ top: -50, behavior: 'auto' });
      break;
    case 'h':
      window.scrollBy({ left: -50, behavior: 'auto' });
      break;
    case 'l':
      window.scrollBy({ left: 50, behavior: 'auto' });
      break;
    case 't':
      cycleTheme();
      break;
    case '=':
      if (e.ctrlKey || e.metaKey) {
        e.preventDefault();
        zoomIn();
      }
      break;
    case '-':
      if (e.ctrlKey || e.metaKey) {
        e.preventDefault();
        zoomOut();
      }
      break;
  }
});

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

  const settingsBtn = document.getElementById('settings-btn');
  const popover = document.getElementById('settings-popover');

  settingsBtn.addEventListener('click', () => {
    if (isPopoverOpen()) {
      closePopover();
    } else {
      openPopover();
    }
  });

  popover.addEventListener('click', (e) => {
    e.stopPropagation();
  });

  document.addEventListener('click', (e) => {
    if (isPopoverOpen() && !popover.contains(e.target) && e.target !== settingsBtn) {
      closePopover();
    }
  });

  document.querySelectorAll('.theme-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      setTheme(btn.dataset.theme);
    });
  });

  document.getElementById('zoom-in-btn').addEventListener('click', () => {
    zoomIn();
  });

  document.getElementById('zoom-out-btn').addEventListener('click', () => {
    zoomOut();
  });

  let hoverTimeoutId = null;

  document.addEventListener('mousemove', (e) => {
    const margin = 80;
    const inZone = e.clientX >= window.innerWidth - margin && e.clientY >= window.innerHeight - margin;

    if (inZone) {
      clearTimeout(hoverTimeoutId);
      settingsBtn.classList.add('visible');
    } else {
      hoverTimeoutId = setTimeout(() => {
        if (!isPopoverOpen()) {
          settingsBtn.classList.remove('visible');
        }
      }, 1500);
    }
  });
});
