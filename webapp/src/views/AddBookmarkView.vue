<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useBookmarksStore } from '@/stores/bookmarks'
import { useTagsStore } from '@/stores/tags'
import AppLayout from '@/components/layout/AppLayout.vue'
import TagSelector from '@/components/ui/TagSelector.vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const router = useRouter()
const bookmarksStore = useBookmarksStore()
const tagsStore = useTagsStore()

// Form fields
const url = ref('')
const title = ref('')
const excerpt = ref('')
const selectedTagIds = ref<number[]>([])
const createArchive = ref(true)
const createEbook = ref(false)
const visibility = ref<'internal' | 'public'>('internal')

// State
const isLoading = ref(false)
const error = ref<string | null>(null)

const handleSubmit = async () => {
    // Validation
    if (!url.value.trim()) {
        error.value = t('bookmarks.please_enter_url')
        return
    }

    // Basic URL validation
    try {
        new URL(url.value)
    } catch {
        error.value = t('bookmarks.please_enter_valid_url')
        return
    }

    isLoading.value = true
    error.value = null

    try {
        // 1. Create bookmark
        const newBookmark = await bookmarksStore.createBookmark(
            url.value.trim(),
            title.value.trim() || undefined,
            excerpt.value.trim() || undefined,
            visibility.value === 'public' ? 1 : 0
        )

        // 2. Associate tags if provided
        if (selectedTagIds.value.length > 0 && newBookmark.id) {
            for (const tagId of selectedTagIds.value) {
                try {
                    await bookmarksStore.addTagToBookmark(newBookmark.id, tagId)
                } catch (tagErr) {
                    console.error(`Failed to add tag with ID ${tagId}:`, tagErr)
                    // Continue with other tags
                }
            }
        }

        // 3. Update bookmark data (archive/ebook) if needed
        if ((createArchive.value || createEbook.value) && newBookmark.id) {
            try {
                await bookmarksStore.updateBookmarkData(newBookmark.id, {
                    updateReadable: true,
                    createArchive: createArchive.value,
                    createEbook: createEbook.value,
                    keepMetadata: false,
                    skipExisting: false
                })
            } catch (dataError) {
                console.warn('Failed to generate bookmark data:', dataError)
                // Don't show this error to user as the bookmark was created successfully
            }
        }

        // Redirect to home page after successful creation
        router.push('/home')
    } catch (err) {
        console.error('Error creating bookmark:', err)
        error.value = err instanceof Error ? err.message : t('bookmarks.failed_to_create_bookmark')
    } finally {
        isLoading.value = false
    }
}

const handleCancel = () => {
    router.push('/home')
}

// Load tags on mount
onMounted(async () => {
    try {
        await tagsStore.fetchTags()
    } catch (err) {
        console.warn('Failed to load tags:', err)
        // Don't block the form if tags fail to load
    }
})
</script>

