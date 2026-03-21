# shiori-api@1.0.0

A TypeScript SDK client for the localhost API.

## Usage

First, install the SDK from npm.

```bash
npm install shiori-api --save
```

Next, try it out.


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


## Documentation

### API Endpoints

All URIs are relative to *http://localhost*

| Class | Method | HTTP request | Description
| ----- | ------ | ------------ | -------------
*AccountsApi* | [**apiV1AccountsGet**](docs/AccountsApi.md#apiv1accountsget) | **GET** /api/v1/accounts | List accounts
*AccountsApi* | [**apiV1AccountsIdDelete**](docs/AccountsApi.md#apiv1accountsiddelete) | **DELETE** /api/v1/accounts/{id} | Delete an account
*AccountsApi* | [**apiV1AccountsIdPatch**](docs/AccountsApi.md#apiv1accountsidpatch) | **PATCH** /api/v1/accounts/{id} | Update an account
*AccountsApi* | [**apiV1AccountsPost**](docs/AccountsApi.md#apiv1accountspost) | **POST** /api/v1/accounts | Create an account
*AuthApi* | [**apiV1AuthAccountPatch**](docs/AuthApi.md#apiv1authaccountpatch) | **PATCH** /api/v1/auth/account | Update account information
*AuthApi* | [**apiV1AuthLoginPost**](docs/AuthApi.md#apiv1authloginpost) | **POST** /api/v1/auth/login | Login to an account using username and password
*AuthApi* | [**apiV1AuthLogoutPost**](docs/AuthApi.md#apiv1authlogoutpost) | **POST** /api/v1/auth/logout | Logout from the current session
*AuthApi* | [**apiV1AuthMeGet**](docs/AuthApi.md#apiv1authmeget) | **GET** /api/v1/auth/me | Get information for the current logged in user
*AuthApi* | [**apiV1AuthRefreshPost**](docs/AuthApi.md#apiv1authrefreshpost) | **POST** /api/v1/auth/refresh | Refresh a token for an account
*BookmarksApi* | [**apiV1BookmarksBulkTagsPut**](docs/BookmarksApi.md#apiv1bookmarksbulktagsput) | **PUT** /api/v1/bookmarks/bulk/tags | Bulk update tags for multiple bookmarks.
*BookmarksApi* | [**apiV1BookmarksCachePut**](docs/BookmarksApi.md#apiv1bookmarkscacheput) | **PUT** /api/v1/bookmarks/cache | Update Cache and Ebook on server.
*BookmarksApi* | [**apiV1BookmarksDelete**](docs/BookmarksApi.md#apiv1bookmarksdelete) | **DELETE** /api/v1/bookmarks | Delete one or more bookmarks.
*BookmarksApi* | [**apiV1BookmarksGet**](docs/BookmarksApi.md#apiv1bookmarksget) | **GET** /api/v1/bookmarks | List bookmarks with optional filtering and pagination.
*BookmarksApi* | [**apiV1BookmarksIdDataGet**](docs/BookmarksApi.md#apiv1bookmarksiddataget) | **GET** /api/v1/bookmarks/{id}/data | Get bookmark data and resource URLs.
*BookmarksApi* | [**apiV1BookmarksIdDataPut**](docs/BookmarksApi.md#apiv1bookmarksiddataput) | **PUT** /api/v1/bookmarks/{id}/data | Update bookmark data (content, archive, ebook).
*BookmarksApi* | [**apiV1BookmarksIdGet**](docs/BookmarksApi.md#apiv1bookmarksidget) | **GET** /api/v1/bookmarks/{id} | Get a bookmark by ID.
*BookmarksApi* | [**apiV1BookmarksIdPut**](docs/BookmarksApi.md#apiv1bookmarksidput) | **PUT** /api/v1/bookmarks/{id} | Update an existing bookmark.
*BookmarksApi* | [**apiV1BookmarksIdReadableGet**](docs/BookmarksApi.md#apiv1bookmarksidreadableget) | **GET** /api/v1/bookmarks/id/readable | Get readable version of bookmark.
*BookmarksApi* | [**apiV1BookmarksIdTagsDelete**](docs/BookmarksApi.md#apiv1bookmarksidtagsdelete) | **DELETE** /api/v1/bookmarks/{id}/tags | Remove a tag from a bookmark.
*BookmarksApi* | [**apiV1BookmarksIdTagsGet**](docs/BookmarksApi.md#apiv1bookmarksidtagsget) | **GET** /api/v1/bookmarks/{id}/tags | Get tags for a bookmark.
*BookmarksApi* | [**apiV1BookmarksIdTagsPost**](docs/BookmarksApi.md#apiv1bookmarksidtagspost) | **POST** /api/v1/bookmarks/{id}/tags | Add a tag to a bookmark.
*BookmarksApi* | [**apiV1BookmarksPost**](docs/BookmarksApi.md#apiv1bookmarkspost) | **POST** /api/v1/bookmarks | Create a new bookmark.
*SystemApi* | [**apiV1SystemInfoGet**](docs/SystemApi.md#apiv1systeminfoget) | **GET** /api/v1/system/info | Get general system information
*SystemApi* | [**apiV1SystemPingGet**](docs/SystemApi.md#apiv1systempingget) | **GET** /api/v1/system/ping | Ping the system
*TagsApi* | [**apiV1TagsGet**](docs/TagsApi.md#apiv1tagsget) | **GET** /api/v1/tags | List tags
*TagsApi* | [**apiV1TagsIdDelete**](docs/TagsApi.md#apiv1tagsiddelete) | **DELETE** /api/v1/tags/{id} | Delete tag
*TagsApi* | [**apiV1TagsIdGet**](docs/TagsApi.md#apiv1tagsidget) | **GET** /api/v1/tags/{id} | Get tag
*TagsApi* | [**apiV1TagsIdPut**](docs/TagsApi.md#apiv1tagsidput) | **PUT** /api/v1/tags/{id} | Update tag
*TagsApi* | [**apiV1TagsPost**](docs/TagsApi.md#apiv1tagspost) | **POST** /api/v1/tags | Create tag


### Models

- [ApiV1BookmarkDataResponse](docs/ApiV1BookmarkDataResponse.md)
- [ApiV1BookmarkTagPayload](docs/ApiV1BookmarkTagPayload.md)
- [ApiV1BulkUpdateBookmarkTagsPayload](docs/ApiV1BulkUpdateBookmarkTagsPayload.md)
- [ApiV1CreateBookmarkOptions](docs/ApiV1CreateBookmarkOptions.md)
- [ApiV1CreateBookmarkPayload](docs/ApiV1CreateBookmarkPayload.md)
- [ApiV1DeleteBookmarksPayload](docs/ApiV1DeleteBookmarksPayload.md)
- [ApiV1InfoResponse](docs/ApiV1InfoResponse.md)
- [ApiV1InfoResponseVersion](docs/ApiV1InfoResponseVersion.md)
- [ApiV1LoginRequestPayload](docs/ApiV1LoginRequestPayload.md)
- [ApiV1LoginResponseMessage](docs/ApiV1LoginResponseMessage.md)
- [ApiV1ReadableResponseMessage](docs/ApiV1ReadableResponseMessage.md)
- [ApiV1UpdateAccountPayload](docs/ApiV1UpdateAccountPayload.md)
- [ApiV1UpdateBookmarkDataPayload](docs/ApiV1UpdateBookmarkDataPayload.md)
- [ApiV1UpdateBookmarkPayload](docs/ApiV1UpdateBookmarkPayload.md)
- [ApiV1UpdateCachePayload](docs/ApiV1UpdateCachePayload.md)
- [ModelAccount](docs/ModelAccount.md)
- [ModelAccountDTO](docs/ModelAccountDTO.md)
- [ModelBookmarkDTO](docs/ModelBookmarkDTO.md)
- [ModelPaginatedResponseModelBookmarkDTO](docs/ModelPaginatedResponseModelBookmarkDTO.md)
- [ModelPaginatedResponseModelTagDTO](docs/ModelPaginatedResponseModelTagDTO.md)
- [ModelTagDTO](docs/ModelTagDTO.md)
- [ModelUserConfig](docs/ModelUserConfig.md)

### Authorization

Endpoints do not require authorization.


## About

This TypeScript SDK client supports the [Fetch API](https://fetch.spec.whatwg.org/)
and is automatically generated by the
[OpenAPI Generator](https://openapi-generator.tech) project:

- API version: `1.0.0`
- Package version: `1.0.0`
- Generator version: `7.20.0`
- Build package: `org.openapitools.codegen.languages.TypeScriptFetchClientCodegen`

The generated npm module supports the following:

- Environments
  * Node.js
  * Webpack
  * Browserify
- Language levels
  * ES5 - you must have a Promises/A+ library installed
  * ES6
- Module systems
  * CommonJS
  * ES6 module system


## Development

### Building

To build the TypeScript source code, you need to have Node.js and npm installed.
After cloning the repository, navigate to the project directory and run:

```bash
npm install
npm run build
```

### Publishing

Once you've built the package, you can publish it to npm:

```bash
npm publish
```

## License

[]()
