<script setup lang="ts">
import AppLayout from '@/components/layout/AppLayout.vue';
import Pagination from '@/components/ui/Pagination.vue';
import { useTagsStore } from '@/stores/tags';
import { useAuthStore } from '@/stores/auth';
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { storeToRefs } from 'pinia';
import { CheckIcon, XIcon, TagIcon, PencilIcon, TrashIcon } from '@/components/icons';

const tagsStore = useTagsStore();
const authStore = useAuthStore();
const router = useRouter();
const { tags, isLoading, error, totalCount, currentPage, pageLimit } = storeToRefs(tagsStore);
const { fetchTags, createTag, updateTag, deleteTag } = tagsStore;

// New tag form
const showNewTagForm = ref(false);
const newTagName = ref('');
const isSubmitting = ref(false);
const formError = ref<string | null>(null);

// Edit tag form
const editingTagId = ref<number | null>(null);
const editTagName = ref('');

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

// Handle new tag submission
const handleCreateTag = async () => {
  if (!newTagName.value.trim()) {
    formError.value = 'Tag name cannot be empty';
    return;
  }

  formError.value = null;
  isSubmitting.value = true;

  try {
    await createTag(newTagName.value.trim());
    newTagName.value = '';
    showNewTagForm.value = false;
  } catch (err) {
    // Check for authentication errors
    handleApiError(err);
  } finally {
    isSubmitting.value = false;
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
  } catch (err) {
    // Check for authentication errors
    handleApiError(err);
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
  } catch (err) {
    // Check for authentication errors
    handleApiError(err);
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
</script>

<template>
  <AppLayout>
    <template #header>
      <div class="flex justify-between items-center">
        <h1 class="text-xl font-bold">Tags</h1>
        <div class="flex space-x-2">
          <button @click="showNewTagForm = !showNewTagForm"
            class="bg-blue-500 text-white px-3 py-1 rounded-md hover:bg-blue-600 transition">
            {{ showNewTagForm ? 'Cancel' : 'New Tag' }}
          </button>
        </div>
      </div>
    </template>

    <!-- New Tag Form -->
    <div v-if="showNewTagForm" class="bg-white p-4 rounded-md shadow-sm mb-6">
      <h2 class="text-lg font-medium mb-3">Create New Tag</h2>
      <form @submit.prevent="handleCreateTag" class="flex flex-col space-y-3">
        <div>
          <label for="tagName" class="block text-sm font-medium text-gray-700 mb-1">Tag Name</label>
          <input id="tagName" v-model="newTagName" type="text"
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="Enter tag name" :disabled="isSubmitting" />
          <p v-if="formError" class="mt-1 text-sm text-red-600">{{ formError }}</p>
        </div>
        <div class="flex justify-end space-x-2">
          <button type="button" @click="showNewTagForm = false"
            class="px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50" :disabled="isSubmitting">
            Cancel
          </button>
          <button type="submit"
            class="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 disabled:opacity-50"
            :disabled="isSubmitting">
            Create
          </button>
        </div>
      </form>
    </div>

    <!-- Error Message -->
    <div v-if="error" class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-md mb-6">
      {{ error }}
    </div>

    <!-- Loading State -->
    <div v-if="isLoading && !tags.length" class="bg-white p-6 rounded-md shadow-sm flex justify-center">
      <div class="animate-pulse text-gray-500">Loading tags...</div>
    </div>

    <!-- Empty State -->
    <div v-else-if="!isLoading && !tags.length" class="bg-white p-6 rounded-md shadow-sm text-center">
      <p class="text-gray-500 mb-4">No tags found. Create your first tag to organize your bookmarks.</p>
      <button v-if="!showNewTagForm" @click="showNewTagForm = true"
        class="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600">
        Create Tag
      </button>
    </div>

    <!-- Tag List -->
    <div v-else class="mt-6">
      <ul class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <li v-for="tag in tags" :key="tag.id"
          class="bg-white p-4 rounded-md shadow-sm hover:shadow-md transition-shadow border border-gray-200">
          <!-- Edit Mode -->
          <div v-if="editingTagId === tag.id" class="flex items-center">
            <input v-model="editTagName" type="text"
              class="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              :disabled="isSubmitting" />
            <div class="flex ml-2 space-x-1">
              <button @click="handleUpdateTag(tag.id!)" class="text-blue-500 hover:text-blue-700 p-1"
                :disabled="isSubmitting" title="Save">
                <CheckIcon class="h-5 w-5" />
              </button>
              <button @click="cancelEdit" class="text-gray-500 hover:text-gray-700 p-1" :disabled="isSubmitting"
                title="Cancel">
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
              <h3 class="font-medium text-lg">{{ tag.name }}</h3>
              <p class="text-sm text-gray-500">{{ tag.bookmarkCount || 0 }} bookmarks</p>
            </div>
            <div class="flex space-x-1">
              <button @click="startEditTag(tag.id!, tag.name!)" class="text-gray-400 hover:text-gray-600 p-1"
                title="Edit">
                <PencilIcon class="h-5 w-5" />
              </button>
              <button @click="confirmDeleteTag(tag.id!)" class="text-gray-400 hover:text-red-500 p-1" title="Delete">
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
      <div class="bg-white rounded-lg p-6 max-w-md w-full">
        <h3 class="text-lg font-medium mb-4">Confirm Delete</h3>
        <p class="mb-6">Are you sure you want to delete this tag? This action cannot be undone.</p>
        <div class="flex justify-end space-x-3">
          <button @click="tagToDelete = null"
            class="px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50">
            Cancel
          </button>
          <button @click="handleDeleteTag" class="px-4 py-2 bg-red-500 text-white rounded-md hover:bg-red-600">
            Delete
          </button>
        </div>
      </div>
    </div>
  </AppLayout>
</template>
