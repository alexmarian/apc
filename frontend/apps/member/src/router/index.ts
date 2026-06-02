import { createRouter, createWebHistory } from 'vue-router'
import TokenView from '@/views/TokenView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/:token', component: TokenView },
    { path: '/', redirect: '/invalid' },
  ],
})

export default router
