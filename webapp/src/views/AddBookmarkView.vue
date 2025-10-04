<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useBookmarksStore } from '@/stores/bookmarks'
import AppLayout from '@/components/layout/AppLayout.vue'

const router = useRouter()
const bookmarksStore = useBookmarksStore()

const url = ref('')
const isLoading = ref(false)
const error = ref<string | null>(null)

const handleSubmit = async () => {
  if (!url.value.trim()) {
    error.value = 'Please enter a URL'
    return
  }

  // Basic URL validation
  try {
    new URL(url.value)
  } catch {
    error.value = 'Please enter a valid URL'
    return
  }

  isLoading.value = true
  error.value = null

  try {
    // Create bookmark with URL as title (as requested)
    const newBookmark = await bookmarksStore.createBookmark(url.value, url.value)

    // After successful creation, request readable content, archive, and ebook
    if (newBookmark.id) {
      try {
        await bookmarksStore.updateBookmarkData(newBookmark.id, {
          updateReadable: true,
          createArchive: true,
          createEbook: true,
          keepMetadata: false,
          skipExisting: false
        })
      } catch (dataError) {
        console.warn('Failed to generate bookmark data:', dataError)
        // Don't show this error to user as the bookmark was created successfully
        // The data generation can be retried later
      }
    }

    // Redirect to home page after successful creation
    router.push('/home')
  } catch (err) {
    console.error('Error creating bookmark:', err)
    error.value = err instanceof Error ? err.message : 'Failed to create bookmark'
  } finally {
    isLoading.value = false
  }
}

const handleCancel = () => {
  router.push('/home')
}
</script>

<template>
  <AppLayout>
    <template #header>
      <div class="flex justify-between items-center">
        <h1 class="text-xl font-bold text-gray-800 dark:text-white">Add Bookmark</h1>
        <button
          @click="handleCancel"
          class="text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200"
        >
          Cancel
        </button>
      </div>
    </template>

    <div class="max-w-2xl mx-auto">
      <form @submit.prevent="handleSubmit" class="space-y-6">
        <!-- URL Input -->
        <div>
          <label for="url" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            URL
          </label>
          <input
            id="url"
            v-model="url"
            type="url"
            placeholder="https://example.com"
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-red-500 focus:border-red-500 dark:bg-gray-700 dark:text-white"
            :disabled="isLoading"
            required
          />
        </div>

        <!-- Error Message -->
        <div v-if="error" class="bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-300 p-4 rounded-md">
          {{ error }}
        </div>

        <!-- Submit Button -->
        <div class="flex justify-end space-x-3">
          <button
            type="button"
            @click="handleCancel"
            class="px-4 py-2 text-gray-700 dark:text-gray-300 bg-gray-200 dark:bg-gray-700 rounded-md hover:bg-gray-300 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-gray-500"
            :disabled="isLoading"
          >
            Cancel
          </button>
          <button
            type="submit"
            class="px-4 py-2 bg-red-500 text-white rounded-md hover:bg-red-600 focus:outline-none focus:ring-2 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="isLoading || !url.trim()"
          >
            <span v-if="isLoading" class="flex items-center">
              <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
              Processing...
            </span>
            <span v-else>Add Bookmark</span>
          </button>
        </div>
      </form>

      <!-- Help Text -->
      <div class="mt-8 p-4 bg-blue-50 dark:bg-blue-900/30 rounded-md">
        <h3 class="text-sm font-medium text-blue-800 dark:text-blue-200 mb-2">Note</h3>
        <p class="text-sm text-blue-700 dark:text-blue-300">
          The URL will be used as both the bookmark URL and title. After creating the bookmark, Shiori will automatically:
        </p>
        <ul class="text-sm text-blue-700 dark:text-blue-300 mt-2 ml-4 list-disc">
          <li>Fetch and generate readable content</li>
          <li>Create an archive of the page</li>
          <li>Generate an ebook version</li>
          <li>Extract the proper title and metadata</li>
        </ul>
      </div>
    </div>
  </AppLayout>
</template>
