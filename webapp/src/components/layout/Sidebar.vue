<script setup lang="ts">
import { RouterLink } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { useRouter } from 'vue-router';
import { ref, onMounted, onUnmounted, type Component } from 'vue';
import { useI18n } from 'vue-i18n';
import {
    BookmarksIcon,
    TagIcon,
    SettingsIcon,
    UserIcon
} from '@/components/icons';

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
    icon: Component;
    route: string;
}

const navItems: NavItem[] = [
    { nameKey: 'navigation.library', icon: BookmarksIcon, route: '/library' },
    { nameKey: 'navigation.tags', icon: TagIcon, route: '/tags' },
    { nameKey: 'navigation.settings', icon: SettingsIcon, route: '/settings' },
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
                        <component :is="item.icon" class="h-6 w-6" />
                        <span class="text-xs mt-1 dark:text-gray-300">{{ t(item.nameKey) }}</span>
                    </RouterLink>

                    <!-- Spacer -->
                    <div class="flex-1"></div>

                    <!-- User Menu -->
                    <div class="relative mt-auto" ref="menuRef">
                        <button @click.stop="toggleMenu"
                            class="text-gray-500 dark:text-gray-400 hover:text-red-500 p-2 rounded-md hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors flex flex-col items-center"
                            :title="t('auth.user')">
                            <UserIcon class="h-6 w-6" />
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
                            <RouterLink to="/settings"
                                class="block px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700">
                                {{ t('navigation.settings') }}
                            </RouterLink>
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
                    <component :is="item.icon" class="h-6 w-6" />
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
