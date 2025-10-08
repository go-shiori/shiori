import { Configuration } from '@/client/runtime'
import { getApiBaseUrl } from './env-config'

/**
 * Get API configuration for the generated TypeScript client
 * Uses environment-based API base URL for flexible development setup
 *
 * @param token - Optional authentication token
 * @returns Configuration object for API clients
 */
export const getApiConfig = (token?: string | null): Configuration => {
  return new Configuration({
    basePath: getApiBaseUrl(),
    accessToken: token || undefined,
    headers: token ? {
      'Authorization': `Bearer ${token}`,
      'X-Shiori-Response-Format': 'new'
    } : undefined
  })
}
