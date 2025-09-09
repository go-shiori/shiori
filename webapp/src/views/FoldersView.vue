<script setup lang="ts">
import AppLayout from '@/components/layout/AppLayout.vue';
import { ref } from 'vue';

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
                <h1 class="text-xl font-bold">Folders</h1>
                <div class="flex space-x-2">
                    <button class="bg-red-500 text-white px-3 py-1 rounded-md hover:bg-red-600">
                        New Folder
                    </button>
                </div>
            </div>
        </template>

        <div class="mt-6">
            <div v-if="folders.length === 0" class="text-center py-8 text-gray-500">
                No folders yet. Create your first folder to organize your bookmarks.
            </div>

            <ul v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                <li v-for="folder in folders" :key="folder.id"
                    class="bg-white p-4 rounded-md shadow-sm hover:shadow-md transition-shadow border border-gray-200">
                    <div class="flex items-center">
                        <div class="mr-3 text-gray-400">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8" fill="none" viewBox="0 0 24 24"
                                stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                    d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                            </svg>
                        </div>
                        <div class="flex-1">
                            <h3 class="font-medium text-lg">{{ folder.name }}</h3>
                            <p class="text-sm text-gray-500">{{ folder.bookmarkCount }} bookmarks</p>
                        </div>
                        <div class="flex space-x-1">
                            <button class="text-gray-400 hover:text-gray-600 p-1" title="Edit">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20"
                                    fill="currentColor">
                                    <path
                                        d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z" />
                                </svg>
                            </button>
                            <button class="text-gray-400 hover:text-red-500 p-1" title="Delete">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20"
                                    fill="currentColor">
                                    <path fill-rule="evenodd"
                                        d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z"
                                        clip-rule="evenodd" />
                                </svg>
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
