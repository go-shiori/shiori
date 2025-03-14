<script setup lang="ts">
import { RouterView } from 'vue-router'
import { onMounted, ref } from 'vue'
import { useAuthStore } from './stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()
const isInitializing = ref(true)

onMounted(async () => {
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
</script>

<template>
  <div class="min-h-screen h-full flex flex-col bg-[var(--background-color)] text-[var(--text-color)]">
    <div v-if="isInitializing"
      class="fixed inset-0 flex items-center justify-center bg-white dark:bg-gray-900 bg-opacity-80 dark:bg-opacity-80 z-50">
      <div class="text-center">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-red-500 mx-auto mb-2"></div>
        <p class="text-gray-700 dark:text-gray-300">Loading...</p>
      </div>
    </div>
    <RouterView v-else class="flex-1" />
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
