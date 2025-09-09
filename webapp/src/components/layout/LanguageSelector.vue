<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { setLanguage } from '@/utils/i18n'
import type { SupportedLocale } from '@/utils/i18n'

const { t, locale } = useI18n()

const languages = [
    { code: 'en' as SupportedLocale, name: 'English' },
    { code: 'es' as SupportedLocale, name: 'Español' },
    { code: 'fr' as SupportedLocale, name: 'Français' },
    { code: 'de' as SupportedLocale, name: 'Deutsch' },
    { code: 'ja' as SupportedLocale, name: '日本語' }
]

const isOpen = ref(false)
const selectedLanguage = ref(locale.value as SupportedLocale)

const toggleDropdown = () => {
    isOpen.value = !isOpen.value
}

const closeDropdown = () => {
    isOpen.value = false
}

const changeLanguage = (langCode: SupportedLocale) => {
    selectedLanguage.value = langCode
    setLanguage(langCode)
    closeDropdown()
}

// Close dropdown when clicking outside
onMounted(() => {
    document.addEventListener('click', (event) => {
        const target = event.target as HTMLElement
        if (!target.closest('.language-selector')) {
            closeDropdown()
        }
    })
})
</script>

<template>
    <div class="language-selector relative">
        <button @click.stop="toggleDropdown"
            class="flex items-center px-3 py-2 text-sm rounded-md hover:bg-gray-100 dark:hover:bg-gray-700"
            aria-haspopup="true" :aria-expanded="isOpen">
            <span class="mr-1">{{languages.find(lang => lang.code === selectedLanguage)?.name}}</span>
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"
                xmlns="http://www.w3.org/2000/svg">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
            </svg>
        </button>

        <div v-if="isOpen" class="absolute right-0 mt-2 w-48 bg-white dark:bg-gray-800 rounded-md shadow-lg z-10 py-1">
            <button v-for="language in languages" :key="language.code" @click="changeLanguage(language.code)"
                class="block w-full text-left px-4 py-2 text-sm hover:bg-gray-100 dark:hover:bg-gray-700"
                :class="{ 'bg-gray-100 dark:bg-gray-700': selectedLanguage === language.code }">
                {{ language.name }}
            </button>
        </div>
    </div>
</template>
