<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';

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
  <header class="bg-white border-b border-gray-200 px-4 py-3 flex items-center justify-between">
    <!-- Logo -->
    <div class="flex items-center">
      <div class="text-red-500 font-bold text-xl">栞</div>
    </div>

    <!-- Search -->
    <div class="flex-1 mx-4">
      <div class="relative">
        <input type="text" placeholder="Search..."
          class="w-full border border-gray-300 rounded-md px-3 py-1 focus:outline-none focus:ring-2 focus:ring-red-500" />
      </div>
    </div>

    <!-- User Menu -->
    <div class="relative" ref="menuRef">
      <button @click="toggleMenu" class="text-gray-500 p-1 hover:bg-gray-100 rounded-full">
        <!-- User menu icon (consistent across mobile and desktop) -->
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M5.121 17.804A13.937 13.937 0 0112 16c2.5 0 4.847.655 6.879 1.804M15 10a3 3 0 11-6 0 3 3 0 016 0zm6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      </button>

      <!-- Dropdown Menu -->
      <div v-if="isMenuOpen"
        class="absolute right-0 mt-2 w-48 bg-white rounded-md shadow-lg py-1 z-50 border border-gray-200">
        <div class="px-4 py-2 text-sm text-gray-500 border-b border-gray-200">
          <div class="font-medium">{{ authStore.user?.username || 'User' }}</div>
        </div>
        <router-link to="/settings" class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">
          Settings
        </router-link>
        <button @click="handleLogout" class="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">
          Logout
        </button>
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
