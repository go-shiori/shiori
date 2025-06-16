import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw, NavigationGuardNext as NavigationGuard, RouteLocationNormalized } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'
import { useAuthStore } from '@/stores/auth'

const routes: Array<RouteRecordRaw> = [
  {
    path: '/',
    redirect: '/home'
  },
  {
    path: '/home',
    name: 'home',
    component: HomeView,
    meta: { requiresAuth: true }
  },
  {
    path: '/login',
    name: 'login',
    component: LoginView,
    props: (route) => ({ dst: route.query.dst })
  },
  {
    path: '/tags',
    name: 'tags',
    component: () => import('../views/TagsView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/folders',
    name: 'folders',
    component: () => import('../views/FoldersView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/archive',
    name: 'archive',
    component: () => import('../views/ArchiveView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/settings',
    name: 'settings',
    component: () => import('../views/SettingsView.vue'),
    meta: { requiresAuth: true }
  },
  // Redirect any unmatched routes to home (which will redirect to login if not authenticated)
  {
    path: '/:pathMatch(.*)*',
    redirect: '/home'
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

// Navigation guard
router.beforeEach(async (to: RouteLocationNormalized, from: RouteLocationNormalized, next: NavigationGuard) => {
  const authStore = useAuthStore()

  // Check if the route requires authentication
  if (to.matched.some((record) => record.meta.requiresAuth)) {
    // If we have a token, validate it
    if (authStore.token) {
      const isValid = await authStore.validateToken()

      if (isValid) {
        // Token is valid, proceed to the requested route
        next()
      } else {
        // Token is invalid, redirect to login with destination
        const destination = to.fullPath
        authStore.setRedirectDestination(destination)
        next({
          name: 'login',
          query: { dst: destination }
        })
      }
    } else {
      // No token, redirect to login with destination
      const destination = to.fullPath
      authStore.setRedirectDestination(destination)
      next({
        name: 'login',
        query: { dst: destination }
      })
    }
  } else {
    // Route doesn't require auth, proceed
    next()
  }
})

export default router
