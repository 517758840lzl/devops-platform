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
        // SSE：避免代理缓冲构建状态推送
        configure: (proxy) => {
          proxy.on('proxyRes', (proxyRes) => {
            if (proxyRes.headers['content-type']?.includes('text/event-stream')) {
              proxyRes.headers['cache-control'] = 'no-cache'
              proxyRes.headers['x-accel-buffering'] = 'no'
            }
          })
        },
      },
      '/uploads': {
        // TODO(deploy): 同上，生产走域名反代
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
