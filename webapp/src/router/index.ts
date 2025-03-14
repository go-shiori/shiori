import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw, NavigationGuardNext as NavigationGuard, RouteLocationNormalized } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'

// Simple auth guard
const isAuthenticated = (): boolean => {
  // For now, just check if there's a token in localStorage
  return !!localStorage.getItem('token')
}

const routes: Array<RouteRecordRaw> = [
  {
    path: '/',
    redirect: '/login'
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
    component: LoginView
  },
  {
    path: '/tags',
    name: 'tags',
    component: () => import('../views/TagsView.vue'),
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
  // Redirect any unmatched routes to login
  {
    path: '/:pathMatch(.*)*',
    redirect: '/login'
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

// Navigation guard
router.beforeEach((to: RouteLocationNormalized, from: RouteLocationNormalized, next) => {
  if (to.matched.some((record: any) => record.meta.requiresAuth)) {
    if (!isAuthenticated()) {
      next({ name: 'login' })
    } else {
      next()
    }
  } else {
    next()
  }
})

export default router
