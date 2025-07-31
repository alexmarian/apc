import { createRouter, createWebHistory } from 'vue-router'
import { routes } from 'vue-router/auto-routes'
import { useAuthStore } from '@/stores/auth'
import { registerRouterService } from '@/services/auth-service'

const authPages = ['/login', '/register', '/reset']

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

// Register router service for logout navigation
registerRouterService(router)

router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()

  console.log(to.path)
  const isAuthPage = authPages.includes(to.path)

  to.meta.isAuthPage = isAuthPage

  if (!isAuthPage && !authStore.isAuthenticated) {
    next('/login')
  } else if (isAuthPage && authStore.isAuthenticated) {
    next('/')
  } else {
    next()
  }
})

export default router
