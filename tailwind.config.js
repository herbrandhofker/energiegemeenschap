/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./internal/web/templates/**/*.html",
    "./internal/web/static/**/*.js",
  ],
  theme: {
    extend: {
      colors: {
        primary: 'var(--primary)',
        secondary: 'var(--secondary)',
        dark: 'var(--dark)',
      },
    },
  },
  plugins: [],
} 