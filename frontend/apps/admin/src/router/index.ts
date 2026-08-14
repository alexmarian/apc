import { createRouter, createWebHistory } from 'vue-router'
import { routes } from 'vue-router/auto-routes'
import { useAuthStore } from '@/stores/auth'
import { registerRouterService } from '@/services/auth-service'
import { isAuthPath } from '@/router/authPages'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

// Register router service for logout navigation
registerRouterService(router)

router.beforeEach((to) => {
  const authStore = useAuthStore()
  const isAuthPage = isAuthPath(to.path)

  if (!isAuthPage && !authStore.isAuthenticated) {
    // Carry the destination so the user resumes where the session ended
    // instead of always landing on the dashboard.
    return { path: '/login', query: to.fullPath === '/' ? {} : { redirect: to.fullPath } }
  }

  if (isAuthPage && authStore.isAuthenticated) {
    return { path: '/' }
  }

  return true
})

export default router
