import { Configuration } from '@/client/runtime'

/**
 * Get API configuration for the generated TypeScript client
 * Uses the current window origin as basePath for flexibility
 *
 * @param token - Optional authentication token
 * @returns Configuration object for API clients
 */
export const getApiConfig = (token?: string | null): Configuration => {
  return new Configuration({
    basePath: window.location.origin,
    accessToken: token || undefined,
    headers: token ? {
      'Authorization': `Bearer ${token}`,
      'X-Shiori-Response-Format': 'new'
    } : undefined
  })
}
