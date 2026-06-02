import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createI18n } from 'vue-i18n'
import App from './App.vue'
import router from './router'

const storedLocale = localStorage.getItem('userLocale')
const browserLocale = (navigator.languages?.[0] ?? navigator.language ?? 'ro').split('-')[0]
const locale = ['ro', 'ru'].includes(storedLocale ?? '') ? storedLocale! : ['ro', 'ru'].includes(browserLocale) ? browserLocale : 'ro'

const i18n = createI18n({
  legacy: false,
  locale,
  fallbackLocale: 'ro',
  messages: {
    ro: {
      linkInvalid: 'Link invalid sau expirat',
      linkInvalidDesc: 'Acest link de vot nu mai este valid. Contactați administratorul asociației pentru o nouă invitație.',
      networkError: 'Eroare de rețea',
      notStartedTitle: 'Votul nu a început',
      notStartedDesc: 'Această adunare este {status}. Reveniți când votul se deschide.',
      endedTitle: 'Votul s-a încheiat',
      endedDesc: 'Votul pentru această adunare s-a închis. Rezultatele sunt în curs de numărare.',
      unavailableTitle: 'Adunare indisponibilă',
      unavailableDesc: 'Această adunare se află într-o stare neașteptată: {status}.',
    },
    ru: {
      linkInvalid: 'Ссылка недействительна или устарела',
      linkInvalidDesc: 'Эта ссылка для голосования больше не действительна. Обратитесь к администратору ассоциации за новым приглашением.',
      networkError: 'Ошибка сети',
      notStartedTitle: 'Голосование ещё не началось',
      notStartedDesc: 'Это собрание находится в статусе {status}. Вернитесь, когда откроется голосование.',
      endedTitle: 'Голосование завершено',
      endedDesc: 'Голосование для этого собрания закрыто. Результаты подсчитываются.',
      unavailableTitle: 'Собрание недоступно',
      unavailableDesc: 'Это собрание находится в неожиданном состоянии: {status}.',
    },
  },
})

createApp(App)
  .use(createPinia())
  .use(router)
  .use(i18n)
  .mount('#app')
