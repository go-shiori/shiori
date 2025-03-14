<script setup lang="ts">
import { ref } from 'vue';

const isMenuOpen = ref(false);

const toggleMenu = () => {
  isMenuOpen.value = !isMenuOpen.value;
};
</script>

<template>
  <header class="bg-white border-b border-gray-200 px-4 py-3 flex items-center justify-between">
    <!-- Logo -->
    <div class="flex items-center">
      <div class="text-red-500 font-bold text-xl">栞 shiori</div>
    </div>

    <!-- Search -->
    <div class="flex-1 mx-4">
      <div class="relative">
        <input
          type="text"
          placeholder="Search..."
          class="w-full border border-gray-300 rounded-md px-3 py-1 focus:outline-none focus:ring-2 focus:ring-red-500"
        />
      </div>
    </div>

    <!-- User Menu -->
    <div>
      <button @click="toggleMenu" class="text-gray-500 p-1">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
        </svg>
      </button>

      <!-- Dropdown Menu -->
      <div v-if="isMenuOpen" class="absolute right-0 mt-2 w-48 bg-white rounded-md shadow-lg py-1 z-10">
        <a href="#" class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">Profile</a>
        <a href="#" class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">Settings</a>
        <a href="#" class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">Logout</a>
      </div>
    </div>
  </header>
</template>
