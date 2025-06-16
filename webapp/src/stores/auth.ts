import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { AuthApi } from '@/client/apis/AuthApi'
import type { ApiV1LoginRequestPayload } from '@/client/models/ApiV1LoginRequestPayload'
import { Configuration } from '@/client/runtime'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const expires = ref<number | null>(Number(localStorage.getItem('expires')) || null)
  const user = ref<any | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const redirectDestination = ref<string | null>(null)

  const isAuthenticated = computed(() => {
    if (!token.value) return false
    if (!expires.value) return false
    return expires.value > Date.now()
  })

  // Create API client with auth token
  const getApiClient = () => {
    const config = new Configuration({
      basePath: 'http://localhost:8080',
      accessToken: token.value || undefined,
      headers: token.value ? {
        'Authorization': `Bearer ${token.value}`,
        'X-Shiori-Response-Format': 'new'
      } : undefined
    })
    return new AuthApi(config)
  }

  // Validate token by fetching user info
  const validateToken = async (): Promise<boolean> => {
    if (!token.value) return false

    loading.value = true
    try {
      const result = await fetchUserInfo()
      loading.value = false
      return !!result
    } catch (err) {
      loading.value = false
      return false
    }
  }

  // Login function
  const login = async (username: string, password: string, rememberMe: boolean = false) => {
    loading.value = true
    error.value = null

    try {
      const payload: ApiV1LoginRequestPayload = {
        username,
        password,
        rememberMe,
      }

      const api = getApiClient()
      const response = await api.apiV1AuthLoginPost({ payload })

      if (response.token) {
        token.value = response.token
        expires.value = response.expires || 0

        // Store in localStorage
        localStorage.setItem('token', response.token)
        localStorage.setItem('expires', String(response.expires))

        // Get user info
        await fetchUserInfo()
        return true
      } else {
        throw new Error('Invalid response from server')
      }
    } catch (err: any) {
      console.error('Login error:', err)

      // Extract error message from response if available
      if (err.response) {
        try {
          // Try to parse the response body as JSON
          const responseBody = await err.response.json()
          if (responseBody && responseBody.message) {
            error.value = responseBody.message
          } else if (responseBody && responseBody.error) {
            error.value = responseBody.error
          } else if (typeof responseBody === 'string') {
            error.value = responseBody
          } else {
            error.value = `Server error: ${err.response.status}`
          }
        } catch (jsonError) {
          // If response is not JSON, use status text
          error.value = err.response.statusText || `Server error: ${err.response.status}`
        }
      } else {
        // If no response object, use the error message
        error.value = err.message || 'Failed to login'
      }

      return false
    } finally {
      loading.value = false
    }
  }

  // Fetch user info
  const fetchUserInfo = async () => {
    if (!token.value) return null

    try {
      // Create a new API client with the current token
      const api = getApiClient()

      // Make the API request with the token in the headers
      const response = await api.apiV1AuthMeGet()

      if (response) {
        user.value = response
        return user.value
      } else {
        throw new Error('Failed to fetch user info')
      }
    } catch (err: any) {
      console.error('Error fetching user info:', err)

      // If we get a 401 Unauthorized, the token is invalid
      if (err.response && err.response.status === 401) {
        // Clear the invalid token
        clearAuth()
      }

      return null
    }
  }

  // Clear authentication data
  const clearAuth = () => {
    token.value = null
    expires.value = null
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('expires')
  }

  // Logout function
  const logout = async () => {
    loading.value = true

    try {
      if (token.value) {
        const api = getApiClient()
        await api.apiV1AuthLogoutPost()
      }
    } catch (err) {
      console.error('Logout error:', err)
    } finally {
      // Clear state regardless of API success
      clearAuth()
      loading.value = false
    }
  }

  // Refresh token
  const refreshToken = async () => {
    if (!token.value) return false

    try {
      const api = getApiClient()
      const response = await api.apiV1AuthRefreshPost()

      if (response.token) {
        token.value = response.token
        expires.value = response.expires || 0

        localStorage.setItem('token', response.token)
        localStorage.setItem('expires', String(response.expires))
        return true
      }
      return false
    } catch (err) {
      console.error('Token refresh error:', err)
      return false
    }
  }

  // Set redirect destination
  const setRedirectDestination = (destination: string | null) => {
    redirectDestination.value = destination
  }

  // Get and clear redirect destination
  const getAndClearRedirectDestination = () => {
    const destination = redirectDestination.value
    redirectDestination.value = null
    return destination
  }

  return {
    token,
    expires,
    user,
    loading,
    error,
    isAuthenticated,
    login,
    logout,
    fetchUserInfo,
    refreshToken,
    validateToken,
    setRedirectDestination,
    getAndClearRedirectDestination,
    clearAuth
  }
})
