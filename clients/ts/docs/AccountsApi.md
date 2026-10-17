# AccountsApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**apiV1AccountsGet**](AccountsApi.md#apiv1accountsget) | **GET** /api/v1/accounts | List accounts |
| [**apiV1AccountsIdDelete**](AccountsApi.md#apiv1accountsiddelete) | **DELETE** /api/v1/accounts/{id} | Delete an account |
| [**apiV1AccountsIdPatch**](AccountsApi.md#apiv1accountsidpatch) | **PATCH** /api/v1/accounts/{id} | Update an account |
| [**apiV1AccountsPost**](AccountsApi.md#apiv1accountspost) | **POST** /api/v1/accounts | Create an account |



## apiV1AccountsGet

> Array&lt;ModelAccountDTO&gt; apiV1AccountsGet()

List accounts

List accounts

### Example

```ts
import {
  Configuration,
  AccountsApi,
} from 'shiori-api';
import type { ApiV1AccountsGetRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new AccountsApi();

  try {
    const data = await api.apiV1AccountsGet();
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters

This endpoint does not need any parameter.

### Return type

[**Array&lt;ModelAccountDTO&gt;**](ModelAccountDTO.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## apiV1AccountsIdDelete

> apiV1AccountsIdDelete(id)

Delete an account

### Example

```ts
import {
  Configuration,
  AccountsApi,
} from 'shiori-api';
import type { ApiV1AccountsIdDeleteRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new AccountsApi();

  const body = {
    // number | Account ID
    id: 56,
  } satisfies ApiV1AccountsIdDeleteRequest;

  try {
    const data = await api.apiV1AccountsIdDelete(body);
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
| **id** | `number` | Account ID | [Defaults to `undefined`] |

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
| **204** | No content |  -  |
| **400** | Invalid ID |  -  |
| **404** | Account not found |  -  |
| **500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## apiV1AccountsIdPatch

> ModelAccountDTO apiV1AccountsIdPatch(id, account)

Update an account

### Example

```ts
import {
  Configuration,
  AccountsApi,
} from 'shiori-api';
import type { ApiV1AccountsIdPatchRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new AccountsApi();

  const body = {
    // number | Account ID
    id: 56,
    // ApiV1UpdateAccountPayload | Account data
    account: ...,
  } satisfies ApiV1AccountsIdPatchRequest;

  try {
    const data = await api.apiV1AccountsIdPatch(body);
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
| **id** | `number` | Account ID | [Defaults to `undefined`] |
| **account** | [ApiV1UpdateAccountPayload](ApiV1UpdateAccountPayload.md) | Account data | |

### Return type

[**ModelAccountDTO**](ModelAccountDTO.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **400** | Invalid ID/data |  -  |
| **404** | Account not found |  -  |
| **409** | Account already exists |  -  |
| **500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## apiV1AccountsPost

> ModelAccountDTO apiV1AccountsPost()

Create an account

### Example

```ts
import {
  Configuration,
  AccountsApi,
} from 'shiori-api';
import type { ApiV1AccountsPostRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new AccountsApi();

  try {
    const data = await api.apiV1AccountsPost();
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters

This endpoint does not need any parameter.

### Return type

[**ModelAccountDTO**](ModelAccountDTO.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | Created |  -  |
| **400** | Bad Request |  -  |
| **409** | Account already exists |  -  |
| **500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)

