/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: "class",
  content: ["./index.html", "./src/**/*.{svelte,js}"],
  theme: {
    extend: {
      colors: {
        primary: "#10b981",
        "background-light": "#f8fafc",
        "background-dark": "#0f172a",
        "card-light": "#ffffff",
        "card-dark": "#1e293b",
        "editor-dark": "#020617",
        "accent-red": "#ef4444",
      },
      fontFamily: {
        display: ["Inter", "sans-serif"],
        mono: ["JetBrains Mono", "monospace"],
      },
      borderRadius: {
        DEFAULT: "0.5rem",
      },
    },
  },
  plugins: [require("@tailwindcss/forms"), require("@tailwindcss/typography")],
};
