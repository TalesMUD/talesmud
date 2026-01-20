import svelte from 'rollup-plugin-svelte';
import resolve from '@rollup/plugin-node-resolve';
import commonjs from '@rollup/plugin-commonjs';
import livereload from 'rollup-plugin-livereload';
import { terser } from 'rollup-plugin-terser';
import css from "rollup-plugin-css-only";

const production = !process.env.ROLLUP_WATCH;

export default {
	input: 'src/main.js',
	output: {
		sourcemap: true,
		format: 'iife',
		name: 'mudclient',
		file: 'public/bundle.js'
	},
	plugins: [
		css({ output: "extra.css" }),

		svelte({
			compilerOptions: {
				dev: !production
			},
			emitCss: true
		}),
		css({ output: 'bundle.css' }),

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
};

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
