// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

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
        "/health/status": {
            "get": {
                "description": "Check health status of the service pinging the database and go metrics",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Health"
                ],
                "summary": "Check health status",
                "operationId": "0986a6ff-aa83-4b06-9a16-7e338eaa50d1",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Health"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/respond.HTTPMessage"
                        }
                    }
                }
            }
        },
        "/users": {
            "get": {
                "description": "List all users",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Users"
                ],
                "summary": "List all users",
                "operationId": "1213ffb2-b9f3-4134-923e-13bb777da62b",
                "parameters": [
                    {
                        "type": "string",
                        "format": "string",
                        "description": "Comma-separated list of fields to sort by. Example: first_name ASC, created_at DESC",
                        "name": "sort",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "format": "string",
                        "description": "Filter field. Example: id=1 AND first_name='John'",
                        "name": "filter",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "format": "string",
                        "description": "Fields to return. Example: id,first_name,last_name",
                        "name": "fields",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "format": "string",
                        "description": "Next cursor",
                        "name": "next_token",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "format": "string",
                        "description": "Previous cursor",
                        "name": "prev_token",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "format": "int",
                        "description": "Limit",
                        "name": "limit",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.ListUsersResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/respond.HTTPMessage"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/respond.HTTPMessage"
                        }
                    }
                }
            },
            "post": {
                "description": "Create a new user from scratch\nIf the id is not provided, it will be generated automatically",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Users"
                ],
                "summary": "Create a new user",
                "operationId": "f71e14db-fc77-4fb3-a21d-292eade431df",
                "parameters": [
                    {
                        "format": "json",
                        "description": "CreateUserRequest",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.CreateUserRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/respond.HTTPMessage"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/respond.HTTPMessage"
                        }
                    },
                    "409": {
                        "description": "Conflict",
                        "schema": {
                            "$ref": "#/definitions/respond.HTTPMessage"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/respond.HTTPMessage"
                        }
                    }
                }
            }
        },
        "/users/{user_id}": {
            "get": {
                "description": "Get a user by ID",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Users"
                ],
                "summary": "Get a user by ID",
                "operationId": "b823ba3c-3b83-4eaa-bdf7-ce1b05237f23",
                "parameters": [
                    {
                        "type": "string",
                        "format": "uuid",
                        "description": "The user ID in UUID format",
                        "name": "user_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.User"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/respond.HTTPMessage"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/respond.HTTPMessage"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/respond.HTTPMessage"
                        }
                    }
                }
            },
            "put": {
                "description": "Update a user",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Users"
                ],
                "summary": "Update a user",
                "operationId": "75165751-045b-465d-ba93-c88a27b6a42e",
                "parameters": [
                    {
                        "type": "string",
                        "format": "uuid",
                        "description": "The user ID in UUID format",
                        "name": "user_id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "format": "json",
                        "description": "User",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.UpdateUserRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/respond.HTTPMessage"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/respond.HTTPMessage"
                        }
                    },
                    "409": {
                        "description": "Conflict",
                        "schema": {
                            "$ref": "#/definitions/respond.HTTPMessage"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/respond.HTTPMessage"
                        }
                    }
                }
            },
            "delete": {
                "description": "Delete a user",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Users"
                ],
                "summary": "Delete a user",
                "operationId": "48e60e0a-ea1c-46d4-8729-c47dd82a4e93",
                "parameters": [
                    {
                        "type": "string",
                        "format": "uuid",
                        "description": "The user ID in UUID format",
                        "name": "user_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content",
                        "schema": {
                            "$ref": "#/definitions/respond.HTTPMessage"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/respond.HTTPMessage"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/respond.HTTPMessage"
                        }
                    }
                }
            }
        },
        "/version": {
            "get": {
                "description": "Get the version of the service",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Version"
                ],
                "summary": "Get the version of the service",
                "operationId": "d85b4a3f-b032-4dd1-b3ab-bc9a00f95eb5",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handler.Version"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/respond.HTTPMessage"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "handler.Version": {
            "type": "object",
            "properties": {
                "build_date": {
                    "type": "string"
                },
                "git_branch": {
                    "type": "string"
                },
                "git_commit": {
                    "type": "string"
                },
                "go_version": {
                    "type": "string"
                },
                "go_version_arch": {
                    "type": "string"
                },
                "go_version_os": {
                    "type": "string"
                },
                "version": {
                    "type": "string"
                }
            }
        },
        "model.Check": {
            "description": "Health check of the service",
            "type": "object",
            "properties": {
                "data": {
                    "type": "object",
                    "format": "map",
                    "additionalProperties": true
                },
                "kind": {
                    "type": "string",
                    "format": "string",
                    "example": "database"
                },
                "name": {
                    "type": "string",
                    "format": "string",
                    "example": "database"
                },
                "status": {
                    "type": "boolean",
                    "format": "boolean",
                    "example": true
                }
            }
        },
        "model.CreateUserRequest": {
            "description": "CreateUserRequest represents the input for the CreateUser method",
            "type": "object",
            "properties": {
                "email": {
                    "type": "string",
                    "format": "email",
                    "example": "my@email.com"
                },
                "first_name": {
                    "type": "string",
                    "format": "string",
                    "example": "John"
                },
                "id": {
                    "type": "string",
                    "format": "uuid",
                    "example": "550e8400-e29b-41d4-a716-446655440000"
                },
                "last_name": {
                    "type": "string",
                    "format": "string",
                    "example": "Doe"
                },
                "password": {
                    "type": "string",
                    "format": "string",
                    "example": "ThisIs4Passw0rd"
                }
            }
        },
        "model.Health": {
            "description": "Health check of the service",
            "type": "object",
            "properties": {
                "checks": {
                    "type": "array",
                    "items": {
                        "format": "array",
                        "$ref": "#/definitions/model.Check"
                    }
                },
                "status": {
                    "type": "boolean",
                    "format": "boolean",
                    "example": true
                }
            }
        },
        "model.ListUsersResponse": {
            "description": "ListUsersResponse represents a list of users",
            "type": "object",
            "properties": {
                "items": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/model.User"
                    }
                },
                "paginator": {
                    "$ref": "#/definitions/paginator.Paginator"
                }
            }
        },
        "model.UpdateUserRequest": {
            "description": "UpdateUserRequest represents the input for the UpdateUser method",
            "type": "object",
            "properties": {
                "disabled": {
                    "type": "boolean",
                    "format": "boolean",
                    "example": false
                },
                "email": {
                    "type": "string",
                    "format": "email",
                    "example": "my@email.com"
                },
                "first_name": {
                    "type": "string",
                    "format": "string",
                    "example": "John"
                },
                "last_name": {
                    "type": "string",
                    "format": "string",
                    "example": "Doe"
                },
                "password": {
                    "type": "string",
                    "format": "string",
                    "example": "ThisIs4Passw0rd"
                }
            }
        },
        "model.User": {
            "description": "User represents a user entity",
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string",
                    "format": "date-time",
                    "example": "2021-01-01T00:00:00Z"
                },
                "disabled": {
                    "type": "boolean",
                    "format": "boolean",
                    "example": false
                },
                "email": {
                    "type": "string",
                    "format": "email",
                    "example": "my@email.com"
                },
                "first_name": {
                    "type": "string",
                    "format": "string",
                    "example": "John"
                },
                "id": {
                    "type": "string",
                    "format": "uuid",
                    "example": "550e8400-e29b-41d4-a716-446655440000"
                },
                "last_name": {
                    "type": "string",
                    "format": "string",
                    "example": "Doe"
                },
                "updated_at": {
                    "type": "string",
                    "format": "date-time",
                    "example": "2021-01-01T00:00:00Z"
                }
            }
        },
        "paginator.Paginator": {
            "description": "Paginator represents a paginator",
            "type": "object",
            "properties": {
                "limit": {
                    "type": "integer",
                    "format": "int",
                    "example": 10
                },
                "next_page": {
                    "type": "string",
                    "format": "string",
                    "example": "http://localhost:8080/users?next_token=ZmZmZmZmZmYtZmZmZi0tZmZmZmZmZmY=\u0026limit=10"
                },
                "next_token": {
                    "type": "string",
                    "format": "string",
                    "example": "ZmZmZmZmZmYtZmZmZi0tZmZmZmZmZmY="
                },
                "prev_page": {
                    "type": "string",
                    "format": "string",
                    "example": "http://localhost:8080/users?prev_token=ZmZmZmZmZmYtZmZmZi0tZmZmZmZmZmY=\u0026limit=10"
                },
                "prev_token": {
                    "type": "string",
                    "format": "string",
                    "example": "ZmZmZmZmZmYtZmZmZi0tZmZmZmZmZmY="
                },
                "size": {
                    "type": "integer",
                    "format": "int",
                    "example": 10
                }
            }
        },
        "respond.HTTPMessage": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                },
                "method": {
                    "type": "string"
                },
                "path": {
                    "type": "string"
                },
                "status_code": {
                    "type": "integer"
                },
                "timestamp": {
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
	Description:      "This is a service template for building RESTful APIs in Go.\nIt uses a PostgreSQL database to store user information.\nThe service provides:\n- CRUD operations for users.\n- Health and version endpoints.\n- Configuration using environment variables or command line arguments.\n- Debug mode to enable debug logging.\n- TLS enabled to secure the communication.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
