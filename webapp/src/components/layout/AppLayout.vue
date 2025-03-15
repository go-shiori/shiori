<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import Sidebar from './Sidebar.vue';
import TopBar from './TopBar.vue';

const isMobile = ref(false);

const checkMobile = () => {
  isMobile.value = window.innerWidth < 768;
};

onMounted(() => {
  checkMobile();
  window.addEventListener('resize', checkMobile);
});

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile);
});
</script>

<template>
  <div class="min-h-screen flex flex-col bg-gray-100 dark:bg-gray-900">
    <!-- Mobile Top Bar (only visible on mobile) -->
    <TopBar v-if="isMobile" />

    <div class="flex flex-1">
      <!-- Sidebar (left on desktop, bottom on mobile) -->
      <Sidebar :is-mobile="isMobile" />

      <!-- Main Content -->
      <main class="flex-1 p-6 pb-24 md:pb-6 overflow-auto">
        <!-- Header slot for page-specific headers -->
        <header v-if="$slots.header" class="mb-6">
          <slot name="header"></slot>
        </header>

        <!-- Default slot for page content -->
        <slot></slot>
      </main>
    </div>

    <!-- Mobile Navigation (only visible on mobile) -->
    <nav v-if="isMobile"
      class="bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 fixed bottom-0 left-0 right-0 z-10">
      <!-- Mobile navigation content will be rendered by Sidebar component -->
    </nav>
  </div>
</template>

<style>
/* Ensure the layout takes up the full viewport height */
html,
body,
#app {
  height: 100%;
  min-height: 100vh;
}
</style>
