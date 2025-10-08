<script setup lang="ts">
import { computed } from 'vue';
import { useAuthStore } from '@/stores/auth';
import AuthenticatedImage from './AuthenticatedImage.vue';
import type { ModelBookmarkDTO } from '@/client';

interface Props {
    bookmark: ModelBookmarkDTO;
    size?: 'small' | 'medium' | 'large';
    class?: string;
}

const props = withDefaults(defineProps<Props>(), {
    size: 'medium',
    class: ''
});

const authStore = useAuthStore();

// Check if thumbnails should be hidden based on user config
const shouldHideThumbnail = computed(() => {
    const config = authStore.user?.config as any;
    return config?.HideThumbnail === true || config?.hideThumbnail === true;
});

// Size classes for different thumbnail sizes
const sizeClasses = computed(() => {
    const sizes = {
        small: 'w-12 h-12',
        medium: 'w-24 h-24',
        large: 'w-32 h-32'
    };
    return sizes[props.size];
});

// Combined classes
const thumbnailClasses = computed(() => {
    return `${sizeClasses.value} object-cover rounded-md ${props.class}`;
});
</script>

<template>
    <div v-if="!shouldHideThumbnail && bookmark.hasThumbnail" :class="thumbnailClasses">
        <AuthenticatedImage :bookmark-id="bookmark.id!" :alt="bookmark.title || 'Bookmark thumbnail'"
            class="w-full h-full object-cover rounded-md" />
    </div>
    <div v-else-if="!shouldHideThumbnail" :class="thumbnailClasses"
        class="bg-gray-200 dark:bg-gray-700 flex items-center justify-center">
        <svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
        </svg>
    </div>
</template>
