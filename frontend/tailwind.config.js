/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}", // Scan React/TS files
    "./src/index.css", // Explicitly scan index.css for @apply usage
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#eef2ff',
          100: '#e0e7ff',
          200: '#c7d2fe',
          300: '#a5b4fc',
          400: '#818cf8',
          500: '#6366f1',
          600: '#4f46e5',
          700: '#4338ca',
          800: '#3730a3',
          900: '#312e81',
          950: '#1e1b4b',
        },
        secondary: {
          50: '#f0fdfa',
          100: '#ccfbf1',
          200: '#99f6e4',
          300: '#5eead4',
          400: '#2dd4bf',
          500: '#14b8a6',
          600: '#0d9488',
          700: '#0f766e',
          800: '#115e59',
          900: '#134e4a',
          950: '#042f2e',
        },
        gray: {
          750: '#2d3748', // 添加额外的灰色深度
        }
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        display: ['Poppins', 'system-ui', 'sans-serif'],
      },
      boxShadow: {
        'inner-lg': 'inset 0 2px 4px 0 rgba(0, 0, 0, 0.06)',
        'soft': '0 2px 15px -3px rgba(0, 0, 0, 0.07), 0 10px 20px -2px rgba(0, 0, 0, 0.04)',
      },
      borderRadius: {
        'xl': '1rem',
        '2xl': '1.5rem',
      },
      scale: {
        '98': '.98',
        '105': '1.05',
      },
      animation: {
        'bounce-slow': 'bounce 3s infinite',
      },
      typography: (theme) => ({
        DEFAULT: {
          css: {
            code: {
              color: theme('colors.primary.500'),
              fontWeight: '500',
              backgroundColor: theme('colors.gray.100'),
              borderRadius: theme('borderRadius.md'),
              padding: '0.2em 0.4em',
            },
            'code::before': {
              content: '""',
            },
            'code::after': {
              content: '""',
            },
            pre: {
              backgroundColor: theme('colors.gray.800'),
              borderRadius: theme('borderRadius.lg'),
              padding: theme('spacing.6'),
              margin: '0',
            },
            a: {
              color: theme('colors.primary.600'),
              textDecoration: 'none',
              fontWeight: '500',
              '&:hover': {
                textDecoration: 'underline',
              },
            },
            h3: {
              fontSize: theme('fontSize.lg')[0],
              marginTop: theme('spacing.6'),
              marginBottom: theme('spacing.2'),
            },
          },
        },
        dark: {
          css: {
            color: theme('colors.gray.300'),
            code: {
              color: theme('colors.primary.400'),
              backgroundColor: theme('colors.gray.700'),
            },
            a: {
              color: theme('colors.primary.400'),
            },
          },
        },
      }),
    },
  },
  plugins: [
    require('@tailwindcss/typography'),
  ],
}