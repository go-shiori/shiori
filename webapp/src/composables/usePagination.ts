import { ref, computed } from 'vue'

export interface PaginationOptions {
  page?: number
  limit?: number
  total?: number
}

export interface PaginationState {
  currentPage: ReturnType<typeof ref<number>>
  pageLimit: ReturnType<typeof ref<number>>
  totalCount: ReturnType<typeof ref<number>>
  isLoading: ReturnType<typeof ref<boolean>>
  error: ReturnType<typeof ref<string | null>>
}

/**
 * Composable for managing pagination state and operations
 */
export function usePagination(initialLimit = 20) {
  const currentPage = ref(1)
  const pageLimit = ref(initialLimit)
  const totalCount = ref(0)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // Computed values
  const totalPages = computed(() => Math.ceil(totalCount.value / pageLimit.value))
  const hasNextPage = computed(() => currentPage.value < totalPages.value)
  const hasPreviousPage = computed(() => currentPage.value > 1)

  // Helper functions
  const setLoading = (loading: boolean) => {
    isLoading.value = loading
  }

  const setError = (err: string | null) => {
    error.value = err
  }

  const resetError = () => {
    error.value = null
  }

  const updatePagination = (options: PaginationOptions) => {
    if (options.page !== undefined) {
      currentPage.value = options.page
    }
    if (options.limit !== undefined) {
      pageLimit.value = options.limit
    }
    if (options.total !== undefined) {
      totalCount.value = options.total
    }
  }

  const resetToPageOne = () => {
    currentPage.value = 1
  }

  const goToNextPage = () => {
    if (hasNextPage.value) {
      currentPage.value++
    }
  }

  const goToPreviousPage = () => {
    if (hasPreviousPage.value) {
      currentPage.value--
    }
  }

  const goToPage = (page: number) => {
    if (page >= 1 && page <= totalPages.value) {
      currentPage.value = page
    }
  }

  return {
    // State
    currentPage,
    pageLimit,
    totalCount,
    isLoading,
    error,

    // Computed
    totalPages,
    hasNextPage,
    hasPreviousPage,

    // Actions
    setLoading,
    setError,
    resetError,
    updatePagination,
    resetToPageOne,
    goToNextPage,
    goToPreviousPage,
    goToPage
  }
}
