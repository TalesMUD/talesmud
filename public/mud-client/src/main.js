import '../node_modules/materialize-css/dist/css/materialize.css'
import '../public/global.css'
import '../node_modules/materialize-css/dist/js/materialize'

import App from './App.svelte';

// Keep ligature names invisible until the local icon font has loaded.
document.fonts.load('24px "Material Icons"', 'shield').then((faces) => {
  if (faces.some((face) => face.status === 'loaded')) {
    const context = document.createElement('canvas').getContext('2d');
    context.font = '24px "Material Icons"';
    const validate = (icon) => {
      const name = icon.textContent.trim();
      icon.classList.toggle('material-icon-unavailable', !name || Math.abs(context.measureText(name).width - 24) > 0.5);
    };
    const scan = (node) => {
      if (node.nodeType === Node.ELEMENT_NODE) {
        if (node.matches('.material-icons')) validate(node);
        node.querySelectorAll('.material-icons').forEach(validate);
      } else if (node.parentElement?.matches('.material-icons')) {
        validate(node.parentElement);
      }
    };
    scan(document.body);
    new MutationObserver((records) => {
      for (const record of records) {
        if (record.type === 'characterData') scan(record.target);
        else {
          scan(record.target);
          record.addedNodes.forEach(scan);
        }
      }
    }).observe(document.body, { childList: true, characterData: true, subtree: true });
    document.documentElement.classList.add('material-icons-ready');
  }
}).catch(() => {});

Object.defineProperty(String.prototype, "hashCode", {
    value: function () {
      var hash = 0,
        i,
        chr;
      for (i = 0; i < this.length; i++) {
        chr = this.charCodeAt(i);
        hash = (hash << 5) - hash + chr;
        hash |= 0; // Convert to 32bit integer
      }
      return hash;
    },
  });

const app = new App({
	target: document.body,
	props: {}
});

M.AutoInit();

export default app;
