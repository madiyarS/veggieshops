/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        veggie: {
          green: '#2d5a27',
          light: '#7cb342',
          dark: '#1b3d17',
        },
      },
    },
  },
  plugins: [],
}
