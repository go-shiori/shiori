<script setup lang="ts">
import BookmarkThumbnail from '@/components/ui/BookmarkThumbnail.vue';
import DeleteConfirmationModal from '@/components/ui/DeleteConfirmationModal.vue';
import Button from '@/components/ui/Button.vue';
import {
  ImageIcon,
  PencilIcon,
  TrashIcon,
  ArchiveIcon,
  BookIcon,
  FileTextIcon,
  ExternalLinkIcon,
} from '@/components/icons';
import type { ModelBookmarkDTO } from '@/client';
import { useI18n } from 'vue-i18n';
import { computed, ref } from 'vue';
import { useAuthStore } from '@/stores/auth';
import { useBookmarksStore } from '@/stores/bookmarks';
import { useToast } from '@/composables/useToast';

interface Props {
  bookmark: ModelBookmarkDTO;
  authToken?: string;
}

interface Emits {
  (e: 'delete', bookmark: ModelBookmarkDTO): void;
  (e: 'edit', bookmark: ModelBookmarkDTO): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

const { t } = useI18n();

const authStore = useAuthStore();
const bookmarksStore = useBookmarksStore();
const { success, error: showErrorToast } = useToast();

const shouldHideExcerpt = computed(
  () => authStore.user?.config?.HideExcerpt === true
);
const shouldHideThumbnail = computed(
  () => authStore.user?.config?.HideThumbnail === true
);

const showDeleteModal = ref(false);
const isDeleting = ref(false);

const handleDeleteClick = () => {
  showDeleteModal.value = true;
};

const handleDeleteConfirm = async () => {
  if (!props.bookmark.id) return;

  isDeleting.value = true;
  try {
    await bookmarksStore.deleteBookmarks([props.bookmark.id]);
    showDeleteModal.value = false;
    success(
      t('bookmarks.toast.deleted_success'),
      t('bookmarks.toast.deleted_success_message')
    );
    // Also emit the delete event for parent components that might need it
    emit('delete', props.bookmark);
  } catch (err) {
    console.error('Failed to delete bookmark:', err);
    showErrorToast(
      t('bookmarks.toast.deleted_error'),
      t('bookmarks.toast.deleted_error_message')
    );
  } finally {
    isDeleting.value = false;
  }
};

const handleEditClick = () => {
  emit('edit', props.bookmark);
};
</script>

<template>
  <div
    class="bg-white dark:bg-gray-800 rounded-lg shadow-sm hover:shadow-md transition-shadow overflow-hidden cursor-pointer flex flex-col h-full"
    @click="$router.push(`/bookmark/${props.bookmark.id}/content`)">
    <!-- Image at the top -->
    <div v-if="!shouldHideThumbnail && props.bookmark.hasThumbnail"
      class="aspect-[2/1] bg-gray-100 dark:bg-gray-700 flex-shrink-0">
      <BookmarkThumbnail :bookmark="props.bookmark" size="large" class="w-full h-full" />
    </div>

    <!-- Content area that grows to fill space -->
    <div class="p-4 flex flex-col flex-grow">
      <!-- Title -->
      <div class="flex items-start justify-between gap-2 mb-2">
        <h3 class="text-blue-600 dark:text-blue-400 font-medium text-sm line-clamp-2 flex-1">
          {{ props.bookmark.title || props.bookmark.url }}
        </h3>
        <!-- Feature icons -->
        <div class="flex items-center gap-1 flex-shrink-0">
          <FileTextIcon v-if="props.bookmark.hasContent" class="h-3 w-3 text-gray-500 dark:text-gray-400"
            :title="t('bookmarks.has_readable_content')" />
          <ArchiveIcon v-if="props.bookmark.hasArchive" class="h-3 w-3 text-gray-500 dark:text-gray-400"
            :title="t('bookmarks.has_archive')" />
          <BookIcon v-if="props.bookmark.hasEbook" class="h-3 w-3 text-gray-500 dark:text-gray-400"
            :title="t('bookmarks.has_ebook')" />
        </div>
      </div>

      <!-- URL -->
      <div class="text-gray-500 dark:text-gray-400 text-xs truncate mb-2">
        {{ props.bookmark.url }}
      </div>

      <!-- Excerpt -->
      <div v-if="props.bookmark.excerpt && !shouldHideExcerpt"
        class="text-gray-600 dark:text-gray-400 text-xs line-clamp-2 flex-grow">
        {{ props.bookmark.excerpt }}
      </div>
    </div>

    <!-- Actions pinned to bottom -->
    <div class="px-4 pb-4 flex justify-end space-x-2">
      <Button v-if="props.bookmark.url" variant="icon" size="xs" :href="props.bookmark.url" target="_blank" @click.stop>
        <span class="sr-only">{{ t('bookmarks.open_original_url') }}</span>
        <ExternalLinkIcon class="h-4 w-4" />
      </Button>
      <Button variant="icon" size="xs" @click.stop="handleEditClick">
        <span class="sr-only">{{ t('bookmarks.edit_bookmark_action') }}</span>
        <PencilIcon class="h-4 w-4" />
      </Button>
      <Button variant="icon" size="xs" @click.stop="handleDeleteClick"
        class="hover:text-red-500 dark:hover:text-red-400">
        <span class="sr-only">{{ t('bookmarks.delete_bookmark_action') }}</span>
        <TrashIcon class="h-4 w-4" />
      </Button>
    </div>
  </div>

  <!-- Delete Confirmation Modal -->
  <DeleteConfirmationModal :is-open="showDeleteModal" :title="t('bookmarks.delete_bookmark')"
    :message="t('bookmarks.confirm_delete_message')" :item-name="props.bookmark.title || props.bookmark.url"
    :is-loading="isDeleting" @close="showDeleteModal = false" @confirm="handleDeleteConfirm" />
</template>

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
