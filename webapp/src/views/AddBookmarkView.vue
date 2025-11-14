<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useBookmarksStore } from '@/stores/bookmarks';
import { useTagsStore } from '@/stores/tags';
import { useToast } from '@/composables/useToast';
import AppLayout from '@/components/layout/AppLayout.vue';
import {
  TextInput,
  Textarea,
  Select,
  Checkbox,
  TagSelector,
  Button,
} from '@/components/ui';
import { useI18n } from 'vue-i18n';
import { useErrorHandler } from '@/utils/errorHandler';

const { t } = useI18n();
const router = useRouter();
const bookmarksStore = useBookmarksStore();
const tagsStore = useTagsStore();
const { success, error } = useToast();
const { handleApiError } = useErrorHandler();

// Form fields
const url = ref('');
const title = ref('');
const excerpt = ref('');
const selectedTagIds = ref<number[]>([]);
const createArchive = ref(true);
const createEbook = ref(false);
const visibility = ref<'internal' | 'public'>('internal');

// State
const isLoading = ref(false);
const formError = ref<string | null>(null);

const handleSubmit = async () => {
  // Validation
  if (!url.value.trim()) {
    formError.value = t('bookmarks.please_enter_url');
    return;
  }

  // Basic URL validation
  try {
    new URL(url.value);
  } catch {
    formError.value = t('bookmarks.please_enter_valid_url');
    return;
  }

  isLoading.value = true;
  formError.value = null;

  try {
    // 1. Create bookmark with options
    const newBookmark = await bookmarksStore.createBookmark(
      url.value.trim(),
      title.value.trim() || undefined,
      excerpt.value.trim() || undefined,
      visibility.value === 'public' ? 1 : 0,
      {
        createArchive: createArchive.value,
        createEbook: createEbook.value,
      }
    );

    // 2. Associate tags if provided
    if (selectedTagIds.value.length > 0 && newBookmark.id) {
      for (const tagId of selectedTagIds.value) {
        try {
          await bookmarksStore.addTagToBookmark(newBookmark.id, tagId);
        } catch (tagErr) {
          console.error(`Failed to add tag with ID ${tagId}:`, tagErr);
          // Continue with other tags
        }
      }
    }

    // Show success toast
    success(
      t('bookmarks.toast.created_success'),
      t('bookmarks.toast.created_success_message')
    );

    // Redirect to library page after successful creation
    router.push('/library');
  } catch (err) {
    console.error('Failed to create bookmark:', err);
    const errorMessage = handleApiError(err as any, 'bookmark');
    formError.value = errorMessage;
    // Don't show toast for form errors - they're displayed inline
  } finally {
    isLoading.value = false;
  }
};

const handleCancel = () => {
  router.push('/library');
};

// Load tags on mount
onMounted(async () => {
  try {
    await tagsStore.fetchTags();
  } catch (err) {
    console.warn('Failed to load tags:', err);
    // Don't block the form if tags fail to load
  }
});
</script>

<template>
  <AppLayout>
    <template #header>
      <div class="flex justify-between items-center">
        <h1 class="text-xl font-bold text-gray-800 dark:text-white">
          {{ t('bookmarks.new_bookmark') }}
        </h1>
        <button @click="handleCancel"
          class="text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200">
          {{ t('common.cancel') }}
        </button>
      </div>
    </template>

    <div class="max-w-2xl mx-auto">
      <div class="bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg shadow-sm">
        <!-- Dialog Header -->
        <div class="bg-gray-800 text-white px-4 py-3 rounded-t-lg">
          <h2 class="text-lg font-semibold uppercase">
            {{ t('bookmarks.create_new_bookmark') }}
          </h2>
        </div>

        <!-- Dialog Body -->
        <div class="p-4 space-y-4">
          <!-- URL Field -->
          <div>
            <label for="url" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {{ t('bookmarks.url_label') }}
            </label>
            <TextInput id="url" v-model="url" type="url" :placeholder="t('bookmarks.url_placeholder')" name="url"
              autocomplete="url" :disabled="isLoading" required />
          </div>

          <!-- Custom Title Field -->
          <div>
            <label for="title" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {{ t('bookmarks.custom_title_label') }}
            </label>
            <TextInput id="title" v-model="title" type="text" :placeholder="t('bookmarks.custom_title_placeholder')"
              name="title" :disabled="isLoading" />
          </div>

          <!-- Custom Excerpt Field -->
          <div>
            <label for="excerpt" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {{ t('bookmarks.custom_excerpt_label') }}
            </label>
            <Textarea id="excerpt" v-model="excerpt" :rows="3" :placeholder="t('bookmarks.custom_excerpt_placeholder')"
              name="excerpt" :disabled="isLoading" />
          </div>

          <!-- Tags Field -->
          <div>
            <label for="tags" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {{ t('bookmarks.tags_label') }}
            </label>
            <TagSelector v-model="selectedTagIds" :disabled="isLoading"
              :placeholder="t('bookmarks.tags_placeholder')" />
          </div>

          <!-- Checkboxes -->
          <div class="space-y-3">
            <label class="flex items-center cursor-pointer">
              <Checkbox v-model="createArchive" :disabled="isLoading" class="mr-2" />
              <span class="text-sm text-gray-700 dark:text-gray-300">
                {{ t('bookmarks.create_archive') }}
              </span>
            </label>

            <label class="flex items-center cursor-pointer">
              <Checkbox v-model="createEbook" :disabled="isLoading" class="mr-2" />
              <span class="text-sm text-gray-700 dark:text-gray-300">
                {{ t('bookmarks.generate_ebook') }}
              </span>
            </label>

            <!-- Visibility Select -->
            <div>
              <label for="visibility" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                {{ t('bookmarks.visibility_label') }}
              </label>
              <Select id="visibility" v-model="visibility" :options="[
                {
                  value: 'internal',
                  label: t('bookmarks.visibility_internal'),
                },
                { value: 'public', label: t('bookmarks.visibility_public') },
              ]" :disabled="isLoading" />
              <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                {{ t('bookmarks.visibility_description') }}
              </p>
            </div>
          </div>
        </div>

        <!-- Dialog Footer -->
        <div :class="[
          'bg-gray-50 dark:bg-gray-700 px-4 py-3 rounded-b-lg border-t border-gray-200 dark:border-gray-600 flex items-center',
          formError ? 'justify-between' : 'justify-end',
        ]">
          <!-- Error Message (left side) -->
          <div v-if="formError" class="flex-1 mr-4">
            <div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-2">
              <p class="text-sm text-red-800 dark:text-red-200">
                {{ formError }}
              </p>
            </div>
          </div>

          <!-- Buttons (right side) -->
          <div class="flex space-x-3">
            <Button type="button" variant="secondary" @click="handleCancel" :disabled="isLoading">
              {{ t('common.cancel') }}
            </Button>
            <Button type="button" variant="primary" @click="handleSubmit" :loading="isLoading" :disabled="!url.trim()">
              {{ t('common.ok') }}
            </Button>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>
