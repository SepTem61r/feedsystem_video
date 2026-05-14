import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/static': 'http://localhost:8080',
      '/account': 'http://localhost:8080',
      '/video': 'http://localhost:8080',
      '/like': 'http://localhost:8080',
      '/comment': 'http://localhost:8080',
      '/social': 'http://localhost:8080',
      '/feed': 'http://localhost:8080',
    },
  },
})
