export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        black: '#000000',
        white: '#ffffff',
        gray: {
          dark: '#333333',
          DEFAULT: '#666666',
          medium: '#666666',
          light: '#cccccc',
          lighter: '#eeeeee',
        },
      },
      fontFamily: {
        mono: ['"Source Code Pro"', 'monospace'],
      },
      fontSize: {
        'xs': '0.75rem',
        'sm': '0.875rem',
        'base': '0.9375rem',
        'lg': '1rem',
        'xl': '1.125rem',
        '2xl': '1.25rem',
        '3xl': '1.5rem',
        '4xl': '2rem',
      },
      letterSpacing: {
        tight: '0.05em',
      },
      borderRadius: {
        'none': '0',
        DEFAULT: '0',
        'sm': '0',
        'md': '0',
        'lg': '0',
        'xl': '0',
        '2xl': '0',
        'full': '50%',
      },
    },
  },
  plugins: [],
}
