# TagsApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**apiV1TagsGet**](TagsApi.md#apiv1tagsget) | **GET** /api/v1/tags | List tags |
| [**apiV1TagsIdDelete**](TagsApi.md#apiv1tagsiddelete) | **DELETE** /api/v1/tags/{id} | Delete tag |
| [**apiV1TagsIdGet**](TagsApi.md#apiv1tagsidget) | **GET** /api/v1/tags/{id} | Get tag |
| [**apiV1TagsIdPut**](TagsApi.md#apiv1tagsidput) | **PUT** /api/v1/tags/{id} | Update tag |
| [**apiV1TagsPost**](TagsApi.md#apiv1tagspost) | **POST** /api/v1/tags | Create tag |



## apiV1TagsGet

> ModelPaginatedResponseModelTagDTO apiV1TagsGet(withBookmarkCount, bookmarkId, search, page, limit)

List tags

List all tags with pagination

### Example

```ts
import {
  Configuration,
  TagsApi,
} from 'shiori-api';
import type { ApiV1TagsGetRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new TagsApi();

  const body = {
    // boolean | Include bookmark count for each tag (optional)
    withBookmarkCount: true,
    // number | Filter tags by bookmark ID (optional)
    bookmarkId: 56,
    // string | Search tags by name (optional)
    search: search_example,
    // number | Page number (default: 1) (optional)
    page: 56,
    // number | Items per page (default: 30, max: 100) (optional)
    limit: 56,
  } satisfies ApiV1TagsGetRequest;

  try {
    const data = await api.apiV1TagsGet(body);
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
| **withBookmarkCount** | `boolean` | Include bookmark count for each tag | [Optional] [Defaults to `undefined`] |
| **bookmarkId** | `number` | Filter tags by bookmark ID | [Optional] [Defaults to `undefined`] |
| **search** | `string` | Search tags by name | [Optional] [Defaults to `undefined`] |
| **page** | `number` | Page number (default: 1) | [Optional] [Defaults to `undefined`] |
| **limit** | `number` | Items per page (default: 30, max: 100) | [Optional] [Defaults to `undefined`] |

### Return type

[**ModelPaginatedResponseModelTagDTO**](ModelPaginatedResponseModelTagDTO.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **403** | Authentication required |  -  |
| **500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## apiV1TagsIdDelete

> apiV1TagsIdDelete(id)

Delete tag

Delete a tag

### Example

```ts
import {
  Configuration,
  TagsApi,
} from 'shiori-api';
import type { ApiV1TagsIdDeleteRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new TagsApi();

  const body = {
    // number | Tag ID
    id: 56,
  } satisfies ApiV1TagsIdDeleteRequest;

  try {
    const data = await api.apiV1TagsIdDelete(body);
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
| **id** | `number` | Tag ID | [Defaults to `undefined`] |

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
| **204** | No Content |  -  |
| **403** | Authentication required |  -  |
| **404** | Tag not found |  -  |
| **500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## apiV1TagsIdGet

> ModelTagDTO apiV1TagsIdGet(id)

Get tag

Get a tag by ID

### Example

```ts
import {
  Configuration,
  TagsApi,
} from 'shiori-api';
import type { ApiV1TagsIdGetRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new TagsApi();

  const body = {
    // number | Tag ID
    id: 56,
  } satisfies ApiV1TagsIdGetRequest;

  try {
    const data = await api.apiV1TagsIdGet(body);
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
| **id** | `number` | Tag ID | [Defaults to `undefined`] |

### Return type

[**ModelTagDTO**](ModelTagDTO.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **403** | Authentication required |  -  |
| **404** | Tag not found |  -  |
| **500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## apiV1TagsIdPut

> ModelTagDTO apiV1TagsIdPut(id, tag)

Update tag

Update an existing tag

### Example

```ts
import {
  Configuration,
  TagsApi,
} from 'shiori-api';
import type { ApiV1TagsIdPutRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new TagsApi();

  const body = {
    // number | Tag ID
    id: 56,
    // ModelTagDTO | Tag data
    tag: ...,
  } satisfies ApiV1TagsIdPutRequest;

  try {
    const data = await api.apiV1TagsIdPut(body);
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
| **id** | `number` | Tag ID | [Defaults to `undefined`] |
| **tag** | [ModelTagDTO](ModelTagDTO.md) | Tag data | |

### Return type

[**ModelTagDTO**](ModelTagDTO.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **400** | Invalid request |  -  |
| **403** | Authentication required |  -  |
| **404** | Tag not found |  -  |
| **500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## apiV1TagsPost

> ModelTagDTO apiV1TagsPost(tag)

Create tag

Create a new tag. If a tag with the same name already exists, returns 204 No Content.

### Example

```ts
import {
  Configuration,
  TagsApi,
} from 'shiori-api';
import type { ApiV1TagsPostRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new TagsApi();

  const body = {
    // ModelTagDTO | Tag data
    tag: ...,
  } satisfies ApiV1TagsPostRequest;

  try {
    const data = await api.apiV1TagsPost(body);
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
| **tag** | [ModelTagDTO](ModelTagDTO.md) | Tag data | |

### Return type

[**ModelTagDTO**](ModelTagDTO.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | Tag created |  -  |
| **204** | Tag already exists |  -  |
| **400** | Invalid request |  -  |
| **403** | Authentication required |  -  |
| **500** | Internal server error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)

