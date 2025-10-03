<script setup lang="ts">
import { computed } from 'vue';
import { ChevronLeftIcon, ChevronRightIcon } from '@/components/icons';

interface Props {
    currentPage: number;
    totalItems: number;
    itemsPerPage?: number;
    maxVisiblePages?: number;
    perPageOptions?: number[];
}

const props = withDefaults(defineProps<Props>(), {
    itemsPerPage: 20,
    maxVisiblePages: 5,
    perPageOptions: () => [10, 20, 50, 100]
});

const emit = defineEmits<{
    'update:currentPage': [page: number];
    'update:itemsPerPage': [perPage: number];
    'page-change': [page: number];
    'per-page-change': [perPage: number];
}>();

const totalPages = computed(() => Math.ceil(props.totalItems / props.itemsPerPage));

const visiblePages = computed(() => {
    const pages: (number | string)[] = [];
    const total = totalPages.value;
    const current = props.currentPage;
    const maxVisible = props.maxVisiblePages;

    if (total <= maxVisible + 2) {
        // Show all pages if total is small
        for (let i = 1; i <= total; i++) {
            pages.push(i);
        }
    } else {
        // Always show first page
        pages.push(1);

        // Calculate range around current page
        let start = Math.max(2, current - Math.floor(maxVisible / 2));
        let end = Math.min(total - 1, start + maxVisible - 1);

        // Adjust start if we're near the end
        if (end === total - 1) {
            start = Math.max(2, end - maxVisible + 1);
        }

        // Add ellipsis after first page if needed
        if (start > 2) {
            pages.push('...');
        }

        // Add middle pages
        for (let i = start; i <= end; i++) {
            pages.push(i);
        }

        // Add ellipsis before last page if needed
        if (end < total - 1) {
            pages.push('...');
        }

        // Always show last page
        pages.push(total);
    }

    return pages;
});

const goToPage = (page: number) => {
    if (page < 1 || page > totalPages.value || page === props.currentPage) return;
    emit('update:currentPage', page);
    emit('page-change', page);
};

const previousPage = () => {
    goToPage(props.currentPage - 1);
};

const nextPage = () => {
    goToPage(props.currentPage + 1);
};

const startItem = computed(() => {
    return (props.currentPage - 1) * props.itemsPerPage + 1;
});

const endItem = computed(() => {
    return Math.min(props.currentPage * props.itemsPerPage, props.totalItems);
});

const changeItemsPerPage = (event: Event) => {
    const newPerPage = parseInt((event.target as HTMLSelectElement).value);
    emit('update:itemsPerPage', newPerPage);
    emit('per-page-change', newPerPage);
    // Reset to first page when changing items per page
    if (props.currentPage !== 1) {
        emit('update:currentPage', 1);
        emit('page-change', 1);
    }
};
</script>

<template>
    <div v-if="totalPages > 1" class="flex flex-col sm:flex-row items-center justify-between gap-4 mt-6">
        <!-- Info text and per page selector -->
        <div class="flex flex-col sm:flex-row items-center gap-3">
            <div class="text-sm text-gray-600 dark:text-gray-400">
                Showing {{ startItem }} to {{ endItem }} of {{ totalItems }} items
            </div>
            <div class="flex items-center gap-2">
                <label for="items-per-page" class="text-sm text-gray-600 dark:text-gray-400">
                    Per page:
                </label>
                <select id="items-per-page" :value="itemsPerPage" @change="changeItemsPerPage"
                    class="px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500">
                    <option v-for="option in perPageOptions" :key="option" :value="option">
                        {{ option }}
                    </option>
                </select>
            </div>
        </div>

        <!-- Pagination controls -->
        <nav class="flex items-center gap-1" role="navigation" aria-label="Pagination">
            <!-- Previous button -->
            <button @click="previousPage" :disabled="currentPage === 1" :class="[
                'px-3 py-2 rounded-md text-sm font-medium transition-colors',
                currentPage === 1
                    ? 'text-gray-400 dark:text-gray-600 cursor-not-allowed'
                    : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
            ]" :aria-label="'Go to previous page'">
                <ChevronLeftIcon class="h-5 w-5" />
            </button>

            <!-- Page numbers -->
            <template v-for="(page, index) in visiblePages" :key="index">
                <span v-if="page === '...'" class="px-3 py-2 text-gray-500 dark:text-gray-400">
                    ...
                </span>
                <button v-else @click="goToPage(page as number)" :class="[
                    'px-3 py-2 rounded-md text-sm font-medium transition-colors',
                    page === currentPage
                        ? 'bg-blue-500 text-white'
                        : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
                ]" :aria-label="'Go to page ' + page" :aria-current="page === currentPage ? 'page' : undefined">
                    {{ page }}
                </button>
            </template>

            <!-- Next button -->
            <button @click="nextPage" :disabled="currentPage === totalPages" :class="[
                'px-3 py-2 rounded-md text-sm font-medium transition-colors',
                currentPage === totalPages
                    ? 'text-gray-400 dark:text-gray-600 cursor-not-allowed'
                    : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
            ]" :aria-label="'Go to next page'">
                <ChevronRightIcon class="h-5 w-5" />
            </button>
        </nav>
    </div>
</template>
