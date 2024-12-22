// Package swagger Code generated by swaggo/swag. DO NOT EDIT
package swagger

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/v1/accounts": {
            "get": {
                "description": "List accounts",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "accounts"
                ],
                "summary": "List accounts",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/model.AccountDTO"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "post": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "accounts"
                ],
                "summary": "Create an account",
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/model.AccountDTO"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/api/v1/accounts/{id}": {
            "delete": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "accounts"
                ],
                "summary": "Delete an account",
                "responses": {
                    "204": {
                        "description": "No content",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "patch": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "accounts"
                ],
                "summary": "Update an account",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/api_v1.updateAccountPayload"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/api/v1/auth/account": {
            "patch": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Auth"
                ],
                "summary": "Update account information",
                "parameters": [
                    {
                        "description": "Account data",
                        "name": "payload",
                        "in": "body",
                        "schema": {
                            "$ref": "#/definitions/api_v1.updateAccountPayload"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Account"
                        }
                    },
                    "403": {
                        "description": "Token not provided/invalid"
                    }
                }
            }
        },
        "/api/v1/auth/login": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Auth"
                ],
                "summary": "Login to an account using username and password",
                "parameters": [
                    {
                        "description": "Login data",
                        "name": "payload",
                        "in": "body",
                        "schema": {
                            "$ref": "#/definitions/api_v1.loginRequestPayload"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Login successful",
                        "schema": {
                            "$ref": "#/definitions/api_v1.loginResponseMessage"
                        }
                    },
                    "400": {
                        "description": "Invalid login data"
                    }
                }
            }
        },
        "/api/v1/auth/logout": {
            "post": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Auth"
                ],
                "summary": "Logout from the current session",
                "responses": {
                    "200": {
                        "description": "Logout successful"
                    },
                    "403": {
                        "description": "Token not provided/invalid"
                    }
                }
            }
        },
        "/api/v1/auth/me": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Auth"
                ],
                "summary": "Get information for the current logged in user",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Account"
                        }
                    },
                    "403": {
                        "description": "Token not provided/invalid"
                    }
                }
            }
        },
        "/api/v1/auth/refresh": {
            "post": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Auth"
                ],
                "summary": "Refresh a token for an account",
                "responses": {
                    "200": {
                        "description": "Refresh successful",
                        "schema": {
                            "$ref": "#/definitions/api_v1.loginResponseMessage"
                        }
                    },
                    "403": {
                        "description": "Token not provided/invalid"
                    }
                }
            }
        },
        "/api/v1/bookmarks/cache": {
            "put": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Auth"
                ],
                "summary": "Update Cache and Ebook on server.",
                "parameters": [
                    {
                        "description": "Update Cache Payload",
                        "name": "payload",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/api_v1.updateCachePayload"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.BookmarkDTO"
                        }
                    },
                    "403": {
                        "description": "Token not provided/invalid"
                    }
                }
            }
        },
        "/api/v1/bookmarks/id/readable": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Auth"
                ],
                "summary": "Get readable version of bookmark.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api_v1.readableResponseMessage"
                        }
                    },
                    "403": {
                        "description": "Token not provided/invalid"
                    }
                }
            }
        },
        "/api/v1/system/info": {
            "get": {
                "description": "Get general system information like Shiori version, database, and OS",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "system"
                ],
                "summary": "Get general system information",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api_v1.infoResponse"
                        }
                    },
                    "403": {
                        "description": "Only owners can access this endpoint"
                    }
                }
            }
        },
        "/api/v1/tags": {
            "get": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Tags"
                ],
                "summary": "List tags",
                "responses": {
                    "200": {
                        "description": "List of tags",
                        "schema": {
                            "$ref": "#/definitions/model.Tag"
                        }
                    },
                    "403": {
                        "description": "Token not provided/invalid"
                    }
                }
            },
            "post": {
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Tags"
                ],
                "summary": "Create tag",
                "responses": {
                    "200": {
                        "description": "Created tag",
                        "schema": {
                            "$ref": "#/definitions/model.Tag"
                        }
                    },
                    "400": {
                        "description": "Token not provided/invalid"
                    },
                    "403": {
                        "description": "Token not provided/invalid"
                    }
                }
            }
        }
    },
    "definitions": {
        "api_v1.infoResponse": {
            "type": "object",
            "properties": {
                "database": {
                    "type": "string"
                },
                "os": {
                    "type": "string"
                },
                "version": {
                    "type": "object",
                    "properties": {
                        "commit": {
                            "type": "string"
                        },
                        "date": {
                            "type": "string"
                        },
                        "tag": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "api_v1.loginRequestPayload": {
            "type": "object",
            "required": [
                "password",
                "username"
            ],
            "properties": {
                "password": {
                    "type": "string"
                },
                "remember_me": {
                    "type": "boolean"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "api_v1.loginResponseMessage": {
            "type": "object",
            "properties": {
                "expires": {
                    "description": "Deprecated, used only for legacy APIs",
                    "type": "integer"
                },
                "session": {
                    "description": "Deprecated, used only for legacy APIs",
                    "type": "string"
                },
                "token": {
                    "type": "string"
                }
            }
        },
        "api_v1.readableResponseMessage": {
            "type": "object",
            "properties": {
                "content": {
                    "type": "string"
                },
                "html": {
                    "type": "string"
                }
            }
        },
        "api_v1.updateAccountPayload": {
            "type": "object",
            "properties": {
                "config": {
                    "$ref": "#/definitions/model.UserConfig"
                },
                "new_password": {
                    "type": "string"
                },
                "old_password": {
                    "type": "string"
                },
                "owner": {
                    "type": "boolean"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "api_v1.updateCachePayload": {
            "type": "object",
            "required": [
                "ids"
            ],
            "properties": {
                "create_archive": {
                    "type": "boolean"
                },
                "create_ebook": {
                    "type": "boolean"
                },
                "ids": {
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                },
                "keep_metadata": {
                    "type": "boolean"
                },
                "skip_exist": {
                    "type": "boolean"
                }
            }
        },
        "model.Account": {
            "type": "object",
            "properties": {
                "config": {
                    "$ref": "#/definitions/model.UserConfig"
                },
                "id": {
                    "type": "integer"
                },
                "owner": {
                    "type": "boolean"
                },
                "password": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "model.AccountDTO": {
            "type": "object",
            "properties": {
                "config": {
                    "$ref": "#/definitions/model.UserConfig"
                },
                "id": {
                    "type": "integer"
                },
                "owner": {
                    "type": "boolean"
                },
                "passowrd": {
                    "description": "Used only to store, not to retrieve",
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "model.BookmarkDTO": {
            "type": "object",
            "properties": {
                "author": {
                    "type": "string"
                },
                "create_archive": {
                    "description": "TODO: migrate outside the DTO",
                    "type": "boolean"
                },
                "create_ebook": {
                    "description": "TODO: migrate outside the DTO",
                    "type": "boolean"
                },
                "createdAt": {
                    "type": "string"
                },
                "excerpt": {
                    "type": "string"
                },
                "hasArchive": {
                    "type": "boolean"
                },
                "hasContent": {
                    "type": "boolean"
                },
                "hasEbook": {
                    "type": "boolean"
                },
                "html": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "imageURL": {
                    "type": "string"
                },
                "modifiedAt": {
                    "type": "string"
                },
                "public": {
                    "type": "integer"
                },
                "tags": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/model.Tag"
                    }
                },
                "title": {
                    "type": "string"
                },
                "url": {
                    "type": "string"
                }
            }
        },
        "model.Tag": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer"
                },
                "nBookmarks": {
                    "type": "integer"
                },
                "name": {
                    "type": "string"
                }
            }
        },
        "model.UserConfig": {
            "type": "object",
            "properties": {
                "CreateEbook": {
                    "type": "boolean"
                },
                "HideExcerpt": {
                    "type": "boolean"
                },
                "HideThumbnail": {
                    "type": "boolean"
                },
                "KeepMetadata": {
                    "type": "boolean"
                },
                "ListMode": {
                    "type": "boolean"
                },
                "MakePublic": {
                    "type": "boolean"
                },
                "ShowId": {
                    "type": "boolean"
                },
                "Theme": {
                    "type": "string"
                },
                "UseArchive": {
                    "type": "boolean"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
