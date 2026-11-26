import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// 产物直接进 web/dist，由 Go 的 embed 打进二进制
export default defineConfig({
  plugins: [vue()],
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    chunkSizeWarningLimit: 1500,
  },
  server: {
    // 只给 pnpm dev 用；生产是后端直接发静态文件，不需要反向代理
    proxy: { '/api': 'http://127.0.0.1:8000' },
  },
})
