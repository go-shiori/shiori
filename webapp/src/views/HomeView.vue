<script setup lang="ts">
import { ref } from 'vue';
import AppLayout from '@/components/layout/AppLayout.vue';

const bookmarks = ref([
  { id: 1, title: 'Example Bookmark 1', url: 'https://example.com/1', tags: ['example', 'first'] },
  { id: 2, title: 'Example Bookmark 2', url: 'https://example.com/2', tags: ['example', 'second'] },
  { id: 3, title: 'Example Bookmark 3', url: 'https://example.com/3', tags: ['example', 'third'] },
]);
</script>

<template>
  <AppLayout>
    <template #header>
      <div class="flex justify-between items-center">
        <h1 class="text-xl font-bold text-gray-800 dark:text-white">My Bookmarks</h1>
        <div class="flex space-x-2">
          <button class="bg-red-500 text-white px-3 py-1 rounded-md hover:bg-red-600">
            Add Bookmark
          </button>
          <div class="relative">
            <input type="text" placeholder="Search..."
              class="border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md px-3 py-1 focus:outline-none focus:ring-2 focus:ring-red-500" />
          </div>
        </div>
      </div>
    </template>

    <div class="mt-6">
      <ul class="space-y-4">
        <li v-for="bookmark in bookmarks" :key="bookmark.id"
          class="bg-white dark:bg-gray-800 p-4 rounded-md shadow-sm hover:shadow-md transition-shadow">
          <div class="flex justify-between">
            <a :href="bookmark.url" target="_blank"
              class="text-blue-600 dark:text-blue-400 hover:underline font-medium">{{
                bookmark.title }}</a>
            <div class="flex space-x-2">
              <button class="text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300">
                <span class="sr-only">Edit</span>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                  <path
                    d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z" />
                </svg>
              </button>
              <button class="text-gray-500 dark:text-gray-400 hover:text-red-500">
                <span class="sr-only">Delete</span>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                  <path fill-rule="evenodd"
                    d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z"
                    clip-rule="evenodd" />
                </svg>
              </button>
            </div>
          </div>
          <div class="text-gray-500 dark:text-gray-400 text-sm mt-1 truncate">{{ bookmark.url }}</div>
          <div class="mt-2 flex flex-wrap gap-1">
            <span v-for="tag in bookmark.tags" :key="tag"
              class="bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 text-xs px-2 py-1 rounded-full">
              {{ tag }}
            </span>
          </div>
        </li>
      </ul>
    </div>
  </AppLayout>
</template>
