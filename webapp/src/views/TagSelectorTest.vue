<script setup lang="ts">
import { ref } from 'vue'
import TagSelector from '@/components/ui/TagSelector.vue'

const selectedTagIds = ref<number[]>([])

const handleTagChange = (newTagIds: number[]) => {
    console.log('Selected tag IDs:', newTagIds)
    selectedTagIds.value = newTagIds
}
</script>

<template>
    <div class="p-8 max-w-2xl mx-auto">
        <h1 class="text-2xl font-bold mb-6">Tag Selector Test</h1>

        <div class="space-y-4">
            <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Test Tag Selector
                </label>
                <TagSelector v-model="selectedTagIds" placeholder="Type to search or create tags..."
                    @update:modelValue="handleTagChange" />
            </div>

            <div class="mt-4 p-4 bg-gray-100 dark:bg-gray-800 rounded-md">
                <h3 class="font-semibold mb-2">Selected Tag IDs:</h3>
                <pre class="text-sm">{{ JSON.stringify(selectedTagIds, null, 2) }}</pre>
            </div>

            <div class="mt-4 p-4 bg-blue-50 dark:bg-blue-900/30 rounded-md">
                <h3 class="font-semibold mb-2">Instructions:</h3>
                <ul class="text-sm space-y-1">
                    <li>• Type to search for existing tags (debounced search with 300ms delay)</li>
                    <li>• Use ↑/↓ arrow keys to navigate suggestions</li>
                    <li>• Press Enter to select highlighted item</li>
                    <li>• Press Enter on non-existing tag to create it</li>
                    <li>• Press Escape to close dropdown</li>
                    <li>• Press Backspace when input is empty to remove last tag</li>
                    <li>• Click X on tag chips to remove them</li>
                    <li>• Search works with server-side API calls</li>
                    <li>• Loading indicator shows during search</li>
                </ul>
            </div>
        </div>
    </div>
</template>
