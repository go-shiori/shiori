<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useBookmarksStore } from '@/stores/bookmarks'
import { useTagsStore } from '@/stores/tags'
import { useToast } from '@/composables/useToast'
import AppLayout from '@/components/layout/AppLayout.vue'
import { Input, Textarea, Select, Checkbox, TagSelector } from '@/components/ui'
import { useI18n } from 'vue-i18n'
import { useErrorHandler } from '@/utils/errorHandler'

const { t } = useI18n()
const router = useRouter()
const bookmarksStore = useBookmarksStore()
const tagsStore = useTagsStore()
const { success, error } = useToast()
const { handleApiError } = useErrorHandler()

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
const formError = ref<string | null>(null)

const handleSubmit = async () => {
    // Validation
    if (!url.value.trim()) {
        formError.value = t('bookmarks.please_enter_url')
        return
    }

    // Basic URL validation
    try {
        new URL(url.value)
    } catch {
        formError.value = t('bookmarks.please_enter_valid_url')
        return
    }

    isLoading.value = true
    formError.value = null

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

        // Show success toast
        success(
            t('bookmarks.toast.created_success'),
            t('bookmarks.toast.created_success_message')
        )

        // Redirect to home page after successful creation
        router.push('/home')
    } catch (err) {
        console.error('Failed to create bookmark:', err)
        const errorMessage = handleApiError(err as any)
        formError.value = errorMessage

        // Show error toast
        error(
            t('bookmarks.toast.created_error'),
            errorMessage
        )
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
                    <!-- Error Message -->
                    <div v-if="formError"
                        class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-3">
                        <p class="text-sm text-red-800 dark:text-red-200">{{ formError }}</p>
                    </div>

                    <!-- URL Field -->
                    <div>
                        <label for="url" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                            {{ t('bookmarks.url_label') }}
                        </label>
                        <Input id="url" v-model="url" type="url" :placeholder="t('bookmarks.url_placeholder')"
                            :disabled="isLoading" required />
                    </div>

                    <!-- Custom Title Field -->
                    <div>
                        <label for="title" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                            {{ t('bookmarks.custom_title_label') }}
                        </label>
                        <Input id="title" v-model="title" type="text"
                            :placeholder="t('bookmarks.custom_title_placeholder')" :disabled="isLoading" />
                    </div>

                    <!-- Custom Excerpt Field -->
                    <div>
                        <label for="excerpt" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                            {{ t('bookmarks.custom_excerpt_label') }}
                        </label>
                        <Textarea id="excerpt" v-model="excerpt"
                            :placeholder="t('bookmarks.custom_excerpt_placeholder')" :rows="3" :disabled="isLoading" />
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
                            <Checkbox v-model="createArchive" :disabled="isLoading" class="mr-2" />
                            <span class="text-sm text-gray-700 dark:text-gray-300">{{ t('bookmarks.create_archive')
                                }}</span>
                        </label>

                        <label class="flex items-center cursor-pointer">
                            <Checkbox v-model="createEbook" :disabled="isLoading" class="mr-2" />
                            <span class="text-sm text-gray-700 dark:text-gray-300">{{ t('bookmarks.generate_ebook')
                                }}</span>
                        </label>

                        <!-- Visibility Select -->
                        <div>
                            <label for="visibility"
                                class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                                {{ t('bookmarks.visibility_label') }}
                            </label>
                            <Select id="visibility" v-model="visibility" :options="[
                                { value: 'internal', label: t('bookmarks.visibility_internal') },
                                { value: 'public', label: t('bookmarks.visibility_public') }
                            ]" :disabled="isLoading" />
                            <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                                {{ t('bookmarks.visibility_description') }}
                            </p>
                        </div>
                    </div>

                    <!-- Error Message -->
                    <div v-if="formError"
                        class="bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-300 p-3 rounded-md">
                        {{ formError }}
                    </div>
                </div>

                <!-- Dialog Footer -->
                <div
                    class="bg-gray-50 dark:bg-gray-700 px-4 py-3 rounded-b-lg border-t border-gray-200 dark:border-gray-600 flex justify-end space-x-3">
                    <button type="button" @click="handleCancel"
                        class="px-4 py-2 text-sm font-semibold text-gray-700 dark:text-gray-300 bg-gray-200 dark:bg-gray-600 rounded-md hover:bg-gray-300 dark:hover:bg-gray-500 focus:outline-none focus:ring-1 focus:ring-gray-500 uppercase"
                        :disabled="isLoading">
                        {{ t('common.cancel') }}
                    </button>
                    <button type="button" @click="handleSubmit"
                        class="px-4 py-2 text-sm font-semibold text-white bg-red-500 rounded-md hover:bg-red-600 focus:outline-none focus:ring-1 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed uppercase"
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
