<script setup lang="ts">
import { RouterLink, useRoute } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { useRouter } from 'vue-router';
import { ref, onMounted, onUnmounted, computed, type Component } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  BookmarksIcon,
  TagIcon,
  UserIcon,
  AdminIcon,
  ChevronDownIcon,
  ChevronRightIcon,
  GlobeIcon,
  FileTextIcon,
  DocumentIcon,
  ArchiveIcon,
  UsersIcon,
  CogIcon,
  ChartBarIcon,
  CheckCircleIcon,
} from '@/components/icons';

// Define props using the compiler macro (no import needed)
defineProps<{
  isMobile: boolean;
}>();

const { t } = useI18n();
const authStore = useAuthStore();
const router = useRouter();
const route = useRoute();
const isMenuOpen = ref(false);
const menuRef = ref<HTMLElement | null>(null);

interface NavItem {
  nameKey: string;
  icon: Component;
  route: string;
}

interface NavSection {
  nameKey: string;
  icon: Component;
  route: string;
  subsections?: NavItem[];
}

const navSections: NavSection[] = [
  {
    nameKey: 'navigation.library',
    icon: BookmarksIcon,
    route: '/library',
    // subsections: [
    //     { nameKey: 'navigation.all_bookmarks', icon: BookmarksIcon, route: '/library' },
    //     { nameKey: 'navigation.websites', icon: GlobeIcon, route: '/library?type=website' },
    //     { nameKey: 'navigation.pdfs', icon: FileTextIcon, route: '/library?type=pdf' },
    //     { nameKey: 'navigation.documents', icon: DocumentIcon, route: '/library?type=document' },
    //     { nameKey: 'navigation.archived', icon: ArchiveIcon, route: '/library?archived=true' }
    // ]
  },
  {
    nameKey: 'navigation.tags',
    icon: TagIcon,
    route: '/tags',
  },
];

// Add admin section only for admin users
const adminSection: NavSection = {
  nameKey: 'navigation.admin',
  icon: AdminIcon,
  route: '/admin',
  subsections: [
    { nameKey: 'admin.system_info', icon: CheckCircleIcon, route: '/admin' },
    {
      nameKey: 'admin.user_management',
      icon: UsersIcon,
      route: '/admin/users',
    },
  ],
};

// Track expanded sections - always expanded on desktop
const expandedSections = ref<Set<string>>(new Set(['library', 'admin']));

// Check if a route is active
const isRouteActive = (routePath: string) => {
  if (routePath.includes('?')) {
    const [path, query] = routePath.split('?');
    const currentPath = route.path;
    const currentQuery = route.query;

    if (path !== currentPath) return false;

    // Check query parameters
    const queryParams = new URLSearchParams(query);
    for (const [key, value] of queryParams.entries()) {
      if (currentQuery[key] !== value) return false;
    }
    return true;
  }
  return route.path === routePath;
};

// Check if a section is active (any of its subsections is active)
const isSectionActive = (section: NavSection) => {
  if (isRouteActive(section.route)) return true;
  if (section.subsections) {
    return section.subsections.some(subsection =>
      isRouteActive(subsection.route)
    );
  }
  return false;
};

// Toggle menu
const toggleMenu = () => {
  isMenuOpen.value = !isMenuOpen.value;
};

// Toggle section expansion
const toggleSection = (sectionKey: string) => {
  if (expandedSections.value.has(sectionKey)) {
    expandedSections.value.delete(sectionKey);
  } else {
    expandedSections.value.add(sectionKey);
  }
};

