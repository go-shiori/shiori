<script setup lang="ts">
interface Props {
  size?: string | number;
  class?: string;
}

withDefaults(defineProps<Props>(), {
  size: 20,
  class: '',
});
</script>

<template>
  <svg
    xmlns="http://www.w3.org/2000/svg"
    :width="size"
    :height="size"
    :class="class"
    fill="none"
    viewBox="0 0 20 20"
    stroke="currentColor"
  >
    <path
      stroke-linecap="round"
      stroke-linejoin="round"
      stroke-width="2"
      d="M8.33 5H5a1.67 1.67 0 00-1.67 1.67v8.33a1.67 1.67 0 001.67 1.67h8.33a1.67 1.67 0 001.67-1.67v-3.33M11.67 3.33h5m0 0v5m0-5L8.33 11.67"
    />
  </svg>
</template>
