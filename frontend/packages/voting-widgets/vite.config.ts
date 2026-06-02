import { resolve } from 'path'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  build: {
    lib: {
      entry: resolve(__dirname, 'src/index.ts'),
      name: 'VotingWidgets',
      fileName: 'voting-widgets'
    },
    rollupOptions: {
      external: ['vue', 'naive-ui', 'vue-i18n'],
      output: {
        globals: {
          vue: 'Vue',
          'naive-ui': 'NaiveUI',
          'vue-i18n': 'VueI18n'
        }
      }
    }
  }
})
