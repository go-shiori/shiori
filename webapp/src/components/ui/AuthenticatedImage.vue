<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
import { getBookmarkThumbnailDataUrl } from '@/utils/image-url';

interface Props {
    bookmarkId: number;
    authToken?: string;
    alt?: string;
    class?: string;
}

const props = withDefaults(defineProps<Props>(), {
    alt: 'Bookmark thumbnail'
});

const imageSrc = ref('');
const isLoading = ref(false);
const hasError = ref(false);

const loadImage = async () => {
    if (!props.bookmarkId) return;

    isLoading.value = true;
    hasError.value = false;

    try {
        const dataUrl = await getBookmarkThumbnailDataUrl(props.bookmarkId, props.authToken);
        if (dataUrl) {
            imageSrc.value = dataUrl;
        } else {
            hasError.value = true;
        }
    } catch (error) {
        console.error('Error loading authenticated image:', error);
        hasError.value = true;
    } finally {
        isLoading.value = false;
    }
};

// Load image when component mounts
onMounted(() => {
    loadImage();
});

// Reload when props change
watch(() => [props.bookmarkId, props.authToken], () => {
    if (props.bookmarkId) {
        loadImage();
    }
});
</script>

<template>
    <div :class="props.class">
        <!-- Loading state -->
        <div v-if="isLoading" class="flex items-center justify-center h-full bg-gray-100 dark:bg-gray-700">
            <div class="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-500"></div>
        </div>

        <!-- Image -->
        <img v-else-if="imageSrc && !hasError" :src="imageSrc" :alt="props.alt" class="w-full h-full object-cover" />

        <!-- Error/placeholder state -->
        <div v-else class="flex items-center justify-center h-full bg-gray-100 dark:bg-gray-700">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-gray-400 dark:text-gray-500" fill="none"
                viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                    d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
        </div>
    </div>
</template>
