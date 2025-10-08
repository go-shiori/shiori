<script setup lang="ts">
import AppLayout from '@/components/layout/AppLayout.vue';
import Pagination from '@/components/ui/Pagination.vue';
import { useTagsStore } from '@/stores/tags';
import { useAuthStore } from '@/stores/auth';
import { useToast } from '@/composables/useToast';
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { storeToRefs } from 'pinia';
import { useI18n } from 'vue-i18n'
import { Input } from '@/components/ui';
import { CheckIcon, XIcon, TagIcon, PencilIcon, TrashIcon, PlusIcon, SearchIcon } from '@/components/icons';
import { useErrorHandler } from '@/utils/errorHandler';

const { t } = useI18n();
const tagsStore = useTagsStore();
const authStore = useAuthStore();
const router = useRouter();
const { handleApiError: handleApiErrorWithI18n } = useErrorHandler();
const { success, error: showErrorToast } = useToast();
const { tags, isLoading, error, totalCount, currentPage, pageLimit } = storeToRefs(tagsStore);
const { fetchTags, createTag, updateTag, deleteTag } = tagsStore;

// Navigation to add tag view
const navigateToAddTag = () => {
  router.push('/add-tag');
};

// Edit tag form
const editingTagId = ref<number | null>(null);
const editTagName = ref('');
const isSubmitting = ref(false);

// Search functionality
const searchQuery = ref('');
const isSearching = ref(false);

// Load tags on component mount
onMounted(async () => {
  try {
    if (authStore.isAuthenticated) {
      await fetchTags();
    } else {
      const isValid = await authStore.validateToken();
      if (isValid) {
        await fetchTags();
        console.log("tags", tags);
      } else {
        authStore.setRedirectDestination('/tags');
        router.push('/login');
      }
    }
  } catch (err) {
    // If we get an authentication error, redirect to login
    if (err instanceof Error && err.message.includes('401')) {
      authStore.setRedirectDestination('/tags');
      router.push('/login');
    }
  }
});

// Handle API errors and authentication
const handleApiError = (err: any) => {
  if (err instanceof Error && err.message.includes('401')) {
    authStore.setRedirectDestination('/tags');
    router.push('/login');
  }
};


// Start editing a tag
const startEditTag = (id: number, name: string) => {
  editingTagId.value = id;
  editTagName.value = name;
};

// Cancel editing
const cancelEdit = () => {
  editingTagId.value = null;
  editTagName.value = '';
};

// Save edited tag
const handleUpdateTag = async (id: number) => {
  if (!editTagName.value.trim()) {
    return;
  }

  isSubmitting.value = true;

  try {
    await updateTag(id, editTagName.value.trim());
    editingTagId.value = null;

    // Show success toast
    success(
      t('tags.toast.updated_success'),
      t('tags.toast.updated_success_message')
    );
  } catch (err) {
    console.error('Failed to update tag:', err);

    // Check for authentication errors
    handleApiError(err);

    // Show error toast
    showErrorToast(
      t('tags.toast.updated_error'),
      t('tags.toast.updated_error_message')
    );
  } finally {
    isSubmitting.value = false;
  }
};

// Delete tag confirmation
const tagToDelete = ref<number | null>(null);
const confirmDeleteTag = (id: number) => {
  tagToDelete.value = id;
};

// Handle tag deletion
const handleDeleteTag = async () => {
  if (tagToDelete.value === null) return;

  try {
    await deleteTag(tagToDelete.value);
    tagToDelete.value = null;

    // Show success toast
    success(
      t('tags.toast.deleted_success'),
      t('tags.toast.deleted_success_message')
    );
  } catch (err) {
    console.error('Failed to delete tag:', err);

    // Check for authentication errors
    handleApiError(err);

    // Show error toast
    showErrorToast(
      t('tags.toast.deleted_error'),
      t('tags.toast.deleted_error_message')
    );
  }
};

// Handle page change
const handlePageChange = async (page: number) => {
  try {
    await fetchTags({ page, limit: pageLimit.value });
  } catch (err) {
    handleApiError(err);
  }
};

// Handle per page change
const handlePerPageChange = async (perPage: number) => {
  try {
    await fetchTags({ page: 1, limit: perPage }); // Reset to page 1 when changing per page
  } catch (err) {
    handleApiError(err);
  }
};

// Debounced search function
let searchTimeout: number;
const handleSearch = async () => {
  clearTimeout(searchTimeout);
  searchTimeout = setTimeout(async () => {
    isSearching.value = true;
    try {
      await fetchTags({
        page: 1,
        limit: pageLimit.value,
        search: searchQuery.value.trim() || undefined
      });
    } catch (err) {
      handleApiError(err);
    } finally {
      isSearching.value = false;
    }
  }, 300);
};

// Clear search
const clearSearch = async () => {
  searchQuery.value = '';
  isSearching.value = true;
  try {
    await fetchTags({ page: 1, limit: pageLimit.value });
  } catch (err) {
    handleApiError(err);
  } finally {
    isSearching.value = false;
  }
};
</script>

