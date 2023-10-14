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
        "/api/v1/bookmaeks/cache": {
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
                            "$ref": "#/definitions/model.Bookmark"
                        }
                    },
                    "403": {
                        "description": "Token not provided/invalid"
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
        "api_v1.updateCachePayload": {
            "type": "object",
            "required": [
                "ids"
            ],
            "properties": {
                "createArchive": {
                    "type": "boolean"
                },
                "createEbook": {
                    "type": "boolean"
                },
                "ids": {
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                },
                "keepMetadata": {
                    "type": "boolean"
                },
                "skipExist": {
                    "type": "boolean"
                }
            }
        },
        "model.Account": {
            "type": "object",
            "properties": {
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
        "model.Bookmark": {
            "type": "object",
            "properties": {
                "author": {
                    "type": "string"
                },
                "createArchive": {
                    "type": "boolean"
                },
                "createEbook": {
                    "type": "boolean"
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
                "modified": {
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
