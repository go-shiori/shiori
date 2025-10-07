<script setup lang="ts">
import AppLayout from '@/components/layout/AppLayout.vue';
import { useI18n } from 'vue-i18n';
import { ref, computed, onMounted, watch } from 'vue';
import { setLanguage } from '@/utils/i18n';
import type { SupportedLocale } from '@/utils/i18n';
import { CheckCircleIcon } from '@/components/icons';
import { useAuthStore } from '@/stores/auth';
import { useTheme } from '@/composables/useTheme';
import type { ModelUserConfig } from '@/client';

const { t, locale } = useI18n();
const auth = useAuthStore();

const languages = [
  { code: 'en' as SupportedLocale, name: 'English' },
  { code: 'es' as SupportedLocale, name: 'Español' },
  { code: 'fr' as SupportedLocale, name: 'Français' },
  { code: 'de' as SupportedLocale, name: 'Deutsch' },
  { code: 'ja' as SupportedLocale, name: '日本語' }
];

const selectedLanguage = ref(locale.value as SupportedLocale);

const changeLanguage = (langCode: SupportedLocale) => {
  selectedLanguage.value = langCode;
  setLanguage(langCode);
};

const currentTheme = computed(() => auth.user?.config?.Theme as 'light' | 'dark' | 'system');
const changeTheme = (theme: 'light' | 'dark' | 'system') => {
  console.log('changeTheme', theme);
};

// ------- Account configuration (replicate old settings) -------
const isOwner = computed(() => !!auth.user?.owner);

// Derive current config or defaults
const defaultConfig: ModelUserConfig = {
  ShowId: false,
  ListMode: false,
  HideThumbnail: false,
  HideExcerpt: false,
  Theme: 'system',
  KeepMetadata: false,
  UseArchive: false,
  CreateEbook: false,
  MakePublic: false
};

const localConfig = ref<ModelUserConfig>({ ...defaultConfig });
const loading = ref(false);
const error = ref<string | null>(null);

const loadUserConfig = () => {
  const cfg = auth.user?.config || {};
  localConfig.value = { ...defaultConfig, ...cfg };
};

onMounted(async () => {
  if (!auth.user) {
    await auth.fetchUserInfo();
  }
  loadUserConfig();
});

// Keep local config in sync if user changes
watch(() => auth.user, () => {
  loadUserConfig();
});

