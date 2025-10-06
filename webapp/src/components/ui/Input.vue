<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  modelValue?: string
  type?: 'text' | 'email' | 'password' | 'url' | 'search'
  placeholder?: string
  disabled?: boolean
  required?: boolean
  id?: string
  size?: 'sm' | 'md' | 'lg'
  variant?: 'default' | 'search'
}

const props = withDefaults(defineProps<Props>(), {
  type: 'text',
  size: 'md',
  variant: 'default'
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  focus: [event: FocusEvent]
  blur: [event: FocusEvent]
  input: [event: Event]
}>()

const sizeClasses = {
  sm: 'px-2 py-1 text-sm',
  md: 'px-3 py-2',
  lg: 'px-4 py-3 text-lg'
}

const baseClasses = 'w-full border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-1 focus:ring-red-500 bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400'

const variantClasses = {
  default: 'shadow-sm focus:border-red-500',
  search: ''
}

const classes = computed(() => [
  baseClasses,
  sizeClasses[props.size],
  variantClasses[props.variant]
].join(' '))
</script>

<template>
  <input :id="id" :type="type" :class="classes" :value="modelValue" :placeholder="placeholder" :disabled="disabled"
    :required="required" @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    @focus="emit('focus', $event)" @blur="emit('blur', $event)" />
</template>
