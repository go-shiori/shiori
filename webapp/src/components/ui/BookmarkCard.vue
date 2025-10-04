<script setup lang="ts">
import AuthenticatedImage from '@/components/ui/AuthenticatedImage.vue';
import { ImageIcon, PencilIcon, TrashIcon, ArchiveIcon, BookIcon, FileTextIcon } from '@/components/icons';
import type { ModelBookmarkDTO } from '@/client';

interface Props {
    bookmark: ModelBookmarkDTO;
    authToken?: string;
}

defineProps<Props>();
</script>

<template>
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-sm hover:shadow-md transition-shadow overflow-hidden">
        <!-- Image at the top -->
        <div class="aspect-[2/1] bg-gray-100 dark:bg-gray-700">
            <div v-if="bookmark.hasThumbnail" class="w-full h-full">
                <AuthenticatedImage :bookmark-id="bookmark.id || 0" :auth-token="authToken"
                    :alt="bookmark.title || 'Bookmark thumbnail'" class="w-full h-full object-cover" />
            </div>
            <div v-else class="w-full h-full flex items-center justify-center">
                <ImageIcon class="h-12 w-12 text-gray-400 dark:text-gray-500" />
            </div>
        </div>

        <!-- Details at the bottom -->
        <div class="p-4">
            <!-- Title -->
            <div class="flex items-start justify-between gap-2 mb-2">
                <a :href="bookmark.url" target="_blank"
                    class="text-blue-600 dark:text-blue-400 hover:underline font-medium text-sm line-clamp-2 flex-1">
                    {{ bookmark.title || bookmark.url }}
                </a>
                <!-- Feature icons -->
                <div class="flex items-center gap-1 flex-shrink-0">
                    <FileTextIcon v-if="bookmark.hasContent" class="h-3 w-3 text-gray-500 dark:text-gray-400" title="Has readable content" />
                    <ArchiveIcon v-if="bookmark.hasArchive" class="h-3 w-3 text-gray-500 dark:text-gray-400" title="Has archive" />
                    <BookIcon v-if="bookmark.hasEbook" class="h-3 w-3 text-gray-500 dark:text-gray-400" title="Has ebook" />
                </div>
            </div>

            <!-- URL -->
            <div class="text-gray-500 dark:text-gray-400 text-xs truncate mb-2">
                {{ bookmark.url }}
            </div>

            <!-- Excerpt -->
            <div v-if="bookmark.excerpt" class="text-gray-600 dark:text-gray-400 text-xs line-clamp-2 mb-3">
                {{ bookmark.excerpt }}
            </div>

            <!-- Actions -->
            <div class="flex justify-end space-x-2">
                <button class="text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300">
                    <span class="sr-only">Edit</span>
                    <PencilIcon class="h-4 w-4" />
                </button>
                <button class="text-gray-500 dark:text-gray-400 hover:text-red-500">
                    <span class="sr-only">Delete</span>
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
