<template>
  <div
    class="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6"
    :class="containerClass"
  >
    <div class="flex items-start justify-end mb-4">
      <div class="flex items-center gap-2">
        <Button
          v-if="showDownloadEbookButton && hasEbook"
          variant="primary"
          @click="$emit('downloadEbook')"
        >
          <DownloadIcon class="h-4 w-4 mr-2" />
          {{ t('bookmarks.download_ebook') }}
        </Button>

        <Button
          v-if="showViewArchiveButton && hasArchive"
          variant="primary"
          @click="$emit('viewArchive')"
        >
          <ArchiveIcon class="h-4 w-4 mr-2" />
          {{ t('bookmarks.view_archive') }}
        </Button>

        <Button
          v-if="showViewContentButton && hasContent"
          variant="primary"
          @click="$emit('viewContent')"
        >
          <FileTextIcon class="h-4 w-4 mr-2" />
          {{ t('bookmarks.view_content') }}
        </Button>
      </div>
    </div>

    <h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-2">
      {{ bookmark.title }}
      <Button
        v-if="bookmark.url"
        variant="ghost"
        @click="$emit('openOriginal')"
        :title="t('bookmarks.open_original')"
      >
        <ExternalLinkIcon class="h-5 w-5" />
      </Button>
    </h1>

    <p v-if="bookmark.excerpt" class="text-gray-600 dark:text-gray-400 mb-4">
      {{ bookmark.excerpt }}
    </p>

    <div
      class="flex items-center gap-4 text-sm text-gray-500 dark:text-gray-400"
    >
      <span>
        {{ t('bookmarks.by') }}
        {{ bookmark.author || t('bookmarks.unknown_author') }}
      </span>
      <span>•</span>
      <span>{{ new Date(bookmark.createdAt || '').toLocaleDateString() }}</span>
      <span v-if="showArchiveIndicator">•</span>
      <span v-if="showArchiveIndicator" class="flex items-center">
        <ArchiveIcon class="h-4 w-4 mr-1" />
        {{ t('bookmarks.archived_version') }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { ModelBookmarkDTO } from '@/client';
import {
  DownloadIcon,
  ArchiveIcon,
  FileTextIcon,
  ExternalLinkIcon,
} from '@/components/icons';
import Button from './ui/Button.vue';

interface Props {
  bookmark: ModelBookmarkDTO;
  showDownloadEbookButton?: boolean;
  showViewArchiveButton?: boolean;
  showViewContentButton?: boolean;
  showArchiveIndicator?: boolean;
  containerClass?: string;
}

const props = withDefaults(defineProps<Props>(), {
  showDownloadEbookButton: false,
  showViewArchiveButton: false,
  showViewContentButton: false,
  showArchiveIndicator: false,
  containerClass: '',
});

const emit = defineEmits<{
  downloadEbook: [];
  viewArchive: [];
  viewContent: [];
  openOriginal: [];
}>();

const { t } = useI18n();

const hasEbook = computed(() => {
  return props.bookmark.hasEbook || false;
});

const hasArchive = computed(() => {
  return props.bookmark.hasArchive || false;
});

const hasContent = computed(() => {
  return props.bookmark.hasContent || false;
});
</script>