<template>
    <AppLayout>
        <template #header>
            <div class="flex justify-between items-center">
                <h1 class="text-xl font-bold text-gray-800 dark:text-white">{{ t('bookmarks.new_bookmark') }}</h1>
                <button @click="handleCancel"
                    class="text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200">
                    {{ t('common.cancel') }}
                </button>
            </div>
        </template>

        <div class="max-w-2xl mx-auto">
            <div class="bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg shadow-sm">
                <!-- Dialog Header -->
                <div class="bg-gray-800 text-white px-4 py-3 rounded-t-lg">
                    <h2 class="text-lg font-semibold uppercase">{{ t('bookmarks.create_new_bookmark') }}</h2>
                </div>

                <!-- Dialog Body -->
                <div class="p-4 space-y-4">
                    <!-- URL Field -->
                    <div>
                        <label for="url" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                            {{ t('bookmarks.url_label') }}
                        </label>
                        <input id="url" v-model="url" type="url" :placeholder="t('bookmarks.url_placeholder')"
                            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-red-500 focus:border-red-500 dark:bg-gray-700 dark:text-white"
                            :disabled="isLoading" required />
                    </div>

                    <!-- Custom Title Field -->
                    <div>
                        <label for="title" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                            {{ t('bookmarks.custom_title_label') }}
                        </label>
                        <input id="title" v-model="title" type="text"
                            :placeholder="t('bookmarks.custom_title_placeholder')"
                            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-red-500 focus:border-red-500 dark:bg-gray-700 dark:text-white"
                            :disabled="isLoading" />
                    </div>

                    <!-- Custom Excerpt Field -->
                    <div>
                        <label for="excerpt" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                            {{ t('bookmarks.custom_excerpt_label') }}
                        </label>
                        <textarea id="excerpt" v-model="excerpt"
                            :placeholder="t('bookmarks.custom_excerpt_placeholder')" rows="3"
                            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-red-500 focus:border-red-500 dark:bg-gray-700 dark:text-white resize-vertical"
                            :disabled="isLoading"></textarea>
                    </div>

                    <!-- Tags Field -->
                    <div>
                        <label for="tags" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                            {{ t('bookmarks.tags_label') }}
                        </label>
                        <TagSelector v-model="selectedTagIds" :disabled="isLoading"
                            :placeholder="t('bookmarks.tags_placeholder')" />
                    </div>

                    <!-- Checkboxes -->
                    <div class="space-y-3">
                        <label class="flex items-center cursor-pointer">
                            <input v-model="createArchive" type="checkbox"
                                class="mr-2 h-4 w-4 text-red-600 focus:ring-red-500 border-gray-300 rounded"
                                :disabled="isLoading" />
                            <span class="text-sm text-gray-700 dark:text-gray-300">{{ t('bookmarks.create_archive')
                                }}</span>
                        </label>

                        <label class="flex items-center cursor-pointer">
                            <input v-model="createEbook" type="checkbox"
                                class="mr-2 h-4 w-4 text-red-600 focus:ring-red-500 border-gray-300 rounded"
                                :disabled="isLoading" />
                            <span class="text-sm text-gray-700 dark:text-gray-300">{{ t('bookmarks.generate_ebook')
                                }}</span>
                        </label>

                        <!-- Visibility Select -->
                        <div>
                            <label for="visibility"
                                class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                                {{ t('bookmarks.visibility_label') }}
                            </label>
                            <select id="visibility" v-model="visibility"
                                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-red-500 focus:border-red-500 dark:bg-gray-700 dark:text-white"
                                :disabled="isLoading">
                                <option value="internal">{{ t('bookmarks.visibility_internal') }}</option>
                                <option value="public">{{ t('bookmarks.visibility_public') }}</option>
                            </select>
                            <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                                {{ t('bookmarks.visibility_description') }}
                            </p>
                        </div>
                    </div>

                    <!-- Error Message -->
                    <div v-if="error"
                        class="bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-300 p-3 rounded-md">
                        {{ error }}
                    </div>
                </div>

                <!-- Dialog Footer -->
                <div
                    class="bg-gray-50 dark:bg-gray-700 px-4 py-3 rounded-b-lg border-t border-gray-200 dark:border-gray-600 flex justify-end space-x-3">
                    <button type="button" @click="handleCancel"
                        class="px-4 py-2 text-sm font-semibold text-gray-700 dark:text-gray-300 bg-gray-200 dark:bg-gray-600 rounded-md hover:bg-gray-300 dark:hover:bg-gray-500 focus:outline-none focus:ring-2 focus:ring-gray-500 uppercase"
                        :disabled="isLoading">
                        {{ t('common.cancel') }}
                    </button>
                    <button type="button" @click="handleSubmit"
                        class="px-4 py-2 text-sm font-semibold text-white bg-red-500 rounded-md hover:bg-red-600 focus:outline-none focus:ring-2 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed uppercase"
                        :disabled="isLoading || !url.trim()">
                        <span v-if="isLoading" class="flex items-center">
                            <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                            {{ t('common.processing') }}
                        </span>
                        <span v-else>{{ t('common.ok') }}</span>
                    </button>
                </div>
            </div>
        </div>
    </AppLayout>
</template>
