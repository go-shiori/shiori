<script setup lang="ts">
import { computed } from 'vue';

interface Props {
  modelValue?: string;
  type?: 'text' | 'email' | 'password' | 'url' | 'search' | 'number' | 'tel';
  placeholder?: string;
  disabled?: boolean;
  required?: boolean;
  readonly?: boolean;
  autocomplete?: string;
  autofocus?: boolean;
  maxlength?: number;
  minlength?: number;
  pattern?: string;
  id?: string;
  name?: string;
  size?: 'xs' | 'sm' | 'md' | 'lg';
  variant?: 'default' | 'search';
  class?: string;
}

const props = withDefaults(defineProps<Props>(), {
  type: 'text',
  size: 'md',
  variant: 'default',
});

const emit = defineEmits<{
  'update:modelValue': [value: string];
  focus: [event: FocusEvent];
  blur: [event: FocusEvent];
  input: [event: Event];
  change: [event: Event];
  keydown: [event: KeyboardEvent];
  keyup: [event: KeyboardEvent];
  keypress: [event: KeyboardEvent];
}>();

const sizeClasses = {
  xs: 'px-1 py-1 text-xs',
  sm: 'px-2 py-1 text-sm',
  md: 'px-3 py-2',
  lg: 'px-4 py-3 text-lg',
};

const variantClasses = {
  default: 'shadow-sm focus:border-red-500',
  search: '',
};

const baseClasses =
  'w-full border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-1 focus:ring-red-500 bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 font-medium';

const classes = computed(() =>
  [
    baseClasses,
    sizeClasses[props.size],
    variantClasses[props.variant],
    props.class,
  ]
    .filter(Boolean)
    .join(' ')
);
</script>

<template>
  <input
    :id="id"
    :name="name"
    :type="type"
    :class="classes"
    :value="modelValue"
    :placeholder="placeholder"
    :disabled="disabled"
    :required="required"
    :readonly="readonly"
    :autocomplete="autocomplete"
    :autofocus="autofocus"
    :maxlength="maxlength"
    :minlength="minlength"
    :pattern="pattern"
    @input="
      emit('update:modelValue', ($event.target as HTMLInputElement).value);
      emit('input', $event);
    "
    @focus="emit('focus', $event)"
    @blur="emit('blur', $event)"
    @keydown="emit('keydown', $event)"
    @keyup="emit('keyup', $event)"
    @keypress="emit('keypress', $event)"
  />
</template>
