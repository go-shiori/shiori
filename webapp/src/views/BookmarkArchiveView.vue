<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useBookmarksStore } from '@/stores/bookmarks';
import { useAuthStore } from '@/stores/auth';
import AppLayout from '@/components/layout/AppLayout.vue';
import { ExternalLinkIcon, DownloadIcon, ArrowLeftIcon, FileTextIcon } from '@/components/icons';
import type { ModelBookmarkDTO } from '@/client';

const route = useRoute();
const router = useRouter();
const bookmarksStore = useBookmarksStore();
const authStore = useAuthStore();

const bookmark = ref<ModelBookmarkDTO | null>(null);
const isLoading = ref(true);
const error = ref<string | null>(null);

const bookmarkId = computed(() => {
  const id = route.params.id;
  return typeof id === 'string' ? parseInt(id) : null;
});

const hasContent = computed(() => bookmark.value?.hasContent ?? false);
const hasEbook = computed(() => bookmark.value?.hasEbook ?? false);

const loadBookmark = async () => {
  if (!bookmarkId.value) {
    error.value = 'Invalid bookmark ID';
    isLoading.value = false;
    return;
  }

  try {
    isLoading.value = true;
    error.value = null;

    bookmark.value = await bookmarksStore.getBookmark(bookmarkId.value);

    if (!bookmark.value?.hasArchive) {
      error.value = 'No archive available for this bookmark';
    }
  } catch (err) {
    error.value = 'Failed to load bookmark';
    console.error('Error loading bookmark:', err);
  } finally {
    isLoading.value = false;
  }
};

const goBack = () => {
  router.back();
};

const goToOriginal = () => {
  if (bookmark.value?.url) {
    window.open(bookmark.value.url, '_blank');
  }
};

const goToContent = () => {
  if (bookmarkId.value) {
    router.push(`/bookmark/${bookmarkId.value}/content`);
  }
};

const downloadEbook = () => {
  if (bookmarkId.value) {
    const ebookUrl = `/bookmark/${bookmarkId.value}/ebook`;
    const link = document.createElement('a');
    link.href = ebookUrl;
    link.download = `${bookmark.value?.title || 'bookmark'}.epub`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  }
};

onMounted(() => {
  loadBookmark();
});
</script>

<template>
  <AppLayout>
    <div class="max-w-6xl mx-auto p-6">
      <!-- Loading state -->
      <div v-if="isLoading" class="text-center py-12">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
        <p class="text-gray-600 dark:text-gray-400">Loading archive...</p>
      </div>

      <!-- Error state -->
      <div v-else-if="error" class="text-center py-12">
        <p class="text-red-600 dark:text-red-400 text-lg">{{ error }}</p>
        <button @click="goBack" class="mt-4 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">
          Go Back
        </button>
      </div>

      <!-- Archive Content -->
      <div v-else-if="bookmark" class="space-y-6">
        <!-- Header -->
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6">
          <div class="flex items-start justify-between mb-4">
            <button @click="goBack" class="flex items-center text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200">
              <ArrowLeftIcon class="h-5 w-5 mr-2" />
              Back
            </button>

            <div class="flex items-center gap-2">
              <button
                v-if="hasContent"
                @click="goToContent"
                class="flex items-center px-3 py-2 bg-green-500 text-white rounded hover:bg-green-600 transition-colors"
              >
                <FileTextIcon class="h-4 w-4 mr-2" />
                View Content
              </button>

              <button
                v-if="hasEbook"
                @click="downloadEbook"
                class="flex items-center px-3 py-2 bg-purple-500 text-white rounded hover:bg-purple-600 transition-colors"
              >
                <DownloadIcon class="h-4 w-4 mr-2" />
                Download eBook
              </button>
            </div>
          </div>

          <h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-2">
            {{ bookmark.title }}
            <button
              v-if="bookmark.url"
              @click="goToOriginal"
              class="ml-2 text-blue-500 hover:text-blue-600 transition-colors"
              title="Open original URL"
            >
              <ExternalLinkIcon class="h-5 w-5 inline" />
            </button>
          </h1>

          <p v-if="bookmark.excerpt" class="text-gray-600 dark:text-gray-400 mb-4">
            {{ bookmark.excerpt }}
          </p>

          <div class="flex items-center gap-4 text-sm text-gray-500 dark:text-gray-400">
            <span>By {{ bookmark.author || 'Unknown' }}</span>
            <span>•</span>
            <span>{{ new Date(bookmark.createdAt || '').toLocaleDateString() }}</span>
            <span>•</span>
            <span class="flex items-center">
              <ArchiveIcon class="h-4 w-4 mr-1" />
              Archived Version
            </span>
          </div>
        </div>

        <!-- Archive Frame -->
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow-sm overflow-hidden">
          <div class="p-4 border-b border-gray-200 dark:border-gray-700">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">Archived Content</h2>
            <p class="text-sm text-gray-600 dark:text-gray-400">This is the offline archive of the original page</p>
          </div>

          <div class="h-screen">
            <iframe
              :src="`/bookmark/${bookmarkId}/archive/file/`"
              class="w-full h-full border-0"
              title="Archived content"
            ></iframe>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>
