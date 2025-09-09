<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { useI18n } from 'vue-i18n';
import LanguageSelector from './LanguageSelector.vue';

const { t } = useI18n();
const isMenuOpen = ref(false);
const authStore = useAuthStore();
const router = useRouter();
const menuRef = ref<HTMLElement | null>(null);

// Toggle menu
const toggleMenu = (event: MouseEvent) => {
  event.stopPropagation(); // Prevent event from bubbling up
  isMenuOpen.value = !isMenuOpen.value;
};

// Handle logout
const handleLogout = async () => {
  await authStore.logout();
  isMenuOpen.value = false;
  router.push('/login');
};

// Close menu when clicking outside
const handleClickOutside = (event: MouseEvent) => {
  if (menuRef.value && !menuRef.value.contains(event.target as Node)) {
    isMenuOpen.value = false;
  }
};

// Add and remove event listeners
onMounted(() => {
  document.addEventListener('click', handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside);
});
</script>

<template>
  <header
    class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 px-4 py-3 flex items-center justify-between">
    <!-- Logo -->
    <div class="flex items-center">
      <div class="text-red-500 font-bold text-xl">栞</div>
    </div>

    <!-- Search -->
    <div class="flex-1 mx-4">
      <div class="relative">
        <input type="text" :placeholder="t('common.search')"
          class="w-full border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md px-3 py-1 focus:outline-none focus:ring-2 focus:ring-red-500" />
      </div>
    </div>

    <!-- Actions -->
    <div class="flex items-center space-x-2">
      <!-- Language Selector -->
      <LanguageSelector />

      <!-- User Menu -->
      <div class="relative" ref="menuRef">
        <button @click="toggleMenu"
          class="text-gray-500 dark:text-gray-300 p-1 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-full">
          <!-- User menu icon (consistent across mobile and desktop) -->
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M5.121 17.804A13.937 13.937 0 0112 16c2.5 0 4.847.655 6.879 1.804M15 10a3 3 0 11-6 0 3 3 0 016 0zm6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </button>

        <!-- Dropdown menu -->
        <div v-if="isMenuOpen"
          class="absolute right-0 mt-2 w-48 bg-white dark:bg-gray-800 rounded-md shadow-lg py-1 z-10 border border-gray-200 dark:border-gray-700">
          <div class="px-4 py-2 text-sm text-gray-700 dark:text-gray-300">
            {{ authStore.user?.username || 'User' }}
          </div>
          <hr class="border-gray-200 dark:border-gray-700">
          <a href="#" @click.prevent="handleLogout"
            class="block px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700">
            {{ t('auth.logout') }}
          </a>
        </div>
      </div>
    </div>
  </header>
</template>

<style scoped>
/* Ensure the dropdown is visible and positioned correctly */
.relative {
  position: relative;
}
</style>
