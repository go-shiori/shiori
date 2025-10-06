<script setup lang="ts">
import { computed } from 'vue'
import type { Toast } from '@/composables/useToast'

interface Props {
    toast: Toast
}

const props = defineProps<Props>()

const emit = defineEmits<{
    remove: [id: string]
}>()

const toastClasses = computed(() => {
    const baseClasses = 'flex items-start p-4 rounded-lg shadow-lg border-l-4 mb-3 transition-all duration-300 ease-in-out transform'

    const typeClasses = {
        success: 'bg-green-100 dark:bg-green-800 border-green-500 text-green-800 dark:text-green-100',
        error: 'bg-red-100 dark:bg-red-800 border-red-500 text-red-800 dark:text-red-100',
        info: 'bg-blue-100 dark:bg-blue-800 border-blue-500 text-blue-800 dark:text-blue-100',
        warning: 'bg-yellow-100 dark:bg-yellow-800 border-yellow-500 text-yellow-800 dark:text-yellow-100'
    }

    return `${baseClasses} ${typeClasses[props.toast.type]}`
})

const iconClasses = computed(() => {
    const baseClasses = 'flex-shrink-0 w-5 h-5 mr-3 mt-0.5'

    const typeClasses = {
        success: 'text-green-500',
        error: 'text-red-500',
        info: 'text-blue-500',
        warning: 'text-yellow-500'
    }

    return `${baseClasses} ${typeClasses[props.toast.type]}`
})

const handleRemove = () => {
    emit('remove', props.toast.id)
}

const getIcon = () => {
    switch (props.toast.type) {
        case 'success':
            return '✓'
        case 'error':
            return '✕'
        case 'info':
            return 'ℹ'
        case 'warning':
            return '⚠'
        default:
            return 'ℹ'
    }
}
</script>

<template>
    <div :class="toastClasses" role="alert">
        <div :class="iconClasses">
            {{ getIcon() }}
        </div>
        <div class="flex-1 min-w-0">
            <h4 class="text-sm font-semibold">{{ toast.title }}</h4>
            <p v-if="toast.message" class="text-sm mt-1 opacity-90">{{ toast.message }}</p>
        </div>
        <button @click="handleRemove"
            class="flex-shrink-0 ml-3 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors"
            aria-label="Close notification">
            <span class="sr-only">Close</span>
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
        </button>
    </div>
</template>
