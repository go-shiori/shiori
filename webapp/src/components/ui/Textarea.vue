<script setup lang="ts">
import { computed } from 'vue';

interface Props {
  modelValue?: string;
  placeholder?: string;
  disabled?: boolean;
  required?: boolean;
  id?: string;
  rows?: number;
  resize?: 'none' | 'vertical' | 'horizontal' | 'both';
  size?: 'sm' | 'md' | 'lg';
}

const props = withDefaults(defineProps<Props>(), {
  rows: 3,
  resize: 'vertical',
  size: 'md',
});

const emit = defineEmits<{
  'update:modelValue': [value: string];
  focus: [event: FocusEvent];
  blur: [event: FocusEvent];
  input: [event: Event];
}>();

const sizeClasses = {
  sm: 'px-2 py-1 text-sm',
  md: 'px-3 py-2',
  lg: 'px-4 py-3 text-lg',
};

const resizeClasses = {
  none: 'resize-none',
  vertical: 'resize-vertical',
  horizontal: 'resize-horizontal',
  both: 'resize',
};

const classes = computed(() =>
  [
    'w-full border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-1 focus:ring-red-500 focus:border-red-500 bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 font-medium',
    sizeClasses[props.size],
    resizeClasses[props.resize],
  ].join(' ')
);
</script>

<template>
  <textarea
    :id="id"
    :class="classes"
    :value="modelValue"
    :placeholder="placeholder"
    :disabled="disabled"
    :required="required"
    :rows="rows"
    @input="
      emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)
    "
    @focus="emit('focus', $event)"
    @blur="emit('blur', $event)"
  />
</template>
