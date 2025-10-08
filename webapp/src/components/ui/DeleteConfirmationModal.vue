<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { XIcon } from '@/components/icons'

interface Props {
    isOpen: boolean
    title: string
    message: string
    itemName?: string
    isLoading?: boolean
}

interface Emits {
    (e: 'close'): void
    (e: 'confirm'): void
}

defineProps<Props>()
defineEmits<Emits>()

const { t } = useI18n()
</script>

<template>
    <!-- Modal backdrop -->
    <Teleport to="body">
        <div v-if="isOpen" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
            <!-- Modal content - full size on mobile, centered on desktop -->
            <div class="bg-white dark:bg-gray-800 rounded-lg w-full max-w-md mx-auto">
                <!-- Header -->
                <div class="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
                    <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100">
                        {{ title }}
                    </h3>
                    <button @click="$emit('close')" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                        :disabled="isLoading">
                        <XIcon class="h-5 w-5" />
                    </button>
                </div>

                <!-- Body -->
                <div class="p-6">
                    <p class="text-gray-700 dark:text-gray-300 mb-6">
                        {{ message }}
                        <span v-if="itemName" class="font-medium">{{ itemName }}</span>?
                    </p>
                </div>

                <!-- Footer -->
                <div class="flex justify-end space-x-3 p-6 border-t border-gray-200 dark:border-gray-700">
                    <button @click="$emit('close')"
                        class="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
                        :disabled="isLoading">
                        {{ t('common.cancel') }}
                    </button>
                    <button @click="$emit('confirm')"
                        class="px-4 py-2 bg-red-500 text-white rounded-md hover:bg-red-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center space-x-2"
                        :disabled="isLoading">
                        <div v-if="isLoading" class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                        <span>{{ t('common.delete') }}</span>
                    </button>
                </div>
            </div>
        </div>
    </Teleport>
</template>

<style scoped>
/* Ensure modal is full size on mobile */
@media (max-width: 640px) {
    .fixed.inset-0 {
        padding: 0;
        align-items: stretch;
        justify-content: stretch;
    }

    .bg-white.dark\\:bg-gray-800 {
        border-radius: 0;
        height: 100vh;
        max-width: none;
        width: 100vw;
        margin: 0;
        display: flex;
        flex-direction: column;
        justify-content: space-between;
    }

    .flex.justify-end.space-x-3 {
        margin-top: auto;
        padding-bottom: env(safe-area-inset-bottom, 0);
    }

    /* Ensure header and body take available space */
    .flex.items-center.justify-between.p-6 {
        flex-shrink: 0;
    }

    .p-6:not(.flex.justify-end) {
        flex: 1;
        display: flex;
        flex-direction: column;
        justify-content: center;
    }
}
</style>
