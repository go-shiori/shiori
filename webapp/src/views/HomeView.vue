<script setup lang="ts">
import { ref, onMounted, computed, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { storeToRefs } from 'pinia';
import AppLayout from '@/components/layout/AppLayout.vue';
import Pagination from '@/components/ui/Pagination.vue';
import ViewSelector from '@/components/ui/ViewSelector.vue';
import BookmarkCard from '@/components/ui/BookmarkCard.vue';
import { useBookmarksStore } from '@/stores/bookmarks';
import { useAuthStore } from '@/stores/auth';
import AuthenticatedImage from '@/components/ui/AuthenticatedImage.vue';
import { ImageIcon, PencilIcon, TrashIcon, ArchiveIcon, BookIcon, FileTextIcon, ExternalLinkIcon } from '@/components/icons';

const bookmarksStore = useBookmarksStore();
const authStore = useAuthStore();
const router = useRouter();

const { bookmarks, isLoading, error, totalCount, currentPage, pageLimit } = storeToRefs(bookmarksStore);
const { fetchBookmarks } = bookmarksStore;

const searchKeyword = ref('');

// Initialize view from localStorage or default to 'list'
const getStoredView = (): 'list' | 'card' => {
  if (typeof window !== 'undefined') {
    const stored = localStorage.getItem('shiori-view-preference');
    return (stored === 'list' || stored === 'card') ? stored : 'list';
  }
  return 'list';
};

const currentView = ref<'list' | 'card'>(getStoredView());
const isMobile = ref(false);

// Detect mobile screen size
const checkMobile = () => {
  isMobile.value = window.innerWidth < 768; // md breakpoint
};

// Handle view change with persistence
const handleViewChange = (view: 'list' | 'card') => {
  currentView.value = view;
  // Store the preference in localStorage
  if (typeof window !== 'undefined') {
    localStorage.setItem('shiori-view-preference', view);
  }
};

// Computed property for effective view (force card on mobile)
const effectiveView = computed(() => {
  return isMobile.value ? 'card' : currentView.value;
});

// Fetch bookmarks on mount
onMounted(async () => {
  try {
    await fetchBookmarks();
    checkMobile(); // Check mobile on mount
    window.addEventListener('resize', checkMobile); // Listen for resize events
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
    await fetchBookmarks({ keyword: searchKeyword.value, page: 1 }); // Reset to page 1 when searching
  } catch (err) {
    console.error('Error searching bookmarks:', err);
  }
};

// Handle page change
const handlePageChange = async (page: number) => {
  try {
    await fetchBookmarks({ page, limit: pageLimit.value, keyword: searchKeyword.value });
  } catch (err) {
    console.error('Error changing page:', err);
  }
};

// Handle per page change
const handlePerPageChange = async (perPage: number) => {
  try {
    await fetchBookmarks({ page: 1, limit: perPage, keyword: searchKeyword.value });
  } catch (err) {
    console.error('Error changing items per page:', err);
  }
};

// Helper to get tag names from bookmark
const getBookmarkTags = (bookmark: any) => {
  // For now, return empty array as tags are separate in v1 API
  // Tags will need to be fetched separately per bookmark
  return [];
};

// Cleanup resize listener
onUnmounted(() => {
  window.removeEventListener('resize', checkMobile);
});

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
      <!-- View Selector -->
      <div class="flex justify-between items-center mb-4">
        <div class="flex items-center space-x-4">
          <!-- Hide view selector on mobile, force card view -->
          <ViewSelector v-if="!isMobile" :current-view="currentView" :on-view-change="handleViewChange" />
          <!-- Mobile: Force card view -->
          <div v-else class="text-sm text-gray-500 dark:text-gray-400">
            Card view
          </div>
        </div>
      </div>

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

      <!-- List View -->
      <ul v-else-if="effectiveView === 'list'" class="space-y-4">
        <li v-for="bookmark in bookmarks" :key="bookmark.id"
          class="bg-white dark:bg-gray-800 p-4 rounded-md shadow-sm hover:shadow-md transition-shadow cursor-pointer"
          @click="$router.push(`/bookmark/${bookmark.id}/content`)">
          <div class="flex gap-4">
            <!-- Thumbnail -->
            <div class="flex-shrink-0">
              <div v-if="bookmark.hasThumbnail"
                class="w-24 h-24 rounded-md overflow-hidden bg-gray-100 dark:bg-gray-700">
                <AuthenticatedImage :bookmark-id="bookmark.id || 0" :auth-token="authStore.token || undefined"
                  :alt="bookmark.title || 'Bookmark thumbnail'" class="w-full h-full" />
              </div>
              <div v-else class="w-24 h-24 rounded-md bg-gray-100 dark:bg-gray-700 flex items-center justify-center">
                <ImageIcon class="h-10 w-10 text-gray-400 dark:text-gray-500" />
              </div>
            </div>

            <!-- Content -->
            <div class="flex-1 min-w-0">
              <div class="flex justify-between items-start">
                <div class="flex items-start gap-2 flex-1 min-w-0">
                  <h3 class="text-blue-600 dark:text-blue-400 font-medium truncate">
                    {{ bookmark.title || bookmark.url }}
                  </h3>
                  <!-- Feature icons -->
                  <div class="flex items-center gap-1 flex-shrink-0">
                    <FileTextIcon v-if="bookmark.hasContent" class="h-4 w-4 text-gray-500 dark:text-gray-400" title="Has readable content" />
                    <ArchiveIcon v-if="bookmark.hasArchive" class="h-4 w-4 text-gray-500 dark:text-gray-400" title="Has archive" />
                    <BookIcon v-if="bookmark.hasEbook" class="h-4 w-4 text-gray-500 dark:text-gray-400" title="Has ebook" />
                  </div>
                </div>
                <div class="flex space-x-2 ml-4 flex-shrink-0">
                  <a v-if="bookmark.url" :href="bookmark.url" target="_blank"
                     @click.stop
                     class="text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300">
                    <span class="sr-only">Open original URL</span>
                    <ExternalLinkIcon class="h-5 w-5" />
                  </a>
                  <button @click.stop class="text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300">
                    <span class="sr-only">Edit</span>
                    <PencilIcon class="h-5 w-5" />
                  </button>
                  <button @click.stop class="text-gray-500 dark:text-gray-400 hover:text-red-500">
                    <span class="sr-only">Delete</span>
                    <TrashIcon class="h-5 w-5" />
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

      <!-- Card View -->
      <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-4">
        <BookmarkCard v-for="bookmark in bookmarks" :key="bookmark.id" :bookmark="bookmark"
          :auth-token="authStore.token || undefined" />
      </div>

      <!-- Pagination -->
      <Pagination v-if="totalCount > pageLimit" :current-page="currentPage" :total-items="totalCount"
        :items-per-page="pageLimit" @page-change="handlePageChange" @per-page-change="handlePerPageChange" />
    </div>
  </AppLayout>
</template>
