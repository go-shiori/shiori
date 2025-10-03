import { defineStore } from 'pinia'
import { ref } from 'vue'
import { BookmarksApi } from '@/client'
import type { ModelBookmarkDTO } from '@/client/models'
import { useAuthStore } from './auth'
import { getApiConfig } from '@/utils/api-config'

export interface BookmarksFilters {
  keyword?: string
  tags?: string
  exclude?: string
  page?: number
  limit?: number
}

export const useBookmarksStore = defineStore('bookmarks', () => {
  const bookmarks = ref<ModelBookmarkDTO[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const totalCount = ref(0)
  const currentPage = ref(1)
  const pageLimit = ref(30)

  // API client
  const getBookmarksApi = () => {
    const authStore = useAuthStore()
    return new BookmarksApi(getApiConfig(authStore.token))
  }

  // Get bookmarks with filters
  const fetchBookmarks = async (filters: BookmarksFilters = {}) => {
    isLoading.value = true
    error.value = null

    try {
      const api = getBookmarksApi()
      const response = await api.apiV1BookmarksGet({
        keyword: filters.keyword,
        tags: filters.tags,
        exclude: filters.exclude,
        page: filters.page || currentPage.value,
        limit: filters.limit || pageLimit.value
      })

      // Ensure response is an array before assigning
      if (Array.isArray(response)) {
        bookmarks.value = response
        totalCount.value = response.length
      } else {
        console.error('Expected array response but got:', typeof response)
        bookmarks.value = []
        totalCount.value = 0
      }

      return bookmarks.value
    } catch (err) {
      console.error('Failed to fetch bookmarks:', err)
      if (err instanceof Error && err.message.includes('401')) {
        error.value = 'Authentication error. Please log in again.'
      } else {
        error.value = 'Failed to load bookmarks. Please try again.'
      }
      throw err
    } finally {
      isLoading.value = false
    }
  }

  // Get single bookmark by ID
  const getBookmark = async (id: number) => {
    isLoading.value = true
    error.value = null

    try {
      const api = getBookmarksApi()
      const bookmark = await api.apiV1BookmarksIdGet({ id })
      return bookmark
    } catch (err) {
      console.error('Failed to get bookmark:', err)
      if (err instanceof Error && err.message.includes('401')) {
        error.value = 'Authentication error. Please log in again.'
      } else {
        error.value = 'Failed to load bookmark. Please try again.'
      }
      throw err
    } finally {
      isLoading.value = false
    }
  }

  // Create a new bookmark
  const createBookmark = async (url: string, title?: string, excerpt?: string, isPublic?: number) => {
    isLoading.value = true
    error.value = null

    try {
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
    } catch (err) {
      console.error('Failed to create bookmark:', err)
      if (err instanceof Error && err.message.includes('401')) {
        error.value = 'Authentication error. Please log in again.'
      } else {
        error.value = 'Failed to create bookmark. Please try again.'
      }
      throw err
    } finally {
      isLoading.value = false
    }
  }

  // Update a bookmark
  const updateBookmark = async (
    id: number,
    updates: { url?: string; title?: string; excerpt?: string; public?: number }
  ) => {
    isLoading.value = true
    error.value = null

    try {
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
    } catch (err) {
      console.error('Failed to update bookmark:', err)
      if (err instanceof Error && err.message.includes('401')) {
        error.value = 'Authentication error. Please log in again.'
      } else {
        error.value = 'Failed to update bookmark. Please try again.'
      }
      throw err
    } finally {
      isLoading.value = false
    }
  }

  // Delete bookmarks
  const deleteBookmarks = async (ids: number[]) => {
    isLoading.value = true
    error.value = null

    try {
      const api = getBookmarksApi()
      await api.apiV1BookmarksDelete({ payload: { ids } })

      bookmarks.value = bookmarks.value.filter(bookmark => !ids.includes(bookmark.id || 0))
      totalCount.value -= ids.length
    } catch (err) {
      console.error('Failed to delete bookmarks:', err)
      if (err instanceof Error && err.message.includes('401')) {
        error.value = 'Authentication error. Please log in again.'
      } else {
        error.value = 'Failed to delete bookmarks. Please try again.'
      }
      throw err
    } finally {
      isLoading.value = false
    }
  }

  // Get bookmark tags
  const getBookmarkTags = async (id: number) => {
    try {
      const api = getBookmarksApi()
      const tags = await api.apiV1BookmarksIdTagsGet({ id })
      return tags
    } catch (err) {
      console.error('Failed to get bookmark tags:', err)
      throw err
    }
  }

  // Add tag to bookmark
  const addTagToBookmark = async (bookmarkId: number, tagId: number) => {
    try {
      const api = getBookmarksApi()
      await api.apiV1BookmarksIdTagsPost({
        id: bookmarkId,
        payload: { tagId }
      })
    } catch (err) {
      console.error('Failed to add tag to bookmark:', err)
      throw err
    }
  }

  // Remove tag from bookmark
  const removeTagFromBookmark = async (bookmarkId: number, tagId: number) => {
    try {
      const api = getBookmarksApi()
      await api.apiV1BookmarksIdTagsDelete({
        id: bookmarkId,
        payload: { tagId }
      })
    } catch (err) {
      console.error('Failed to remove tag from bookmark:', err)
      throw err
    }
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
    removeTagFromBookmark
  }
})
