<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useBookmarksStore } from '@/stores/bookmarks';
import { useAuthStore } from '@/stores/auth';
import AppLayout from '@/components/layout/AppLayout.vue';
import { ArrowLeftIcon } from '@/components/icons';
import BookmarkDetailHeader from '@/components/BookmarkDetailHeader.vue';
import { useI18n } from 'vue-i18n';
import type { ModelBookmarkDTO } from '@/client';

const route = useRoute();
const router = useRouter();
const bookmarksStore = useBookmarksStore();
const authStore = useAuthStore();
const { t } = useI18n();

const bookmark = ref<ModelBookmarkDTO | null>(null);
const isLoading = ref(true);
const error = ref<string | null>(null);

const bookmarkId = computed(() => {
  const id = route.params.id;
  return typeof id === 'string' ? parseInt(id) : null;
});

const hasContent = computed(() => bookmark.value?.hasContent ?? false);
const hasArchive = computed(() => bookmark.value?.hasArchive ?? false);
const hasEbook = computed(() => bookmark.value?.hasEbook ?? false);

const loadBookmark = async () => {
  if (!bookmarkId.value) {
    error.value = t('bookmarks.error.invalid_id');
    isLoading.value = false;
    return;
  }

  try {
    isLoading.value = true;
    error.value = null;

    const bookmarkData = await bookmarksStore.getBookmarkData(bookmarkId.value);
    bookmark.value = await bookmarksStore.getBookmark(bookmarkId.value);
  } catch (err) {
    error.value = t('bookmarks.error.load_failed');
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

const goToArchive = () => {
  if (bookmarkId.value) {
    router.push(`/bookmark/${bookmarkId.value}/archive`);
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
    <div class="w-full">
      <!-- Loading state -->
      <div v-if="isLoading" class="text-center py-12">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
        <p class="text-gray-600 dark:text-gray-400">{{ t('common.loading') }}</p>
      </div>

      <!-- Error state -->
      <div v-else-if="error" class="text-center py-12">
        <p class="text-red-600 dark:text-red-400 text-lg">{{ error }}</p>
        <button @click="goBack" class="mt-4 px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600">
          {{ t('common.back') }}
        </button>
      </div>

      <!-- Content -->
      <div v-else-if="bookmark" class="space-y-6">
        <!-- Header -->
        <BookmarkDetailHeader :bookmark="bookmark" :show-download-ebook-button="true" :show-view-archive-button="true"
          @download-ebook="downloadEbook" @view-archive="goToArchive" @open-original="goToOriginal" />

        <!-- Content -->
        <div class=" max-w-4xl mx-auto p-6">
          <div v-if="hasContent" class="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6">
            <div class="prose-content max-w-none" v-html="bookmark.html"></div>
          </div>

          <!-- No content message -->
          <div v-else class="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6 text-center">
            <p class="text-gray-600 dark:text-gray-400 mb-4">{{ t('bookmarks.no_readable_content') }}</p>
            <button @click="goToOriginal"
              class="px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600 transition-colors">
              {{ t('bookmarks.view_original_page') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style>
.prose-content {
  color: rgb(17 24 39);
  /* text-gray-900 */
  line-height: 1.7;
  font-size: 1rem;
  max-width: none;
}

/* Dark mode support - both class-based and system preference */
.dark .prose-content {
  color: rgb(243 244 246);
  /* text-gray-100 */
  line-height: 1.7;
}

@media (prefers-color-scheme: dark) {
  .prose-content {
    color: rgb(243 244 246);
    /* text-gray-100 */
    line-height: 1.7;
  }
}

/* Specific overrides for text elements only */
.dark .prose-content p {
  color: rgb(243 244 246);
}

@media (prefers-color-scheme: dark) {
  .prose-content p {
    color: rgb(243 244 246);
  }
}

.prose-content h1,
.prose-content h2,
.prose-content h3,
.prose-content h4,
.prose-content h5,
.prose-content h6 {
  color: rgb(17 24 39);
  /* text-gray-900 */
  font-weight: 600;
  /* font-semibold */
  margin-bottom: 0.5rem;
  /* mb-2 */
  margin-top: 1rem;
  /* mt-4 */
}

.dark .prose-content h1,
.dark .prose-content h2,
.dark .prose-content h3,
.dark .prose-content h4,
.dark .prose-content h5,
.dark .prose-content h6 {
  color: rgb(243 244 246);
  /* text-gray-100 */
}

@media (prefers-color-scheme: dark) {

  .prose-content h1,
  .prose-content h2,
  .prose-content h3,
  .prose-content h4,
  .prose-content h5,
  .prose-content h6 {
    color: rgb(243 244 246);
    /* text-gray-100 */
  }
}

.prose-content h1 {
  font-size: 1.5rem;
  line-height: 2rem;
}

/* text-2xl */
.prose-content h2 {
  font-size: 1.25rem;
  line-height: 1.75rem;
}

/* text-xl */
.prose-content h3 {
  font-size: 1.125rem;
  line-height: 1.75rem;
}

/* text-lg */

.prose-content p {
  margin-bottom: 1rem;
  /* mb-4 */
  line-height: 1.7;
  /* leading-relaxed */
  color: inherit;
}

.dark .prose-content p {
  color: rgb(243 244 246);
}

@media (prefers-color-scheme: dark) {
  .prose-content p {
    color: rgb(243 244 246);
  }
}

.prose-content a {
  color: rgb(37 99 235);
  /* text-blue-600 */
  text-decoration: underline;
}

.dark .prose-content a {
  color: rgb(96 165 250);
  /* text-blue-400 */
}

.prose-content a:hover {
  color: rgb(30 64 175);
  /* hover:text-blue-800 */
}

.dark .prose-content a:hover {
  color: rgb(147 197 253);
  /* hover:text-blue-300 */
}

.prose-content blockquote {
  border-left: 4px solid rgb(209 213 219);
  /* border-l-4 border-gray-300 */
  background-color: rgb(249 250 251);
  /* bg-gray-50 */
  padding: 1rem;
  /* p-4 */
  margin: 1rem 0;
  /* my-4 */
  font-style: italic;
}

.dark .prose-content blockquote {
  border-left-color: rgb(75 85 99);
  /* border-gray-600 */
  background-color: rgb(55 65 81);
  /* bg-gray-700 */
}

.prose-content code {
  background-color: rgb(243 244 246);
  /* bg-gray-100 */
  color: rgb(31 41 55);
  /* text-gray-800 */
  padding: 0.125rem 0.25rem;
  /* px-1 py-0.5 */
  border-radius: 0.25rem;
  /* rounded */
  font-size: 0.875rem;
  /* text-sm */
}

.dark .prose-content code {
  background-color: rgb(55 65 81);
  /* bg-gray-700 */
  color: rgb(229 231 235);
  /* text-gray-200 */
}

.prose-content pre {
  background-color: rgb(243 244 246);
  /* bg-gray-100 */
  padding: 1rem;
  /* p-4 */
  border-radius: 0.25rem;
  /* rounded */
  overflow-x: auto;
  /* overflow-x-auto */
  margin: 1rem 0;
  /* my-4 */
}

.dark .prose-content pre {
  background-color: rgb(55 65 81);
  /* bg-gray-700 */
}

.prose-content pre code {
  background-color: transparent;
  padding: 0;
}

.prose-content ul,
.prose-content ol {
  margin-bottom: 1rem;
  /* mb-4 */
  padding-left: 1.5rem;
  /* pl-6 */
}

.prose-content li {
  margin-bottom: 0.25rem;
  /* mb-1 */
}

.prose-content img {
  max-width: 100%;
  height: auto;
  border-radius: 0.25rem;
  /* rounded */
  margin: 1rem 0;
  /* my-4 */
}
</style>
