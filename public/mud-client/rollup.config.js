import svelte from 'rollup-plugin-svelte';
import resolve from '@rollup/plugin-node-resolve';
import commonjs from '@rollup/plugin-commonjs';
import livereload from 'rollup-plugin-livereload';
import { terser } from 'rollup-plugin-terser';
import css from "rollup-plugin-css-only";

const production = !process.env.ROLLUP_WATCH;

// Auth0 credentials come from deployment build arguments, shared by both pages.
function auth0EnvReplace() {
	const replacements = {
		__VITE_AUTH0_DOMAIN__: process.env.VITE_AUTH0_DOMAIN || "",
		__VITE_AUTH0_CLIENT_ID__: process.env.VITE_AUTH0_CLIENT_ID || "",
		__VITE_AUTH0_AUDIENCE__: process.env.VITE_AUTH0_AUDIENCE || "",
	};
	return {
		name: "auth0-env-replace",
		transform(code, id) {
			if (!id.endsWith("/authConfig.js")) return null;
			let next = code;
			for (const [token, value] of Object.entries(replacements)) {
				next = next.split(JSON.stringify(token)).join(JSON.stringify(value));
			}
			return { code: next, map: null };
		},
	};
}

export default [{
	input: 'src/main.js',
	output: {
		sourcemap: true,
		format: 'iife',
		name: 'mudclient',
		file: 'public/bundle.js'
	},
	plugins: [
		auth0EnvReplace(),
		// Single CSS sink: imported sheets (materialize/xterm/global) + Svelte emitCss.
		// A second css({output:'bundle.css'}) after svelte left component styles in an
		// unlinked file while extra.css stayed stale — scoped hashes then diverged from
		// bundle.js (useon1 inventory grid + item-card overlay regression).
		svelte({
			compilerOptions: {
				dev: !production
			},
			emitCss: true
		}),
		css({ output: 'extra.css' }),

		resolve({
			browser: true,
			dedupe: ['svelte'],
			extensions: ['.svelte', '.mjs', '.js', '.json', '.node']
		}),
		commonjs(),

		// In dev mode, call `npm run start` once the bundle has been generated
		!production && serve(),

		// Watch the `public` directory and refresh the browser on changes
		!production && livereload('public'),

		// Minify for production
		production && terser()
	],
	watch: {
		clearScreen: false
	}
}, {
	// Standalone /activate page (SSH device sign-in). Component CSS stays inside
	// the JS so it never touches the play client's extra.css.
	input: 'src/activate/main.js',
	output: { sourcemap: true, format: 'iife', name: 'activatepage', file: 'public/activate.js' },
	plugins: [
		auth0EnvReplace(),
		svelte({ compilerOptions: { dev: !production, css: true }, emitCss: false }),
		resolve({ browser: true, dedupe: ['svelte'], extensions: ['.svelte', '.mjs', '.js', '.json'] }),
		commonjs(),
		production && terser()
	]
}, {
 input:'src/game/widgets/worldmapWorker.js',
 output:{file:'public/worldmap-worker.js',format:'iife',sourcemap:true},
 plugins:[resolve({browser:true}),production&&terser()]
}];

function serve() {
	let started = false;

	return {
		writeBundle() {
			if (!started) {
				started = true;

				require('child_process').spawn('npm', ['run', 'start', '--', '--dev'], {
					stdio: ['ignore', 'inherit', 'inherit'],
					shell: true
				});
			}
		}
	};
}
