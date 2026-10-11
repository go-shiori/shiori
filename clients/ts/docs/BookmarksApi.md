# BookmarksApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**apiV1BookmarksBulkTagsPut**](BookmarksApi.md#apiv1bookmarksbulktagsput) | **PUT** /api/v1/bookmarks/bulk/tags | Bulk update tags for multiple bookmarks. |
| [**apiV1BookmarksCachePut**](BookmarksApi.md#apiv1bookmarkscacheput) | **PUT** /api/v1/bookmarks/cache | Update Cache and Ebook on server. |
| [**apiV1BookmarksDelete**](BookmarksApi.md#apiv1bookmarksdelete) | **DELETE** /api/v1/bookmarks | Delete one or more bookmarks. |
| [**apiV1BookmarksGet**](BookmarksApi.md#apiv1bookmarksget) | **GET** /api/v1/bookmarks | List bookmarks with optional filtering and pagination. |
| [**apiV1BookmarksIdDataGet**](BookmarksApi.md#apiv1bookmarksiddataget) | **GET** /api/v1/bookmarks/{id}/data | Get bookmark data and resource URLs. |
| [**apiV1BookmarksIdDataPut**](BookmarksApi.md#apiv1bookmarksiddataput) | **PUT** /api/v1/bookmarks/{id}/data | Update bookmark data (content, archive, ebook). |
| [**apiV1BookmarksIdGet**](BookmarksApi.md#apiv1bookmarksidget) | **GET** /api/v1/bookmarks/{id} | Get a bookmark by ID. |
| [**apiV1BookmarksIdPut**](BookmarksApi.md#apiv1bookmarksidput) | **PUT** /api/v1/bookmarks/{id} | Update an existing bookmark. |
| [**apiV1BookmarksIdReadableGet**](BookmarksApi.md#apiv1bookmarksidreadableget) | **GET** /api/v1/bookmarks/{id}/readable | Get readable version of bookmark. |
| [**apiV1BookmarksIdTagsDelete**](BookmarksApi.md#apiv1bookmarksidtagsdelete) | **DELETE** /api/v1/bookmarks/{id}/tags | Remove a tag from a bookmark. |
| [**apiV1BookmarksIdTagsGet**](BookmarksApi.md#apiv1bookmarksidtagsget) | **GET** /api/v1/bookmarks/{id}/tags | Get tags for a bookmark. |
| [**apiV1BookmarksIdTagsPost**](BookmarksApi.md#apiv1bookmarksidtagspost) | **POST** /api/v1/bookmarks/{id}/tags | Add a tag to a bookmark. |
| [**apiV1BookmarksPost**](BookmarksApi.md#apiv1bookmarkspost) | **POST** /api/v1/bookmarks | Create a new bookmark. |



## apiV1BookmarksBulkTagsPut

> Array&lt;ModelBookmarkDTO&gt; apiV1BookmarksBulkTagsPut(payload)

Bulk update tags for multiple bookmarks.

### Example

```ts
import {
  Configuration,
  BookmarksApi,
} from 'shiori-api';
import type { ApiV1BookmarksBulkTagsPutRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new BookmarksApi();

  const body = {
    // ApiV1BulkUpdateBookmarkTagsPayload | Bulk Update Bookmark Tags Payload
    payload: ...,
  } satisfies ApiV1BookmarksBulkTagsPutRequest;

  try {
    const data = await api.apiV1BookmarksBulkTagsPut(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **payload** | [ApiV1BulkUpdateBookmarkTagsPayload](ApiV1BulkUpdateBookmarkTagsPayload.md) | Bulk Update Bookmark Tags Payload | |

### Return type

[**Array&lt;ModelBookmarkDTO&gt;**](ModelBookmarkDTO.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **400** | Invalid request payload |  -  |
| **403** | Token not provided/invalid |  -  |
| **404** | No bookmarks found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## apiV1BookmarksCachePut

> ModelBookmarkDTO apiV1BookmarksCachePut(payload)

Update Cache and Ebook on server.

### Example

```ts
import {
  Configuration,
  BookmarksApi,
} from 'shiori-api';
import type { ApiV1BookmarksCachePutRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new BookmarksApi();

  const body = {
    // ApiV1UpdateCachePayload | Update Cache Payload
    payload: ...,
  } satisfies ApiV1BookmarksCachePutRequest;

  try {
    const data = await api.apiV1BookmarksCachePut(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **payload** | [ApiV1UpdateCachePayload](ApiV1UpdateCachePayload.md) | Update Cache Payload | |

### Return type

[**ModelBookmarkDTO**](ModelBookmarkDTO.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **403** | Token not provided/invalid |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## apiV1BookmarksDelete

> apiV1BookmarksDelete(payload)

Delete one or more bookmarks.

### Example

```ts
import {
  Configuration,
  BookmarksApi,
} from 'shiori-api';
import type { ApiV1BookmarksDeleteRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new BookmarksApi();

  const body = {
    // ApiV1DeleteBookmarksPayload | Delete Bookmarks Payload
    payload: ...,
  } satisfies ApiV1BookmarksDeleteRequest;

  try {
    const data = await api.apiV1BookmarksDelete(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **payload** | [ApiV1DeleteBookmarksPayload](ApiV1DeleteBookmarksPayload.md) | Delete Bookmarks Payload | |

### Return type

`void` (Empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **400** | Invalid request payload |  -  |
| **403** | Token not provided/invalid |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## apiV1BookmarksGet

> ModelPaginatedResponseModelBookmarkDTO apiV1BookmarksGet(keyword, tags, exclude, page, limit)

List bookmarks with optional filtering and pagination.

### Example

```ts
import {
  Configuration,
  BookmarksApi,
} from 'shiori-api';
import type { ApiV1BookmarksGetRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new BookmarksApi();

  const body = {
    // string | Search keyword (optional)
    keyword: keyword_example,
    // string | Comma-separated list of tags to include (optional)
    tags: tags_example,
    // string | Comma-separated list of tags to exclude (optional)
    exclude: exclude_example,
    // number | Page number (default: 1) (optional)
    page: 56,
    // number | Items per page (default: 30, max: 100) (optional)
    limit: 56,
  } satisfies ApiV1BookmarksGetRequest;

  try {
    const data = await api.apiV1BookmarksGet(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **keyword** | `string` | Search keyword | [Optional] [Defaults to `undefined`] |
| **tags** | `string` | Comma-separated list of tags to include | [Optional] [Defaults to `undefined`] |
| **exclude** | `string` | Comma-separated list of tags to exclude | [Optional] [Defaults to `undefined`] |
| **page** | `number` | Page number (default: 1) | [Optional] [Defaults to `undefined`] |
| **limit** | `number` | Items per page (default: 30, max: 100) | [Optional] [Defaults to `undefined`] |

### Return type

[**ModelPaginatedResponseModelBookmarkDTO**](ModelPaginatedResponseModelBookmarkDTO.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **403** | Token not provided/invalid |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## apiV1BookmarksIdDataGet

> ApiV1BookmarkDataResponse apiV1BookmarksIdDataGet(id)

Get bookmark data and resource URLs.

### Example

```ts
import {
  Configuration,
  BookmarksApi,
} from 'shiori-api';
import type { ApiV1BookmarksIdDataGetRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new BookmarksApi();

  const body = {
    // number | Bookmark ID
    id: 56,
  } satisfies ApiV1BookmarksIdDataGetRequest;

  try {
    const data = await api.apiV1BookmarksIdDataGet(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | `number` | Bookmark ID | [Defaults to `undefined`] |

### Return type

[**ApiV1BookmarkDataResponse**](ApiV1BookmarkDataResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **400** | Invalid bookmark ID |  -  |
| **403** | Token not provided/invalid |  -  |
| **404** | Bookmark not found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## apiV1BookmarksIdDataPut

> ApiV1BookmarkDataResponse apiV1BookmarksIdDataPut(id, payload)

Update bookmark data (content, archive, ebook).

### Example

```ts
import {
  Configuration,
  BookmarksApi,
} from 'shiori-api';
import type { ApiV1BookmarksIdDataPutRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new BookmarksApi();

  const body = {
    // number | Bookmark ID
    id: 56,
    // ApiV1UpdateBookmarkDataPayload | Update Bookmark Data Payload
    payload: ...,
  } satisfies ApiV1BookmarksIdDataPutRequest;

  try {
    const data = await api.apiV1BookmarksIdDataPut(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | `number` | Bookmark ID | [Defaults to `undefined`] |
| **payload** | [ApiV1UpdateBookmarkDataPayload](ApiV1UpdateBookmarkDataPayload.md) | Update Bookmark Data Payload | |

### Return type

[**ApiV1BookmarkDataResponse**](ApiV1BookmarkDataResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **400** | Invalid bookmark ID or payload |  -  |
| **403** | Token not provided/invalid |  -  |
| **404** | Bookmark not found |  -  |
| **500** | Failed to update bookmark data |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## apiV1BookmarksIdGet

> ModelBookmarkDTO apiV1BookmarksIdGet(id)

Get a bookmark by ID.

### Example

```ts
import {
  Configuration,
  BookmarksApi,
} from 'shiori-api';
import type { ApiV1BookmarksIdGetRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new BookmarksApi();

  const body = {
    // number | Bookmark ID
    id: 56,
  } satisfies ApiV1BookmarksIdGetRequest;

  try {
    const data = await api.apiV1BookmarksIdGet(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | `number` | Bookmark ID | [Defaults to `undefined`] |

### Return type

[**ModelBookmarkDTO**](ModelBookmarkDTO.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **403** | Token not provided/invalid |  -  |
| **404** | Bookmark not found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## apiV1BookmarksIdPut

> ModelBookmarkDTO apiV1BookmarksIdPut(id, payload)

Update an existing bookmark.

### Example

```ts
import {
  Configuration,
  BookmarksApi,
} from 'shiori-api';
import type { ApiV1BookmarksIdPutRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new BookmarksApi();

  const body = {
    // number | Bookmark ID
    id: 56,
    // ApiV1UpdateBookmarkPayload | Update Bookmark Payload
    payload: ...,
  } satisfies ApiV1BookmarksIdPutRequest;

  try {
    const data = await api.apiV1BookmarksIdPut(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | `number` | Bookmark ID | [Defaults to `undefined`] |
| **payload** | [ApiV1UpdateBookmarkPayload](ApiV1UpdateBookmarkPayload.md) | Update Bookmark Payload | |

### Return type

[**ModelBookmarkDTO**](ModelBookmarkDTO.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **400** | Invalid request payload |  -  |
| **403** | Token not provided/invalid |  -  |
| **404** | Bookmark not found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## apiV1BookmarksIdReadableGet

> ApiV1ReadableResponseMessage apiV1BookmarksIdReadableGet(id)

Get readable version of bookmark.

### Example

```ts
import {
  Configuration,
  BookmarksApi,
} from 'shiori-api';
import type { ApiV1BookmarksIdReadableGetRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new BookmarksApi();

  const body = {
    // number | Bookmark ID
    id: 56,
  } satisfies ApiV1BookmarksIdReadableGetRequest;

  try {
    const data = await api.apiV1BookmarksIdReadableGet(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | `number` | Bookmark ID | [Defaults to `undefined`] |

### Return type

[**ApiV1ReadableResponseMessage**](ApiV1ReadableResponseMessage.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **403** | Token not provided/invalid |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## apiV1BookmarksIdTagsDelete

> apiV1BookmarksIdTagsDelete(id, payload)

Remove a tag from a bookmark.

### Example

```ts
import {
  Configuration,
  BookmarksApi,
} from 'shiori-api';
import type { ApiV1BookmarksIdTagsDeleteRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new BookmarksApi();

  const body = {
    // number | Bookmark ID
    id: 56,
    // ApiV1BookmarkTagPayload | Remove Tag Payload
    payload: ...,
  } satisfies ApiV1BookmarksIdTagsDeleteRequest;

  try {
    const data = await api.apiV1BookmarksIdTagsDelete(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | `number` | Bookmark ID | [Defaults to `undefined`] |
| **payload** | [ApiV1BookmarkTagPayload](ApiV1BookmarkTagPayload.md) | Remove Tag Payload | |

### Return type

`void` (Empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **403** | Token not provided/invalid |  -  |
| **404** | Bookmark not found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## apiV1BookmarksIdTagsGet

> Array&lt;ModelTagDTO&gt; apiV1BookmarksIdTagsGet(id)

Get tags for a bookmark.

### Example

```ts
import {
  Configuration,
  BookmarksApi,
} from 'shiori-api';
import type { ApiV1BookmarksIdTagsGetRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new BookmarksApi();

  const body = {
    // number | Bookmark ID
    id: 56,
  } satisfies ApiV1BookmarksIdTagsGetRequest;

  try {
    const data = await api.apiV1BookmarksIdTagsGet(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | `number` | Bookmark ID | [Defaults to `undefined`] |

### Return type

[**Array&lt;ModelTagDTO&gt;**](ModelTagDTO.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **403** | Token not provided/invalid |  -  |
| **404** | Bookmark not found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## apiV1BookmarksIdTagsPost

> apiV1BookmarksIdTagsPost(id, payload)

Add a tag to a bookmark.

### Example

```ts
import {
  Configuration,
  BookmarksApi,
} from 'shiori-api';
import type { ApiV1BookmarksIdTagsPostRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new BookmarksApi();

  const body = {
    // number | Bookmark ID
    id: 56,
    // ApiV1BookmarkTagPayload | Add Tag Payload
    payload: ...,
  } satisfies ApiV1BookmarksIdTagsPostRequest;

  try {
    const data = await api.apiV1BookmarksIdTagsPost(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **id** | `number` | Bookmark ID | [Defaults to `undefined`] |
| **payload** | [ApiV1BookmarkTagPayload](ApiV1BookmarkTagPayload.md) | Add Tag Payload | |

### Return type

`void` (Empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **403** | Token not provided/invalid |  -  |
| **404** | Bookmark or tag not found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## apiV1BookmarksPost

> ModelBookmarkDTO apiV1BookmarksPost(payload)

Create a new bookmark.

### Example

```ts
import {
  Configuration,
  BookmarksApi,
} from 'shiori-api';
import type { ApiV1BookmarksPostRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new BookmarksApi();

  const body = {
    // ApiV1CreateBookmarkPayload | Create Bookmark Payload
    payload: ...,
  } satisfies ApiV1BookmarksPostRequest;

  try {
    const data = await api.apiV1BookmarksPost(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **payload** | [ApiV1CreateBookmarkPayload](ApiV1CreateBookmarkPayload.md) | Create Bookmark Payload | |

### Return type

[**ModelBookmarkDTO**](ModelBookmarkDTO.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | Created |  -  |
| **400** | Invalid request payload |  -  |
| **403** | Token not provided/invalid |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)

