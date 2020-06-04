import '../node_modules/materialize-css/dist/css/materialize.css'
import '../public/global.css'
import '../node_modules/materialize-css/dist/js/materialize'

import App from './App.svelte';

const app = new App({
	target: document.body,
	props: {}
});

M.AutoInit();

export default app;