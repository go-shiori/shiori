<script setup lang="ts">
import { useToast } from '@/composables/useToast'
import Toast from './Toast.vue'

const { toasts, removeToast } = useToast()
</script>

<template>
    <div class="fixed top-4 right-4 z-50 max-w-sm w-full">
        <TransitionGroup name="toast" tag="div" class="space-y-2">
            <Toast v-for="toast in toasts" :key="toast.id" :toast="toast" @remove="removeToast" />
        </TransitionGroup>
    </div>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
    transition: all 0.3s ease;
}

.toast-enter-from {
    opacity: 0;
    transform: translateX(100%);
}

.toast-leave-to {
    opacity: 0;
    transform: translateX(100%);
}

.toast-move {
    transition: transform 0.3s ease;
}
</style>
