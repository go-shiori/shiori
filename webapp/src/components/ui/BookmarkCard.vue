<script setup lang="ts">
import BookmarkThumbnail from '@/components/ui/BookmarkThumbnail.vue';
import { ImageIcon, PencilIcon, TrashIcon, ArchiveIcon, BookIcon, FileTextIcon, ExternalLinkIcon } from '@/components/icons';
import type { ModelBookmarkDTO } from '@/client';
import { useI18n } from 'vue-i18n';
import { computed } from 'vue';
import { useAuthStore } from '@/stores/auth';

interface Props {
    bookmark: ModelBookmarkDTO;
    authToken?: string;
}

defineProps<Props>();

const { t } = useI18n();

const authStore = useAuthStore();
const shouldHideExcerpt = computed(() => authStore.user?.config?.HideExcerpt === true);
const shouldHideThumbnail = computed(() => authStore.user?.config?.HideThumbnail === true);
</script>

<template>
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-sm hover:shadow-md transition-shadow overflow-hidden cursor-pointer"
        @click="$router.push(`/bookmark/${bookmark.id}/content`)">
        <!-- Image at the top -->
        <div v-if="!shouldHideThumbnail && bookmark.hasThumbnail" class="aspect-[2/1] bg-gray-100 dark:bg-gray-700">
            <BookmarkThumbnail :bookmark="bookmark" size="large" class="w-full h-full" />
        </div>

        <!-- Details at the bottom -->
        <div class="p-4">
            <!-- Title -->
            <div class="flex items-start justify-between gap-2 mb-2">
                <h3 class="text-blue-600 dark:text-blue-400 font-medium text-sm line-clamp-2 flex-1">
                    {{ bookmark.title || bookmark.url }}
                </h3>
                <!-- Feature icons -->
                <div class="flex items-center gap-1 flex-shrink-0">
                    <FileTextIcon v-if="bookmark.hasContent" class="h-3 w-3 text-gray-500 dark:text-gray-400"
                        :title="t('bookmarks.has_readable_content')" />
                    <ArchiveIcon v-if="bookmark.hasArchive" class="h-3 w-3 text-gray-500 dark:text-gray-400"
                        :title="t('bookmarks.has_archive')" />
                    <BookIcon v-if="bookmark.hasEbook" class="h-3 w-3 text-gray-500 dark:text-gray-400"
                        :title="t('bookmarks.has_ebook')" />
                </div>
            </div>

            <!-- URL -->
            <div class="text-gray-500 dark:text-gray-400 text-xs truncate mb-2">
                {{ bookmark.url }}
            </div>

            <!-- Excerpt -->
            <div v-if="bookmark.excerpt && !shouldHideExcerpt"
                class="text-gray-600 dark:text-gray-400 text-xs line-clamp-2 mb-3">
                {{ bookmark.excerpt }}
            </div>

            <!-- Actions -->
            <div class="flex justify-end space-x-2">
                <a v-if="bookmark.url" :href="bookmark.url" target="_blank" @click.stop
                    class="text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300">
                    <span class="sr-only">{{ t('bookmarks.open_original_url') }}</span>
                    <ExternalLinkIcon class="h-4 w-4" />
                </a>
                <button @click.stop
                    class="text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300">
                    <span class="sr-only">{{ t('bookmarks.edit_bookmark_action') }}</span>
                    <PencilIcon class="h-4 w-4" />
                </button>
                <button @click.stop class="text-gray-500 dark:text-gray-400 hover:text-red-500">
                    <span class="sr-only">{{ t('bookmarks.delete_bookmark_action') }}</span>
                    <TrashIcon class="h-4 w-4" />
                </button>
            </div>
        </div>
    </div>
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
