import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// A URL da API vem de VITE_API_URL (veja .env.example).
// O servidor de desenvolvimento sobe em 5173, porta já liberada no CORS do backend.
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    strictPort: false,
  },
})
