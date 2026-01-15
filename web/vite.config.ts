import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  
  return {
    plugins: [react()],
    server: {
      host: '0.0.0.0',
      port: parseInt(env.NOFX_FRONTEND_PORT || '3300'),
      proxy: {
        '/api': {
          target: `http://${env.API_HOST || 'localhost'}:${env.NOFX_BACKEND_PORT || '8888'}`,
          changeOrigin: true,
        },
      },
    },
  }
})