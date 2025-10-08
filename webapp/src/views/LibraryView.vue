<script setup lang="ts">
import { ref, onMounted, computed, onUnmounted, watch } from 'vue';
import { useRouter } from 'vue-router';
import { storeToRefs } from 'pinia';
import { useI18n } from 'vue-i18n'
import { Input } from '@/components/ui';
import AppLayout from '@/components/layout/AppLayout.vue';
import Pagination from '@/components/ui/Pagination.vue';
import ViewSelector from '@/components/ui/ViewSelector.vue';
import BookmarkCard from '@/components/ui/BookmarkCard.vue';
import BookmarkThumbnail from '@/components/ui/BookmarkThumbnail.vue';
import DeleteConfirmationModal from '@/components/ui/DeleteConfirmationModal.vue';
import { useBookmarksStore } from '@/stores/bookmarks';
import { useAuthStore } from '@/stores/auth';
import { useToast } from '@/composables/useToast';
import { ImageIcon, PencilIcon, TrashIcon, ArchiveIcon, BookIcon, FileTextIcon, ExternalLinkIcon, PlusIcon } from '@/components/icons';

const bookmarksStore = useBookmarksStore();
const authStore = useAuthStore();
const router = useRouter();
const { t } = useI18n();
const { success, error: showErrorToast } = useToast();

const { bookmarks, isLoading, error, totalCount, currentPage, pageLimit } = storeToRefs(bookmarksStore);
const { fetchBookmarks, deleteBookmarks } = bookmarksStore;

const searchKeyword = ref('');

// Delete state
const bookmarkToDelete = ref<any>(null);
const isDeleting = ref(false);

// Respect user setting to hide excerpts
const shouldHideExcerpt = computed(() => authStore.user?.config?.HideExcerpt === true);

// Initialize view from user config or localStorage or default to 'list'
const getStoredView = (): 'list' | 'card' => {
    if (typeof window !== 'undefined') {
        // First check user configuration (handle both PascalCase and camelCase)
        const config = authStore.user?.config as any;
        const listMode = config?.ListMode ?? config?.listMode;
        if (listMode !== undefined) {
            return listMode ? 'list' : 'card';
        }
        // Fallback to localStorage
        const stored = localStorage.getItem('shiori-view-preference');
        return (stored === 'list' || stored === 'card') ? stored : 'list';
    }
    return 'list';
};

const currentView = ref<'list' | 'card'>(getStoredView());
const isMobile = ref(false);

// Detect mobile screen size
const checkMobile = () => {
    isMobile.value = window.innerWidth < 768; // md breakpoint
};

// Handle view change with persistence
const handleViewChange = async (view: 'list' | 'card') => {
    currentView.value = view;
    // Store the preference in localStorage
    if (typeof window !== 'undefined') {
        localStorage.setItem('shiori-view-preference', view);
    }

    // Update user configuration with listMode preference
    if (authStore.isAuthenticated && authStore.user) {
        try {
            const listMode = view === 'list';
            const currentConfig = authStore.user.config as any || {};
            // Ensure we send the complete config object with the updated field
            const apiConfig = {
                ShowId: currentConfig.ShowId || false,
                ListMode: listMode,
                HideThumbnail: currentConfig.HideThumbnail || false,
                HideExcerpt: currentConfig.HideExcerpt || false,
                Theme: currentConfig.Theme || 'system',
                KeepMetadata: currentConfig.KeepMetadata || false,
                UseArchive: currentConfig.UseArchive || false,
                CreateEbook: currentConfig.CreateEbook || false,
                MakePublic: currentConfig.MakePublic || false,
            };
            console.log('Updating user config with complete object:', apiConfig);
            await authStore.updateUserConfig(apiConfig as any);
        } catch (error) {
            console.error('Failed to update user configuration:', error);
        }
    } else {
        console.log('Not authenticated or no user:', {
            isAuthenticated: authStore.isAuthenticated,
            hasUser: !!authStore.user,
            hasToken: !!authStore.token,
            expires: authStore.expires,
            currentTime: Date.now()
        });
    }
};

// Watch for changes in user configuration to update view
watch(() => {
    const config = authStore.user?.config as any;
    return config?.ListMode ?? config?.listMode;
}, (newListMode) => {
    if (newListMode !== undefined) {
        currentView.value = newListMode ? 'list' : 'card';
    }
});

