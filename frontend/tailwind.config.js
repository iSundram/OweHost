/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'blue-eclipse': {
          DEFAULT: '#2E5E99',
          50: '#E7F0FA',
          100: '#E7F0FA',
          200: '#7BA4D0',
          300: '#7BA4D0',
          400: '#7BA4D0',
          500: '#2E5E99',
          600: '#2E5E99',
          700: '#0D2440',
          800: '#0D2440',
          900: '#0D2440',
        },
      },
    },
  },
  plugins: [],
}

