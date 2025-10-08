import { useI18n } from 'vue-i18n'

export interface ApiError {
  response?: {
    status: number
    data?: {
      error?: string
      message?: string
    }
  }
  message?: string
}

/**
 * Handles API errors and returns a localized error message
 * @param error The error object from the API
 * @param t The i18n translation function
 * @returns A localized error message
 */
export function handleApiError(error: ApiError, t: (key: string) => string, resourceHint?: 'bookmark' | 'tag'): string {
  // Check if it's an API error with a response
  if (error.response?.data?.error) {
    const errorKey = error.response.data.error

    // Handle namespaced error keys (e.g., "already_exists.bookmark")
    if (errorKey.includes('.')) {
      const parts = errorKey.split('.')
      if (parts.length >= 2) {
        const [errorType, resource] = parts

        // Map error types to appropriate sections
        const sectionMap: Record<string, string> = {
          'bookmark': 'bookmarks',
          'tag': 'tags'
        }

        const section = resource ? sectionMap[resource] : undefined
        if (section) {
          const translationKey = `${section}.errors.${errorType}.${resource}`
          const translatedMessage = t(translationKey)

          // If translation exists and is different from the key, use it
          if (translatedMessage !== translationKey) {
            return translatedMessage
          }
        }
      }
    }

    // Handle legacy error keys for backward compatibility
    switch (errorKey) {
      case 'BOOKMARK_ALREADY_EXISTS':
        return t('bookmarks.errors.already_exists.bookmark')
      case 'TAG_ALREADY_EXISTS':
        return t('tags.errors.already_exists.tag')
      default:
        // Return the error as-is if no specific translation is found
        return errorKey
    }
  }

  // Fallback to generic error message
  if (error.response?.status && resourceHint) {
    // Infer from status + resource when backend sent a JSON error we couldn't parse upstream
    const sectionMap: Record<string, string> = { bookmark: 'bookmarks', tag: 'tags' }
    const section = sectionMap[resourceHint]
    if (error.response.status === 409 && section) {
      const translationKey = `${section}.errors.already_exists.${resourceHint}`
      const translatedMessage = t(translationKey)
      if (translatedMessage !== translationKey) return translatedMessage
    }
  }

  if (error.message) return error.message

  return t('common.error_occurred') || 'An error occurred'
}

/**
 * Composable for handling API errors with i18n
 */
export function useErrorHandler() {
  const { t } = useI18n()

  return {
    handleApiError: (error: ApiError, resourceHint?: 'bookmark' | 'tag') => handleApiError(error, t, resourceHint)
  }
}