// Computed property for effective view (force card on mobile)
const effectiveView = computed(() => {
    return isMobile.value ? 'card' : currentView.value;
});

// Fetch bookmarks on mount
onMounted(async () => {
    try {
        // Ensure user is loaded
        if (!authStore.user) {
            await authStore.fetchUserInfo();
        }

        await fetchBookmarks();
        checkMobile(); // Check mobile on mount
        window.addEventListener('resize', checkMobile); // Listen for resize events
    } catch (err) {
        console.error('Error loading bookmarks:', err);
        // Handle authentication errors
        if (err instanceof Error && err.message.includes('401')) {
            authStore.clearAuth();
            router.push('/login');
        }
    }
});

// Search bookmarks
const handleSearch = async () => {
    try {
        await fetchBookmarks({ keyword: searchKeyword.value, page: 1 }); // Reset to page 1 when searching
    } catch (err) {
        console.error('Error searching bookmarks:', err);
    }
};

// Handle page change
const handlePageChange = async (page: number) => {
    try {
        await fetchBookmarks({ page, limit: pageLimit.value, keyword: searchKeyword.value });
    } catch (err) {
        console.error('Error changing page:', err);
    }
};

// Handle per page change
const handlePerPageChange = async (perPage: number) => {
    try {
        await fetchBookmarks({ page: 1, limit: perPage, keyword: searchKeyword.value });
    } catch (err) {
        console.error('Error changing items per page:', err);
    }
};

// Helper to get tag names from bookmark
const getBookmarkTags = (bookmark: any) => {
    // For now, return empty array as tags are separate in v1 API
    // Tags will need to be fetched separately per bookmark
    return [];
};

// Delete handlers
const handleDeleteBookmark = (bookmark: any) => {
    bookmarkToDelete.value = bookmark;
};

const confirmDeleteBookmark = async () => {
    if (!bookmarkToDelete.value) return;

    isDeleting.value = true;
    try {
        await deleteBookmarks([bookmarkToDelete.value.id]);
        bookmarkToDelete.value = null;

        success(
            t('bookmarks.toast.deleted_success'),
            t('bookmarks.toast.deleted_success_message')
        );
    } catch (err) {
        console.error('Failed to delete bookmark:', err);
        showErrorToast(
            t('bookmarks.toast.deleted_error'),
            t('bookmarks.toast.deleted_error_message')
        );
    } finally {
        isDeleting.value = false;
    }
};

const cancelDeleteBookmark = () => {
    bookmarkToDelete.value = null;
};

// Cleanup resize listener
onUnmounted(() => {
    window.removeEventListener('resize', checkMobile);
});

</script>

