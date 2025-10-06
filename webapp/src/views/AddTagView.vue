<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useTagsStore } from '@/stores/tags'
import { useToast } from '@/composables/useToast'
import AppLayout from '@/components/layout/AppLayout.vue'
import { Input } from '@/components/ui'
import { useI18n } from 'vue-i18n'
import { useErrorHandler } from '@/utils/errorHandler'

const { t } = useI18n()
const router = useRouter()
const tagsStore = useTagsStore()
const { success } = useToast()
const { handleApiError } = useErrorHandler()

// Form fields
const name = ref('')

// State
const isLoading = ref(false)
const formError = ref<string | null>(null)

const handleSubmit = async () => {
    // Validation
    if (!name.value.trim()) {
        formError.value = t('tags.tag_name_required')
        return
    }

    isLoading.value = true
    formError.value = null

    try {
        // Create the tag
        await tagsStore.createTag(name.value.trim())

        // Show success toast
        success(
            t('tags.toast.created_success'),
            t('tags.toast.created_success_message')
        )

        // Redirect to tags page after successful creation
        router.push('/tags')
    } catch (err) {
        console.error('Failed to create tag:', err)
        const errorMessage = handleApiError(err as any, 'tag')
        formError.value = errorMessage
    } finally {
        isLoading.value = false
    }
}

const handleCancel = () => {
    router.push('/tags')
}
</script>

<template>
    <AppLayout>
        <!-- Header -->
        <div class="mb-6">
            <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">
                {{ t('tags.add_tag') }}
            </h1>
            <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">
                {{ t('tags.add_tag_description') }}
            </p>
        </div>

        <div class="max-w-2xl mx-auto">
            <div class="bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg shadow-sm">
                <!-- Dialog Header -->
                <div class="bg-gray-800 text-white px-4 py-3 rounded-t-lg">
                    <h2 class="text-lg font-semibold uppercase">{{ t('tags.create_new_tag') }}</h2>
                </div>

                <!-- Dialog Body -->
                <div class="p-4 space-y-4">
                    <!-- Name Field -->
                    <div>
                        <label for="name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                            {{ t('tags.name') }}
                        </label>
                        <Input id="name" v-model="name" type="text" :placeholder="t('tags.name_placeholder')"
                            :disabled="isLoading" class="w-full" />
                    </div>
                </div>

                <!-- Dialog Footer -->
                <div
                    :class="['bg-gray-50 dark:bg-gray-700 px-4 py-3 rounded-b-lg border-t border-gray-200 dark:border-gray-600 flex items-center', formError ? 'justify-between' : 'justify-end']">
                    <!-- Error Message (left side) -->
                    <div v-if="formError" class="flex-1 mr-4">
                        <div
                            class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-2">
                            <p class="text-sm text-red-800 dark:text-red-200">{{ formError }}</p>
                        </div>
                    </div>

                    <!-- Buttons (right side) -->
                    <div class="flex space-x-3">
                        <button type="button" @click="handleCancel"
                            class="px-4 py-2 text-sm font-semibold text-gray-700 dark:text-gray-300 bg-gray-200 dark:bg-gray-600 rounded-md hover:bg-gray-300 dark:hover:bg-gray-500 focus:outline-none focus:ring-1 focus:ring-gray-500 uppercase"
                            :disabled="isLoading">
                            {{ t('common.cancel') }}
                        </button>
                        <button type="button" @click="handleSubmit"
                            class="px-4 py-2 text-sm font-semibold text-white bg-red-500 rounded-md hover:bg-red-600 focus:outline-none focus:ring-1 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed uppercase"
                            :disabled="isLoading || !name.trim()">
                            <span v-if="isLoading" class="flex items-center">
                                <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                                {{ t('common.processing') }}
                            </span>
                            <span v-else>{{ t('common.save') }}</span>
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </AppLayout>
</template>
