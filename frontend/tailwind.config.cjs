module.exports = {
  content: ['./index.html', './src/**/*.{js,jsx}'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['"IBM Plex Sans"', '"Noto Sans SC"', 'sans-serif'],
      },
      boxShadow: {
        card: '0 16px 35px -18px rgba(15, 23, 42, 0.35)',
      },
      backgroundImage: {
        board: 'radial-gradient(circle at 10% 10%, rgba(52, 211, 153, 0.28), transparent 34%), radial-gradient(circle at 85% 0%, rgba(56, 189, 248, 0.25), transparent 38%), linear-gradient(140deg, #f8fafc 0%, #ecfeff 45%, #f0fdf4 100%)',
      },
    },
  },
  plugins: [require('daisyui')],
  daisyui: {
    themes: [
      {
        quizlab: {
          primary: '#0f766e',
          secondary: '#0369a1',
          accent: '#f59e0b',
          neutral: '#1f2937',
          'base-100': '#ffffff',
          'base-200': '#f1f5f9',
          info: '#0284c7',
          success: '#15803d',
          warning: '#d97706',
          error: '#dc2626',
        },
      },
    ],
  },
};
