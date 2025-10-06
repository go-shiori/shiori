<script setup lang="ts">
interface Props {
    size?: string | number;
    class?: string;
}

withDefaults(defineProps<Props>(), {
    size: 20,
    class: ''
});
</script>

<template>
    <svg xmlns="http://www.w3.org/2000/svg" :width="size" :height="size" :class="class" viewBox="0 0 24 24" fill="none"
        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
        class="icon icon-tabler icons-tabler-outline icon-tabler-bookmarks">
        <path stroke="none" d="M0 0h24v24H0z" fill="none" />
        <path d="M15 10v11l-5 -3l-5 3v-11a3 3 0 0 1 3 -3h4a3 3 0 0 1 3 3z" />
        <path d="M11 3h5a3 3 0 0 1 3 3v11" />
    </svg>
</template>
