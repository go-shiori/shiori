<script setup lang="ts">
import { RouterLink } from 'vue-router';

// Define props using the compiler macro (no import needed)
defineProps<{
    isMobile: boolean;
}>();

interface NavItem {
    name: string;
    icon: 'home' | 'tag' | 'archive' | 'settings';
    route: string;
}

const navItems: NavItem[] = [
    { name: 'Home', icon: 'home', route: '/home' },
    { name: 'Tags', icon: 'tag', route: '/tags' },
    { name: 'Archive', icon: 'archive', route: '/archive' },
    { name: 'Settings', icon: 'settings', route: '/settings' },
];

// SVG icons mapping
const icons = {
    home: `<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
  </svg>`,
    tag: `<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
  </svg>`,
    archive: `<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
  </svg>`,
    settings: `<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
  </svg>`,
};
</script>

<template>
    <!-- Desktop Sidebar -->
    <aside v-if="!isMobile" class="w-16 bg-white border-r border-gray-200 flex flex-col items-center py-6">
        <!-- Logo -->
        <div class="mb-8">
            <div class="text-red-500 font-bold text-2xl">栞</div>
        </div>

        <!-- Navigation -->
        <nav class="flex flex-col items-center space-y-6">
            <RouterLink v-for="item in navItems" :key="item.name" :to="item.route"
                class="text-gray-500 hover:text-red-500 p-2 rounded-md hover:bg-gray-100 transition-colors"
                v-slot="{ isActive }">
                <div class="flex flex-col items-center">
                    <div :class="{ 'text-red-500': isActive }" v-html="icons[item.icon]"></div>
                    <span class="text-xs mt-1" :class="{ 'font-medium': isActive }">{{ item.name }}</span>
                </div>
            </RouterLink>
        </nav>

        <!-- User Profile at bottom -->
        <div class="mt-auto">
            <button class="text-gray-500 hover:text-red-500 p-2 rounded-full hover:bg-gray-100">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24"
                    stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                </svg>
            </button>
        </div>
    </aside>

    <!-- Mobile Bottom Navigation -->
    <nav v-else class="fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 px-4 py-2 flex justify-around">
        <RouterLink v-for="item in navItems" :key="item.name" :to="item.route"
            class="text-gray-500 hover:text-red-500 p-2 flex flex-col items-center" v-slot="{ isActive }">
            <div :class="{ 'text-red-500': isActive }" v-html="icons[item.icon]"></div>
            <span class="text-xs mt-1" :class="{ 'font-medium': isActive }">{{ item.name }}</span>
        </RouterLink>
    </nav>
</template>
