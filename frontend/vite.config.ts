import react from '@vitejs/plugin-react'
import { resolve } from 'path'
import { defineConfig } from 'vite'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  base: './',
  
  // 开发服务器配置
  server: {
    port: 1420,
    host: '0.0.0.0',
    strictPort: true,
  },

  // 构建配置
  build: {
    outDir: 'dist',
    sourcemap: false,
    minify: 'esbuild',
    target: 'esnext',
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'index.html'),
      },
    },
  },

  // 路径别名
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },

  // 环境变量前缀
  envPrefix: 'VITE_',

  // 清除控制台
  clearScreen: false,
})
