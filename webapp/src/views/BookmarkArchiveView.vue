<script setup lang="ts">
import { ref, onMounted, computed, nextTick } from 'vue';
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
const iframeRef = ref<HTMLIFrameElement | null>(null);

const bookmarkId = computed(() => {
  const id = route.params.id;
  return typeof id === 'string' ? parseInt(id) : null;
});

const hasContent = computed(() => bookmark.value?.hasContent ?? false);
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

    bookmark.value = await bookmarksStore.getBookmark(bookmarkId.value);

    if (!bookmark.value?.hasArchive) {
      error.value = t('bookmarks.error.no_archive');
    }
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

const resizeIframe = () => {
  if (iframeRef.value) {
    try {
      const iframe = iframeRef.value;
      const iframeDoc = iframe.contentDocument || iframe.contentWindow?.document;

      if (iframeDoc) {
        // Get the height of the content
        const height = Math.max(
          iframeDoc.body.scrollHeight,
          iframeDoc.body.offsetHeight,
          iframeDoc.documentElement.clientHeight,
          iframeDoc.documentElement.scrollHeight,
          iframeDoc.documentElement.offsetHeight
        );

        // Set iframe height with some padding
        iframe.style.height = `${height + 20}px`;
      }
    } catch (error) {
      // Cross-origin restrictions might prevent access
      console.log('Cannot resize iframe due to cross-origin restrictions');
    }
  }
};

const onIframeLoad = () => {
  // Try to resize after a short delay to ensure content is loaded
  setTimeout(resizeIframe, 100);
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

      <!-- Archive Content -->
      <div v-else-if="bookmark" class="space-y-6">
        <!-- Header -->
        <BookmarkDetailHeader :bookmark="bookmark" :show-download-ebook-button="true" :show-view-content-button="true"
          :show-archive-indicator="true" container-class="mx-6" @download-ebook="downloadEbook"
          @view-content="goToContent" @open-original="goToOriginal" />

        <!-- Archive Frame -->
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow-sm overflow-hidden mx-6">
          <div class="min-h-[600px]">
            <iframe ref="iframeRef" :src="`/bookmark/${bookmarkId}/archive/file/`" class="w-full border-0"
              :title="t('bookmarks.archived_content')" @load="onIframeLoad"></iframe>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>
