import { defineStore } from 'pinia'
import { ref } from 'vue'
import { BookmarksApi } from '@/client'
import type { ModelBookmarkDTO } from '@/client'
import type { ModelPaginatedResponseModelBookmarkDTO } from '@/client/models/ModelPaginatedResponseModelBookmarkDTO'
import { usePagination } from '@/composables/usePagination'
import { useApiStore, createApiClient } from '@/composables/useApiStore'

export interface BookmarksFilters {
  keyword?: string
  tags?: string
  exclude?: string
  page?: number
  limit?: number
}

export const useBookmarksStore = defineStore('bookmarks', () => {
  const bookmarks = ref<ModelBookmarkDTO[]>([])

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
  const getBookmarksApi = () => {
    const { getAuthToken } = useApiStore()
    return createApiClient(BookmarksApi, getAuthToken())
  }

  // Get bookmarks with filters
  const fetchBookmarks = async (filters: BookmarksFilters = {}) => {
    return executeWithLoading(
      isLoading,
      error,
      async () => {
        const api = getBookmarksApi()
        const page = filters.page || currentPage.value
        const limit = filters.limit || pageLimit.value

        const response = await api.apiV1BookmarksGet({
          keyword: filters.keyword,
          tags: filters.tags,
          exclude: filters.exclude,
          page: page,
          limit: limit
        })

        // Update pagination state
        updatePagination({ page, limit, total: (response as any).total || 0 })

        // Response is now a paginated response with items and total
        if (response && (response as any).items) {
          bookmarks.value = (response as any).items
        } else {
          console.error('Unexpected response format:', response)
          bookmarks.value = []
        }

        return bookmarks.value
      },
      'Failed to load bookmarks. Please try again.'
    )
  }

  // Get single bookmark by ID
  const getBookmark = async (id: number) => {
    return executeWithLoading(
      isLoading,
      error,
      async () => {
        const api = getBookmarksApi()
        return await api.apiV1BookmarksIdGet({ id })
      },
      'Failed to load bookmark. Please try again.'
    )
  }

  // Create a new bookmark
  const createBookmark = async (url: string, title?: string, excerpt?: string, isPublic?: number) => {
    return executeWithLoading(
      isLoading,
      error,
      async () => {
        const api = getBookmarksApi()
        const newBookmark = await api.apiV1BookmarksPost({
          payload: {
            url,
            title,
            excerpt,
            _public: isPublic
          }
        })

        bookmarks.value.unshift(newBookmark)
        totalCount.value++
        return newBookmark
      },
      'Failed to create bookmark. Please try again.'
    )
  }

  // Update a bookmark
  const updateBookmark = async (
    id: number,
    updates: { url?: string; title?: string; excerpt?: string; public?: number }
  ) => {
    return executeWithLoading(
      isLoading,
      error,
      async () => {
        const api = getBookmarksApi()
        const updatedBookmark = await api.apiV1BookmarksIdPut({
          id,
          payload: updates
        })

        const index = bookmarks.value.findIndex(bookmark => bookmark.id === id)
        if (index !== -1) {
          bookmarks.value[index] = updatedBookmark
        }

        return updatedBookmark
      },
      'Failed to update bookmark. Please try again.'
    )
  }

  // Delete bookmarks
  const deleteBookmarks = async (ids: number[]) => {
    return executeWithLoading(
      isLoading,
      error,
      async () => {
        const api = getBookmarksApi()
        await api.apiV1BookmarksDelete({ payload: { ids } })

        bookmarks.value = bookmarks.value.filter(bookmark => !ids.includes(bookmark.id || 0))
        totalCount.value -= ids.length
      },
      'Failed to delete bookmarks. Please try again.'
    )
  }

  // Get bookmark tags
  const getBookmarkTags = async (id: number) => {
    return executeWithLoading(
      isLoading,
      error,
      async () => {
        const api = getBookmarksApi()
        const tags = await api.apiV1BookmarksIdTagsGet({ id })
        return tags
      },
      'Failed to get bookmark tags. Please try again.'
    )
  }

  // Add tag to bookmark
  const addTagToBookmark = async (bookmarkId: number, tagId: number) => {
    return executeWithLoading(
      isLoading,
      error,
      async () => {
        const api = getBookmarksApi()
        await api.apiV1BookmarksIdTagsPost({
          id: bookmarkId,
          payload: { tagId }
        })
      },
      'Failed to add tag to bookmark. Please try again.'
    )
  }

  // Remove tag from bookmark
  const removeTagFromBookmark = async (bookmarkId: number, tagId: number) => {
    return executeWithLoading(
      isLoading,
      error,
      async () => {
        const api = getBookmarksApi()
        await api.apiV1BookmarksIdTagsDelete({
          id: bookmarkId,
          payload: { tagId }
        })
      },
      'Failed to remove tag to bookmark. Please try again.'
    )
  }

  // Get bookmark data (content, archive, ebook info)
  const getBookmarkData = async (id: number) => {
    return executeWithLoading(
      isLoading,
      error,
      async () => {
        const api = getBookmarksApi()
        const data = await api.apiV1BookmarksIdDataGet({ id })
        return data
      },
      'Failed to get bookmark data. Please try again.'
    )
  }

  // Update bookmark data (generate/update readable content, archive, ebook)
  const updateBookmarkData = async (
    id: number,
    options: {
      updateReadable?: boolean
      createArchive?: boolean
      createEbook?: boolean
      keepMetadata?: boolean
      skipExisting?: boolean
    }
  ) => {
    return executeWithLoading(
      isLoading,
      error,
      async () => {
        const api = getBookmarksApi()
        const data = await api.apiV1BookmarksIdDataPut({
          id,
          payload: {
            updateReadable: options.updateReadable || false,
            createArchive: options.createArchive || false,
            createEbook: options.createEbook || false,
            keepMetadata: options.keepMetadata || false,
            skipExisting: options.skipExisting || false
          }
        })
        return data
      },
      'Failed to update bookmark data. Please try again.'
    )
  }

  return {
    bookmarks,
    isLoading,
    error,
    totalCount,
    currentPage,
    pageLimit,
    fetchBookmarks,
    getBookmark,
    createBookmark,
    updateBookmark,
    deleteBookmarks,
    getBookmarkTags,
    addTagToBookmark,
    removeTagFromBookmark,
    getBookmarkData,
    updateBookmarkData
  }
})