<template>
    <AppLayout>
        <template #header>
            <div class="flex justify-between items-center">
                <h1 class="text-xl font-bold text-gray-800 dark:text-white">{{ t('bookmarks.my_bookmarks') }}</h1>
                <div class="flex space-x-2">
                    <button @click="$router.push('/add-bookmark')"
                        class="bg-red-500 text-white px-3 py-1 rounded-md hover:bg-red-600 flex items-center space-x-2">
                        <PlusIcon size="16" />
                        <span>{{ t('bookmarks.add_bookmark') }}</span>
                    </button>
                    <div class="relative">
                        <Input v-model="searchKeyword" @keyup.enter="handleSearch" type="search" variant="search"
                            size="sm" :placeholder="t('bookmarks.search_placeholder')" />
                    </div>
                </div>
            </div>
        </template>

        <div class="mt-6">
            <!-- View Selector -->
            <div class="flex justify-between items-center mb-4">
                <div class="flex items-center space-x-4">
                    <!-- Hide view selector on mobile, force card view -->
                    <ViewSelector v-if="!isMobile" :current-view="currentView" :on-view-change="handleViewChange" />
                    <!-- Mobile: Force card view -->
                    <div v-else class="text-sm text-gray-500 dark:text-gray-400">
                        {{ t('bookmarks.card_view') }}
                    </div>
                </div>
            </div>

            <!-- Loading state -->
            <div v-if="isLoading" class="text-center py-8">
                <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-red-500"></div>
                <p class="mt-2 text-gray-600 dark:text-gray-400">{{ t('bookmarks.loading_bookmarks') }}</p>
            </div>

            <!-- Error state -->
            <div v-else-if="error" class="bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-300 p-4 rounded-md">
                {{ error }}
            </div>

            <!-- Empty state -->
            <div v-else-if="bookmarks.length === 0" class="text-center py-12">
                <p class="text-gray-600 dark:text-gray-400 text-lg">{{ t('bookmarks.no_bookmarks_found') }}</p>
                <p class="text-gray-500 dark:text-gray-500 text-sm mt-2">{{ t('bookmarks.create_first_bookmark') }}</p>
            </div>

            <!-- List View -->
            <ul v-else-if="effectiveView === 'list'" class="space-y-4">
                <li v-for="bookmark in bookmarks" :key="bookmark.id"
                    class="bg-white dark:bg-gray-800 p-4 rounded-md shadow-sm hover:shadow-md transition-shadow cursor-pointer"
                    @click="$router.push(`/bookmark/${bookmark.id}/content`)">
                    <div class="flex gap-4">
                        <!-- Thumbnail -->
                        <div class="flex-shrink-0">
                            <BookmarkThumbnail :bookmark="bookmark" size="medium" />
                        </div>

                        <!-- Content -->
                        <div class="flex-1 min-w-0">
                            <div class="flex justify-between items-start">
                                <div class="flex items-start gap-2 flex-1 min-w-0">
                                    <h3 class="text-blue-600 dark:text-blue-400 font-medium truncate">
                                        {{ bookmark.title || bookmark.url }}
                                    </h3>
                                    <!-- Feature icons -->
                                    <div class="flex items-center gap-1 flex-shrink-0">
                                        <FileTextIcon v-if="bookmark.hasContent"
                                            class="h-4 w-4 text-gray-500 dark:text-gray-400"
                                            :title="t('bookmarks.has_readable_content')" />
                                        <ArchiveIcon v-if="bookmark.hasArchive"
                                            class="h-4 w-4 text-gray-500 dark:text-gray-400"
                                            :title="t('bookmarks.has_archive')" />
                                        <BookIcon v-if="bookmark.hasEbook"
                                            class="h-4 w-4 text-gray-500 dark:text-gray-400"
                                            :title="t('bookmarks.has_ebook')" />
                                    </div>
                                </div>
                                <div class="flex space-x-2 ml-4 flex-shrink-0">
                                    <a v-if="bookmark.url" :href="bookmark.url" target="_blank" @click.stop
                                        class="text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300">
                                        <span class="sr-only">{{ t('bookmarks.open_original_url') }}</span>
                                        <ExternalLinkIcon class="h-5 w-5" />
                                    </a>
                                    <button @click.stop
                                        class="text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300">
                                        <span class="sr-only">{{ t('bookmarks.edit_bookmark_action') }}</span>
                                        <PencilIcon class="h-5 w-5" />
                                    </button>
                                    <button @click.stop="handleDeleteBookmark(bookmark)"
                                        class="text-gray-500 dark:text-gray-400 hover:text-red-500">
                                        <span class="sr-only">{{ t('bookmarks.delete_bookmark_action') }}</span>
                                        <TrashIcon class="h-5 w-5" />
                                    </button>
                                </div>
                            </div>
                            <div class="text-gray-500 dark:text-gray-400 text-sm mt-1 truncate">{{ bookmark.url }}</div>
                            <div v-if="bookmark.excerpt && !shouldHideExcerpt"
                                class="text-gray-600 dark:text-gray-400 text-sm mt-2 line-clamp-2">
                                {{ bookmark.excerpt }}
                            </div>
                        </div>
                    </div>
                </li>
            </ul>

            <!-- Card View -->
            <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-4">
                <BookmarkCard v-for="bookmark in bookmarks" :key="bookmark.id" :bookmark="bookmark"
                    :auth-token="authStore.token || undefined" @delete="handleDeleteBookmark" />
            </div>

            <!-- Pagination -->
            <Pagination v-if="totalCount > pageLimit" :current-page="currentPage" :total-items="totalCount"
                :items-per-page="pageLimit" @page-change="handlePageChange" @per-page-change="handlePerPageChange" />
        </div>

        <!-- Delete Confirmation Modal -->
        <DeleteConfirmationModal :is-open="bookmarkToDelete !== null" :title="t('bookmarks.delete_bookmark')"
            :message="t('bookmarks.confirm_delete_message')"
            :item-name="bookmarkToDelete?.title || bookmarkToDelete?.url" :is-loading="isDeleting"
            @close="cancelDeleteBookmark" @confirm="confirmDeleteBookmark" />
    </AppLayout>
</template>
