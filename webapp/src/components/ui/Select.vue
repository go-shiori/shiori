<script setup lang="ts">
import { computed } from 'vue'

interface SelectOption {
  value: string | number
  label: string
  disabled?: boolean
}

interface Props {
  modelValue?: string | number
  options: SelectOption[]
  placeholder?: string
  disabled?: boolean
  required?: boolean
  id?: string
  size?: 'sm' | 'md' | 'lg'
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md'
})

const emit = defineEmits<{
  'update:modelValue': [value: string | number]
  change: [event: Event]
}>()

const sizeClasses = {
  sm: 'px-2 py-1 text-sm',
  md: 'px-3 py-2',
  lg: 'px-4 py-3 text-lg'
}

const classes = computed(() => [
  'w-full border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-1 focus:ring-red-500 focus:border-red-500 bg-white dark:bg-gray-700 text-gray-900 dark:text-white',
  sizeClasses[props.size]
].join(' '))
</script>

<template>
  <select
    :id="id"
    :class="classes"
    :value="modelValue"
    :disabled="disabled"
    :required="required"
    @change="emit('update:modelValue', ($event.target as HTMLSelectElement).value); emit('change', $event)"
  >
    <option v-if="placeholder" value="" disabled>{{ placeholder }}</option>
    <option
      v-for="option in options"
      :key="option.value"
      :value="option.value"
      :disabled="option.disabled"
    >
      {{ option.label }}
    </option>
  </select>
</template>
