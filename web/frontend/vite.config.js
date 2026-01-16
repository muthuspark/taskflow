import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  base: '/taskflow/',
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/taskflow/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/taskflow/setup': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/health': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/taskflow-app': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      }
    }
  },
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    minify: false,
  }
})
