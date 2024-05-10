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
        "/health": {
            "get": {
                "description": "Get the health of the service",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "health"
                ],
                "summary": "Get the health of the service",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Health"
                        }
                    }
                }
            }
        },
        "/healthz": {
            "get": {
                "description": "Get the health of the service",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "health"
                ],
                "summary": "Get the health of the service",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Health"
                        }
                    }
                }
            }
        },
        "/status": {
            "get": {
                "description": "Get the health of the service",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "health"
                ],
                "summary": "Get the health of the service",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Health"
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
                    "users"
                ],
                "summary": "List all users",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Sort field",
                        "name": "sort",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Filter field",
                        "name": "filter",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Fields to return",
                        "name": "fields",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Query string",
                        "name": "query",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Next cursor",
                        "name": "next",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Previous cursor",
                        "name": "prev",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "Limit",
                        "name": "limit",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.ListUserResponse"
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
                "description": "Create a new user",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "Create a new user",
                "parameters": [
                    {
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
                            "$ref": "#/definitions/model.CreateUserRequest"
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
        "/users/{id}": {
            "get": {
                "description": "Get a user by ID",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "Get a user by ID",
                "parameters": [
                    {
                        "type": "string",
                        "description": "User ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Query string",
                        "name": "query",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.User"
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
            "put": {
                "description": "Update a user",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "Update a user",
                "parameters": [
                    {
                        "type": "string",
                        "description": "User ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "User",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.User"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "delete": {
                "description": "Delete a user",
                "tags": [
                    "users"
                ],
                "summary": "Delete a user",
                "parameters": [
                    {
                        "type": "string",
                        "description": "User ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "model.Check": {
            "type": "object",
            "properties": {
                "data": {
                    "description": "Data is an optional field that can be used to provide additional information about the check.",
                    "type": "object",
                    "additionalProperties": true
                },
                "kind": {
                    "description": "Kind is the kind of check.",
                    "type": "string"
                },
                "name": {
                    "description": "Name is the name of the check.",
                    "type": "string"
                },
                "status": {
                    "description": "Status is the status of the check.",
                    "type": "boolean"
                }
            }
        },
        "model.CreateUserRequest": {
            "type": "object",
            "properties": {
                "email": {
                    "description": "Email is the email address of the user.",
                    "type": "string"
                },
                "first_name": {
                    "description": "FirstName is the first name of the user.",
                    "type": "string"
                },
                "id": {
                    "description": "ID is the unique identifier of the user.",
                    "type": "string"
                },
                "last_name": {
                    "description": "LastName is the last name of the user.",
                    "type": "string"
                }
            }
        },
        "model.Health": {
            "type": "object",
            "properties": {
                "checks": {
                    "description": "Checks is a list of health checks.",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/model.Check"
                    }
                },
                "status": {
                    "description": "Status is the status of the health check.",
                    "type": "boolean"
                }
            }
        },
        "model.ListUserResponse": {
            "type": "object",
            "properties": {
                "data": {
                    "description": "Items is a list of users.",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/model.User"
                    }
                },
                "paginator": {
                    "description": "Paginator is the paginator for the list of users.",
                    "allOf": [
                        {
                            "$ref": "#/definitions/paginator.Paginator"
                        }
                    ]
                }
            }
        },
        "model.User": {
            "type": "object",
            "properties": {
                "created_at": {
                    "description": "Email is the email address of the user.",
                    "type": "string"
                },
                "email": {
                    "description": "Email is the email address of the user.",
                    "type": "string"
                },
                "first_name": {
                    "description": "FirstName is the first name of the user.",
                    "type": "string"
                },
                "id": {
                    "description": "ID is the unique identifier of the user.",
                    "type": "string"
                },
                "last_name": {
                    "description": "LastName is the last name of the user.",
                    "type": "string"
                },
                "updated_at": {
                    "description": "UpdatedAt is the time the user was last updated.",
                    "type": "string"
                }
            }
        },
        "paginator.Paginator": {
            "type": "object",
            "properties": {
                "limit": {
                    "description": "Limit is the maximum number of elements to return.",
                    "type": "integer"
                },
                "next": {
                    "description": "Next is the cursor token to the next page.",
                    "type": "string"
                },
                "previous": {
                    "description": "Prev is the cursor token to the previous page.",
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0.0",
	Host:             "localhost:8080",
	BasePath:         "/",
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
