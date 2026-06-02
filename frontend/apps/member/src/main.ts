import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createI18n } from 'vue-i18n'
import App from './App.vue'
import router from './router'

const storedLocale = localStorage.getItem('userLocale')
const browserLocale = (navigator.languages?.[0] ?? navigator.language ?? 'ro').split('-')[0]
const locale = storedLocale || browserLocale

const i18n = createI18n({
  legacy: false,
  locale,
  fallbackLocale: 'en',
  messages: {},
})

createApp(App)
  .use(createPinia())
  .use(router)
  .use(i18n)
  .mount('#app')
