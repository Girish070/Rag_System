/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        dark: "#0d1117",
        card: "#161b22",
        accent: "#58a6ff",
        border: "#30363d",
      }
    },
  },
  plugins: [],
}
