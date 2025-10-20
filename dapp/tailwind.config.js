/** @type {import('tailwindcss').Config} */
import config from 'konsta/config';

export default config({
  konsta: {
  },
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      keyframes: {
        'swing': {
          '0%,100%' : { transform: 'rotate(15deg)' },
          '50%' : { transform: 'rotate(-15deg)' },
        },
        'zoom': {
          '0%,100%' : { transform: 'scale(1.0)' },
          '50%' : { transform: 'scale(0)' },
        },
        'blink': {
          '100%' : { transform: 'scaleY(1.0)' },
          '0%' : { transform: 'scaleY(0)' },
        },
        'rotation': {
          '0%' : { transform: 'rotate(0deg)' },
          '100%' : { transform: 'rotate(360deg)' },
        },
      },
      animation: {
        'swing': 'swing 0.2s',
        'zoom': 'zoom 0.2s',
        'blink': 'blink 0.4s',
        'rotation': 'rotation 1.0s'
      }
    },
  },
  plugins: [],
})
