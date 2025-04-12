import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { resolve } from 'path'

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => ({
  plugins: [react()],
  define: {
    'process.env.NODE_ENV': JSON.stringify(mode)
  },
  server: {
    port: 5173,
    host: '0.0.0.0',
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
        secure: false,
        ws: true,
      }
    }
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    assetsDir: 'assets',
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true
      }
    },
    rollupOptions: {
      output: {
        manualChunks: (id) => {
          // 核心 React 相关库
          if (id.includes('node_modules/react') || 
              id.includes('node_modules/react-dom') || 
              id.includes('node_modules/react-router-dom')) {
            return 'vendor-react';
          }
          
          // Ant Design 相关库
          if (id.includes('node_modules/antd') || 
              id.includes('node_modules/@ant-design')) {
            return 'vendor-antd';
          }
          
          // 其他 UI 相关库
          if (id.includes('node_modules/@emotion') || 
              id.includes('node_modules/framer-motion') ||
              id.includes('node_modules/styled-components')) {
            return 'vendor-ui';
          }
          
          // 工具库
          if (id.includes('node_modules/lodash') || 
              id.includes('node_modules/axios') ||
              id.includes('node_modules/dayjs') ||
              id.includes('node_modules/i18next')) {
            return 'vendor-utils';
          }
          
          // 图表相关库
          if (id.includes('node_modules/echarts') || 
              id.includes('node_modules/d3') ||
              id.includes('node_modules/chart.js')) {
            return 'vendor-charts';
          }
          
          // 其他第三方库
          if (id.includes('node_modules')) {
            return 'vendor-others';
          }
        }
      }
    },
    sourcemap: false,
    cssCodeSplit: true,
    cssMinify: true,
    chunkSizeWarningLimit: 2000,
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  }
}))
