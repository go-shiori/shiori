import { defineStore } from 'pinia'
import { ref } from 'vue'
import { Configuration, TagsApi } from '@/client'
import type { ModelTagDTO } from '@/client/models'
import { useAuthStore } from './auth'

export const useTagsStore = defineStore('tags', () => {
  const tags = ref<ModelTagDTO[]>([])
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // API client
  const getTagsApi = () => {
    const authStore = useAuthStore()
    const token = authStore.token

    const config = new Configuration({
      basePath: 'http://localhost:8080',
      accessToken: token || undefined,
      headers: token ? {
        'Authorization': `Bearer ${token}`,
        'X-Shiori-Response-Format': 'new'
      } : undefined
    })
    return new TagsApi(config)
  }

  // Get all tags
  const fetchTags = async (withBookmarkCount = true) => {
    isLoading.value = true
    error.value = null

    try {
      const api = getTagsApi()
      const response = await api.apiV1TagsGet({ withBookmarkCount })

      // Ensure response is an array before assigning
      if (Array.isArray(response)) {
        tags.value = response
      } else {
        console.error('Expected array response but got:', typeof response)
        tags.value = []
      }

      return tags.value
    } catch (err) {
      console.error('Failed to fetch tags:', err)
      if (err instanceof Error && err.message.includes('401')) {
        error.value = 'Authentication error. Please log in again.'
      } else {
        error.value = 'Failed to load tags. Please try again.'
      }
      throw err
    } finally {
      isLoading.value = false
    }

  }

  // Create a new tag
  const createTag = async (name: string) => {
    isLoading.value = true
    error.value = null

    try {
      const api = getTagsApi()
      const newTag = await api.apiV1TagsPost({ tag: { name } })
      tags.value.push(newTag)
      return newTag
    } catch (err) {
      console.error('Failed to create tag:', err)
      if (err instanceof Error && err.message.includes('401')) {
        error.value = 'Authentication error. Please log in again.'
      } else {
        error.value = 'Failed to create tag. Please try again.'
      }
      throw err
    } finally {
      isLoading.value = false
    }
  }

  // Update a tag
  const updateTag = async (id: number, name: string) => {
    isLoading.value = true
    error.value = null

    try {
      const api = getTagsApi()
      const updatedTag = await api.apiV1TagsIdPut({ id, tag: { id, name } })

      const index = tags.value.findIndex(tag => tag.id === id)
      if (index !== -1) {
        tags.value[index] = updatedTag
      }

      return updatedTag
    } catch (err) {
      console.error('Failed to update tag:', err)
      if (err instanceof Error && err.message.includes('401')) {
        error.value = 'Authentication error. Please log in again.'
      } else {
        error.value = 'Failed to update tag. Please try again.'
      }
      throw err
    } finally {
      isLoading.value = false
    }
  }

  // Delete a tag
  const deleteTag = async (id: number) => {
    isLoading.value = true
    error.value = null

    try {
      const api = getTagsApi()
      await api.apiV1TagsIdDelete({ id })
      tags.value = tags.value.filter(tag => tag.id !== id)
    } catch (err) {
      console.error('Failed to delete tag:', err)
      if (err instanceof Error && err.message.includes('401')) {
        error.value = 'Authentication error. Please log in again.'
      } else {
        error.value = 'Failed to delete tag. Please try again.'
      }
      throw err
    } finally {
      isLoading.value = false
    }
  }

  return {
    tags,
    isLoading,
    error,
    fetchTags,
    createTag,
    updateTag,
    deleteTag
  }
})
