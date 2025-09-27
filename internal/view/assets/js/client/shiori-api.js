(() => {
  var __defProp = Object.defineProperty;
  var __export = (target, all) => {
    for (var name in all)
      __defProp(target, name, {
        get: all[name],
        enumerable: true,
        configurable: true,
        set: (newValue) => all[name] = () => newValue
      });
  };

  // dist/esm/index.js
  var exports_esm = {};
  __export(exports_esm, {
    querystring: () => querystring,
    mapValues: () => mapValues,
    instanceOfModelUserConfig: () => instanceOfModelUserConfig,
    instanceOfModelTagDTO: () => instanceOfModelTagDTO,
    instanceOfModelBookmarkDTO: () => instanceOfModelBookmarkDTO,
    instanceOfModelAccountDTO: () => instanceOfModelAccountDTO,
    instanceOfModelAccount: () => instanceOfModelAccount,
    instanceOfApiV1UpdateCachePayload: () => instanceOfApiV1UpdateCachePayload,
    instanceOfApiV1UpdateBookmarkPayload: () => instanceOfApiV1UpdateBookmarkPayload,
    instanceOfApiV1UpdateAccountPayload: () => instanceOfApiV1UpdateAccountPayload,
    instanceOfApiV1ReadableResponseMessage: () => instanceOfApiV1ReadableResponseMessage,
    instanceOfApiV1LoginResponseMessage: () => instanceOfApiV1LoginResponseMessage,
    instanceOfApiV1LoginRequestPayload: () => instanceOfApiV1LoginRequestPayload,
    instanceOfApiV1InfoResponseVersion: () => instanceOfApiV1InfoResponseVersion,
    instanceOfApiV1InfoResponse: () => instanceOfApiV1InfoResponse,
    instanceOfApiV1DeleteBookmarksPayload: () => instanceOfApiV1DeleteBookmarksPayload,
    instanceOfApiV1CreateBookmarkPayload: () => instanceOfApiV1CreateBookmarkPayload,
    instanceOfApiV1BulkUpdateBookmarkTagsPayload: () => instanceOfApiV1BulkUpdateBookmarkTagsPayload,
    instanceOfApiV1BookmarkTagPayload: () => instanceOfApiV1BookmarkTagPayload,
    exists: () => exists,
    canConsumeForm: () => canConsumeForm,
    VoidApiResponse: () => VoidApiResponse,
    TextApiResponse: () => TextApiResponse,
    TagsApi: () => TagsApi,
    SystemApi: () => SystemApi,
    ResponseError: () => ResponseError,
    RequiredError: () => RequiredError,
    ModelUserConfigToJSONTyped: () => ModelUserConfigToJSONTyped,
    ModelUserConfigToJSON: () => ModelUserConfigToJSON,
    ModelUserConfigFromJSONTyped: () => ModelUserConfigFromJSONTyped,
    ModelUserConfigFromJSON: () => ModelUserConfigFromJSON,
    ModelTagDTOToJSONTyped: () => ModelTagDTOToJSONTyped,
    ModelTagDTOToJSON: () => ModelTagDTOToJSON,
    ModelTagDTOFromJSONTyped: () => ModelTagDTOFromJSONTyped,
    ModelTagDTOFromJSON: () => ModelTagDTOFromJSON,
    ModelBookmarkDTOToJSONTyped: () => ModelBookmarkDTOToJSONTyped,
    ModelBookmarkDTOToJSON: () => ModelBookmarkDTOToJSON,
    ModelBookmarkDTOFromJSONTyped: () => ModelBookmarkDTOFromJSONTyped,
    ModelBookmarkDTOFromJSON: () => ModelBookmarkDTOFromJSON,
    ModelAccountToJSONTyped: () => ModelAccountToJSONTyped,
    ModelAccountToJSON: () => ModelAccountToJSON,
    ModelAccountFromJSONTyped: () => ModelAccountFromJSONTyped,
    ModelAccountFromJSON: () => ModelAccountFromJSON,
    ModelAccountDTOToJSONTyped: () => ModelAccountDTOToJSONTyped,
    ModelAccountDTOToJSON: () => ModelAccountDTOToJSON,
    ModelAccountDTOFromJSONTyped: () => ModelAccountDTOFromJSONTyped,
    ModelAccountDTOFromJSON: () => ModelAccountDTOFromJSON,
    JSONApiResponse: () => JSONApiResponse,
    FetchError: () => FetchError,
    DefaultConfig: () => DefaultConfig,
    Configuration: () => Configuration,
    COLLECTION_FORMATS: () => COLLECTION_FORMATS,
    BookmarksApi: () => BookmarksApi,
    BlobApiResponse: () => BlobApiResponse,
    BaseAPI: () => BaseAPI,
    BASE_PATH: () => BASE_PATH,
    AuthApi: () => AuthApi,
    ApiV1UpdateCachePayloadToJSONTyped: () => ApiV1UpdateCachePayloadToJSONTyped,
    ApiV1UpdateCachePayloadToJSON: () => ApiV1UpdateCachePayloadToJSON,
    ApiV1UpdateCachePayloadFromJSONTyped: () => ApiV1UpdateCachePayloadFromJSONTyped,
    ApiV1UpdateCachePayloadFromJSON: () => ApiV1UpdateCachePayloadFromJSON,
    ApiV1UpdateBookmarkPayloadToJSONTyped: () => ApiV1UpdateBookmarkPayloadToJSONTyped,
    ApiV1UpdateBookmarkPayloadToJSON: () => ApiV1UpdateBookmarkPayloadToJSON,
    ApiV1UpdateBookmarkPayloadFromJSONTyped: () => ApiV1UpdateBookmarkPayloadFromJSONTyped,
    ApiV1UpdateBookmarkPayloadFromJSON: () => ApiV1UpdateBookmarkPayloadFromJSON,
    ApiV1UpdateAccountPayloadToJSONTyped: () => ApiV1UpdateAccountPayloadToJSONTyped,
    ApiV1UpdateAccountPayloadToJSON: () => ApiV1UpdateAccountPayloadToJSON,
    ApiV1UpdateAccountPayloadFromJSONTyped: () => ApiV1UpdateAccountPayloadFromJSONTyped,
    ApiV1UpdateAccountPayloadFromJSON: () => ApiV1UpdateAccountPayloadFromJSON,
    ApiV1ReadableResponseMessageToJSONTyped: () => ApiV1ReadableResponseMessageToJSONTyped,
    ApiV1ReadableResponseMessageToJSON: () => ApiV1ReadableResponseMessageToJSON,
    ApiV1ReadableResponseMessageFromJSONTyped: () => ApiV1ReadableResponseMessageFromJSONTyped,
    ApiV1ReadableResponseMessageFromJSON: () => ApiV1ReadableResponseMessageFromJSON,
    ApiV1LoginResponseMessageToJSONTyped: () => ApiV1LoginResponseMessageToJSONTyped,
    ApiV1LoginResponseMessageToJSON: () => ApiV1LoginResponseMessageToJSON,
    ApiV1LoginResponseMessageFromJSONTyped: () => ApiV1LoginResponseMessageFromJSONTyped,
    ApiV1LoginResponseMessageFromJSON: () => ApiV1LoginResponseMessageFromJSON,
    ApiV1LoginRequestPayloadToJSONTyped: () => ApiV1LoginRequestPayloadToJSONTyped,
    ApiV1LoginRequestPayloadToJSON: () => ApiV1LoginRequestPayloadToJSON,
    ApiV1LoginRequestPayloadFromJSONTyped: () => ApiV1LoginRequestPayloadFromJSONTyped,
    ApiV1LoginRequestPayloadFromJSON: () => ApiV1LoginRequestPayloadFromJSON,
    ApiV1InfoResponseVersionToJSONTyped: () => ApiV1InfoResponseVersionToJSONTyped,
    ApiV1InfoResponseVersionToJSON: () => ApiV1InfoResponseVersionToJSON,
    ApiV1InfoResponseVersionFromJSONTyped: () => ApiV1InfoResponseVersionFromJSONTyped,
    ApiV1InfoResponseVersionFromJSON: () => ApiV1InfoResponseVersionFromJSON,
    ApiV1InfoResponseToJSONTyped: () => ApiV1InfoResponseToJSONTyped,
    ApiV1InfoResponseToJSON: () => ApiV1InfoResponseToJSON,
    ApiV1InfoResponseFromJSONTyped: () => ApiV1InfoResponseFromJSONTyped,
    ApiV1InfoResponseFromJSON: () => ApiV1InfoResponseFromJSON,
    ApiV1DeleteBookmarksPayloadToJSONTyped: () => ApiV1DeleteBookmarksPayloadToJSONTyped,
    ApiV1DeleteBookmarksPayloadToJSON: () => ApiV1DeleteBookmarksPayloadToJSON,
    ApiV1DeleteBookmarksPayloadFromJSONTyped: () => ApiV1DeleteBookmarksPayloadFromJSONTyped,
    ApiV1DeleteBookmarksPayloadFromJSON: () => ApiV1DeleteBookmarksPayloadFromJSON,
    ApiV1CreateBookmarkPayloadToJSONTyped: () => ApiV1CreateBookmarkPayloadToJSONTyped,
    ApiV1CreateBookmarkPayloadToJSON: () => ApiV1CreateBookmarkPayloadToJSON,
    ApiV1CreateBookmarkPayloadFromJSONTyped: () => ApiV1CreateBookmarkPayloadFromJSONTyped,
    ApiV1CreateBookmarkPayloadFromJSON: () => ApiV1CreateBookmarkPayloadFromJSON,
    ApiV1BulkUpdateBookmarkTagsPayloadToJSONTyped: () => ApiV1BulkUpdateBookmarkTagsPayloadToJSONTyped,
    ApiV1BulkUpdateBookmarkTagsPayloadToJSON: () => ApiV1BulkUpdateBookmarkTagsPayloadToJSON,
    ApiV1BulkUpdateBookmarkTagsPayloadFromJSONTyped: () => ApiV1BulkUpdateBookmarkTagsPayloadFromJSONTyped,
    ApiV1BulkUpdateBookmarkTagsPayloadFromJSON: () => ApiV1BulkUpdateBookmarkTagsPayloadFromJSON,
    ApiV1BookmarkTagPayloadToJSONTyped: () => ApiV1BookmarkTagPayloadToJSONTyped,
    ApiV1BookmarkTagPayloadToJSON: () => ApiV1BookmarkTagPayloadToJSON,
    ApiV1BookmarkTagPayloadFromJSONTyped: () => ApiV1BookmarkTagPayloadFromJSONTyped,
    ApiV1BookmarkTagPayloadFromJSON: () => ApiV1BookmarkTagPayloadFromJSON,
    AccountsApi: () => AccountsApi
  });

  // dist/esm/runtime.js
  var __awaiter = function(thisArg, _arguments, P, generator) {
    function adopt(value) {
      return value instanceof P ? value : new P(function(resolve) {
        resolve(value);
      });
    }
    return new (P || (P = Promise))(function(resolve, reject) {
      function fulfilled(value) {
        try {
          step(generator.next(value));
        } catch (e) {
          reject(e);
        }
      }
      function rejected(value) {
        try {
          step(generator["throw"](value));
        } catch (e) {
          reject(e);
        }
      }
      function step(result) {
        result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected);
      }
      step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
  };
  var BASE_PATH = "http://localhost".replace(/\/+$/, "");

  class Configuration {
    constructor(configuration = {}) {
      this.configuration = configuration;
    }
    set config(configuration) {
      this.configuration = configuration;
    }
    get basePath() {
      return this.configuration.basePath != null ? this.configuration.basePath : BASE_PATH;
    }
    get fetchApi() {
      return this.configuration.fetchApi;
    }
    get middleware() {
      return this.configuration.middleware || [];
    }
    get queryParamsStringify() {
      return this.configuration.queryParamsStringify || querystring;
    }
    get username() {
      return this.configuration.username;
    }
    get password() {
      return this.configuration.password;
    }
    get apiKey() {
      const apiKey = this.configuration.apiKey;
      if (apiKey) {
        return typeof apiKey === "function" ? apiKey : () => apiKey;
      }
      return;
    }
    get accessToken() {
      const accessToken = this.configuration.accessToken;
      if (accessToken) {
        return typeof accessToken === "function" ? accessToken : () => __awaiter(this, undefined, undefined, function* () {
          return accessToken;
        });
      }
      return;
    }
    get headers() {
      return this.configuration.headers;
    }
    get credentials() {
      return this.configuration.credentials;
    }
  }
  var DefaultConfig = new Configuration;

  class BaseAPI {
    constructor(configuration = DefaultConfig) {
      this.configuration = configuration;
      this.fetchApi = (url, init) => __awaiter(this, undefined, undefined, function* () {
        let fetchParams = { url, init };
        for (const middleware of this.middleware) {
          if (middleware.pre) {
            fetchParams = (yield middleware.pre(Object.assign({ fetch: this.fetchApi }, fetchParams))) || fetchParams;
          }
        }
        let response = undefined;
        try {
          response = yield (this.configuration.fetchApi || fetch)(fetchParams.url, fetchParams.init);
        } catch (e) {
          for (const middleware of this.middleware) {
            if (middleware.onError) {
              response = (yield middleware.onError({
                fetch: this.fetchApi,
                url: fetchParams.url,
                init: fetchParams.init,
                error: e,
                response: response ? response.clone() : undefined
              })) || response;
            }
          }
          if (response === undefined) {
            if (e instanceof Error) {
              throw new FetchError(e, "The request failed and the interceptors did not return an alternative response");
            } else {
              throw e;
            }
          }
        }
        for (const middleware of this.middleware) {
          if (middleware.post) {
            response = (yield middleware.post({
              fetch: this.fetchApi,
              url: fetchParams.url,
              init: fetchParams.init,
              response: response.clone()
            })) || response;
          }
        }
        return response;
      });
      this.middleware = configuration.middleware;
    }
    withMiddleware(...middlewares) {
      const next = this.clone();
      next.middleware = next.middleware.concat(...middlewares);
      return next;
    }
    withPreMiddleware(...preMiddlewares) {
      const middlewares = preMiddlewares.map((pre) => ({ pre }));
      return this.withMiddleware(...middlewares);
    }
    withPostMiddleware(...postMiddlewares) {
      const middlewares = postMiddlewares.map((post) => ({ post }));
      return this.withMiddleware(...middlewares);
    }
    isJsonMime(mime) {
      if (!mime) {
        return false;
      }
      return BaseAPI.jsonRegex.test(mime);
    }
    request(context, initOverrides) {
      return __awaiter(this, undefined, undefined, function* () {
        const { url, init } = yield this.createFetchParams(context, initOverrides);
        const response = yield this.fetchApi(url, init);
        if (response && (response.status >= 200 && response.status < 300)) {
          return response;
        }
        throw new ResponseError(response, "Response returned an error code");
      });
    }
    createFetchParams(context, initOverrides) {
      return __awaiter(this, undefined, undefined, function* () {
        let url = this.configuration.basePath + context.path;
        if (context.query !== undefined && Object.keys(context.query).length !== 0) {
          url += "?" + this.configuration.queryParamsStringify(context.query);
        }
        const headers = Object.assign({}, this.configuration.headers, context.headers);
        Object.keys(headers).forEach((key) => headers[key] === undefined ? delete headers[key] : {});
        const initOverrideFn = typeof initOverrides === "function" ? initOverrides : () => __awaiter(this, undefined, undefined, function* () {
          return initOverrides;
        });
        const initParams = {
          method: context.method,
          headers,
          body: context.body,
          credentials: this.configuration.credentials
        };
        const overriddenInit = Object.assign(Object.assign({}, initParams), yield initOverrideFn({
          init: initParams,
          context
        }));
        let body;
        if (isFormData(overriddenInit.body) || overriddenInit.body instanceof URLSearchParams || isBlob(overriddenInit.body)) {
          body = overriddenInit.body;
        } else if (this.isJsonMime(headers["Content-Type"])) {
          body = JSON.stringify(overriddenInit.body);
        } else {
          body = overriddenInit.body;
        }
        const init = Object.assign(Object.assign({}, overriddenInit), { body });
        return { url, init };
      });
    }
    clone() {
      const constructor = this.constructor;
      const next = new constructor(this.configuration);
      next.middleware = this.middleware.slice();
      return next;
    }
  }
  BaseAPI.jsonRegex = new RegExp("^(:?application/json|[^;/ \t]+/[^;/ \t]+[+]json)[ \t]*(:?;.*)?$", "i");
  function isBlob(value) {
    return typeof Blob !== "undefined" && value instanceof Blob;
  }
  function isFormData(value) {
    return typeof FormData !== "undefined" && value instanceof FormData;
  }

  class ResponseError extends Error {
    constructor(response, msg) {
      super(msg);
      this.response = response;
      this.name = "ResponseError";
    }
  }

  class FetchError extends Error {
    constructor(cause, msg) {
      super(msg);
      this.cause = cause;
      this.name = "FetchError";
    }
  }

  class RequiredError extends Error {
    constructor(field, msg) {
      super(msg);
      this.field = field;
      this.name = "RequiredError";
    }
  }
  var COLLECTION_FORMATS = {
    csv: ",",
    ssv: " ",
    tsv: "\t",
    pipes: "|"
  };
  function querystring(params, prefix = "") {
    return Object.keys(params).map((key) => querystringSingleKey(key, params[key], prefix)).filter((part) => part.length > 0).join("&");
  }
  function querystringSingleKey(key, value, keyPrefix = "") {
    const fullKey = keyPrefix + (keyPrefix.length ? `[${key}]` : key);
    if (value instanceof Array) {
      const multiValue = value.map((singleValue) => encodeURIComponent(String(singleValue))).join(`&${encodeURIComponent(fullKey)}=`);
      return `${encodeURIComponent(fullKey)}=${multiValue}`;
    }
    if (value instanceof Set) {
      const valueAsArray = Array.from(value);
      return querystringSingleKey(key, valueAsArray, keyPrefix);
    }
    if (value instanceof Date) {
      return `${encodeURIComponent(fullKey)}=${encodeURIComponent(value.toISOString())}`;
    }
    if (value instanceof Object) {
      return querystring(value, fullKey);
    }
    return `${encodeURIComponent(fullKey)}=${encodeURIComponent(String(value))}`;
  }
  function exists(json, key) {
    const value = json[key];
    return value !== null && value !== undefined;
  }
  function mapValues(data, fn) {
    const result = {};
    for (const key of Object.keys(data)) {
      result[key] = fn(data[key]);
    }
    return result;
  }
  function canConsumeForm(consumes) {
    for (const consume of consumes) {
      if (consume.contentType === "multipart/form-data") {
        return true;
      }
    }
    return false;
  }

  class JSONApiResponse {
    constructor(raw, transformer = (jsonValue) => jsonValue) {
      this.raw = raw;
      this.transformer = transformer;
    }
    value() {
      return __awaiter(this, undefined, undefined, function* () {
        return this.transformer(yield this.raw.json());
      });
    }
  }

  class VoidApiResponse {
    constructor(raw) {
      this.raw = raw;
    }
    value() {
      return __awaiter(this, undefined, undefined, function* () {
        return;
      });
    }
  }

  class BlobApiResponse {
    constructor(raw) {
      this.raw = raw;
    }
    value() {
      return __awaiter(this, undefined, undefined, function* () {
        return yield this.raw.blob();
      });
    }
  }

  class TextApiResponse {
    constructor(raw) {
      this.raw = raw;
    }
    value() {
      return __awaiter(this, undefined, undefined, function* () {
        return yield this.raw.text();
      });
    }
  }
  // dist/esm/models/ApiV1BookmarkTagPayload.js
  function instanceOfApiV1BookmarkTagPayload(value) {
    if (!("tagId" in value) || value["tagId"] === undefined)
      return false;
    return true;
  }
  function ApiV1BookmarkTagPayloadFromJSON(json) {
    return ApiV1BookmarkTagPayloadFromJSONTyped(json, false);
  }
  function ApiV1BookmarkTagPayloadFromJSONTyped(json, ignoreDiscriminator) {
    if (json == null) {
      return json;
    }
    return {
      tagId: json["tag_id"]
    };
  }
  function ApiV1BookmarkTagPayloadToJSON(json) {
    return ApiV1BookmarkTagPayloadToJSONTyped(json, false);
  }
  function ApiV1BookmarkTagPayloadToJSONTyped(value, ignoreDiscriminator = false) {
    if (value == null) {
      return value;
    }
    return {
      tag_id: value["tagId"]
    };
  }

  // dist/esm/models/ApiV1BulkUpdateBookmarkTagsPayload.js
  function instanceOfApiV1BulkUpdateBookmarkTagsPayload(value) {
    if (!("bookmarkIds" in value) || value["bookmarkIds"] === undefined)
      return false;
    if (!("tagIds" in value) || value["tagIds"] === undefined)
      return false;
    return true;
  }
  function ApiV1BulkUpdateBookmarkTagsPayloadFromJSON(json) {
    return ApiV1BulkUpdateBookmarkTagsPayloadFromJSONTyped(json, false);
  }
  function ApiV1BulkUpdateBookmarkTagsPayloadFromJSONTyped(json, ignoreDiscriminator) {
    if (json == null) {
      return json;
    }
    return {
      bookmarkIds: json["bookmark_ids"],
      tagIds: json["tag_ids"]
    };
  }
  function ApiV1BulkUpdateBookmarkTagsPayloadToJSON(json) {
    return ApiV1BulkUpdateBookmarkTagsPayloadToJSONTyped(json, false);
  }
  function ApiV1BulkUpdateBookmarkTagsPayloadToJSONTyped(value, ignoreDiscriminator = false) {
    if (value == null) {
      return value;
    }
    return {
      bookmark_ids: value["bookmarkIds"],
      tag_ids: value["tagIds"]
    };
  }

  // dist/esm/models/ApiV1CreateBookmarkPayload.js
  function instanceOfApiV1CreateBookmarkPayload(value) {
    if (!("url" in value) || value["url"] === undefined)
      return false;
    return true;
  }
  function ApiV1CreateBookmarkPayloadFromJSON(json) {
    return ApiV1CreateBookmarkPayloadFromJSONTyped(json, false);
  }
  function ApiV1CreateBookmarkPayloadFromJSONTyped(json, ignoreDiscriminator) {
    if (json == null) {
      return json;
    }
    return {
      createEbook: json["create_ebook"] == null ? undefined : json["create_ebook"],
      excerpt: json["excerpt"] == null ? undefined : json["excerpt"],
      _public: json["public"] == null ? undefined : json["public"],
      tags: json["tags"] == null ? undefined : json["tags"],
      title: json["title"] == null ? undefined : json["title"],
      url: json["url"]
    };
  }
  function ApiV1CreateBookmarkPayloadToJSON(json) {
    return ApiV1CreateBookmarkPayloadToJSONTyped(json, false);
  }
  function ApiV1CreateBookmarkPayloadToJSONTyped(value, ignoreDiscriminator = false) {
    if (value == null) {
      return value;
    }
    return {
      create_ebook: value["createEbook"],
      excerpt: value["excerpt"],
      public: value["_public"],
      tags: value["tags"],
      title: value["title"],
      url: value["url"]
    };
  }

  // dist/esm/models/ApiV1DeleteBookmarksPayload.js
  function instanceOfApiV1DeleteBookmarksPayload(value) {
    if (!("ids" in value) || value["ids"] === undefined)
      return false;
    return true;
  }
  function ApiV1DeleteBookmarksPayloadFromJSON(json) {
    return ApiV1DeleteBookmarksPayloadFromJSONTyped(json, false);
  }
  function ApiV1DeleteBookmarksPayloadFromJSONTyped(json, ignoreDiscriminator) {
    if (json == null) {
      return json;
    }
    return {
      ids: json["ids"]
    };
  }
  function ApiV1DeleteBookmarksPayloadToJSON(json) {
    return ApiV1DeleteBookmarksPayloadToJSONTyped(json, false);
  }
  function ApiV1DeleteBookmarksPayloadToJSONTyped(value, ignoreDiscriminator = false) {
    if (value == null) {
      return value;
    }
    return {
      ids: value["ids"]
    };
  }

  // dist/esm/models/ApiV1InfoResponseVersion.js
  function instanceOfApiV1InfoResponseVersion(value) {
    return true;
  }
  function ApiV1InfoResponseVersionFromJSON(json) {
    return ApiV1InfoResponseVersionFromJSONTyped(json, false);
  }
  function ApiV1InfoResponseVersionFromJSONTyped(json, ignoreDiscriminator) {
    if (json == null) {
      return json;
    }
    return {
      commit: json["commit"] == null ? undefined : json["commit"],
      date: json["date"] == null ? undefined : json["date"],
      tag: json["tag"] == null ? undefined : json["tag"]
    };
  }
  function ApiV1InfoResponseVersionToJSON(json) {
    return ApiV1InfoResponseVersionToJSONTyped(json, false);
  }
  function ApiV1InfoResponseVersionToJSONTyped(value, ignoreDiscriminator = false) {
    if (value == null) {
      return value;
    }
    return {
      commit: value["commit"],
      date: value["date"],
      tag: value["tag"]
    };
  }

  // dist/esm/models/ApiV1InfoResponse.js
  function instanceOfApiV1InfoResponse(value) {
    return true;
  }
  function ApiV1InfoResponseFromJSON(json) {
    return ApiV1InfoResponseFromJSONTyped(json, false);
  }
  function ApiV1InfoResponseFromJSONTyped(json, ignoreDiscriminator) {
    if (json == null) {
      return json;
    }
    return {
      database: json["database"] == null ? undefined : json["database"],
      os: json["os"] == null ? undefined : json["os"],
      version: json["version"] == null ? undefined : ApiV1InfoResponseVersionFromJSON(json["version"])
    };
  }
  function ApiV1InfoResponseToJSON(json) {
    return ApiV1InfoResponseToJSONTyped(json, false);
  }
  function ApiV1InfoResponseToJSONTyped(value, ignoreDiscriminator = false) {
    if (value == null) {
      return value;
    }
    return {
      database: value["database"],
      os: value["os"],
      version: ApiV1InfoResponseVersionToJSON(value["version"])
    };
  }

  // dist/esm/models/ApiV1LoginRequestPayload.js
  function instanceOfApiV1LoginRequestPayload(value) {
    return true;
  }
  function ApiV1LoginRequestPayloadFromJSON(json) {
    return ApiV1LoginRequestPayloadFromJSONTyped(json, false);
  }
  function ApiV1LoginRequestPayloadFromJSONTyped(json, ignoreDiscriminator) {
    if (json == null) {
      return json;
    }
    return {
      password: json["password"] == null ? undefined : json["password"],
      rememberMe: json["remember_me"] == null ? undefined : json["remember_me"],
      username: json["username"] == null ? undefined : json["username"]
    };
  }
  function ApiV1LoginRequestPayloadToJSON(json) {
    return ApiV1LoginRequestPayloadToJSONTyped(json, false);
  }
  function ApiV1LoginRequestPayloadToJSONTyped(value, ignoreDiscriminator = false) {
    if (value == null) {
      return value;
    }
    return {
      password: value["password"],
      remember_me: value["rememberMe"],
      username: value["username"]
    };
  }

  // dist/esm/models/ApiV1LoginResponseMessage.js
  function instanceOfApiV1LoginResponseMessage(value) {
    return true;
  }
  function ApiV1LoginResponseMessageFromJSON(json) {
    return ApiV1LoginResponseMessageFromJSONTyped(json, false);
  }
  function ApiV1LoginResponseMessageFromJSONTyped(json, ignoreDiscriminator) {
    if (json == null) {
      return json;
    }
    return {
      expires: json["expires"] == null ? undefined : json["expires"],
      token: json["token"] == null ? undefined : json["token"]
    };
  }
  function ApiV1LoginResponseMessageToJSON(json) {
    return ApiV1LoginResponseMessageToJSONTyped(json, false);
  }
  function ApiV1LoginResponseMessageToJSONTyped(value, ignoreDiscriminator = false) {
    if (value == null) {
      return value;
    }
    return {
      expires: value["expires"],
      token: value["token"]
    };
  }

  // dist/esm/models/ApiV1ReadableResponseMessage.js
  function instanceOfApiV1ReadableResponseMessage(value) {
    return true;
  }
  function ApiV1ReadableResponseMessageFromJSON(json) {
    return ApiV1ReadableResponseMessageFromJSONTyped(json, false);
  }
  function ApiV1ReadableResponseMessageFromJSONTyped(json, ignoreDiscriminator) {
    if (json == null) {
      return json;
    }
    return {
      content: json["content"] == null ? undefined : json["content"],
      html: json["html"] == null ? undefined : json["html"]
    };
  }
  function ApiV1ReadableResponseMessageToJSON(json) {
    return ApiV1ReadableResponseMessageToJSONTyped(json, false);
  }
  function ApiV1ReadableResponseMessageToJSONTyped(value, ignoreDiscriminator = false) {
    if (value == null) {
      return value;
    }
    return {
      content: value["content"],
      html: value["html"]
    };
  }

  // dist/esm/models/ModelUserConfig.js
  function instanceOfModelUserConfig(value) {
    return true;
  }
  function ModelUserConfigFromJSON(json) {
    return ModelUserConfigFromJSONTyped(json, false);
  }
  function ModelUserConfigFromJSONTyped(json, ignoreDiscriminator) {
    if (json == null) {
      return json;
    }
    return {
      createEbook: json["createEbook"] == null ? undefined : json["createEbook"],
      hideExcerpt: json["hideExcerpt"] == null ? undefined : json["hideExcerpt"],
      hideThumbnail: json["hideThumbnail"] == null ? undefined : json["hideThumbnail"],
      keepMetadata: json["keepMetadata"] == null ? undefined : json["keepMetadata"],
      listMode: json["listMode"] == null ? undefined : json["listMode"],
      makePublic: json["makePublic"] == null ? undefined : json["makePublic"],
      showId: json["showId"] == null ? undefined : json["showId"],
      theme: json["theme"] == null ? undefined : json["theme"],
      useArchive: json["useArchive"] == null ? undefined : json["useArchive"]
    };
  }
  function ModelUserConfigToJSON(json) {
    return ModelUserConfigToJSONTyped(json, false);
  }
  function ModelUserConfigToJSONTyped(value, ignoreDiscriminator = false) {
    if (value == null) {
      return value;
    }
    return {
      createEbook: value["createEbook"],
      hideExcerpt: value["hideExcerpt"],
      hideThumbnail: value["hideThumbnail"],
      keepMetadata: value["keepMetadata"],
      listMode: value["listMode"],
      makePublic: value["makePublic"],
      showId: value["showId"],
      theme: value["theme"],
      useArchive: value["useArchive"]
    };
  }

  // dist/esm/models/ApiV1UpdateAccountPayload.js
  function instanceOfApiV1UpdateAccountPayload(value) {
    return true;
  }
  function ApiV1UpdateAccountPayloadFromJSON(json) {
    return ApiV1UpdateAccountPayloadFromJSONTyped(json, false);
  }
  function ApiV1UpdateAccountPayloadFromJSONTyped(json, ignoreDiscriminator) {
    if (json == null) {
      return json;
    }
    return {
      config: json["config"] == null ? undefined : ModelUserConfigFromJSON(json["config"]),
      newPassword: json["new_password"] == null ? undefined : json["new_password"],
      oldPassword: json["old_password"] == null ? undefined : json["old_password"],
      owner: json["owner"] == null ? undefined : json["owner"],
      username: json["username"] == null ? undefined : json["username"]
    };
  }
  function ApiV1UpdateAccountPayloadToJSON(json) {
    return ApiV1UpdateAccountPayloadToJSONTyped(json, false);
  }
  function ApiV1UpdateAccountPayloadToJSONTyped(value, ignoreDiscriminator = false) {
    if (value == null) {
      return value;
    }
    return {
      config: ModelUserConfigToJSON(value["config"]),
      new_password: value["newPassword"],
      old_password: value["oldPassword"],
      owner: value["owner"],
      username: value["username"]
    };
  }

  // dist/esm/models/ApiV1UpdateBookmarkPayload.js
  function instanceOfApiV1UpdateBookmarkPayload(value) {
    return true;
  }
  function ApiV1UpdateBookmarkPayloadFromJSON(json) {
    return ApiV1UpdateBookmarkPayloadFromJSONTyped(json, false);
  }
  function ApiV1UpdateBookmarkPayloadFromJSONTyped(json, ignoreDiscriminator) {
    if (json == null) {
      return json;
    }
    return {
      createEbook: json["create_ebook"] == null ? undefined : json["create_ebook"],
      excerpt: json["excerpt"] == null ? undefined : json["excerpt"],
      _public: json["public"] == null ? undefined : json["public"],
      tags: json["tags"] == null ? undefined : json["tags"],
      title: json["title"] == null ? undefined : json["title"],
      url: json["url"] == null ? undefined : json["url"]
    };
  }
  function ApiV1UpdateBookmarkPayloadToJSON(json) {
    return ApiV1UpdateBookmarkPayloadToJSONTyped(json, false);
  }
  function ApiV1UpdateBookmarkPayloadToJSONTyped(value, ignoreDiscriminator = false) {
    if (value == null) {
      return value;
    }
    return {
      create_ebook: value["createEbook"],
      excerpt: value["excerpt"],
      public: value["_public"],
      tags: value["tags"],
      title: value["title"],
      url: value["url"]
    };
  }

  // dist/esm/models/ApiV1UpdateCachePayload.js
  function instanceOfApiV1UpdateCachePayload(value) {
    if (!("ids" in value) || value["ids"] === undefined)
      return false;
    return true;
  }
  function ApiV1UpdateCachePayloadFromJSON(json) {
    return ApiV1UpdateCachePayloadFromJSONTyped(json, false);
  }
  function ApiV1UpdateCachePayloadFromJSONTyped(json, ignoreDiscriminator) {
    if (json == null) {
      return json;
    }
    return {
      createArchive: json["create_archive"] == null ? undefined : json["create_archive"],
      createEbook: json["create_ebook"] == null ? undefined : json["create_ebook"],
      ids: json["ids"],
      keepMetadata: json["keep_metadata"] == null ? undefined : json["keep_metadata"],
      skipExist: json["skip_exist"] == null ? undefined : json["skip_exist"]
    };
  }
  function ApiV1UpdateCachePayloadToJSON(json) {
    return ApiV1UpdateCachePayloadToJSONTyped(json, false);
  }
  function ApiV1UpdateCachePayloadToJSONTyped(value, ignoreDiscriminator = false) {
    if (value == null) {
      return value;
    }
    return {
      create_archive: value["createArchive"],
      create_ebook: value["createEbook"],
      ids: value["ids"],
      keep_metadata: value["keepMetadata"],
      skip_exist: value["skipExist"]
    };
  }

  // dist/esm/models/ModelAccount.js
  function instanceOfModelAccount(value) {
    return true;
  }
  function ModelAccountFromJSON(json) {
    return ModelAccountFromJSONTyped(json, false);
  }
  function ModelAccountFromJSONTyped(json, ignoreDiscriminator) {
    if (json == null) {
      return json;
    }
    return {
      config: json["config"] == null ? undefined : ModelUserConfigFromJSON(json["config"]),
      id: json["id"] == null ? undefined : json["id"],
      owner: json["owner"] == null ? undefined : json["owner"],
      password: json["password"] == null ? undefined : json["password"],
      username: json["username"] == null ? undefined : json["username"]
    };
  }
  function ModelAccountToJSON(json) {
    return ModelAccountToJSONTyped(json, false);
  }
  function ModelAccountToJSONTyped(value, ignoreDiscriminator = false) {
    if (value == null) {
      return value;
    }
    return {
      config: ModelUserConfigToJSON(value["config"]),
      id: value["id"],
      owner: value["owner"],
      password: value["password"],
      username: value["username"]
    };
  }

  // dist/esm/models/ModelAccountDTO.js
  function instanceOfModelAccountDTO(value) {
    return true;
  }
  function ModelAccountDTOFromJSON(json) {
    return ModelAccountDTOFromJSONTyped(json, false);
  }
  function ModelAccountDTOFromJSONTyped(json, ignoreDiscriminator) {
    if (json == null) {
      return json;
    }
    return {
      config: json["config"] == null ? undefined : ModelUserConfigFromJSON(json["config"]),
      id: json["id"] == null ? undefined : json["id"],
      owner: json["owner"] == null ? undefined : json["owner"],
      passowrd: json["passowrd"] == null ? undefined : json["passowrd"],
      username: json["username"] == null ? undefined : json["username"]
    };
  }
  function ModelAccountDTOToJSON(json) {
    return ModelAccountDTOToJSONTyped(json, false);
  }
  function ModelAccountDTOToJSONTyped(value, ignoreDiscriminator = false) {
    if (value == null) {
      return value;
    }
    return {
      config: ModelUserConfigToJSON(value["config"]),
      id: value["id"],
      owner: value["owner"],
      passowrd: value["passowrd"],
      username: value["username"]
    };
  }

  // dist/esm/models/ModelTagDTO.js
  function instanceOfModelTagDTO(value) {
    return true;
  }
  function ModelTagDTOFromJSON(json) {
    return ModelTagDTOFromJSONTyped(json, false);
  }
  function ModelTagDTOFromJSONTyped(json, ignoreDiscriminator) {
    if (json == null) {
      return json;
    }
    return {
      bookmarkCount: json["bookmark_count"] == null ? undefined : json["bookmark_count"],
      deleted: json["deleted"] == null ? undefined : json["deleted"],
      id: json["id"] == null ? undefined : json["id"],
      name: json["name"] == null ? undefined : json["name"]
    };
  }
  function ModelTagDTOToJSON(json) {
    return ModelTagDTOToJSONTyped(json, false);
  }
  function ModelTagDTOToJSONTyped(value, ignoreDiscriminator = false) {
    if (value == null) {
      return value;
    }
    return {
      bookmark_count: value["bookmarkCount"],
      deleted: value["deleted"],
      id: value["id"],
      name: value["name"]
    };
  }

  // dist/esm/models/ModelBookmarkDTO.js
  function instanceOfModelBookmarkDTO(value) {
    return true;
  }
  function ModelBookmarkDTOFromJSON(json) {
    return ModelBookmarkDTOFromJSONTyped(json, false);
  }
  function ModelBookmarkDTOFromJSONTyped(json, ignoreDiscriminator) {
    if (json == null) {
      return json;
    }
    return {
      author: json["author"] == null ? undefined : json["author"],
      createArchive: json["create_archive"] == null ? undefined : json["create_archive"],
      createEbook: json["create_ebook"] == null ? undefined : json["create_ebook"],
      createdAt: json["createdAt"] == null ? undefined : json["createdAt"],
      excerpt: json["excerpt"] == null ? undefined : json["excerpt"],
      hasArchive: json["hasArchive"] == null ? undefined : json["hasArchive"],
      hasContent: json["hasContent"] == null ? undefined : json["hasContent"],
      hasEbook: json["hasEbook"] == null ? undefined : json["hasEbook"],
      html: json["html"] == null ? undefined : json["html"],
      id: json["id"] == null ? undefined : json["id"],
      imageURL: json["imageURL"] == null ? undefined : json["imageURL"],
      modifiedAt: json["modifiedAt"] == null ? undefined : json["modifiedAt"],
      _public: json["public"] == null ? undefined : json["public"],
      tags: json["tags"] == null ? undefined : json["tags"].map(ModelTagDTOFromJSON),
      title: json["title"] == null ? undefined : json["title"],
      url: json["url"] == null ? undefined : json["url"]
    };
  }
  function ModelBookmarkDTOToJSON(json) {
    return ModelBookmarkDTOToJSONTyped(json, false);
  }
  function ModelBookmarkDTOToJSONTyped(value, ignoreDiscriminator = false) {
    if (value == null) {
      return value;
    }
    return {
      author: value["author"],
      create_archive: value["createArchive"],
      create_ebook: value["createEbook"],
      createdAt: value["createdAt"],
      excerpt: value["excerpt"],
      hasArchive: value["hasArchive"],
      hasContent: value["hasContent"],
      hasEbook: value["hasEbook"],
      html: value["html"],
      id: value["id"],
      imageURL: value["imageURL"],
      modifiedAt: value["modifiedAt"],
      public: value["_public"],
      tags: value["tags"] == null ? undefined : value["tags"].map(ModelTagDTOToJSON),
      title: value["title"],
      url: value["url"]
    };
  }

  // dist/esm/apis/AccountsApi.js
  var __awaiter2 = function(thisArg, _arguments, P, generator) {
    function adopt(value) {
      return value instanceof P ? value : new P(function(resolve) {
        resolve(value);
      });
    }
    return new (P || (P = Promise))(function(resolve, reject) {
      function fulfilled(value) {
        try {
          step(generator.next(value));
        } catch (e) {
          reject(e);
        }
      }
      function rejected(value) {
        try {
          step(generator["throw"](value));
        } catch (e) {
          reject(e);
        }
      }
      function step(result) {
        result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected);
      }
      step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
  };

  class AccountsApi extends BaseAPI {
    apiV1AccountsGetRaw(initOverrides) {
      return __awaiter2(this, undefined, undefined, function* () {
        const queryParameters = {};
        const headerParameters = {};
        let urlPath = `/api/v1/accounts`;
        const response = yield this.request({
          path: urlPath,
          method: "GET",
          headers: headerParameters,
          query: queryParameters
        }, initOverrides);
        return new JSONApiResponse(response, (jsonValue) => jsonValue.map(ModelAccountDTOFromJSON));
      });
    }
    apiV1AccountsGet(initOverrides) {
      return __awaiter2(this, undefined, undefined, function* () {
        const response = yield this.apiV1AccountsGetRaw(initOverrides);
        return yield response.value();
      });
    }
    apiV1AccountsIdDeleteRaw(requestParameters, initOverrides) {
      return __awaiter2(this, undefined, undefined, function* () {
        if (requestParameters["id"] == null) {
          throw new RequiredError("id", 'Required parameter "id" was null or undefined when calling apiV1AccountsIdDelete().');
        }
        const queryParameters = {};
        const headerParameters = {};
        let urlPath = `/api/v1/accounts/{id}`;
        urlPath = urlPath.replace(`{${"id"}}`, encodeURIComponent(String(requestParameters["id"])));
        const response = yield this.request({
          path: urlPath,
          method: "DELETE",
          headers: headerParameters,
          query: queryParameters
        }, initOverrides);
        return new VoidApiResponse(response);
      });
    }
    apiV1AccountsIdDelete(requestParameters, initOverrides) {
      return __awaiter2(this, undefined, undefined, function* () {
        yield this.apiV1AccountsIdDeleteRaw(requestParameters, initOverrides);
      });
    }
    apiV1AccountsIdPatchRaw(requestParameters, initOverrides) {
      return __awaiter2(this, undefined, undefined, function* () {
        if (requestParameters["id"] == null) {
          throw new RequiredError("id", 'Required parameter "id" was null or undefined when calling apiV1AccountsIdPatch().');
        }
        if (requestParameters["account"] == null) {
          throw new RequiredError("account", 'Required parameter "account" was null or undefined when calling apiV1AccountsIdPatch().');
        }
        const queryParameters = {};
        const headerParameters = {};
        headerParameters["Content-Type"] = "application/json";
        let urlPath = `/api/v1/accounts/{id}`;
        urlPath = urlPath.replace(`{${"id"}}`, encodeURIComponent(String(requestParameters["id"])));
        const response = yield this.request({
          path: urlPath,
          method: "PATCH",
          headers: headerParameters,
          query: queryParameters,
          body: ApiV1UpdateAccountPayloadToJSON(requestParameters["account"])
        }, initOverrides);
        return new JSONApiResponse(response, (jsonValue) => ModelAccountDTOFromJSON(jsonValue));
      });
    }
    apiV1AccountsIdPatch(requestParameters, initOverrides) {
      return __awaiter2(this, undefined, undefined, function* () {
        const response = yield this.apiV1AccountsIdPatchRaw(requestParameters, initOverrides);
        return yield response.value();
      });
    }
    apiV1AccountsPostRaw(initOverrides) {
      return __awaiter2(this, undefined, undefined, function* () {
        const queryParameters = {};
        const headerParameters = {};
        let urlPath = `/api/v1/accounts`;
        const response = yield this.request({
          path: urlPath,
          method: "POST",
          headers: headerParameters,
          query: queryParameters
        }, initOverrides);
        return new JSONApiResponse(response, (jsonValue) => ModelAccountDTOFromJSON(jsonValue));
      });
    }
    apiV1AccountsPost(initOverrides) {
      return __awaiter2(this, undefined, undefined, function* () {
        const response = yield this.apiV1AccountsPostRaw(initOverrides);
        return yield response.value();
      });
    }
  }

  // dist/esm/apis/AuthApi.js
  var __awaiter3 = function(thisArg, _arguments, P, generator) {
    function adopt(value) {
      return value instanceof P ? value : new P(function(resolve) {
        resolve(value);
      });
    }
    return new (P || (P = Promise))(function(resolve, reject) {
      function fulfilled(value) {
        try {
          step(generator.next(value));
        } catch (e) {
          reject(e);
        }
      }
      function rejected(value) {
        try {
          step(generator["throw"](value));
        } catch (e) {
          reject(e);
        }
      }
      function step(result) {
        result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected);
      }
      step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
  };

  class AuthApi extends BaseAPI {
    apiV1AuthAccountPatchRaw(requestParameters, initOverrides) {
      return __awaiter3(this, undefined, undefined, function* () {
        const queryParameters = {};
        const headerParameters = {};
        headerParameters["Content-Type"] = "application/json";
        let urlPath = `/api/v1/auth/account`;
        const response = yield this.request({
          path: urlPath,
          method: "PATCH",
          headers: headerParameters,
          query: queryParameters,
          body: ApiV1UpdateAccountPayloadToJSON(requestParameters["payload"])
        }, initOverrides);
        return new JSONApiResponse(response, (jsonValue) => ModelAccountFromJSON(jsonValue));
      });
    }
    apiV1AuthAccountPatch() {
      return __awaiter3(this, arguments, undefined, function* (requestParameters = {}, initOverrides) {
        const response = yield this.apiV1AuthAccountPatchRaw(requestParameters, initOverrides);
        return yield response.value();
      });
    }
    apiV1AuthLoginPostRaw(requestParameters, initOverrides) {
      return __awaiter3(this, undefined, undefined, function* () {
        const queryParameters = {};
        const headerParameters = {};
        headerParameters["Content-Type"] = "application/json";
        let urlPath = `/api/v1/auth/login`;
        const response = yield this.request({
          path: urlPath,
          method: "POST",
          headers: headerParameters,
          query: queryParameters,
          body: ApiV1LoginRequestPayloadToJSON(requestParameters["payload"])
        }, initOverrides);
        return new JSONApiResponse(response, (jsonValue) => ApiV1LoginResponseMessageFromJSON(jsonValue));
      });
    }
    apiV1AuthLoginPost() {
      return __awaiter3(this, arguments, undefined, function* (requestParameters = {}, initOverrides) {
        const response = yield this.apiV1AuthLoginPostRaw(requestParameters, initOverrides);
        return yield response.value();
      });
    }
    apiV1AuthLogoutPostRaw(initOverrides) {
      return __awaiter3(this, undefined, undefined, function* () {
        const queryParameters = {};
        const headerParameters = {};
        let urlPath = `/api/v1/auth/logout`;
        const response = yield this.request({
          path: urlPath,
          method: "POST",
          headers: headerParameters,
          query: queryParameters
        }, initOverrides);
        return new VoidApiResponse(response);
      });
    }
    apiV1AuthLogoutPost(initOverrides) {
      return __awaiter3(this, undefined, undefined, function* () {
        yield this.apiV1AuthLogoutPostRaw(initOverrides);
      });
    }
    apiV1AuthMeGetRaw(initOverrides) {
      return __awaiter3(this, undefined, undefined, function* () {
        const queryParameters = {};
        const headerParameters = {};
        let urlPath = `/api/v1/auth/me`;
        const response = yield this.request({
          path: urlPath,
          method: "GET",
          headers: headerParameters,
          query: queryParameters
        }, initOverrides);
        return new JSONApiResponse(response, (jsonValue) => ModelAccountFromJSON(jsonValue));
      });
    }
    apiV1AuthMeGet(initOverrides) {
      return __awaiter3(this, undefined, undefined, function* () {
        const response = yield this.apiV1AuthMeGetRaw(initOverrides);
        return yield response.value();
      });
    }
    apiV1AuthRefreshPostRaw(initOverrides) {
      return __awaiter3(this, undefined, undefined, function* () {
        const queryParameters = {};
        const headerParameters = {};
        let urlPath = `/api/v1/auth/refresh`;
        const response = yield this.request({
          path: urlPath,
          method: "POST",
          headers: headerParameters,
          query: queryParameters
        }, initOverrides);
        return new JSONApiResponse(response, (jsonValue) => ApiV1LoginResponseMessageFromJSON(jsonValue));
      });
    }
    apiV1AuthRefreshPost(initOverrides) {
      return __awaiter3(this, undefined, undefined, function* () {
        const response = yield this.apiV1AuthRefreshPostRaw(initOverrides);
        return yield response.value();
      });
    }
  }

  // dist/esm/apis/BookmarksApi.js
  var __awaiter4 = function(thisArg, _arguments, P, generator) {
    function adopt(value) {
      return value instanceof P ? value : new P(function(resolve) {
        resolve(value);
      });
    }
    return new (P || (P = Promise))(function(resolve, reject) {
      function fulfilled(value) {
        try {
          step(generator.next(value));
        } catch (e) {
          reject(e);
        }
      }
      function rejected(value) {
        try {
          step(generator["throw"](value));
        } catch (e) {
          reject(e);
        }
      }
      function step(result) {
        result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected);
      }
      step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
  };

  class BookmarksApi extends BaseAPI {
    apiV1BookmarksBulkTagsPutRaw(requestParameters, initOverrides) {
      return __awaiter4(this, undefined, undefined, function* () {
        if (requestParameters["payload"] == null) {
          throw new RequiredError("payload", 'Required parameter "payload" was null or undefined when calling apiV1BookmarksBulkTagsPut().');
        }
        const queryParameters = {};
        const headerParameters = {};
        headerParameters["Content-Type"] = "application/json";
        let urlPath = `/api/v1/bookmarks/bulk/tags`;
        const response = yield this.request({
          path: urlPath,
          method: "PUT",
          headers: headerParameters,
          query: queryParameters,
          body: ApiV1BulkUpdateBookmarkTagsPayloadToJSON(requestParameters["payload"])
        }, initOverrides);
        return new JSONApiResponse(response, (jsonValue) => jsonValue.map(ModelBookmarkDTOFromJSON));
      });
    }
    apiV1BookmarksBulkTagsPut(requestParameters, initOverrides) {
      return __awaiter4(this, undefined, undefined, function* () {
        const response = yield this.apiV1BookmarksBulkTagsPutRaw(requestParameters, initOverrides);
        return yield response.value();
      });
    }
    apiV1BookmarksCachePutRaw(requestParameters, initOverrides) {
      return __awaiter4(this, undefined, undefined, function* () {
        if (requestParameters["payload"] == null) {
          throw new RequiredError("payload", 'Required parameter "payload" was null or undefined when calling apiV1BookmarksCachePut().');
        }
        const queryParameters = {};
        const headerParameters = {};
        headerParameters["Content-Type"] = "application/json";
        let urlPath = `/api/v1/bookmarks/cache`;
        const response = yield this.request({
          path: urlPath,
          method: "PUT",
          headers: headerParameters,
          query: queryParameters,
          body: ApiV1UpdateCachePayloadToJSON(requestParameters["payload"])
        }, initOverrides);
        return new JSONApiResponse(response, (jsonValue) => ModelBookmarkDTOFromJSON(jsonValue));
      });
    }
    apiV1BookmarksCachePut(requestParameters, initOverrides) {
      return __awaiter4(this, undefined, undefined, function* () {
        const response = yield this.apiV1BookmarksCachePutRaw(requestParameters, initOverrides);
        return yield response.value();
      });
    }
    apiV1BookmarksDeleteRaw(requestParameters, initOverrides) {
      return __awaiter4(this, undefined, undefined, function* () {
        if (requestParameters["payload"] == null) {
          throw new RequiredError("payload", 'Required parameter "payload" was null or undefined when calling apiV1BookmarksDelete().');
        }
        const queryParameters = {};
        const headerParameters = {};
        headerParameters["Content-Type"] = "application/json";
        let urlPath = `/api/v1/bookmarks`;
        const response = yield this.request({
          path: urlPath,
          method: "DELETE",
          headers: headerParameters,
          query: queryParameters,
          body: ApiV1DeleteBookmarksPayloadToJSON(requestParameters["payload"])
        }, initOverrides);
        return new VoidApiResponse(response);
      });
    }
    apiV1BookmarksDelete(requestParameters, initOverrides) {
      return __awaiter4(this, undefined, undefined, function* () {
        yield this.apiV1BookmarksDeleteRaw(requestParameters, initOverrides);
      });
    }
    apiV1BookmarksGetRaw(requestParameters, initOverrides) {
      return __awaiter4(this, undefined, undefined, function* () {
        const queryParameters = {};
        if (requestParameters["keyword"] != null) {
          queryParameters["keyword"] = requestParameters["keyword"];
        }
        if (requestParameters["tags"] != null) {
          queryParameters["tags"] = requestParameters["tags"];
        }
        if (requestParameters["exclude"] != null) {
          queryParameters["exclude"] = requestParameters["exclude"];
        }
        if (requestParameters["page"] != null) {
          queryParameters["page"] = requestParameters["page"];
        }
        if (requestParameters["limit"] != null) {
          queryParameters["limit"] = requestParameters["limit"];
        }
        const headerParameters = {};
        let urlPath = `/api/v1/bookmarks`;
        const response = yield this.request({
          path: urlPath,
          method: "GET",
          headers: headerParameters,
          query: queryParameters
        }, initOverrides);
        return new JSONApiResponse(response, (jsonValue) => jsonValue.map(ModelBookmarkDTOFromJSON));
      });
    }
    apiV1BookmarksGet() {
      return __awaiter4(this, arguments, undefined, function* (requestParameters = {}, initOverrides) {
        const response = yield this.apiV1BookmarksGetRaw(requestParameters, initOverrides);
        return yield response.value();
      });
    }
    apiV1BookmarksIdGetRaw(requestParameters, initOverrides) {
      return __awaiter4(this, undefined, undefined, function* () {
        if (requestParameters["id"] == null) {
          throw new RequiredError("id", 'Required parameter "id" was null or undefined when calling apiV1BookmarksIdGet().');
        }
        const queryParameters = {};
        const headerParameters = {};
        let urlPath = `/api/v1/bookmarks/{id}`;
        urlPath = urlPath.replace(`{${"id"}}`, encodeURIComponent(String(requestParameters["id"])));
        const response = yield this.request({
          path: urlPath,
          method: "GET",
          headers: headerParameters,
          query: queryParameters
        }, initOverrides);
        return new JSONApiResponse(response, (jsonValue) => ModelBookmarkDTOFromJSON(jsonValue));
      });
    }
    apiV1BookmarksIdGet(requestParameters, initOverrides) {
      return __awaiter4(this, undefined, undefined, function* () {
        const response = yield this.apiV1BookmarksIdGetRaw(requestParameters, initOverrides);
        return yield response.value();
      });
    }
    apiV1BookmarksIdPutRaw(requestParameters, initOverrides) {
      return __awaiter4(this, undefined, undefined, function* () {
        if (requestParameters["id"] == null) {
          throw new RequiredError("id", 'Required parameter "id" was null or undefined when calling apiV1BookmarksIdPut().');
        }
        if (requestParameters["payload"] == null) {
          throw new RequiredError("payload", 'Required parameter "payload" was null or undefined when calling apiV1BookmarksIdPut().');
        }
        const queryParameters = {};
        const headerParameters = {};
        headerParameters["Content-Type"] = "application/json";
        let urlPath = `/api/v1/bookmarks/{id}`;
        urlPath = urlPath.replace(`{${"id"}}`, encodeURIComponent(String(requestParameters["id"])));
        const response = yield this.request({
          path: urlPath,
          method: "PUT",
          headers: headerParameters,
          query: queryParameters,
          body: ApiV1UpdateBookmarkPayloadToJSON(requestParameters["payload"])
        }, initOverrides);
        return new JSONApiResponse(response, (jsonValue) => ModelBookmarkDTOFromJSON(jsonValue));
      });
    }
    apiV1BookmarksIdPut(requestParameters, initOverrides) {
      return __awaiter4(this, undefined, undefined, function* () {
        const response = yield this.apiV1BookmarksIdPutRaw(requestParameters, initOverrides);
        return yield response.value();
      });
    }
    apiV1BookmarksIdReadableGetRaw(initOverrides) {
      return __awaiter4(this, undefined, undefined, function* () {
        const queryParameters = {};
        const headerParameters = {};
        let urlPath = `/api/v1/bookmarks/id/readable`;
        const response = yield this.request({
          path: urlPath,
          method: "GET",
          headers: headerParameters,
          query: queryParameters
        }, initOverrides);
        return new JSONApiResponse(response, (jsonValue) => ApiV1ReadableResponseMessageFromJSON(jsonValue));
      });
    }
    apiV1BookmarksIdReadableGet(initOverrides) {
      return __awaiter4(this, undefined, undefined, function* () {
        const response = yield this.apiV1BookmarksIdReadableGetRaw(initOverrides);
        return yield response.value();
      });
    }
    apiV1BookmarksIdTagsDeleteRaw(requestParameters, initOverrides) {
      return __awaiter4(this, undefined, undefined, function* () {
        if (requestParameters["id"] == null) {
          throw new RequiredError("id", 'Required parameter "id" was null or undefined when calling apiV1BookmarksIdTagsDelete().');
        }
        if (requestParameters["payload"] == null) {
          throw new RequiredError("payload", 'Required parameter "payload" was null or undefined when calling apiV1BookmarksIdTagsDelete().');
        }
        const queryParameters = {};
        const headerParameters = {};
        headerParameters["Content-Type"] = "application/json";
        let urlPath = `/api/v1/bookmarks/{id}/tags`;
        urlPath = urlPath.replace(`{${"id"}}`, encodeURIComponent(String(requestParameters["id"])));
        const response = yield this.request({
          path: urlPath,
          method: "DELETE",
          headers: headerParameters,
          query: queryParameters,
          body: ApiV1BookmarkTagPayloadToJSON(requestParameters["payload"])
        }, initOverrides);
        return new VoidApiResponse(response);
      });
    }
    apiV1BookmarksIdTagsDelete(requestParameters, initOverrides) {
      return __awaiter4(this, undefined, undefined, function* () {
        yield this.apiV1BookmarksIdTagsDeleteRaw(requestParameters, initOverrides);
      });
    }
    apiV1BookmarksIdTagsGetRaw(requestParameters, initOverrides) {
      return __awaiter4(this, undefined, undefined, function* () {
        if (requestParameters["id"] == null) {
          throw new RequiredError("id", 'Required parameter "id" was null or undefined when calling apiV1BookmarksIdTagsGet().');
        }
        const queryParameters = {};
        const headerParameters = {};
        let urlPath = `/api/v1/bookmarks/{id}/tags`;
        urlPath = urlPath.replace(`{${"id"}}`, encodeURIComponent(String(requestParameters["id"])));
        const response = yield this.request({
          path: urlPath,
          method: "GET",
          headers: headerParameters,
          query: queryParameters
        }, initOverrides);
        return new JSONApiResponse(response, (jsonValue) => jsonValue.map(ModelTagDTOFromJSON));
      });
    }
    apiV1BookmarksIdTagsGet(requestParameters, initOverrides) {
      return __awaiter4(this, undefined, undefined, function* () {
        const response = yield this.apiV1BookmarksIdTagsGetRaw(requestParameters, initOverrides);
        return yield response.value();
      });
    }
    apiV1BookmarksIdTagsPostRaw(requestParameters, initOverrides) {
      return __awaiter4(this, undefined, undefined, function* () {
        if (requestParameters["id"] == null) {
          throw new RequiredError("id", 'Required parameter "id" was null or undefined when calling apiV1BookmarksIdTagsPost().');
        }
        if (requestParameters["payload"] == null) {
          throw new RequiredError("payload", 'Required parameter "payload" was null or undefined when calling apiV1BookmarksIdTagsPost().');
        }
        const queryParameters = {};
        const headerParameters = {};
        headerParameters["Content-Type"] = "application/json";
        let urlPath = `/api/v1/bookmarks/{id}/tags`;
        urlPath = urlPath.replace(`{${"id"}}`, encodeURIComponent(String(requestParameters["id"])));
        const response = yield this.request({
          path: urlPath,
          method: "POST",
          headers: headerParameters,
          query: queryParameters,
          body: ApiV1BookmarkTagPayloadToJSON(requestParameters["payload"])
        }, initOverrides);
        return new VoidApiResponse(response);
      });
    }
    apiV1BookmarksIdTagsPost(requestParameters, initOverrides) {
      return __awaiter4(this, undefined, undefined, function* () {
        yield this.apiV1BookmarksIdTagsPostRaw(requestParameters, initOverrides);
      });
    }
    apiV1BookmarksPostRaw(requestParameters, initOverrides) {
      return __awaiter4(this, undefined, undefined, function* () {
        if (requestParameters["payload"] == null) {
          throw new RequiredError("payload", 'Required parameter "payload" was null or undefined when calling apiV1BookmarksPost().');
        }
        const queryParameters = {};
        const headerParameters = {};
        headerParameters["Content-Type"] = "application/json";
        let urlPath = `/api/v1/bookmarks`;
        const response = yield this.request({
          path: urlPath,
          method: "POST",
          headers: headerParameters,
          query: queryParameters,
          body: ApiV1CreateBookmarkPayloadToJSON(requestParameters["payload"])
        }, initOverrides);
        return new JSONApiResponse(response, (jsonValue) => ModelBookmarkDTOFromJSON(jsonValue));
      });
    }
    apiV1BookmarksPost(requestParameters, initOverrides) {
      return __awaiter4(this, undefined, undefined, function* () {
        const response = yield this.apiV1BookmarksPostRaw(requestParameters, initOverrides);
        return yield response.value();
      });
    }
  }

  // dist/esm/apis/SystemApi.js
  var __awaiter5 = function(thisArg, _arguments, P, generator) {
    function adopt(value) {
      return value instanceof P ? value : new P(function(resolve) {
        resolve(value);
      });
    }
    return new (P || (P = Promise))(function(resolve, reject) {
      function fulfilled(value) {
        try {
          step(generator.next(value));
        } catch (e) {
          reject(e);
        }
      }
      function rejected(value) {
        try {
          step(generator["throw"](value));
        } catch (e) {
          reject(e);
        }
      }
      function step(result) {
        result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected);
      }
      step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
  };

  class SystemApi extends BaseAPI {
    apiV1SystemInfoGetRaw(initOverrides) {
      return __awaiter5(this, undefined, undefined, function* () {
        const queryParameters = {};
        const headerParameters = {};
        let urlPath = `/api/v1/system/info`;
        const response = yield this.request({
          path: urlPath,
          method: "GET",
          headers: headerParameters,
          query: queryParameters
        }, initOverrides);
        return new JSONApiResponse(response, (jsonValue) => ApiV1InfoResponseFromJSON(jsonValue));
      });
    }
    apiV1SystemInfoGet(initOverrides) {
      return __awaiter5(this, undefined, undefined, function* () {
        const response = yield this.apiV1SystemInfoGetRaw(initOverrides);
        return yield response.value();
      });
    }
  }

  // dist/esm/apis/TagsApi.js
  var __awaiter6 = function(thisArg, _arguments, P, generator) {
    function adopt(value) {
      return value instanceof P ? value : new P(function(resolve) {
        resolve(value);
      });
    }
    return new (P || (P = Promise))(function(resolve, reject) {
      function fulfilled(value) {
        try {
          step(generator.next(value));
        } catch (e) {
          reject(e);
        }
      }
      function rejected(value) {
        try {
          step(generator["throw"](value));
        } catch (e) {
          reject(e);
        }
      }
      function step(result) {
        result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected);
      }
      step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
  };

  class TagsApi extends BaseAPI {
    apiV1TagsGetRaw(requestParameters, initOverrides) {
      return __awaiter6(this, undefined, undefined, function* () {
        const queryParameters = {};
        if (requestParameters["withBookmarkCount"] != null) {
          queryParameters["with_bookmark_count"] = requestParameters["withBookmarkCount"];
        }
        if (requestParameters["bookmarkId"] != null) {
          queryParameters["bookmark_id"] = requestParameters["bookmarkId"];
        }
        if (requestParameters["search"] != null) {
          queryParameters["search"] = requestParameters["search"];
        }
        const headerParameters = {};
        let urlPath = `/api/v1/tags`;
        const response = yield this.request({
          path: urlPath,
          method: "GET",
          headers: headerParameters,
          query: queryParameters
        }, initOverrides);
        return new JSONApiResponse(response, (jsonValue) => jsonValue.map(ModelTagDTOFromJSON));
      });
    }
    apiV1TagsGet() {
      return __awaiter6(this, arguments, undefined, function* (requestParameters = {}, initOverrides) {
        const response = yield this.apiV1TagsGetRaw(requestParameters, initOverrides);
        return yield response.value();
      });
    }
    apiV1TagsIdDeleteRaw(requestParameters, initOverrides) {
      return __awaiter6(this, undefined, undefined, function* () {
        if (requestParameters["id"] == null) {
          throw new RequiredError("id", 'Required parameter "id" was null or undefined when calling apiV1TagsIdDelete().');
        }
        const queryParameters = {};
        const headerParameters = {};
        let urlPath = `/api/v1/tags/{id}`;
        urlPath = urlPath.replace(`{${"id"}}`, encodeURIComponent(String(requestParameters["id"])));
        const response = yield this.request({
          path: urlPath,
          method: "DELETE",
          headers: headerParameters,
          query: queryParameters
        }, initOverrides);
        return new VoidApiResponse(response);
      });
    }
    apiV1TagsIdDelete(requestParameters, initOverrides) {
      return __awaiter6(this, undefined, undefined, function* () {
        yield this.apiV1TagsIdDeleteRaw(requestParameters, initOverrides);
      });
    }
    apiV1TagsIdGetRaw(requestParameters, initOverrides) {
      return __awaiter6(this, undefined, undefined, function* () {
        if (requestParameters["id"] == null) {
          throw new RequiredError("id", 'Required parameter "id" was null or undefined when calling apiV1TagsIdGet().');
        }
        const queryParameters = {};
        const headerParameters = {};
        let urlPath = `/api/v1/tags/{id}`;
        urlPath = urlPath.replace(`{${"id"}}`, encodeURIComponent(String(requestParameters["id"])));
        const response = yield this.request({
          path: urlPath,
          method: "GET",
          headers: headerParameters,
          query: queryParameters
        }, initOverrides);
        return new JSONApiResponse(response, (jsonValue) => ModelTagDTOFromJSON(jsonValue));
      });
    }
    apiV1TagsIdGet(requestParameters, initOverrides) {
      return __awaiter6(this, undefined, undefined, function* () {
        const response = yield this.apiV1TagsIdGetRaw(requestParameters, initOverrides);
        return yield response.value();
      });
    }
    apiV1TagsIdPutRaw(requestParameters, initOverrides) {
      return __awaiter6(this, undefined, undefined, function* () {
        if (requestParameters["id"] == null) {
          throw new RequiredError("id", 'Required parameter "id" was null or undefined when calling apiV1TagsIdPut().');
        }
        if (requestParameters["tag"] == null) {
          throw new RequiredError("tag", 'Required parameter "tag" was null or undefined when calling apiV1TagsIdPut().');
        }
        const queryParameters = {};
        const headerParameters = {};
        headerParameters["Content-Type"] = "application/json";
        let urlPath = `/api/v1/tags/{id}`;
        urlPath = urlPath.replace(`{${"id"}}`, encodeURIComponent(String(requestParameters["id"])));
        const response = yield this.request({
          path: urlPath,
          method: "PUT",
          headers: headerParameters,
          query: queryParameters,
          body: ModelTagDTOToJSON(requestParameters["tag"])
        }, initOverrides);
        return new JSONApiResponse(response, (jsonValue) => ModelTagDTOFromJSON(jsonValue));
      });
    }
    apiV1TagsIdPut(requestParameters, initOverrides) {
      return __awaiter6(this, undefined, undefined, function* () {
        const response = yield this.apiV1TagsIdPutRaw(requestParameters, initOverrides);
        return yield response.value();
      });
    }
    apiV1TagsPostRaw(requestParameters, initOverrides) {
      return __awaiter6(this, undefined, undefined, function* () {
        if (requestParameters["tag"] == null) {
          throw new RequiredError("tag", 'Required parameter "tag" was null or undefined when calling apiV1TagsPost().');
        }
        const queryParameters = {};
        const headerParameters = {};
        headerParameters["Content-Type"] = "application/json";
        let urlPath = `/api/v1/tags`;
        const response = yield this.request({
          path: urlPath,
          method: "POST",
          headers: headerParameters,
          query: queryParameters,
          body: ModelTagDTOToJSON(requestParameters["tag"])
        }, initOverrides);
        return new JSONApiResponse(response, (jsonValue) => ModelTagDTOFromJSON(jsonValue));
      });
    }
    apiV1TagsPost(requestParameters, initOverrides) {
      return __awaiter6(this, undefined, undefined, function* () {
        const response = yield this.apiV1TagsPostRaw(requestParameters, initOverrides);
        return yield response.value();
      });
    }
  }
  // wrapper.js
  window.ShioriAPI = exports_esm;
})();
