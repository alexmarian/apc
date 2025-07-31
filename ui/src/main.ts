import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import naive from 'naive-ui'
import i18n from '@/i18n'
import App from '@/App.vue'
import router from '@/router'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(i18n)
app.use(naive)

app.mount('#app')

// Import auth tests in development mode
if (import.meta.env.DEV) {
  import('@/utils/auth-test')
}
