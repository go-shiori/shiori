import { useAuthStore } from '@/stores/auth'
import { getApiConfig } from '@/utils/api-config'

/**
 * Composable for common API store patterns and error handling
 */
export function useApiStore() {
  const authStore = useAuthStore()

  const getAuthToken = () => authStore.token || undefined

  const isAuthenticationError = (err: any): boolean => {
    return err instanceof Error && err.message.includes('401')
  }

  const handleApiError = (err: any, defaultMessage: string): void => {
    console.error('API Error:', err)
    if (isAuthenticationError(err)) {
      throw new Error('Authentication error. Please log in again.')
    }
    throw new Error(defaultMessage)
  }

  const executeWithLoading = async <T>(
    isLoading: { value: boolean },
    error: { value: string | null },
    operation: () => Promise<T>,
    defaultErrorMessage: string = 'Operation failed'
  ): Promise<T> => {
    isLoading.value = true
    error.value = null

    try {
      return await operation()
    } catch (err) {
      handleApiError(err, defaultErrorMessage)
      throw err
    } finally {
      isLoading.value = false
    }
  }

  return {
    getAuthToken,
    isAuthenticationError,
    handleApiError,
    executeWithLoading
  }
}

/**
 * Helper to create API client instance with authentication
 */
export function createApiClient<T extends new (...args: any[]) => any>(
  ApiClass: T,
  token?: string
): InstanceType<T> {
  return new ApiClass(getApiConfig(token))
}
