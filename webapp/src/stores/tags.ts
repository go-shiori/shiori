import { defineStore } from 'pinia'
import { ref } from 'vue'
import { TagsApi } from '@/client'
import type { ModelTagDTO } from '@/client'
import type { ModelPaginatedResponseModelTagDTO } from '@/client/models/ModelPaginatedResponseModelTagDTO'
import { usePagination } from '@/composables/usePagination'
import { useApiStore, createApiClient } from '@/composables/useApiStore'

export const useTagsStore = defineStore('tags', () => {
  const tags = ref<ModelTagDTO[]>([])

  // Use standardized pagination composable
  const {
    currentPage,
    pageLimit,
    totalCount,
    isLoading,
    error,
    updatePagination
  } = usePagination(20)

  // Use API store composable
  const { executeWithLoading } = useApiStore()

  // API client
  const getTagsApi = () => {
    const { getAuthToken } = useApiStore()
    return createApiClient(TagsApi, getAuthToken())
  }

  // Get all tags
  const fetchTags = async (options?: {
    withBookmarkCount?: boolean;
    page?: number;
    limit?: number;
    search?: string;
  }) => {
    return executeWithLoading(
      isLoading,
      error,
      async () => {
        const api = getTagsApi()
        const page = options?.page || currentPage.value
        const limit = options?.limit || pageLimit.value

        const response = await api.apiV1TagsGet({
          withBookmarkCount: options?.withBookmarkCount ?? true,
          page,
          limit,
          search: options?.search
        })

        // Update pagination state
        updatePagination({ page, limit, total: (response as any).total || 0 })

        // Response is now a paginated response with items and total
        if (response && (response as any).items) {
          tags.value = (response as any).items
        } else {
          console.error('Unexpected response format:', response)
          tags.value = []
        }

        return tags.value
      },
      'Failed to load tags. Please try again.'
    )
  }

  // Search tags with debouncing
  const searchTags = async (searchTerm: string) => {
    if (!searchTerm.trim()) {
      // If no search term, fetch all tags
      return fetchTags({ limit: 1000 }) // Fetch more tags for better UX
    }

    return executeWithLoading(
      isLoading,
      error,
      async () => {
        const api = getTagsApi()
        const response = await api.apiV1TagsGet({
          withBookmarkCount: true,
          search: searchTerm,
          limit: 100 // Limit search results
        })

        if (response && (response as any).items) {
          return (response as any).items
        } else {
          console.error('Unexpected response format:', response)
          return []
        }
      },
      'Failed to search tags. Please try again.'
    )
  }

  // Create a new tag
  const createTag = async (name: string) => {
    isLoading.value = true
    error.value = null

    try {
      const api = getTagsApi()
      const newTag = await api.apiV1TagsPost({ tag: { name } })
      if (newTag) {
        tags.value.push(newTag)
        totalCount.value++
      }
      return newTag
    } catch (err) {
      // Normalize fetch client errors to include response.data.error
      const maybeResponse = (err as any)?.response
      if (maybeResponse && typeof maybeResponse.json === 'function') {
        try {
          const data = await maybeResponse.json()
          // Re-throw with a standardized shape expected by the error handler
          throw { response: { status: maybeResponse.status, data } }
        } catch (_) {
          // If parsing fails, fall through to rethrow original error
        }
      }
      throw err
    } finally {
      isLoading.value = false
    }
  }

  // Update a tag
  const updateTag = async (id: number, name: string) => {
    return executeWithLoading(
      isLoading,
      error,
      async () => {
        const api = getTagsApi()
        const updatedTag = await api.apiV1TagsIdPut({ id, tag: { id, name } })

        const index = tags.value.findIndex(tag => tag.id === id)
        if (index !== -1) {
          tags.value[index] = updatedTag
        }

        return updatedTag
      },
      'Failed to update tag. Please try again.'
    )
  }

  // Delete a tag
  const deleteTag = async (id: number) => {
    return executeWithLoading(
      isLoading,
      error,
      async () => {
        const api = getTagsApi()
        await api.apiV1TagsIdDelete({ id })

        tags.value = tags.value.filter(tag => tag.id !== id)
        totalCount.value--

        return true
      },
      'Failed to delete tag. Please try again.'
    )
  }

  return {
    tags,
    isLoading,
    error,
    totalCount,
    currentPage,
    pageLimit,
    fetchTags,
    searchTags,
    createTag,
    updateTag,
    deleteTag
  }
})
