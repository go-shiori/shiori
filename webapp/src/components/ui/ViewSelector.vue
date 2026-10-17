<script setup lang="ts">
import { ViewGridIcon, ViewListIcon } from '@/components/icons';
import Button from './Button.vue';

interface Props {
  currentView: 'list' | 'card';
  onViewChange: (view: 'list' | 'card') => void;
}

defineProps<Props>();
</script>

<template>
  <div
    class="flex items-center space-x-1 bg-gray-100 dark:bg-gray-700 rounded-lg p-1"
  >
    <Button
      variant="ghost"
      size="sm"
      @click="onViewChange('list')"
      :class="[
        currentView === 'list'
          ? 'bg-white dark:bg-gray-600 text-gray-900 dark:text-white shadow-sm'
          : '',
      ]"
      :title="'List view'"
    >
      <ViewListIcon class="h-5 w-5" />
    </Button>
    <Button
      variant="ghost"
      size="sm"
      @click="onViewChange('card')"
      :class="[
        currentView === 'card'
          ? 'bg-white dark:bg-gray-600 text-gray-900 dark:text-white shadow-sm'
          : '',
      ]"
      :title="'Card view'"
    >
      <ViewGridIcon class="h-5 w-5" />
    </Button>
  </div>
</template>
