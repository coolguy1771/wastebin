import { sveltekit } from '@sveltejs/kit/vite'

const config = {
  plugins: [sveltekit()],
  optimizeDeps: {
    include: ['highlight.js', 'highlight.js/lib/core'],
  },
}

export default config
