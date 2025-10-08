<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  modelValue?: boolean
  disabled?: boolean
  required?: boolean
  id?: string
  size?: 'sm' | 'md' | 'lg'
  color?: 'red' | 'blue' | 'gray'
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md',
  color: 'red'
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  change: [event: Event]
}>()

const sizeClasses = {
  sm: 'h-3 w-3',
  md: 'h-4 w-4',
  lg: 'h-5 w-5'
}

const colorClasses = {
  red: 'text-red-600 focus:ring-red-500',
  blue: 'text-blue-600 focus:ring-blue-500',
  gray: 'text-gray-600 focus:ring-gray-500'
}

const classes = computed(() => [
  'focus:ring-1 border-gray-300 dark:border-gray-600 rounded',
  sizeClasses[props.size],
  colorClasses[props.color]
].join(' '))
</script>

<template>
  <input :id="id" type="checkbox" :class="classes" :checked="modelValue" :disabled="disabled" :required="required"
    @change="emit('update:modelValue', ($event.target as HTMLInputElement).checked); emit('change', $event)" />
</template>
