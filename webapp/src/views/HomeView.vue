<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { storeToRefs } from 'pinia';
import AppLayout from '@/components/layout/AppLayout.vue';
import { useBookmarksStore } from '@/stores/bookmarks';
import { useAuthStore } from '@/stores/auth';
import AuthenticatedImage from '@/components/ui/AuthenticatedImage.vue';

const bookmarksStore = useBookmarksStore();
const authStore = useAuthStore();
const router = useRouter();

const { bookmarks, isLoading, error } = storeToRefs(bookmarksStore);
const { fetchBookmarks } = bookmarksStore;

const searchKeyword = ref('');

// Fetch bookmarks on mount
onMounted(async () => {
  try {
    await fetchBookmarks();
  } catch (err) {
    console.error('Error loading bookmarks:', err);
    // Handle authentication errors
    if (err instanceof Error && err.message.includes('401')) {
      authStore.clearAuth();
      router.push('/login');
    }
  }
});

// Search bookmarks
const handleSearch = async () => {
  try {
    await fetchBookmarks({ keyword: searchKeyword.value });
  } catch (err) {
    console.error('Error searching bookmarks:', err);
  }
};

// Helper to get tag names from bookmark
const getBookmarkTags = (bookmark: any) => {
  // For now, return empty array as tags are separate in v1 API
  // Tags will need to be fetched separately per bookmark
  return [];
};

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
            <input v-model="searchKeyword" @keyup.enter="handleSearch" type="text" placeholder="Search..."
              class="border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-md px-3 py-1 focus:outline-none focus:ring-2 focus:ring-red-500" />
          </div>
        </div>
      </div>
    </template>

    <div class="mt-6">
      <!-- Loading state -->
      <div v-if="isLoading" class="text-center py-8">
        <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-red-500"></div>
        <p class="mt-2 text-gray-600 dark:text-gray-400">Loading bookmarks...</p>
      </div>

      <!-- Error state -->
      <div v-else-if="error" class="bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-300 p-4 rounded-md">
        {{ error }}
      </div>

      <!-- Empty state -->
      <div v-else-if="bookmarks.length === 0" class="text-center py-12">
        <p class="text-gray-600 dark:text-gray-400 text-lg">No bookmarks found</p>
        <p class="text-gray-500 dark:text-gray-500 text-sm mt-2">Create your first bookmark to get started</p>
      </div>

      <!-- Bookmarks list -->
      <ul v-else class="space-y-4">
        <li v-for="bookmark in bookmarks" :key="bookmark.id"
          class="bg-white dark:bg-gray-800 p-4 rounded-md shadow-sm hover:shadow-md transition-shadow">
          <div class="flex gap-4">
            <!-- Thumbnail -->
            <div class="flex-shrink-0">
              <div v-if="bookmark.hasThumbnail"
                class="w-24 h-24 rounded-md overflow-hidden bg-gray-100 dark:bg-gray-700">
                <AuthenticatedImage :bookmark-id="bookmark.id || 0" :auth-token="authStore.token || undefined"
                  :alt="bookmark.title || 'Bookmark thumbnail'" class="w-full h-full" />
              </div>
              <div v-else class="w-24 h-24 rounded-md bg-gray-100 dark:bg-gray-700 flex items-center justify-center">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-gray-400 dark:text-gray-500" fill="none"
                  viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
              </div>
            </div>

            <!-- Content -->
            <div class="flex-1 min-w-0">
              <div class="flex justify-between">
                <a :href="bookmark.url" target="_blank"
                  class="text-blue-600 dark:text-blue-400 hover:underline font-medium truncate">
                  {{ bookmark.title || bookmark.url }}
                </a>
                <div class="flex space-x-2 ml-4 flex-shrink-0">
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
              <div v-if="bookmark.excerpt" class="text-gray-600 dark:text-gray-400 text-sm mt-2 line-clamp-2">
                {{ bookmark.excerpt }}
              </div>
            </div>
          </div>
        </li>
      </ul>
    </div>
  </AppLayout>
</template>
