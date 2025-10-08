import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw, NavigationGuardNext as NavigationGuard, RouteLocationNormalized } from 'vue-router'
import LibraryView from '../views/LibraryView.vue'
import LoginView from '../views/LoginView.vue'
import { useAuthStore } from '@/stores/auth'

const routes: Array<RouteRecordRaw> = [
  {
    path: '/',
    redirect: '/library'
  },
  {
    path: '/library',
    name: 'library',
    component: LibraryView,
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
    path: '/settings',
    name: 'settings',
    component: () => import('../views/SettingsView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/bookmark/:id/content',
    name: 'bookmark-content',
    component: () => import('../views/BookmarkContentView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/bookmark/:id/archive',
    name: 'bookmark-archive',
    component: () => import('../views/BookmarkArchiveView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/add-bookmark',
    name: 'add-bookmark',
    component: () => import('../views/AddBookmarkView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/add-tag',
    name: 'add-tag',
    component: () => import('../views/AddTagView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/test-tag-selector',
    name: 'test-tag-selector',
    component: () => import('../views/TagSelectorTest.vue'),
    meta: { requiresAuth: true }
  },
  // Redirect any unmatched routes to library (which will redirect to login if not authenticated)
  {
    path: '/:pathMatch(.*)*',
    redirect: '/library'
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
