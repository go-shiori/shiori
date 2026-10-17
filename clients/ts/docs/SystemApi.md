# SystemApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**apiV1SystemInfoGet**](SystemApi.md#apiv1systeminfoget) | **GET** /api/v1/system/info | Get general system information |
| [**apiV1SystemPingGet**](SystemApi.md#apiv1systempingget) | **GET** /api/v1/system/ping | Ping the system |



## apiV1SystemInfoGet

> ApiV1InfoResponse apiV1SystemInfoGet()

Get general system information

Get general system information like Shiori version, database, and OS

### Example

```ts
import {
  Configuration,
  SystemApi,
} from 'shiori-api';
import type { ApiV1SystemInfoGetRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new SystemApi();

  try {
    const data = await api.apiV1SystemInfoGet();
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

[**ApiV1InfoResponse**](ApiV1InfoResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | OK |  -  |
| **403** | Only owners can access this endpoint |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## apiV1SystemPingGet

> apiV1SystemPingGet()

Ping the system

Simple ping endpoint that returns a 200 status code

### Example

```ts
import {
  Configuration,
  SystemApi,
} from 'shiori-api';
import type { ApiV1SystemPingGetRequest } from 'shiori-api';

async function example() {
  console.log("🚀 Testing shiori-api SDK...");
  const api = new SystemApi();

  try {
    const data = await api.apiV1SystemPingGet();
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
| **200** | Pong |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)