<template>
  <AppLayout>
    <template #header>
      <div class="flex justify-between items-center">
        <h1 class="text-xl font-bold">{{ t('tags.title') }}</h1>
        <div class="flex space-x-2">
          <button @click="navigateToAddTag"
            class="bg-red-500 text-white px-3 py-1 rounded-md hover:bg-red-600 transition flex items-center space-x-2 h-8 text-sm">
            <PlusIcon size="16" />
            <span>{{ t('tags.add_tag') }}</span>
          </button>
          <div class="relative">
            <Input v-model="searchQuery" @input="handleSearch" type="search" variant="search" size="sm" class="h-8"
              :placeholder="t('tags.search_placeholder')" />
            <div v-if="isSearching" class="absolute inset-y-0 right-0 pr-3 flex items-center pointer-events-none">
              <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-red-500"></div>
            </div>
          </div>
        </div>
      </div>
    </template>


    <!-- Error Message -->
    <div v-if="error"
      class="bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-300 px-4 py-3 rounded-md mb-6">
      {{ error }}
    </div>

    <!-- Loading State -->
    <div v-if="isLoading && !tags.length"
      class="bg-white dark:bg-gray-800 p-6 rounded-md shadow-sm flex justify-center">
      <div class="animate-pulse text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</div>
    </div>

    <!-- Empty State -->
    <div v-else-if="!isLoading && !tags.length" class="bg-white dark:bg-gray-800 p-6 rounded-md shadow-sm text-center">
      <p class="text-gray-500 dark:text-gray-400 mb-4">{{ t('tags.create_first_tag') }}</p>
      <button @click="navigateToAddTag"
        class="px-4 py-2 bg-red-500 text-white rounded-md hover:bg-red-600 flex items-center space-x-2">
        <PlusIcon size="16" />
        <span>{{ t('tags.add_tag') }}</span>
      </button>
    </div>

    <!-- Tag List -->
    <div v-else class="mt-6">
      <ul class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <li v-for="tag in tags" :key="tag.id"
          class="bg-white dark:bg-gray-800 p-4 rounded-md shadow-sm hover:shadow-md transition-shadow border border-gray-200 dark:border-gray-700">
          <!-- Edit Mode -->
          <div v-if="editingTagId === tag.id" class="flex items-center">
            <Input v-model="editTagName" type="text" :disabled="isSubmitting" />
            <div class="flex ml-2 space-x-1">
              <button @click="handleUpdateTag(tag.id!)"
                class="text-red-500 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300 p-1"
                :disabled="isSubmitting" title="Save">
                <CheckIcon class="h-5 w-5" />
              </button>
              <button @click="cancelEdit"
                class="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300 p-1"
                :disabled="isSubmitting" title="Cancel">
                <XIcon class="h-5 w-5" />
              </button>
            </div>
          </div>

          <!-- View Mode -->
          <div v-else class="flex items-center">
            <div class="mr-3 text-blue-400">
              <TagIcon class="h-6 w-6" />
            </div>
            <div class="flex-1">
              <h3 class="font-medium text-lg text-gray-900 dark:text-gray-100">{{ tag.name }}</h3>
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ tag.bookmarkCount || 0 }} {{
                t('tags.bookmarks_count')
                }}</p>
            </div>
            <div class="flex space-x-1">
              <button @click="startEditTag(tag.id!, tag.name!)"
                class="text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300 p-1"
                :title="t('common.edit')">
                <PencilIcon class="h-5 w-5" />
              </button>
              <button @click="confirmDeleteTag(tag.id!)"
                class="text-gray-400 hover:text-red-500 dark:text-gray-500 dark:hover:text-red-400 p-1"
                :title="t('common.delete')">
                <TrashIcon class="h-5 w-5" />
              </button>
            </div>
          </div>
        </li>
      </ul>

      <!-- Pagination -->
      <Pagination :current-page="currentPage" :total-items="totalCount" :items-per-page="pageLimit"
        @page-change="handlePageChange" @per-page-change="handlePerPageChange" />
    </div>

    <!-- Delete Confirmation Modal -->
    <div v-if="tagToDelete !== null" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white dark:bg-gray-800 rounded-lg p-6 max-w-md w-full">
        <h3 class="text-lg font-medium mb-4 text-gray-900 dark:text-gray-100">{{ t('tags.delete_tag') }}</h3>
        <p class="mb-6 text-gray-700 dark:text-gray-300">{{ t('tags.confirm_delete') }}</p>
        <div class="flex justify-end space-x-3">
          <button @click="tagToDelete = null"
            class="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700">
            {{ t('common.cancel') }}
          </button>
          <button @click="handleDeleteTag" class="px-4 py-2 bg-red-500 text-white rounded-md hover:bg-red-600">
            {{ t('common.delete') }}
          </button>
        </div>
      </div>
    </div>
  </AppLayout>
</template>
