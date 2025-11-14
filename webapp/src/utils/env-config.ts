/**
 * Environment configuration utilities for the Shiori webapp
 * Reads environment variables and provides configuration values
 */

/**
 * Get the API base URL from environment variables
 * Falls back to current window.location.origin if not specified
 *
 * @returns The API base URL to use for backend requests
 */
export const getApiBaseUrl = (): string => {
  // Check for VITE_API_BASE_URL environment variable
  const envBaseUrl = import.meta.env?.VITE_API_BASE_URL;

  if (envBaseUrl && typeof envBaseUrl === 'string') {
    return envBaseUrl;
  }

  // Fallback to current origin (works for production and proxied development)
  // Check if we're in browser context to safely access window
  if (typeof window !== 'undefined' && window.location) {
    return window.location.origin;
  }

  // Fallback for SSR or Node.js contexts
  return '';
};

/**
 * Check if we're running in development mode
 * @returns true if in development mode
 */
export const isDevelopment = (): boolean => {
  return import.meta.env?.DEV === true;
};
