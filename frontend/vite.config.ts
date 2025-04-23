import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: { // 添加服务器配置
    proxy: {
      // 将 /api/v1 开头的请求代理到 Go 后端
      '/api/v1': {
        target: 'http://localhost:8080', // Go 后端地址
        changeOrigin: true, // 对于虚拟主机站点是必需的
        // 通常不需要重写路径，因为后端 API 路径匹配
        // rewrite: (path) => path.replace(/^\/api\/v1/, '/api/v1')
      },
    },
  },
})