const saveConfig = async () => {
  if (!auth.isAuthenticated) return;
  loading.value = true;
  error.value = null;
  try {
    // Ensure we have all fields by merging with defaults
    const completeConfig = { ...defaultConfig, ...localConfig.value };
    await auth.updateUserConfig(completeConfig);
  } catch (e: any) {
    error.value = e?.message || 'Failed to save settings';
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <AppLayout>
    <template #header>
      <div class="flex justify-between items-center">
        <h1 class="text-xl font-bold">{{ t('settings.title') }}</h1>
      </div>
    </template>

    <div class="space-y-6">
      <!-- Appearance Settings -->
      <div class="bg-white dark:bg-gray-800 p-6 rounded-md shadow-sm">
        <div class="flex items-center mb-4">
          <div class="w-8 h-8 bg-purple-100 dark:bg-purple-900/30 rounded-lg flex items-center justify-center mr-3">
            <svg class="w-5 h-5 text-purple-600 dark:text-purple-400" fill="none" stroke="currentColor"
              viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zM21 5a2 2 0 00-2-2h-4a2 2 0 00-2 2v12a4 4 0 004 4h4a2 2 0 002-2V5z" />
            </svg>
          </div>
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('settings.appearance') }}</h2>
        </div>
        <p class="text-sm text-gray-600 dark:text-gray-400 mb-4">{{ t('settings.appearance_description') }}</p>
        <div class="space-y-6">
          <!-- Theme Selection (cards like language selector) -->
          <div>
            <div class="mb-3">
              <label class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('settings.theme') }}</label>
              <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('settings.theme_description') }}</p>
            </div>
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <div @click="changeTheme('system')"
                class="border rounded-lg p-4 cursor-pointer transition-all duration-200 hover:shadow-md"
                :class="currentTheme === 'system' ? 'border-red-500 bg-red-50 dark:bg-red-900/20 shadow-md' : 'border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700'">
                <div class="flex items-center">
                  <div class="flex-1">
                    <div class="font-medium text-gray-900 dark:text-white">{{ t('settings.system') }}</div>
                  </div>
                  <div v-if="currentTheme === 'system'" class="text-red-500">
                    <CheckCircleIcon class="h-5 w-5" />
                  </div>
                </div>
              </div>
              <div @click="changeTheme('light')"
                class="border rounded-lg p-4 cursor-pointer transition-all duration-200 hover:shadow-md"
                :class="currentTheme === 'light' ? 'border-red-500 bg-red-50 dark:bg-red-900/20 shadow-md' : 'border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700'">
                <div class="flex items-center">
                  <div class="flex-1">
                    <div class="font-medium text-gray-900 dark:text-white">{{ t('settings.light') }}</div>
                  </div>
                  <div v-if="currentTheme === 'light'" class="text-red-500">
                    <CheckCircleIcon class="h-5 w-5" />
                  </div>
                </div>
              </div>
              <div @click="changeTheme('dark')"
                class="border rounded-lg p-4 cursor-pointer transition-all duration-200 hover:shadow-md"
                :class="currentTheme === 'dark' ? 'border-red-500 bg-red-50 dark:bg-red-900/20 shadow-md' : 'border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700'">
                <div class="flex items-center">
                  <div class="flex-1">
                    <div class="font-medium text-gray-900 dark:text-white">{{ t('settings.dark') }}</div>
                  </div>
                  <div v-if="currentTheme === 'dark'" class="text-red-500">
                    <CheckCircleIcon class="h-5 w-5" />
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Display Options -->
          <div class="space-y-4">
            <h3 class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('settings.display_options') }}</h3>
            <div class="space-y-3">
              <label
                class="flex items-center space-x-3 p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">
                <input type="checkbox" v-model="localConfig.HideThumbnail" @change="saveConfig"
                  class="w-4 h-4 text-red-600 bg-gray-100 border-gray-300 rounded focus:ring-red-500 dark:focus:ring-red-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600" />
                <div>
                  <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('settings.hide_thumbnail')
                  }}</span>
                  <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('settings.hide_thumbnail_description') }}</p>
                </div>
              </label>
              <label
                class="flex items-center space-x-3 p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">
                <input type="checkbox" v-model="localConfig.HideExcerpt" @change="saveConfig"
                  class="w-4 h-4 text-red-600 bg-gray-100 border-gray-300 rounded focus:ring-red-500 dark:focus:ring-red-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600" />
                <div>
                  <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('settings.hide_excerpt')
                  }}</span>
                  <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('settings.hide_excerpt_description') }}</p>
                </div>
              </label>
            </div>
          </div>
        </div>
      </div>

      <!-- Bookmarks Settings (owner only) -->
      <div v-if="isOwner" class="bg-white dark:bg-gray-800 p-6 rounded-md shadow-sm">
        <div class="flex items-center mb-4">
          <div class="w-8 h-8 bg-green-100 dark:bg-green-900/30 rounded-lg flex items-center justify-center mr-3">
            <svg class="w-5 h-5 text-green-600 dark:text-green-400" fill="none" stroke="currentColor"
              viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
            </svg>
          </div>
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('settings.bookmarks') }}</h2>
        </div>
        <p class="text-sm text-gray-600 dark:text-gray-400 mb-4">{{ t('settings.bookmarks_description') }}</p>
        <div class="space-y-3">
          <label
            class="flex items-center space-x-3 p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">
            <input type="checkbox" v-model="localConfig.KeepMetadata" @change="saveConfig"
              class="w-4 h-4 text-red-600 bg-gray-100 border-gray-300 rounded focus:ring-red-500 dark:focus:ring-red-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600" />
            <div>
              <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('settings.keep_metadata')
              }}</span>
              <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('settings.keep_metadata_description') }}</p>
            </div>
          </label>
          <label
            class="flex items-center space-x-3 p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">
            <input type="checkbox" v-model="localConfig.UseArchive" @change="saveConfig"
              class="w-4 h-4 text-red-600 bg-gray-100 border-gray-300 rounded focus:ring-red-500 dark:focus:ring-red-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600" />
            <div>
              <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('settings.use_archive') }}</span>
              <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('settings.use_archive_description') }}</p>
            </div>
          </label>
          <label
            class="flex items-center space-x-3 p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">
            <input type="checkbox" v-model="localConfig.CreateEbook" @change="saveConfig"
              class="w-4 h-4 text-red-600 bg-gray-100 border-gray-300 rounded focus:ring-red-500 dark:focus:ring-red-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600" />
            <div>
              <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('settings.create_ebook') }}</span>
              <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('settings.create_ebook_description') }}</p>
            </div>
          </label>
          <label
            class="flex items-center space-x-3 p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">
            <input type="checkbox" v-model="localConfig.MakePublic" @change="saveConfig"
              class="w-4 h-4 text-red-600 bg-gray-100 border-gray-300 rounded focus:ring-red-500 dark:focus:ring-red-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600" />
            <div>
              <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('settings.make_public') }}</span>
              <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('settings.make_public_description') }}</p>
            </div>
          </label>
        </div>
      </div>

      <!-- Language Settings -->
      <div class="bg-white dark:bg-gray-800 p-6 rounded-md shadow-sm">
        <div class="flex items-center mb-4">
          <div class="w-8 h-8 bg-blue-100 dark:bg-blue-900/30 rounded-lg flex items-center justify-center mr-3">
            <svg class="w-5 h-5 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                d="M3 5h12M9 3v2m1.048 9.5A18.022 18.022 0 016.412 9m6.088 9h7M11 21l5-10 5 10M12.751 5C11.783 10.77 8.07 15.61 3 18.129" />
            </svg>
          </div>
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('settings.language') }}</h2>
        </div>
        <p class="text-sm text-gray-600 dark:text-gray-400 mb-4">{{ t('settings.language_description') }}</p>
        <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
          <div v-for="language in languages" :key="language.code" @click="changeLanguage(language.code)"
            class="border rounded-lg p-4 cursor-pointer transition-all duration-200 hover:shadow-md" :class="selectedLanguage === language.code ?
              'border-red-500 bg-red-50 dark:bg-red-900/20 shadow-md' :
              'border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700'">
            <div class="flex items-center">
              <div class="flex-1">
                <div class="font-medium text-gray-900 dark:text-white">{{ language.name }}</div>
              </div>
              <div v-if="selectedLanguage === language.code" class="text-red-500">
                <CheckCircleIcon class="h-5 w-5" />
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Status Messages -->
      <div v-if="loading"
        class="bg-blue-50 dark:bg-blue-900/30 border border-blue-200 dark:border-blue-800 text-blue-700 dark:text-blue-300 px-4 py-3 rounded-md">
        <div class="flex items-center">
          <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-blue-500" xmlns="http://www.w3.org/2000/svg" fill="none"
            viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z">
            </path>
          </svg>
          {{ t('common.loading') }}
        </div>
      </div>
      <div v-if="error"
        class="bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-300 px-4 py-3 rounded-md">
        <div class="flex items-center">
          <svg class="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd"
              d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
              clip-rule="evenodd" />
          </svg>
          {{ error }}
        </div>
      </div>
    </div>
  </AppLayout>
</template>
