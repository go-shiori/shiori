<script setup lang="ts">
import { RouterView } from 'vue-router'
import { onMounted, ref, watch } from 'vue'
import { useAuthStore } from './stores/auth'
import { useRouter } from 'vue-router'
import ToastContainer from '@/components/ui/ToastContainer.vue'
import { useI18n } from 'vue-i18n'
import { useTheme } from '@/composables/useTheme'

const authStore = useAuthStore()
const router = useRouter()
const isInitializing = ref(true)
const { t } = useI18n()
const { apply, init, destroy } = useTheme()

onMounted(async () => {
  // Apply theme immediately before any async auth work
  const pref = (authStore.user?.config?.Theme as any) || (localStorage.getItem('shiori-theme') as any) || 'system'
  apply(pref)
  init()
  // If we have a token, validate it
  if (authStore.token) {
    try {
      // Validate the token by fetching user info
      await authStore.validateToken()
    } catch (error) {
      console.error('Failed to validate token:', error)
    }
  }
  isInitializing.value = false
})

// React to user config theme changes
watch(() => authStore.user?.config?.Theme, (newPref) => {
  apply((newPref as any) || 'system')
})
</script>

<template>
  <div class="min-h-screen h-full flex flex-col bg-[var(--background-color)] text-[var(--text-color)]">
    <div v-if="isInitializing"
      class="fixed inset-0 flex items-center justify-center bg-white dark:bg-gray-900 bg-opacity-80 dark:bg-opacity-80 z-50">
      <div class="text-center">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-red-500 mx-auto mb-2"></div>
        <p class="text-gray-700 dark:text-gray-300">{{ t('common.loading') }}</p>
      </div>
    </div>
    <RouterView v-else class="flex-1" />

    <!-- Toast Notifications -->
    <ToastContainer />
  </div>
</template>

<style>
/* Global styles */
html,
body,
#app {
  height: 100%;
  min-height: 100vh;
  margin: 0;
  padding: 0;
}

body {
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
  color: var(--text-color);
  background-color: var(--background-color);
}
</style>
