import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [vue()],
  server: {
    // TODO(deploy): 开发期局域网访问；生产用 Nginx/域名反代，不再依赖 Vite host
    host: true,
    port: 5173,
    proxy: {
      '/api': {
        // TODO(deploy): 生产环境由网关把 /api 反代到后端，此处仅本地开发
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/uploads': {
        // TODO(deploy): 同上，生产走域名反代
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
