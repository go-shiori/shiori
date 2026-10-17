<script setup lang="ts">
import { computed } from 'vue';

interface Props {
  type?: 'button' | 'submit' | 'reset';
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost' | 'icon' | 'link';
  size?: 'xs' | 'sm' | 'md' | 'lg';
  disabled?: boolean;
  loading?: boolean;
  fullWidth?: boolean;
  href?: string;
  target?: string;
  external?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  type: 'button',
  variant: 'primary',
  size: 'md',
  disabled: false,
  loading: false,
  fullWidth: false,
  external: false,
});

const emit = defineEmits<{
  click: [event: MouseEvent];
}>();

const sizeClasses = {
  xs: 'px-1 py-1 text-xs',
  sm: 'px-2 py-1 text-sm',
  md: 'px-3 py-2',
  lg: 'px-4 py-3 text-lg',
};

const variantClasses = {
  primary:
    'bg-red-500 text-white hover:bg-red-600 focus:ring-red-500 shadow-sm',
  secondary:
    'border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600 focus:ring-gray-500 shadow-sm',
  danger: 'bg-red-600 text-white hover:bg-red-700 focus:ring-red-500 shadow-sm',
  ghost:
    'text-gray-500 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 focus:ring-gray-500',
  icon: 'text-gray-500 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 focus:ring-gray-500',
  link: 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded',
};

const baseClasses =
  'inline-flex items-center justify-center rounded-md font-medium transition-colors focus:outline-none focus:ring-1 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed whitespace-nowrap';

const classes = computed(() =>
  [
    baseClasses,
    sizeClasses[props.size],
    variantClasses[props.variant],
    props.fullWidth ? 'w-full' : '',
    props.loading ? 'cursor-wait' : '',
  ]
    .filter(Boolean)
    .join(' ')
);

const handleClick = (event: MouseEvent) => {
  if (!props.disabled && !props.loading) {
    emit('click', event);
  }
};

const isLink = computed(() => props.href || props.variant === 'link');
const Component = computed(() => (isLink.value ? 'a' : 'button'));
</script>

<template>
  <component
    :is="Component"
    :type="isLink ? undefined : type"
    :href="isLink ? href : undefined"
    :target="isLink ? target : undefined"
    :class="classes"
    :disabled="disabled || loading"
    @click="handleClick"
  >
    <div
      v-if="loading"
      class="animate-spin rounded-full h-4 w-4 border-b-2 border-current mr-2"
    ></div>
    <slot />
  </component>
</template>
