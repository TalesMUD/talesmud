import '../node_modules/materialize-css/dist/css/materialize.css'
import '../public/global.css'
import '../node_modules/materialize-css/dist/js/materialize'

import App from './App.svelte';

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
