# AuthApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**apiV1AuthAccountPatch**](AuthApi.md#apiv1authaccountpatch) | **PATCH** /api/v1/auth/account | Update account information |
| [**apiV1AuthLoginPost**](AuthApi.md#apiv1authloginpost) | **POST** /api/v1/auth/login | Login to an account using username and password |
| [**apiV1AuthLogoutPost**](AuthApi.md#apiv1authlogoutpost) | **POST** /api/v1/auth/logout | Logout from the current session |
| [**apiV1AuthMeGet**](AuthApi.md#apiv1authmeget) | **GET** /api/v1/auth/me | Get information for the current logged in user |
| [**apiV1AuthRefreshPost**](AuthApi.md#apiv1authrefreshpost) | **POST** /api/v1/auth/refresh | Refresh a token for an account |



## apiV1AuthAccountPatch

> ModelAccount apiV1AuthAccountPatch(payload)

Update account information

### Example

```ts
import {
  Configuration,
  AuthApi,
} from 'shiori-api';
import type { ApiV1AuthAccountPatchRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new AuthApi();

  const body = {
    // ApiV1UpdateAccountPayload | Account data (optional)
    payload: ...,
  } satisfies ApiV1AuthAccountPatchRequest;

  try {
    const data = await api.apiV1AuthAccountPatch(body);
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
| **payload** | [ApiV1UpdateAccountPayload](ApiV1UpdateAccountPayload.md) | Account data | [Optional] |

### Return type

[**ModelAccount**](ModelAccount.md)

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


## apiV1AuthLoginPost

> ApiV1LoginResponseMessage apiV1AuthLoginPost(payload)

Login to an account using username and password

### Example

```ts
import {
  Configuration,
  AuthApi,
} from 'shiori-api';
import type { ApiV1AuthLoginPostRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new AuthApi();

  const body = {
    // ApiV1LoginRequestPayload | Login data (optional)
    payload: ...,
  } satisfies ApiV1AuthLoginPostRequest;

  try {
    const data = await api.apiV1AuthLoginPost(body);
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
| **payload** | [ApiV1LoginRequestPayload](ApiV1LoginRequestPayload.md) | Login data | [Optional] |

### Return type

[**ApiV1LoginResponseMessage**](ApiV1LoginResponseMessage.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Login successful |  -  |
| **400** | Invalid login data |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## apiV1AuthLogoutPost

> apiV1AuthLogoutPost()

Logout from the current session

### Example

```ts
import {
  Configuration,
  AuthApi,
} from 'shiori-api';
import type { ApiV1AuthLogoutPostRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new AuthApi();

  try {
    const data = await api.apiV1AuthLogoutPost();
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

`void` (Empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Logout successful |  -  |
| **403** | Token not provided/invalid |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## apiV1AuthMeGet

> ModelAccount apiV1AuthMeGet()

Get information for the current logged in user

### Example

```ts
import {
  Configuration,
  AuthApi,
} from 'shiori-api';
import type { ApiV1AuthMeGetRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new AuthApi();

  try {
    const data = await api.apiV1AuthMeGet();
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

[**ModelAccount**](ModelAccount.md)

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


## apiV1AuthRefreshPost

> ApiV1LoginResponseMessage apiV1AuthRefreshPost()

Refresh a token for an account

### Example

```ts
import {
  Configuration,
  AuthApi,
} from 'shiori-api';
import type { ApiV1AuthRefreshPostRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new AuthApi();

  try {
    const data = await api.apiV1AuthRefreshPost();
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

[**ApiV1LoginResponseMessage**](ApiV1LoginResponseMessage.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Refresh successful |  -  |
| **403** | Token not provided/invalid |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)

