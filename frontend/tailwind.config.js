/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Poppins', 'system-ui', '-apple-system', 'sans-serif'],
      },
      colors: {
        gray: {
          50: '#FFFFFF',
          100: '#D4D4D4',
          200: '#B3B3B3',
          300: '#8F8F8F',
          400: '#6B6B6B',
          500: '#4B4B4B',
          600: '#3B3B3B',
          700: '#2B2B2B',
          800: '#1B1B1B',
          900: '#0B0B0B',
        },
        success: {
          DEFAULT: '#22C55E',
          light: '#4ADE80',
          dark: '#16A34A',
        },
        warning: {
          DEFAULT: '#EAB308',
          light: '#FACC15',
          dark: '#CA8A04',
        },
        error: {
          DEFAULT: '#EF4444',
          light: '#F87171',
          dark: '#DC2626',
        },
        info: {
          DEFAULT: '#3B82F6',
          light: '#60A5FA',
          dark: '#2563EB',
        },
      },
    },
  },
  plugins: [],
}

