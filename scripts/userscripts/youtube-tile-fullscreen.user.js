// ==UserScript==
// @name         YouTube Tile Fullscreen (Hyprland)
// @namespace    hairglasses-studio/dotfiles
// @version      1.0.0
// @description  Make YouTube's F key do CSS pseudo-fullscreen inside the window instead of requesting XDG fullscreen. Keeps the Hyprland tile intact.
// @match        https://www.youtube.com/*
// @match        https://music.youtube.com/*
// @run-at       document-end
// @grant        none
// ==/UserScript==

(() => {
  const STYLE_ID = '__hg_pseudo_fs_style';
  const CLASS = 'hg-pseudo-fullscreen';

  const injectStyle = () => {
    if (document.getElementById(STYLE_ID)) return;
    const s = document.createElement('style');
    s.id = STYLE_ID;
    s.textContent = `
      .${CLASS} {
        position: fixed !important;
        inset: 0 !important;
        width: 100vw !important;
        height: 100vh !important;
        z-index: 2147483647 !important;
        background: #000 !important;
        margin: 0 !important;
      }
      html:has(.${CLASS}), body:has(.${CLASS}) { overflow: hidden !important; }
    `;
    (document.head || document.documentElement).appendChild(s);
  };

  let fakeEl = null;

  const fireChange = () => {
    document.dispatchEvent(new Event('fullscreenchange', { bubbles: true }));
    document.dispatchEvent(new Event('webkitfullscreenchange', { bubbles: true }));
  };

  const request = function () {
    injectStyle();
    fakeEl = this;
    this.classList.add(CLASS);
    queueMicrotask(fireChange);
    return Promise.resolve();
  };

  const exit = function () {
    if (fakeEl) fakeEl.classList.remove(CLASS);
    fakeEl = null;
    queueMicrotask(fireChange);
    return Promise.resolve();
  };

  Element.prototype.requestFullscreen = request;
  Element.prototype.webkitRequestFullscreen = request;
  Element.prototype.webkitRequestFullScreen = request;

  Document.prototype.exitFullscreen = exit;
  Document.prototype.webkitExitFullscreen = exit;

  const defineSpoof = (obj, prop, getter) => {
    try {
      Object.defineProperty(obj, prop, {
        configurable: true,
        get: getter,
        set: () => {},
      });
    } catch (_) { /* already non-configurable; ignore */ }
  };

  defineSpoof(document, 'fullscreenElement', () => fakeEl);
  defineSpoof(document, 'webkitFullscreenElement', () => fakeEl);
  defineSpoof(document, 'fullscreenEnabled', () => true);
  defineSpoof(document, 'webkitFullscreenEnabled', () => true);
})();
