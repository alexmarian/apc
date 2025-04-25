import { createRouter, createWebHistory } from 'vue-router'
import { routes } from 'vue-router/auto-routes'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()

  // Check if route requires authentication
  const isAuthPage = ['/login', '/register'].includes(to.path)

  if (!isAuthPage && !authStore.isAuthenticated) {
    next('/login')
  } else if (isAuthPage && authStore.isAuthenticated) {
    next('/')
  } else {
    next()
  }
})
export default router
