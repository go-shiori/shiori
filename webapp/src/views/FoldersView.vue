<script setup lang="ts">
import AppLayout from '@/components/layout/AppLayout.vue';
import { ref } from 'vue';
import { FolderIcon, PencilIcon, TrashIcon } from '@/components/icons';

// Sample folders data (to be replaced with actual API data later)
const folders = ref([
    { id: 1, name: 'Work', bookmarkCount: 12 },
    { id: 2, name: 'Personal', bookmarkCount: 8 },
    { id: 3, name: 'Research', bookmarkCount: 15 },
    { id: 4, name: 'Reading List', bookmarkCount: 23 },
]);
</script>

<template>
    <AppLayout>
        <template #header>
            <div class="flex justify-between items-center">
                <h1 class="text-xl font-bold text-gray-800 dark:text-white">Folders</h1>
                <div class="flex space-x-2">
                    <button class="bg-red-500 text-white px-3 py-1 rounded-md hover:bg-red-600">
                        New Folder
                    </button>
                </div>
            </div>
        </template>

        <div class="mt-6">
            <div v-if="folders.length === 0" class="text-center py-8 text-gray-500 dark:text-gray-400">
                No folders yet. Create your first folder to organize your bookmarks.
            </div>

            <ul v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                <li v-for="folder in folders" :key="folder.id"
                    class="bg-white dark:bg-gray-800 p-4 rounded-md shadow-sm hover:shadow-md transition-shadow border border-gray-200 dark:border-gray-700">
                    <div class="flex items-center">
                        <div class="mr-3 text-gray-400 dark:text-gray-500">
                            <FolderIcon class="h-8 w-8" />
                        </div>
                        <div class="flex-1">
                            <h3 class="font-medium text-lg text-gray-900 dark:text-gray-100">{{ folder.name }}</h3>
                            <p class="text-sm text-gray-500 dark:text-gray-400">{{ folder.bookmarkCount }} bookmarks</p>
                        </div>
                        <div class="flex space-x-1">
                            <button
                                class="text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300 p-1"
                                title="Edit">
                                <PencilIcon class="h-5 w-5" />
                            </button>
                            <button
                                class="text-gray-400 hover:text-red-500 dark:text-gray-500 dark:hover:text-red-400 p-1"
                                title="Delete">
                                <TrashIcon class="h-5 w-5" />
                            </button>
                        </div>
                    </div>
                </li>
            </ul>
        </div>
    </AppLayout>
</template>

<style scoped>
/* Additional custom styles if needed */
</style>
