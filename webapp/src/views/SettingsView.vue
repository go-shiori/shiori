<script setup lang="ts">
import AppLayout from '@/components/layout/AppLayout.vue';
import { useI18n } from 'vue-i18n';
import { ref } from 'vue';
import { setLanguage } from '@/utils/i18n';
import type { SupportedLocale } from '@/utils/i18n';

const { t, locale } = useI18n();

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
</script>

<template>
  <AppLayout>
    <template #header>
      <div class="flex justify-between items-center">
        <h1 class="text-xl font-bold">{{ t('settings.title') }}</h1>
      </div>
    </template>

    <div class="bg-white dark:bg-gray-800 p-6 rounded-md shadow-sm">
      <div class="mb-6">
        <h2 class="text-lg font-semibold mb-4">{{ t('settings.language') }}</h2>
        <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
          <div v-for="language in languages" :key="language.code" @click="changeLanguage(language.code)"
            class="border rounded-md p-4 cursor-pointer transition-colors" :class="selectedLanguage === language.code ?
              'border-red-500 bg-red-50 dark:bg-red-900/20' :
              'border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700'">
            <div class="flex items-center">
              <div class="flex-1">
                <div class="font-medium">{{ language.name }}</div>
              </div>
              <div v-if="selectedLanguage === language.code" class="text-red-500">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                  <path fill-rule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                    clip-rule="evenodd" />
                </svg>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>
