<script setup lang="ts">
interface Props {
    size?: string | number;
    class?: string;
}

withDefaults(defineProps<Props>(), {
    size: 24,
    class: ''
});
</script>

<template>
    <svg xmlns="http://www.w3.org/2000/svg" :width="size" :height="size" :class="class" fill="none" viewBox="0 0 24 24"
        stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M4 6h16M4 10h16M4 14h16M4 18h16" />
    </svg>
</template>
