<script setup lang="ts">
import { RouterLink } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { useRouter } from 'vue-router';
import { ref, onMounted, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';

// Define props using the compiler macro (no import needed)
defineProps<{
    isMobile: boolean;
}>();

const { t } = useI18n();
const authStore = useAuthStore();
const router = useRouter();
const isMenuOpen = ref(false);
const menuRef = ref<HTMLElement | null>(null);

interface NavItem {
    nameKey: string;
    icon: 'home' | 'tag' | 'folder' | 'archive' | 'settings';
    route: string;
}

const navItems: NavItem[] = [
    { nameKey: 'navigation.home', icon: 'home', route: '/home' },
    { nameKey: 'navigation.tags', icon: 'tag', route: '/tags' },
    { nameKey: 'navigation.folders', icon: 'folder', route: '/folders' },
    { nameKey: 'navigation.archive', icon: 'archive', route: '/archive' },
    { nameKey: 'navigation.settings', icon: 'settings', route: '/settings' },
];

// Toggle menu
const toggleMenu = () => {
    isMenuOpen.value = !isMenuOpen.value;
};

// Handle logout
const handleLogout = async () => {
    await authStore.logout();
    isMenuOpen.value = false;
    router.push('/login');
};

// Close menu when clicking outside
const handleClickOutside = (event: MouseEvent) => {
    if (menuRef.value && !menuRef.value.contains(event.target as Node)) {
        isMenuOpen.value = false;
    }
};

// Add and remove event listeners
onMounted(() => {
    document.addEventListener('click', handleClickOutside);
});

onUnmounted(() => {
    document.removeEventListener('click', handleClickOutside);
});

// SVG icons mapping
const icons = {
    home: `<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
  </svg>`,
    tag: `<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
  </svg>`,
    folder: `<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
  </svg>`,
    archive: `<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
  </svg>`,
    settings: `<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
  </svg>`,
    user: `<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
      d="M5.121 17.804A13.937 13.937 0 0112 16c2.5 0 4.847.655 6.879 1.804M15 10a3 3 0 11-6 0 3 3 0 016 0zm6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
  </svg>`,
};
</script>

<template>
    <div>
        <template v-if="!isMobile">
            <!-- Desktop Sidebar -->
            <aside
                class="w-20 h-screen bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 flex flex-col items-center py-6 sticky top-0">
                <!-- Logo -->
                <div class="mb-8 flex flex-col items-center">
                    <div class="text-red-500 font-bold text-2xl">栞</div>
                    <span class="text-xs mt-1 dark:text-gray-300">shiori</span>
                </div>

                <!-- Navigation -->
                <nav class="flex flex-col items-center space-y-6 flex-1">
                    <RouterLink v-for="item in navItems" :key="item.nameKey" :to="item.route"
                        class="text-gray-500 dark:text-gray-400 hover:text-red-500 p-2 rounded-md hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors flex flex-col items-center"
                        :title="t(item.nameKey)">
                        <div v-html="icons[item.icon]"></div>
                        <span class="text-xs mt-1 dark:text-gray-300">{{ t(item.nameKey) }}</span>
                    </RouterLink>

                    <!-- Spacer -->
                    <div class="flex-1"></div>

                    <!-- User Menu -->
                    <div class="relative mt-auto" ref="menuRef">
                        <button @click.stop="toggleMenu"
                            class="text-gray-500 dark:text-gray-400 hover:text-red-500 p-2 rounded-md hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors flex flex-col items-center"
                            :title="t('auth.user')">
                            <div v-html="icons.user"></div>
                            <span class="text-xs mt-1 dark:text-gray-300">{{ authStore.user?.username || t('auth.user')
                                }}</span>
                        </button>

                        <!-- Dropdown Menu -->
                        <div v-if="isMenuOpen"
                            class="absolute left-20 bottom-0 w-48 bg-white dark:bg-gray-800 rounded-md shadow-lg py-1 z-50 border border-gray-200 dark:border-gray-700">
                            <div
                                class="px-4 py-2 text-sm text-gray-500 dark:text-gray-400 border-b border-gray-200 dark:border-gray-700">
                                <div class="font-medium dark:text-gray-300">{{ authStore.user?.username ||
                                    t('auth.user') }}
                                </div>
                            </div>
                            <router-link to="/settings"
                                class="block px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700">
                                {{ t('navigation.settings') }}
                            </router-link>
                            <button @click="handleLogout"
                                class="block w-full text-left px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700">
                                {{ t('auth.logout') }}
                            </button>
                        </div>
                    </div>
                </nav>
            </aside>
        </template>

        <template v-else>
            <!-- Mobile Bottom Navigation -->
            <nav
                class="fixed bottom-0 left-0 right-0 bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 flex justify-around py-2 z-10">
                <RouterLink v-for="item in navItems" :key="item.nameKey" :to="item.route"
                    class="text-gray-500 dark:text-gray-400 hover:text-red-500 p-2 flex flex-col items-center">
                    <div v-html="icons[item.icon]"></div>
                    <span class="text-xs mt-1 dark:text-gray-300">{{ t(item.nameKey) }}</span>
                </RouterLink>
            </nav>
        </template>
    </div>
</template>

<style scoped>
/* Ensure the dropdown is visible and positioned correctly */
.relative {
    position: relative;
}
</style>
