import { getApiBaseUrl } from './env-config';

/**
 * Utility functions for handling image URLs
 */

/**
 * Constructs the thumbnail URL for a bookmark using the API endpoint.
 * This creates URLs pointing to the Shiori backend thumbnail endpoint.
 *
 * In development, this respects VITE_API_BASE_URL environment variable.
 * In production, this uses the current origin.
 *
 * @param bookmarkId - The bookmark ID
 * @returns Full URL string for the thumbnail endpoint
 *
 * @example
 * // Development with default proxy:
 * getBookmarkThumbnailUrl(123) // Returns "http://localhost:5173/bookmark/123/thumb" (proxied to backend)
 *
 * // Development with custom backend:
 * // VITE_API_BASE_URL=http://localhost:8080
 * getBookmarkThumbnailUrl(123) // Returns "http://localhost:8080/bookmark/123/thumb"
 *
 * // Production:
 * getBookmarkThumbnailUrl(123) // Returns "https://your-shiori-domain.com/bookmark/123/thumb"
 */
export const getBookmarkThumbnailUrl = (bookmarkId: number): string => {
  const baseUrl = getApiBaseUrl();
  return `${baseUrl}/bookmark/${bookmarkId}/thumb`;
};

/**
 * Converts a bookmark thumbnail URL to a data URL for authenticated access.
 * This fetches the image with auth headers and converts it to a data URL.
 *
 * @param bookmarkId - The bookmark ID
 * @param authToken - The authentication token
 * @returns Promise that resolves to a data URL string
 */
export const getBookmarkThumbnailDataUrl = async (bookmarkId: number, authToken?: string): Promise<string> => {
  try {
    const thumbnailUrl = getBookmarkThumbnailUrl(bookmarkId);

    const headers: HeadersInit = {
      'Accept': 'image/*'
    };

    if (authToken) {
      headers['Authorization'] = `Bearer ${authToken}`;
      headers['X-Shiori-Response-Format'] = 'new';
    }

    const response = await fetch(thumbnailUrl, {
      method: 'GET',
      headers,
      credentials: 'include' // Include cookies for session-based auth if needed
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch thumbnail: ${response.status}`);
    }

    const blob = await response.blob();
    const reader = new FileReader();

    return new Promise((resolve, reject) => {
      reader.onload = () => {
        resolve(reader.result as string);
      };
      reader.onerror = () => {
        reject(new Error('Failed to convert image to data URL'));
      };
      reader.readAsDataURL(blob);
    });
  } catch (error) {
    console.error('Error fetching bookmark thumbnail:', error);
    return ''; // Return empty string on error, component will show placeholder
  }
};
