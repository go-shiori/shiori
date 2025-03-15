import { createI18n } from 'vue-i18n'
import en from '@/locales/en.json'
import es from '@/locales/es.json'
import fr from '@/locales/fr.json'
import de from '@/locales/de.json'
import ja from '@/locales/ja.json'

// Define supported languages
export type SupportedLocale = 'en' | 'es' | 'fr' | 'de' | 'ja';

// Get the browser language or use English as fallback
const getBrowserLanguage = (): SupportedLocale => {
  const browserLang = navigator.language.split('-')[0]
  return ['en', 'es', 'fr', 'de', 'ja'].includes(browserLang) ? browserLang as SupportedLocale : 'en'
}

// Get the stored language preference or use browser language
const getStoredLanguage = (): SupportedLocale => {
  const storedLang = localStorage.getItem('shiori-language')
  return (storedLang && ['en', 'es', 'fr', 'de', 'ja'].includes(storedLang))
    ? storedLang as SupportedLocale
    : getBrowserLanguage()
}

// Create the i18n instance
const i18n = createI18n({
  legacy: false, // Use Composition API
  locale: getStoredLanguage(),
  fallbackLocale: 'en',
  messages: {
    en,
    es,
    fr,
    de,
    ja
  }
})

// Function to change the language
export const setLanguage = (lang: SupportedLocale): void => {
  i18n.global.locale.value = lang
  localStorage.setItem('shiori-language', lang)
  document.querySelector('html')?.setAttribute('lang', lang)
}

// Initialize HTML lang attribute
document.querySelector('html')?.setAttribute('lang', getStoredLanguage())

export default i18n