// Check if section is expanded
const isSectionExpanded = (sectionKey: string) => {
  return expandedSections.value.has(sectionKey);
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
        class="w-64 h-screen bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 flex flex-col py-6 sticky top-0"
      >
        <!-- Logo -->
        <div class="mb-8 px-6">
          <div class="flex items-center space-x-3">
            <div class="text-red-500 font-bold text-2xl">栞</div>
            <span class="text-lg font-semibold dark:text-gray-300">shiori</span>
          </div>
        </div>

        <!-- Navigation -->
        <nav class="flex flex-col space-y-1 flex-1 px-3">
          <!-- Main Navigation Sections -->
          <div
            v-for="section in navSections"
            :key="section.nameKey"
            class="space-y-1"
          >
            <!-- Section Header -->
            <div v-if="section.subsections" class="space-y-1">
              <RouterLink
                :to="section.route"
                class="flex items-center space-x-3 px-3 py-2 rounded-md transition-colors"
                :class="
                  isSectionActive(section)
                    ? 'text-red-500 bg-red-50 dark:bg-red-900/20 dark:text-red-400'
                    : 'text-gray-600 dark:text-gray-400 hover:text-red-500 hover:bg-gray-100 dark:hover:bg-gray-700'
                "
              >
                <component :is="section.icon" class="h-5 w-5 flex-shrink-0" />
                <span class="text-sm font-medium dark:text-gray-300">
                  {{ t(section.nameKey) }}
                </span>
              </RouterLink>

              <!-- Subsections - Always visible on desktop -->
              <div class="ml-6 space-y-1">
                <RouterLink
                  v-for="subsection in section.subsections"
                  :key="subsection.nameKey"
                  :to="subsection.route"
                  :class="[
                    'px-3 py-2 rounded-md transition-colors flex items-center space-x-3',
                    isRouteActive(subsection.route)
                      ? 'text-red-500 bg-red-50 dark:bg-red-900/20 dark:text-red-400'
                      : 'text-gray-500 dark:text-gray-500 hover:text-red-500 hover:bg-gray-100 dark:hover:bg-gray-700',
                  ]"
                  :title="t(subsection.nameKey)"
                >
                  <component
                    :is="subsection.icon"
                    class="h-4 w-4 flex-shrink-0"
                  />
                  <span class="text-xs font-medium dark:text-gray-400">
                    {{ t(subsection.nameKey) }}
                  </span>
                </RouterLink>
              </div>
            </div>

            <!-- Simple Section (no subsections) -->
            <RouterLink
              v-else
              :to="section.route"
              :class="[
                'px-3 py-2 rounded-md transition-colors flex items-center space-x-3',
                isRouteActive(section.route)
                  ? 'text-red-500 bg-red-50 dark:bg-red-900/20 dark:text-red-400'
                  : 'text-gray-600 dark:text-gray-400 hover:text-red-500 hover:bg-gray-100 dark:hover:bg-gray-700',
              ]"
              :title="t(section.nameKey)"
            >
              <component :is="section.icon" class="h-5 w-5 flex-shrink-0" />
              <span class="text-sm font-medium dark:text-gray-300">
                {{ t(section.nameKey) }}
              </span>
            </RouterLink>
          </div>

          <!-- Admin Section (only for admin users) -->
          <div v-if="authStore.user?.owner" class="space-y-1">
            <RouterLink
              :to="adminSection.route"
              class="flex items-center space-x-3 px-3 py-2 rounded-md transition-colors"
              :class="
                isSectionActive(adminSection)
                  ? 'text-red-500 bg-red-50 dark:bg-red-900/20 dark:text-red-400'
                  : 'text-gray-600 dark:text-gray-400 hover:text-red-500 hover:bg-gray-100 dark:hover:bg-gray-700'
              "
            >
              <component
                :is="adminSection.icon"
                class="h-5 w-5 flex-shrink-0"
              />
              <span class="text-sm font-medium dark:text-gray-300">
                {{ t(adminSection.nameKey) }}
              </span>
            </RouterLink>

            <!-- Admin Subsections - Always visible on desktop -->
            <div class="ml-6 space-y-1">
              <RouterLink
                v-for="subsection in adminSection.subsections"
                :key="subsection.nameKey"
                :to="subsection.route"
                :class="[
                  'px-3 py-2 rounded-md transition-colors flex items-center space-x-3',
                  isRouteActive(subsection.route)
                    ? 'text-red-500 bg-red-50 dark:bg-red-900/20 dark:text-red-400'
                    : 'text-gray-500 dark:text-gray-500 hover:text-red-500 hover:bg-gray-100 dark:hover:bg-gray-700',
                ]"
                :title="t(subsection.nameKey)"
              >
                <component
                  :is="subsection.icon"
                  class="h-4 w-4 flex-shrink-0"
                />
                <span class="text-xs font-medium dark:text-gray-400">
                  {{ t(subsection.nameKey) }}
                </span>
              </RouterLink>
            </div>
          </div>

          <!-- Spacer -->
          <div class="flex-1"></div>

          <!-- User Menu -->
          <div class="relative mt-auto px-3" ref="menuRef">
            <button
              @click.stop="toggleMenu"
              class="w-full text-gray-600 dark:text-gray-400 hover:text-red-500 hover:bg-gray-100 dark:hover:bg-gray-700 px-3 py-2 rounded-md transition-colors flex items-center space-x-3"
              :title="t('auth.user')"
            >
              <UserIcon class="h-5 w-5 flex-shrink-0" />
              <span class="text-sm font-medium dark:text-gray-300">
                {{ authStore.user?.username || t('auth.user') }}
              </span>
            </button>

            <!-- Dropdown Menu -->
            <div
              v-if="isMenuOpen"
              class="absolute left-3 bottom-12 w-48 bg-white dark:bg-gray-800 rounded-md shadow-lg py-1 z-50 border border-gray-200 dark:border-gray-700"
            >
              <div
                class="px-4 py-2 text-sm text-gray-500 dark:text-gray-400 border-b border-gray-200 dark:border-gray-700"
              >
                <div class="font-medium dark:text-gray-300">
                  {{ authStore.user?.username || t('auth.user') }}
                </div>
              </div>
              <RouterLink
                to="/settings"
                class="block px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700"
              >
                {{ t('navigation.settings') }}
              </RouterLink>
              <button
                @click="handleLogout"
                class="block w-full text-left px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700"
              >
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
        class="fixed bottom-0 left-0 right-0 bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 flex justify-around py-2 z-10"
      >
        <RouterLink
          v-for="section in navSections"
          :key="section.nameKey"
          :to="section.route"
          class="text-gray-500 dark:text-gray-400 hover:text-red-500 p-2 flex flex-col items-center"
        >
          <component :is="section.icon" class="h-6 w-6" />
          <span class="text-xs mt-1 dark:text-gray-300">
            {{ t(section.nameKey) }}
          </span>
        </RouterLink>
        <RouterLink
          v-if="authStore.user?.owner"
          :to="adminSection.route"
          class="text-gray-500 dark:text-gray-400 hover:text-red-500 p-2 flex flex-col items-center"
        >
          <component :is="adminSection.icon" class="h-6 w-6" />
          <span class="text-xs mt-1 dark:text-gray-300">
            {{ t(adminSection.nameKey) }}
          </span>
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
