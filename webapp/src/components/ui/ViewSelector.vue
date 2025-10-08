<script setup lang="ts">
import { ViewGridIcon, ViewListIcon } from '@/components/icons';

interface Props {
    currentView: 'list' | 'card';
    onViewChange: (view: 'list' | 'card') => void;
}

defineProps<Props>();
</script>

<template>
    <div class="flex items-center space-x-1 bg-gray-100 dark:bg-gray-700 rounded-lg p-1">
        <button
            @click="onViewChange('list')"
            :class="[
                'p-2 rounded-md transition-colors',
                currentView === 'list'
                    ? 'bg-white dark:bg-gray-600 text-gray-900 dark:text-white shadow-sm'
                    : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
            ]"
            :title="'List view'"
        >
            <ViewListIcon class="h-5 w-5" />
        </button>
        <button
            @click="onViewChange('card')"
            :class="[
                'p-2 rounded-md transition-colors',
                currentView === 'card'
                    ? 'bg-white dark:bg-gray-600 text-gray-900 dark:text-white shadow-sm'
                    : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
            ]"
            :title="'Card view'"
        >
            <ViewGridIcon class="h-5 w-5" />
        </button>
    </div>
</template>
