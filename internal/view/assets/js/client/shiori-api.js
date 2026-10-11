(() => {
	// src/runtime.ts
	var BASE_PATH = "http://localhost".replace(/\/+$/, "");

	class Configuration {
		configuration;
		constructor(configuration = {}) {
			this.configuration = configuration;
		}
		set config(configuration) {
			this.configuration = configuration;
		}
		get basePath() {
			return this.configuration.basePath != null
				? this.configuration.basePath
				: BASE_PATH;
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
				return typeof accessToken === "function"
					? accessToken
					: async () => accessToken;
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
	var DefaultConfig = new Configuration();

	class BaseAPI {
		configuration;
		static jsonRegex = new RegExp(
			"^(:?application/json|[^;/ \t]+/[^;/ \t]+[+]json)[ \t]*(:?;.*)?$",
			"i",
		);
		middleware;
		constructor(configuration = DefaultConfig) {
			this.configuration = configuration;
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
		async request(context, initOverrides) {
			const { url, init } = await this.createFetchParams(
				context,
				initOverrides,
			);
			const response = await this.fetchApi(url, init);
			if (response && response.status >= 200 && response.status < 300) {
				return response;
			}
			throw new ResponseError(response, "Response returned an error code");
		}
		async createFetchParams(context, initOverrides) {
			let url = this.configuration.basePath + context.path;
			if (
				context.query !== undefined &&
				Object.keys(context.query).length !== 0
			) {
				url += "?" + this.configuration.queryParamsStringify(context.query);
			}
			const headers = Object.assign(
				{},
				this.configuration.headers,
				context.headers,
			);
			Object.keys(headers).forEach((key) =>
				headers[key] === undefined ? delete headers[key] : {},
			);
			const initOverrideFn =
				typeof initOverrides === "function"
					? initOverrides
					: async () => initOverrides;
			const initParams = {
				method: context.method,
				headers,
				body: context.body,
				credentials: this.configuration.credentials,
			};
			const overriddenInit = {
				...initParams,
				...(await initOverrideFn({
					init: initParams,
					context,
				})),
			};
			let body;
			if (
				isFormData(overriddenInit.body) ||
				overriddenInit.body instanceof URLSearchParams ||
				isBlob(overriddenInit.body)
			) {
				body = overriddenInit.body;
			} else if (this.isJsonMime(headers["Content-Type"])) {
				body = JSON.stringify(overriddenInit.body);
			} else {
				body = overriddenInit.body;
			}
			const init = {
				...overriddenInit,
				body,
			};
			return { url, init };
		}
		fetchApi = async (url, init) => {
			let fetchParams = { url, init };
			for (const middleware of this.middleware) {
				if (middleware.pre) {
					fetchParams =
						(await middleware.pre({
							fetch: this.fetchApi,
							...fetchParams,
						})) || fetchParams;
				}
			}
			let response = undefined;
			try {
				response = await (this.configuration.fetchApi || fetch)(
					fetchParams.url,
					fetchParams.init,
				);
			} catch (e) {
				for (const middleware of this.middleware) {
					if (middleware.onError) {
						response =
							(await middleware.onError({
								fetch: this.fetchApi,
								url: fetchParams.url,
								init: fetchParams.init,
								error: e,
								response: response ? response.clone() : undefined,
							})) || response;
					}
				}
				if (response === undefined) {
					if (e instanceof Error) {
						throw new FetchError(
							e,
							"The request failed and the interceptors did not return an alternative response",
						);
					} else {
						throw e;
					}
				}
			}
			for (const middleware of this.middleware) {
				if (middleware.post) {
					response =
						(await middleware.post({
							fetch: this.fetchApi,
							url: fetchParams.url,
							init: fetchParams.init,
							response: response.clone(),
						})) || response;
				}
			}
			return response;
		};
		clone() {
			const constructor = this.constructor;
			const next = new constructor(this.configuration);
			next.middleware = this.middleware.slice();
			return next;
		}
	}
	function isBlob(value) {
		return typeof Blob !== "undefined" && value instanceof Blob;
	}
	function isFormData(value) {
		return typeof FormData !== "undefined" && value instanceof FormData;
	}

	class ResponseError extends Error {
		response;
		name = "ResponseError";
		constructor(response, msg) {
			super(msg);
			this.response = response;
		}
	}

	class FetchError extends Error {
		cause;
		name = "FetchError";
		constructor(cause, msg) {
			super(msg);
			this.cause = cause;
		}
	}

	class RequiredError extends Error {
		field;
		name = "RequiredError";
		constructor(field, msg) {
			super(msg);
			this.field = field;
		}
	}
	function querystring(params, prefix = "") {
		return Object.keys(params)
			.map((key) => querystringSingleKey(key, params[key], prefix))
			.filter((part) => part.length > 0)
			.join("&");
	}
	function querystringSingleKey(key, value, keyPrefix = "") {
		const fullKey = keyPrefix + (keyPrefix.length ? `[${key}]` : key);
		if (value instanceof Array) {
			const multiValue = value
				.map((singleValue) => encodeURIComponent(String(singleValue)))
				.join(`&${encodeURIComponent(fullKey)}=`);
			return `${encodeURIComponent(fullKey)}=${multiValue}`;
		}
		if (value instanceof Set) {
			const valueAsArray = Array.from(value);
			return querystringSingleKey(key, valueAsArray, keyPrefix);
		}
		if (value instanceof Date) {
			return `${encodeURIComponent(fullKey)}=${encodeURIComponent(
				value.toISOString(),
			)}`;
		}
		if (value instanceof Object) {
			return querystring(value, fullKey);
		}
		return `${encodeURIComponent(fullKey)}=${encodeURIComponent(
			String(value),
		)}`;
	}
	class JSONApiResponse {
		raw;
		transformer;
		constructor(raw, transformer = (jsonValue) => jsonValue) {
			this.raw = raw;
			this.transformer = transformer;
		}
		async value() {
			return this.transformer(await this.raw.json());
		}
	}

	class VoidApiResponse {
		raw;
		constructor(raw) {
			this.raw = raw;
		}
		async value() {
			return;
		}
	}

	// src/models/ApiV1BookmarkDataResponse.ts
	function ApiV1BookmarkDataResponseFromJSON(json) {
		return ApiV1BookmarkDataResponseFromJSONTyped(json, false);
	}
	function ApiV1BookmarkDataResponseFromJSONTyped(json, ignoreDiscriminator) {
		if (json == null) {
			return json;
		}
		return {
			ArchiveURL: json["ArchiveURL"] == null ? undefined : json["ArchiveURL"],
			Content: json["Content"] == null ? undefined : json["Content"],
			EbookURL: json["EbookURL"] == null ? undefined : json["EbookURL"],
			HasArchive: json["HasArchive"] == null ? undefined : json["HasArchive"],
			HasContent: json["HasContent"] == null ? undefined : json["HasContent"],
			HasEbook: json["HasEbook"] == null ? undefined : json["HasEbook"],
			HTML: json["HTML"] == null ? undefined : json["HTML"],
			ImageURL: json["ImageURL"] == null ? undefined : json["ImageURL"],
		};
	}

	// src/models/ApiV1BookmarkTagPayload.ts
	function ApiV1BookmarkTagPayloadToJSON(json) {
		return ApiV1BookmarkTagPayloadToJSONTyped(json, false);
	}
	function ApiV1BookmarkTagPayloadToJSONTyped(
		value,
		ignoreDiscriminator = false,
	) {
		if (value == null) {
			return value;
		}
		return {
			TagID: value["TagID"],
		};
	}

	// src/models/ApiV1BulkUpdateBookmarkTagsPayload.ts
	function ApiV1BulkUpdateBookmarkTagsPayloadToJSON(json) {
		return ApiV1BulkUpdateBookmarkTagsPayloadToJSONTyped(json, false);
	}
	function ApiV1BulkUpdateBookmarkTagsPayloadToJSONTyped(
		value,
		ignoreDiscriminator = false,
	) {
		if (value == null) {
			return value;
		}
		return {
			BookmarkIDs: value["BookmarkIDs"],
			TagIDs: value["TagIDs"],
		};
	}

	// src/models/ApiV1CreateBookmarkPayload.ts
	function ApiV1CreateBookmarkPayloadToJSON(json) {
		return ApiV1CreateBookmarkPayloadToJSONTyped(json, false);
	}
	function ApiV1CreateBookmarkPayloadToJSONTyped(
		value,
		ignoreDiscriminator = false,
	) {
		if (value == null) {
			return value;
		}
		return {
			Excerpt: value["Excerpt"],
			Public: value["Public"],
			Options: value["Options"],
			Title: value["Title"],
			URL: value["URL"],
		};
	}

	// src/models/ApiV1DeleteBookmarksPayload.ts
	function ApiV1DeleteBookmarksPayloadToJSON(json) {
		return ApiV1DeleteBookmarksPayloadToJSONTyped(json, false);
	}
	function ApiV1DeleteBookmarksPayloadToJSONTyped(
		value,
		ignoreDiscriminator = false,
	) {
		if (value == null) {
			return value;
		}
		return {
			IDs: value["IDs"],
		};
	}

	// src/models/ApiV1InfoResponseVersion.ts
	function ApiV1InfoResponseVersionFromJSON(json) {
		return ApiV1InfoResponseVersionFromJSONTyped(json, false);
	}
	function ApiV1InfoResponseVersionFromJSONTyped(json, ignoreDiscriminator) {
		if (json == null) {
			return json;
		}
		return {
			Commit: json["Commit"] == null ? undefined : json["Commit"],
			Date: json["Date"] == null ? undefined : json["Date"],
			Tag: json["Tag"] == null ? undefined : json["Tag"],
		};
	}

	// src/models/ApiV1InfoResponse.ts
	function ApiV1InfoResponseFromJSON(json) {
		return ApiV1InfoResponseFromJSONTyped(json, false);
	}
	function ApiV1InfoResponseFromJSONTyped(json, ignoreDiscriminator) {
		if (json == null) {
			return json;
		}
		return {
			Database: json["Database"] == null ? undefined : json["Database"],
			OS: json["OS"] == null ? undefined : json["OS"],
			Version:
				json["Version"] == null
					? undefined
					: ApiV1InfoResponseVersionFromJSON(json["Version"]),
		};
	}

	// src/models/ApiV1LoginRequestPayload.ts
	function ApiV1LoginRequestPayloadToJSON(json) {
		return ApiV1LoginRequestPayloadToJSONTyped(json, false);
	}
	function ApiV1LoginRequestPayloadToJSONTyped(
		value,
		ignoreDiscriminator = false,
	) {
		if (value == null) {
			return value;
		}
		return {
			Password: value["Password"],
			RememberMe: value["RememberMe"],
			Username: value["Username"],
		};
	}

	// src/models/ApiV1LoginResponseMessage.ts
	function ApiV1LoginResponseMessageFromJSON(json) {
		return ApiV1LoginResponseMessageFromJSONTyped(json, false);
	}
	function ApiV1LoginResponseMessageFromJSONTyped(json, ignoreDiscriminator) {
		if (json == null) {
			return json;
		}
		return {
			Expiration: json["Expiration"] == null ? undefined : json["Expiration"],
			Token: json["Token"] == null ? undefined : json["Token"],
		};
	}

	// src/models/ApiV1ReadableResponseMessage.ts
	function ApiV1ReadableResponseMessageFromJSON(json) {
		return ApiV1ReadableResponseMessageFromJSONTyped(json, false);
	}
	function ApiV1ReadableResponseMessageFromJSONTyped(
		json,
		ignoreDiscriminator,
	) {
		if (json == null) {
			return json;
		}
		return {
			Content: json["Content"] == null ? undefined : json["Content"],
			HTML: json["HTML"] == null ? undefined : json["HTML"],
		};
	}

	// src/models/ModelUserConfig.ts
	function ModelUserConfigFromJSON(json) {
		return ModelUserConfigFromJSONTyped(json, false);
	}
	function ModelUserConfigFromJSONTyped(json, ignoreDiscriminator) {
		if (json == null) {
			return json;
		}
		return {
			CreateEbook:
				json["CreateEbook"] == null ? undefined : json["CreateEbook"],
			HideExcerpt:
				json["HideExcerpt"] == null ? undefined : json["HideExcerpt"],
			HideThumbnail:
				json["HideThumbnail"] == null ? undefined : json["HideThumbnail"],
			KeepMetadata:
				json["KeepMetadata"] == null ? undefined : json["KeepMetadata"],
			ListMode: json["ListMode"] == null ? undefined : json["ListMode"],
			MakePublic: json["MakePublic"] == null ? undefined : json["MakePublic"],
			ShowId: json["ShowId"] == null ? undefined : json["ShowId"],
			Theme: json["Theme"] == null ? undefined : json["Theme"],
			UseArchive: json["UseArchive"] == null ? undefined : json["UseArchive"],
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
			CreateEbook: value["CreateEbook"],
			HideExcerpt: value["HideExcerpt"],
			HideThumbnail: value["HideThumbnail"],
			KeepMetadata: value["KeepMetadata"],
			ListMode: value["ListMode"],
			MakePublic: value["MakePublic"],
			ShowId: value["ShowId"],
			Theme: value["Theme"],
			UseArchive: value["UseArchive"],
		};
	}

	// src/models/ApiV1UpdateAccountPayload.ts
	function ApiV1UpdateAccountPayloadToJSON(json) {
		return ApiV1UpdateAccountPayloadToJSONTyped(json, false);
	}
	function ApiV1UpdateAccountPayloadToJSONTyped(
		value,
		ignoreDiscriminator = false,
	) {
		if (value == null) {
			return value;
		}
		return {
			Config: ModelUserConfigToJSON(value["Config"]),
			NewPassword: value["NewPassword"],
			OldPassword: value["OldPassword"],
			Owner: value["Owner"],
			Username: value["Username"],
		};
	}

	// src/models/ApiV1UpdateBookmarkDataPayload.ts
	function ApiV1UpdateBookmarkDataPayloadToJSON(json) {
		return ApiV1UpdateBookmarkDataPayloadToJSONTyped(json, false);
	}
	function ApiV1UpdateBookmarkDataPayloadToJSONTyped(
		value,
		ignoreDiscriminator = false,
	) {
		if (value == null) {
			return value;
		}
		return {
			CreateArchive: value["CreateArchive"],
			CreateEbook: value["CreateEbook"],
			KeepMetadata: value["KeepMetadata"],
			SkipExisting: value["SkipExisting"],
			UpdateReadable: value["UpdateReadable"],
		};
	}

	// src/models/ApiV1UpdateBookmarkPayload.ts
	function ApiV1UpdateBookmarkPayloadToJSON(json) {
		return ApiV1UpdateBookmarkPayloadToJSONTyped(json, false);
	}
	function ApiV1UpdateBookmarkPayloadToJSONTyped(
		value,
		ignoreDiscriminator = false,
	) {
		if (value == null) {
			return value;
		}
		return {
			Excerpt: value["Excerpt"],
			Public: value["Public"],
			Title: value["Title"],
			URL: value["URL"],
		};
	}

	// src/models/ApiV1UpdateCachePayload.ts
	function ApiV1UpdateCachePayloadToJSON(json) {
		return ApiV1UpdateCachePayloadToJSONTyped(json, false);
	}
	function ApiV1UpdateCachePayloadToJSONTyped(
		value,
		ignoreDiscriminator = false,
	) {
		if (value == null) {
			return value;
		}
		return {
			CreateArchive: value["CreateArchive"],
			CreateEbook: value["CreateEbook"],
			IDs: value["IDs"],
			KeepMetadata: value["KeepMetadata"],
			SkipExist: value["SkipExist"],
		};
	}

	// src/models/ModelAccount.ts
	function ModelAccountFromJSON(json) {
		return ModelAccountFromJSONTyped(json, false);
	}
	function ModelAccountFromJSONTyped(json, ignoreDiscriminator) {
		if (json == null) {
			return json;
		}
		return {
			Config:
				json["Config"] == null
					? undefined
					: ModelUserConfigFromJSON(json["Config"]),
			ID: json["ID"] == null ? undefined : json["ID"],
			Owner: json["Owner"] == null ? undefined : json["Owner"],
			Password: json["Password"] == null ? undefined : json["Password"],
			Username: json["Username"] == null ? undefined : json["Username"],
		};
	}

	// src/models/ModelAccountDTO.ts
	function ModelAccountDTOFromJSON(json) {
		return ModelAccountDTOFromJSONTyped(json, false);
	}
	function ModelAccountDTOFromJSONTyped(json, ignoreDiscriminator) {
		if (json == null) {
			return json;
		}
		return {
			Config:
				json["Config"] == null
					? undefined
					: ModelUserConfigFromJSON(json["Config"]),
			ID: json["ID"] == null ? undefined : json["ID"],
			Owner: json["Owner"] == null ? undefined : json["Owner"],
			Password: json["Password"] == null ? undefined : json["Password"],
			Username: json["Username"] == null ? undefined : json["Username"],
		};
	}

	// src/models/ModelTagDTO.ts
	function ModelTagDTOFromJSON(json) {
		return ModelTagDTOFromJSONTyped(json, false);
	}
	function ModelTagDTOFromJSONTyped(json, ignoreDiscriminator) {
		if (json == null) {
			return json;
		}
		return {
			BookmarkCount:
				json["BookmarkCount"] == null ? undefined : json["BookmarkCount"],
			Deleted: json["Deleted"] == null ? undefined : json["Deleted"],
			ID: json["ID"] == null ? undefined : json["ID"],
			Name: json["Name"] == null ? undefined : json["Name"],
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
			BookmarkCount: value["BookmarkCount"],
			Deleted: value["Deleted"],
			ID: value["ID"],
			Name: value["Name"],
		};
	}

	// src/models/ModelBookmarkDTO.ts
	function ModelBookmarkDTOFromJSON(json) {
		return ModelBookmarkDTOFromJSONTyped(json, false);
	}
	function ModelBookmarkDTOFromJSONTyped(json, ignoreDiscriminator) {
		if (json == null) {
			return json;
		}
		return {
			Author: json["Author"] == null ? undefined : json["Author"],
			CreateArchive:
				json["CreateArchive"] == null ? undefined : json["CreateArchive"],
			CreateEbook:
				json["CreateEbook"] == null ? undefined : json["CreateEbook"],
			CreatedAt: json["CreatedAt"] == null ? undefined : json["CreatedAt"],
			Excerpt: json["Excerpt"] == null ? undefined : json["Excerpt"],
			HasArchive: json["HasArchive"] == null ? undefined : json["HasArchive"],
			HasContent: json["HasContent"] == null ? undefined : json["HasContent"],
			HasEbook: json["HasEbook"] == null ? undefined : json["HasEbook"],
			HTML: json["HTML"] == null ? undefined : json["HTML"],
			ID: json["ID"] == null ? undefined : json["ID"],
			ImageURL: json["ImageURL"] == null ? undefined : json["ImageURL"],
			ModifiedAt: json["ModifiedAt"] == null ? undefined : json["ModifiedAt"],
			Public: json["Public"] == null ? undefined : json["Public"],
			Tags:
				json["Tags"] == null
					? undefined
					: json["Tags"].map(ModelTagDTOFromJSON),
			Title: json["Title"] == null ? undefined : json["Title"],
			URL: json["URL"] == null ? undefined : json["URL"],
		};
	}

	// src/apis/AccountsApi.ts
	class AccountsApi extends BaseAPI {
		async apiV1AccountsGetRaw(initOverrides) {
			const queryParameters = {};
			const headerParameters = {};
			let urlPath = `/api/v1/accounts`;
			const response = await this.request(
				{
					path: urlPath,
					method: "GET",
					headers: headerParameters,
					query: queryParameters,
				},
				initOverrides,
			);
			return new JSONApiResponse(response, (jsonValue) =>
				jsonValue.map(ModelAccountDTOFromJSON),
			);
		}
		async apiV1AccountsGet(initOverrides) {
			const response = await this.apiV1AccountsGetRaw(initOverrides);
			return await response.value();
		}
		async apiV1AccountsIdDeleteRaw(requestParameters, initOverrides) {
			if (requestParameters["id"] == null) {
				throw new RequiredError(
					"id",
					'Required parameter "id" was null or undefined when calling apiV1AccountsIdDelete().',
				);
			}
			const queryParameters = {};
			const headerParameters = {};
			let urlPath = `/api/v1/accounts/{id}`;
			urlPath = urlPath.replace(
				`{${"id"}}`,
				encodeURIComponent(String(requestParameters["id"])),
			);
			const response = await this.request(
				{
					path: urlPath,
					method: "DELETE",
					headers: headerParameters,
					query: queryParameters,
				},
				initOverrides,
			);
			return new VoidApiResponse(response);
		}
		async apiV1AccountsIdDelete(requestParameters, initOverrides) {
			await this.apiV1AccountsIdDeleteRaw(requestParameters, initOverrides);
		}
		async apiV1AccountsIdPatchRaw(requestParameters, initOverrides) {
			if (requestParameters["id"] == null) {
				throw new RequiredError(
					"id",
					'Required parameter "id" was null or undefined when calling apiV1AccountsIdPatch().',
				);
			}
			if (requestParameters["account"] == null) {
				throw new RequiredError(
					"account",
					'Required parameter "account" was null or undefined when calling apiV1AccountsIdPatch().',
				);
			}
			const queryParameters = {};
			const headerParameters = {};
			headerParameters["Content-Type"] = "application/json";
			let urlPath = `/api/v1/accounts/{id}`;
			urlPath = urlPath.replace(
				`{${"id"}}`,
				encodeURIComponent(String(requestParameters["id"])),
			);
			const response = await this.request(
				{
					path: urlPath,
					method: "PATCH",
					headers: headerParameters,
					query: queryParameters,
					body: ApiV1UpdateAccountPayloadToJSON(requestParameters["account"]),
				},
				initOverrides,
			);
			return new JSONApiResponse(response, (jsonValue) =>
				ModelAccountDTOFromJSON(jsonValue),
			);
		}
		async apiV1AccountsIdPatch(requestParameters, initOverrides) {
			const response = await this.apiV1AccountsIdPatchRaw(
				requestParameters,
				initOverrides,
			);
			return await response.value();
		}
		async apiV1AccountsPostRaw(requestParameters, initOverrides) {
			if (requestParameters["payload"] == null) {
				throw new RequiredError(
					"payload",
					'Required parameter "payload" was null or undefined when calling apiV1AccountsPost().',
				);
			}
			const queryParameters = {};
			const headerParameters = {};
			headerParameters["Content-Type"] = "application/json";
			let urlPath = `/api/v1/accounts`;
			const response = await this.request(
				{
					path: urlPath,
					method: "POST",
					headers: headerParameters,
					query: queryParameters,
					body: requestParameters["payload"],
				},
				initOverrides,
			);
			return new JSONApiResponse(response, (jsonValue) =>
				ModelAccountDTOFromJSON(jsonValue),
			);
		}
		async apiV1AccountsPost(requestParameters, initOverrides) {
			const response = await this.apiV1AccountsPostRaw(requestParameters, initOverrides);
			return await response.value();
		}
	}

	// src/apis/AuthApi.ts
	class AuthApi extends BaseAPI {
		async apiV1AuthAccountPatchRaw(requestParameters, initOverrides) {
			const queryParameters = {};
			const headerParameters = {};
			headerParameters["Content-Type"] = "application/json";
			let urlPath = `/api/v1/auth/account`;
			const response = await this.request(
				{
					path: urlPath,
					method: "PATCH",
					headers: headerParameters,
					query: queryParameters,
					body: ApiV1UpdateAccountPayloadToJSON(requestParameters["payload"]),
				},
				initOverrides,
			);
			return new JSONApiResponse(response, (jsonValue) =>
				ModelAccountFromJSON(jsonValue),
			);
		}
		async apiV1AuthAccountPatch(requestParameters = {}, initOverrides) {
			const response = await this.apiV1AuthAccountPatchRaw(
				requestParameters,
				initOverrides,
			);
			return await response.value();
		}
		async apiV1AuthLoginPostRaw(requestParameters, initOverrides) {
			const queryParameters = {};
			const headerParameters = {};
			headerParameters["Content-Type"] = "application/json";
			let urlPath = `/api/v1/auth/login`;
			const response = await this.request(
				{
					path: urlPath,
					method: "POST",
					headers: headerParameters,
					query: queryParameters,
					body: ApiV1LoginRequestPayloadToJSON(requestParameters["payload"]),
				},
				initOverrides,
			);
			return new JSONApiResponse(response, (jsonValue) =>
				ApiV1LoginResponseMessageFromJSON(jsonValue),
			);
		}
		async apiV1AuthLoginPost(requestParameters = {}, initOverrides) {
			const response = await this.apiV1AuthLoginPostRaw(
				requestParameters,
				initOverrides,
			);
			return await response.value();
		}
		async apiV1AuthLogoutPostRaw(initOverrides) {
			const queryParameters = {};
			const headerParameters = {};
			let urlPath = `/api/v1/auth/logout`;
			const response = await this.request(
				{
					path: urlPath,
					method: "POST",
					headers: headerParameters,
					query: queryParameters,
				},
				initOverrides,
			);
			return new VoidApiResponse(response);
		}
		async apiV1AuthLogoutPost(initOverrides) {
			await this.apiV1AuthLogoutPostRaw(initOverrides);
		}
		async apiV1AuthMeGetRaw(initOverrides) {
			const queryParameters = {};
			const headerParameters = {};
			let urlPath = `/api/v1/auth/me`;
			const response = await this.request(
				{
					path: urlPath,
					method: "GET",
					headers: headerParameters,
					query: queryParameters,
				},
				initOverrides,
			);
			return new JSONApiResponse(response, (jsonValue) =>
				ModelAccountFromJSON(jsonValue),
			);
		}
		async apiV1AuthMeGet(initOverrides) {
			const response = await this.apiV1AuthMeGetRaw(initOverrides);
			return await response.value();
		}
		async apiV1AuthRefreshPostRaw(initOverrides) {
			const queryParameters = {};
			const headerParameters = {};
			let urlPath = `/api/v1/auth/refresh`;
			const response = await this.request(
				{
					path: urlPath,
					method: "POST",
					headers: headerParameters,
					query: queryParameters,
				},
				initOverrides,
			);
			return new JSONApiResponse(response, (jsonValue) =>
				ApiV1LoginResponseMessageFromJSON(jsonValue),
			);
		}
		async apiV1AuthRefreshPost(initOverrides) {
			const response = await this.apiV1AuthRefreshPostRaw(initOverrides);
			return await response.value();
		}
	}

	// src/apis/BookmarksApi.ts
	class BookmarksApi extends BaseAPI {
		async apiV1BookmarksBulkTagsPutRaw(requestParameters, initOverrides) {
			if (requestParameters["payload"] == null) {
				throw new RequiredError(
					"payload",
					'Required parameter "payload" was null or undefined when calling apiV1BookmarksBulkTagsPut().',
				);
			}
			const queryParameters = {};
			const headerParameters = {};
			headerParameters["Content-Type"] = "application/json";
			let urlPath = `/api/v1/bookmarks/bulk/tags`;
			const response = await this.request(
				{
					path: urlPath,
					method: "PUT",
					headers: headerParameters,
					query: queryParameters,
					body: ApiV1BulkUpdateBookmarkTagsPayloadToJSON(
						requestParameters["payload"],
					),
				},
				initOverrides,
			);
			return new JSONApiResponse(response, (jsonValue) =>
				jsonValue.map(ModelBookmarkDTOFromJSON),
			);
		}
		async apiV1BookmarksBulkTagsPut(requestParameters, initOverrides) {
			const response = await this.apiV1BookmarksBulkTagsPutRaw(
				requestParameters,
				initOverrides,
			);
			return await response.value();
		}
		async apiV1BookmarksCachePutRaw(requestParameters, initOverrides) {
			if (requestParameters["payload"] == null) {
				throw new RequiredError(
					"payload",
					'Required parameter "payload" was null or undefined when calling apiV1BookmarksCachePut().',
				);
			}
			const queryParameters = {};
			const headerParameters = {};
			headerParameters["Content-Type"] = "application/json";
			let urlPath = `/api/v1/bookmarks/cache`;
			const response = await this.request(
				{
					path: urlPath,
					method: "PUT",
					headers: headerParameters,
					query: queryParameters,
					body: ApiV1UpdateCachePayloadToJSON(requestParameters["payload"]),
				},
				initOverrides,
			);
			return new JSONApiResponse(response, (jsonValue) =>
				jsonValue.map(ModelBookmarkDTOFromJSON),
			);
		}
		async apiV1BookmarksCachePut(requestParameters, initOverrides) {
			const response = await this.apiV1BookmarksCachePutRaw(
				requestParameters,
				initOverrides,
			);
			return await response.value();
		}
		async apiV1BookmarksDeleteRaw(requestParameters, initOverrides) {
			if (requestParameters["payload"] == null) {
				throw new RequiredError(
					"payload",
					'Required parameter "payload" was null or undefined when calling apiV1BookmarksDelete().',
				);
			}
			const queryParameters = {};
			const headerParameters = {};
			headerParameters["Content-Type"] = "application/json";
			let urlPath = `/api/v1/bookmarks`;
			const response = await this.request(
				{
					path: urlPath,
					method: "DELETE",
					headers: headerParameters,
					query: queryParameters,
					body: ApiV1DeleteBookmarksPayloadToJSON(requestParameters["payload"]),
				},
				initOverrides,
			);
			return new VoidApiResponse(response);
		}
		async apiV1BookmarksDelete(requestParameters, initOverrides) {
			await this.apiV1BookmarksDeleteRaw(requestParameters, initOverrides);
		}
		async apiV1BookmarksGetRaw(requestParameters, initOverrides) {
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
			const response = await this.request(
				{
					path: urlPath,
					method: "GET",
					headers: headerParameters,
					query: queryParameters,
				},
				initOverrides,
			);
			return new JSONApiResponse(response, (jsonValue) => ({
				items: (jsonValue.Items || []).map(ModelBookmarkDTOFromJSON),
				total: jsonValue.Total || 0,
			}));
		}
		async apiV1BookmarksGet(requestParameters = {}, initOverrides) {
			const response = await this.apiV1BookmarksGetRaw(
				requestParameters,
				initOverrides,
			);
			return await response.value();
		}
		async apiV1BookmarksIdDataGetRaw(requestParameters, initOverrides) {
			if (requestParameters["id"] == null) {
				throw new RequiredError(
					"id",
					'Required parameter "id" was null or undefined when calling apiV1BookmarksIdDataGet().',
				);
			}
			const queryParameters = {};
			const headerParameters = {};
			let urlPath = `/api/v1/bookmarks/{id}/data`;
			urlPath = urlPath.replace(
				`{${"id"}}`,
				encodeURIComponent(String(requestParameters["id"])),
			);
			const response = await this.request(
				{
					path: urlPath,
					method: "GET",
					headers: headerParameters,
					query: queryParameters,
				},
				initOverrides,
			);
			return new JSONApiResponse(response, (jsonValue) =>
				ApiV1BookmarkDataResponseFromJSON(jsonValue),
			);
		}
		async apiV1BookmarksIdDataGet(requestParameters, initOverrides) {
			const response = await this.apiV1BookmarksIdDataGetRaw(
				requestParameters,
				initOverrides,
			);
			return await response.value();
		}
		async apiV1BookmarksIdDataPutRaw(requestParameters, initOverrides) {
			if (requestParameters["id"] == null) {
				throw new RequiredError(
					"id",
					'Required parameter "id" was null or undefined when calling apiV1BookmarksIdDataPut().',
				);
			}
			if (requestParameters["payload"] == null) {
				throw new RequiredError(
					"payload",
					'Required parameter "payload" was null or undefined when calling apiV1BookmarksIdDataPut().',
				);
			}
			const queryParameters = {};
			const headerParameters = {};
			headerParameters["Content-Type"] = "application/json";
			let urlPath = `/api/v1/bookmarks/{id}/data`;
			urlPath = urlPath.replace(
				`{${"id"}}`,
				encodeURIComponent(String(requestParameters["id"])),
			);
			const response = await this.request(
				{
					path: urlPath,
					method: "PUT",
					headers: headerParameters,
					query: queryParameters,
					body: ApiV1UpdateBookmarkDataPayloadToJSON(
						requestParameters["payload"],
					),
				},
				initOverrides,
			);
			return new JSONApiResponse(response, (jsonValue) =>
				ApiV1BookmarkDataResponseFromJSON(jsonValue),
			);
		}
		async apiV1BookmarksIdDataPut(requestParameters, initOverrides) {
			const response = await this.apiV1BookmarksIdDataPutRaw(
				requestParameters,
				initOverrides,
			);
			return await response.value();
		}
		async apiV1BookmarksIdGetRaw(requestParameters, initOverrides) {
			if (requestParameters["id"] == null) {
				throw new RequiredError(
					"id",
					'Required parameter "id" was null or undefined when calling apiV1BookmarksIdGet().',
				);
			}
			const queryParameters = {};
			const headerParameters = {};
			let urlPath = `/api/v1/bookmarks/{id}`;
			urlPath = urlPath.replace(
				`{${"id"}}`,
				encodeURIComponent(String(requestParameters["id"])),
			);
			const response = await this.request(
				{
					path: urlPath,
					method: "GET",
					headers: headerParameters,
					query: queryParameters,
				},
				initOverrides,
			);
			return new JSONApiResponse(response, (jsonValue) =>
				ModelBookmarkDTOFromJSON(jsonValue),
			);
		}
		async apiV1BookmarksIdGet(requestParameters, initOverrides) {
			const response = await this.apiV1BookmarksIdGetRaw(
				requestParameters,
				initOverrides,
			);
			return await response.value();
		}
		async apiV1BookmarksIdPutRaw(requestParameters, initOverrides) {
			if (requestParameters["id"] == null) {
				throw new RequiredError(
					"id",
					'Required parameter "id" was null or undefined when calling apiV1BookmarksIdPut().',
				);
			}
			if (requestParameters["payload"] == null) {
				throw new RequiredError(
					"payload",
					'Required parameter "payload" was null or undefined when calling apiV1BookmarksIdPut().',
				);
			}
			const queryParameters = {};
			const headerParameters = {};
			headerParameters["Content-Type"] = "application/json";
			let urlPath = `/api/v1/bookmarks/{id}`;
			urlPath = urlPath.replace(
				`{${"id"}}`,
				encodeURIComponent(String(requestParameters["id"])),
			);
			const response = await this.request(
				{
					path: urlPath,
					method: "PUT",
					headers: headerParameters,
					query: queryParameters,
					body: ApiV1UpdateBookmarkPayloadToJSON(requestParameters["payload"]),
				},
				initOverrides,
			);
			return new JSONApiResponse(response, (jsonValue) =>
				ModelBookmarkDTOFromJSON(jsonValue),
			);
		}
		async apiV1BookmarksIdPut(requestParameters, initOverrides) {
			const response = await this.apiV1BookmarksIdPutRaw(
				requestParameters,
				initOverrides,
			);
			return await response.value();
		}
		async apiV1BookmarksIdReadableGetRaw(requestParameters, initOverrides) {
			if (requestParameters["id"] == null) {
				throw new RequiredError(
					"id",
					'Required parameter "id" was null or undefined when calling apiV1BookmarksIdReadableGet().',
				);
			}
			const queryParameters = {};
			const headerParameters = {};
			let urlPath = `/api/v1/bookmarks/{id}/readable`;
			urlPath = urlPath.replace(
				`{${"id"}}`,
				encodeURIComponent(String(requestParameters["id"])),
			);
			const response = await this.request(
				{
					path: urlPath,
					method: "GET",
					headers: headerParameters,
					query: queryParameters,
				},
				initOverrides,
			);
			return new JSONApiResponse(response, (jsonValue) =>
				ApiV1ReadableResponseMessageFromJSON(jsonValue),
			);
		}
		async apiV1BookmarksIdReadableGet(requestParameters, initOverrides) {
			const response = await this.apiV1BookmarksIdReadableGetRaw(requestParameters, initOverrides);
			return await response.value();
		}
		async apiV1BookmarksIdTagsDeleteRaw(requestParameters, initOverrides) {
			if (requestParameters["id"] == null) {
				throw new RequiredError(
					"id",
					'Required parameter "id" was null or undefined when calling apiV1BookmarksIdTagsDelete().',
				);
			}
			if (requestParameters["payload"] == null) {
				throw new RequiredError(
					"payload",
					'Required parameter "payload" was null or undefined when calling apiV1BookmarksIdTagsDelete().',
				);
			}
			const queryParameters = {};
			const headerParameters = {};
			headerParameters["Content-Type"] = "application/json";
			let urlPath = `/api/v1/bookmarks/{id}/tags`;
			urlPath = urlPath.replace(
				`{${"id"}}`,
				encodeURIComponent(String(requestParameters["id"])),
			);
			const response = await this.request(
				{
					path: urlPath,
					method: "DELETE",
					headers: headerParameters,
					query: queryParameters,
					body: ApiV1BookmarkTagPayloadToJSON(requestParameters["payload"]),
				},
				initOverrides,
			);
			return new VoidApiResponse(response);
		}
		async apiV1BookmarksIdTagsDelete(requestParameters, initOverrides) {
			await this.apiV1BookmarksIdTagsDeleteRaw(
				requestParameters,
				initOverrides,
			);
		}
		async apiV1BookmarksIdTagsGetRaw(requestParameters, initOverrides) {
			if (requestParameters["id"] == null) {
				throw new RequiredError(
					"id",
					'Required parameter "id" was null or undefined when calling apiV1BookmarksIdTagsGet().',
				);
			}
			const queryParameters = {};
			const headerParameters = {};
			let urlPath = `/api/v1/bookmarks/{id}/tags`;
			urlPath = urlPath.replace(
				`{${"id"}}`,
				encodeURIComponent(String(requestParameters["id"])),
			);
			const response = await this.request(
				{
					path: urlPath,
					method: "GET",
					headers: headerParameters,
					query: queryParameters,
				},
				initOverrides,
			);
			return new JSONApiResponse(response, (jsonValue) =>
				jsonValue.map(ModelTagDTOFromJSON),
			);
		}
		async apiV1BookmarksIdTagsGet(requestParameters, initOverrides) {
			const response = await this.apiV1BookmarksIdTagsGetRaw(
				requestParameters,
				initOverrides,
			);
			return await response.value();
		}
		async apiV1BookmarksIdTagsPostRaw(requestParameters, initOverrides) {
			if (requestParameters["id"] == null) {
				throw new RequiredError(
					"id",
					'Required parameter "id" was null or undefined when calling apiV1BookmarksIdTagsPost().',
				);
			}
			if (requestParameters["payload"] == null) {
				throw new RequiredError(
					"payload",
					'Required parameter "payload" was null or undefined when calling apiV1BookmarksIdTagsPost().',
				);
			}
			const queryParameters = {};
			const headerParameters = {};
			headerParameters["Content-Type"] = "application/json";
			let urlPath = `/api/v1/bookmarks/{id}/tags`;
			urlPath = urlPath.replace(
				`{${"id"}}`,
				encodeURIComponent(String(requestParameters["id"])),
			);
			const response = await this.request(
				{
					path: urlPath,
					method: "POST",
					headers: headerParameters,
					query: queryParameters,
					body: ApiV1BookmarkTagPayloadToJSON(requestParameters["payload"]),
				},
				initOverrides,
			);
			return new VoidApiResponse(response);
		}
		async apiV1BookmarksIdTagsPost(requestParameters, initOverrides) {
			await this.apiV1BookmarksIdTagsPostRaw(requestParameters, initOverrides);
		}
		async apiV1BookmarksPostRaw(requestParameters, initOverrides) {
			if (requestParameters["payload"] == null) {
				throw new RequiredError(
					"payload",
					'Required parameter "payload" was null or undefined when calling apiV1BookmarksPost().',
				);
			}
			const queryParameters = {};
			const headerParameters = {};
			headerParameters["Content-Type"] = "application/json";
			let urlPath = `/api/v1/bookmarks`;
			const response = await this.request(
				{
					path: urlPath,
					method: "POST",
					headers: headerParameters,
					query: queryParameters,
					body: ApiV1CreateBookmarkPayloadToJSON(requestParameters["payload"]),
				},
				initOverrides,
			);
			return new JSONApiResponse(response, (jsonValue) =>
				ModelBookmarkDTOFromJSON(jsonValue),
			);
		}
		async apiV1BookmarksPost(requestParameters, initOverrides) {
			const response = await this.apiV1BookmarksPostRaw(
				requestParameters,
				initOverrides,
			);
			return await response.value();
		}
	}

	// src/apis/SystemApi.ts
	class SystemApi extends BaseAPI {
		async apiV1SystemInfoGetRaw(initOverrides) {
			const queryParameters = {};
			const headerParameters = {};
			let urlPath = `/api/v1/system/info`;
			const response = await this.request(
				{
					path: urlPath,
					method: "GET",
					headers: headerParameters,
					query: queryParameters,
				},
				initOverrides,
			);
			return new JSONApiResponse(response, (jsonValue) =>
				ApiV1InfoResponseFromJSON(jsonValue),
			);
		}
		async apiV1SystemInfoGet(initOverrides) {
			const response = await this.apiV1SystemInfoGetRaw(initOverrides);
			return await response.value();
		}
	}

	// src/apis/TagsApi.ts
	class TagsApi extends BaseAPI {
		async apiV1TagsGetRaw(requestParameters, initOverrides) {
			const queryParameters = {};
			if (requestParameters["withBookmarkCount"] != null) {
				queryParameters["with_bookmark_count"] =
					requestParameters["withBookmarkCount"];
			}
			if (requestParameters["bookmarkId"] != null) {
				queryParameters["bookmark_id"] = requestParameters["bookmarkId"];
			}
			if (requestParameters["search"] != null) {
				queryParameters["search"] = requestParameters["search"];
			}
			if (requestParameters["page"] != null) {
				queryParameters["page"] = requestParameters["page"];
			}
			if (requestParameters["limit"] != null) {
				queryParameters["limit"] = requestParameters["limit"];
			}
			const headerParameters = {};
			let urlPath = `/api/v1/tags`;
			const response = await this.request(
				{
					path: urlPath,
					method: "GET",
					headers: headerParameters,
					query: queryParameters,
				},
				initOverrides,
			);
			return new JSONApiResponse(response, (jsonValue) => ({
				items: (jsonValue.Items || []).map(ModelTagDTOFromJSON),
				total: jsonValue.Total || 0,
			}));
		}
		async apiV1TagsGet(requestParameters = {}, initOverrides) {
			const response = await this.apiV1TagsGetRaw(
				requestParameters,
				initOverrides,
			);
			return await response.value();
		}
		async apiV1TagsIdDeleteRaw(requestParameters, initOverrides) {
			if (requestParameters["id"] == null) {
				throw new RequiredError(
					"id",
					'Required parameter "id" was null or undefined when calling apiV1TagsIdDelete().',
				);
			}
			const queryParameters = {};
			const headerParameters = {};
			let urlPath = `/api/v1/tags/{id}`;
			urlPath = urlPath.replace(
				`{${"id"}}`,
				encodeURIComponent(String(requestParameters["id"])),
			);
			const response = await this.request(
				{
					path: urlPath,
					method: "DELETE",
					headers: headerParameters,
					query: queryParameters,
				},
				initOverrides,
			);
			return new VoidApiResponse(response);
		}
		async apiV1TagsIdDelete(requestParameters, initOverrides) {
			await this.apiV1TagsIdDeleteRaw(requestParameters, initOverrides);
		}
		async apiV1TagsIdGetRaw(requestParameters, initOverrides) {
			if (requestParameters["id"] == null) {
				throw new RequiredError(
					"id",
					'Required parameter "id" was null or undefined when calling apiV1TagsIdGet().',
				);
			}
			const queryParameters = {};
			const headerParameters = {};
			let urlPath = `/api/v1/tags/{id}`;
			urlPath = urlPath.replace(
				`{${"id"}}`,
				encodeURIComponent(String(requestParameters["id"])),
			);
			const response = await this.request(
				{
					path: urlPath,
					method: "GET",
					headers: headerParameters,
					query: queryParameters,
				},
				initOverrides,
			);
			return new JSONApiResponse(response, (jsonValue) =>
				ModelTagDTOFromJSON(jsonValue),
			);
		}
		async apiV1TagsIdGet(requestParameters, initOverrides) {
			const response = await this.apiV1TagsIdGetRaw(
				requestParameters,
				initOverrides,
			);
			return await response.value();
		}
		async apiV1TagsIdPutRaw(requestParameters, initOverrides) {
			if (requestParameters["id"] == null) {
				throw new RequiredError(
					"id",
					'Required parameter "id" was null or undefined when calling apiV1TagsIdPut().',
				);
			}
			if (requestParameters["tag"] == null) {
				throw new RequiredError(
					"tag",
					'Required parameter "tag" was null or undefined when calling apiV1TagsIdPut().',
				);
			}
			const queryParameters = {};
			const headerParameters = {};
			headerParameters["Content-Type"] = "application/json";
			let urlPath = `/api/v1/tags/{id}`;
			urlPath = urlPath.replace(
				`{${"id"}}`,
				encodeURIComponent(String(requestParameters["id"])),
			);
			const response = await this.request(
				{
					path: urlPath,
					method: "PUT",
					headers: headerParameters,
					query: queryParameters,
					body: ModelTagDTOToJSON(requestParameters["tag"]),
				},
				initOverrides,
			);
			return new JSONApiResponse(response, (jsonValue) =>
				ModelTagDTOFromJSON(jsonValue),
			);
		}
		async apiV1TagsIdPut(requestParameters, initOverrides) {
			const response = await this.apiV1TagsIdPutRaw(
				requestParameters,
				initOverrides,
			);
			return await response.value();
		}
		async apiV1TagsPostRaw(requestParameters, initOverrides) {
			if (requestParameters["tag"] == null) {
				throw new RequiredError(
					"tag",
					'Required parameter "tag" was null or undefined when calling apiV1TagsPost().',
				);
			}
			const queryParameters = {};
			const headerParameters = {};
			headerParameters["Content-Type"] = "application/json";
			let urlPath = `/api/v1/tags`;
			const response = await this.request(
				{
					path: urlPath,
					method: "POST",
					headers: headerParameters,
					query: queryParameters,
					body: ModelTagDTOToJSON(requestParameters["tag"]),
				},
				initOverrides,
			);
			return new JSONApiResponse(response, (jsonValue) =>
				ModelTagDTOFromJSON(jsonValue),
			);
		}
		async apiV1TagsPost(requestParameters, initOverrides) {
			const response = await this.apiV1TagsPostRaw(
				requestParameters,
				initOverrides,
			);
			switch (response.raw.status) {
				case 201:
					return await response.value();
				case 204:
					return null;
				default:
					return await response.value();
			}
		}
	}

	// wrapper.js
	window.ShioriAPI = {
		Configuration,
		AccountsApi,
		AuthApi,
		BookmarksApi,
		SystemApi,
		TagsApi,
	};
})();
