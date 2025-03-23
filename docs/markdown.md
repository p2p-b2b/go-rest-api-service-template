


# Go REST API Service Template
This is a service template for building RESTful APIs in Go.
It uses a PostgreSQL database to store user information.
The service provides:
- CRUD operations for users.
- Health and version endpoints.
- Configuration using environment variables or command line arguments.
- Debug mode to enable debug logging.
- TLS enabled to secure the communication.
  

## Informations

### Version

v1

### Contact

API Support info@qu3ry.me https://qu3ry.me

## Content negotiation

### URI Schemes
  * http

### Consumes
  * application/json

### Produces
  * application/json

## All endpoints

###  health

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /health/status | [0986a6ff aa83 4b06 9a16 7e338eaa50d1](#0986a6ff-aa83-4b06-9a16-7e338eaa50d1) | Check health status |
  


###  users

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| DELETE | /users/{user_id} | [48e60e0a ea1c 46d4 8729 c47dd82a4e93](#48e60e0a-ea1c-46d4-8729-c47dd82a4e93) | Delete a user |
| POST | /users | [8a1488b0 2d2c 42a0 a57a 6560aaf3ec76](#8a1488b0-2d2c-42a0-a57a-6560aaf3ec76) | Create user |
| PUT | /users/{user_id} | [a7979074 e16c 4aec 86e0 e5a154bbfc51](#a7979074-e16c-4aec-86e0-e5a154bbfc51) | Update a user |
| GET | /users | [b51b8ab6 4bb4 4b37 af5c 9825ba7e71e5](#b51b8ab6-4bb4-4b37-af5c-9825ba7e71e5) | List users |
| GET | /users/{user_id} | [b823ba3c 3b83 4eaa bdf7 ce1b05237f23](#b823ba3c-3b83-4eaa-bdf7-ce1b05237f23) | Get a user by ID |
  


###  version

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /version | [d85b4a3f b032 4dd1 b3ab bc9a00f95eb5](#d85b4a3f-b032-4dd1-b3ab-bc9a00f95eb5) | Get the version of the service |
  


## Paths

### <span id="0986a6ff-aa83-4b06-9a16-7e338eaa50d1"></span> Check health status (*0986a6ff-aa83-4b06-9a16-7e338eaa50d1*)

```
GET /health/status
```

Check health status of the service pinging the database and go metrics

#### Produces
  * application/json

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#0986a6ff-aa83-4b06-9a16-7e338eaa50d1-200) | OK | OK |  | [schema](#0986a6ff-aa83-4b06-9a16-7e338eaa50d1-200-schema) |
| [500](#0986a6ff-aa83-4b06-9a16-7e338eaa50d1-500) | Internal Server Error | Internal Server Error |  | [schema](#0986a6ff-aa83-4b06-9a16-7e338eaa50d1-500-schema) |

#### Responses


##### <span id="0986a6ff-aa83-4b06-9a16-7e338eaa50d1-200"></span> 200 - OK
Status: OK

###### <span id="0986a6ff-aa83-4b06-9a16-7e338eaa50d1-200-schema"></span> Schema
   
  

[ModelHealth](#model-health)

##### <span id="0986a6ff-aa83-4b06-9a16-7e338eaa50d1-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0986a6ff-aa83-4b06-9a16-7e338eaa50d1-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="48e60e0a-ea1c-46d4-8729-c47dd82a4e93"></span> Delete a user (*48e60e0a-ea1c-46d4-8729-c47dd82a4e93*)

```
DELETE /users/{user_id}
```

Delete a user

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| user_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The user ID in UUID format |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#48e60e0a-ea1c-46d4-8729-c47dd82a4e93-200) | OK | OK |  | [schema](#48e60e0a-ea1c-46d4-8729-c47dd82a4e93-200-schema) |
| [400](#48e60e0a-ea1c-46d4-8729-c47dd82a4e93-400) | Bad Request | Bad Request |  | [schema](#48e60e0a-ea1c-46d4-8729-c47dd82a4e93-400-schema) |
| [500](#48e60e0a-ea1c-46d4-8729-c47dd82a4e93-500) | Internal Server Error | Internal Server Error |  | [schema](#48e60e0a-ea1c-46d4-8729-c47dd82a4e93-500-schema) |

#### Responses


##### <span id="48e60e0a-ea1c-46d4-8729-c47dd82a4e93-200"></span> 200 - OK
Status: OK

###### <span id="48e60e0a-ea1c-46d4-8729-c47dd82a4e93-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="48e60e0a-ea1c-46d4-8729-c47dd82a4e93-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="48e60e0a-ea1c-46d4-8729-c47dd82a4e93-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="48e60e0a-ea1c-46d4-8729-c47dd82a4e93-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="48e60e0a-ea1c-46d4-8729-c47dd82a4e93-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="8a1488b0-2d2c-42a0-a57a-6560aaf3ec76"></span> Create user (*8a1488b0-2d2c-42a0-a57a-6560aaf3ec76*)

```
POST /users
```

Create new user from scratch.
If the id is not provided, it will be generated automatically.

#### Consumes
  * application/json

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| body | `body` | [ModelCreateUserRequest](#model-create-user-request) | `models.ModelCreateUserRequest` | | ✓ | | Create user request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [201](#8a1488b0-2d2c-42a0-a57a-6560aaf3ec76-201) | Created | Created |  | [schema](#8a1488b0-2d2c-42a0-a57a-6560aaf3ec76-201-schema) |
| [400](#8a1488b0-2d2c-42a0-a57a-6560aaf3ec76-400) | Bad Request | Bad Request |  | [schema](#8a1488b0-2d2c-42a0-a57a-6560aaf3ec76-400-schema) |
| [409](#8a1488b0-2d2c-42a0-a57a-6560aaf3ec76-409) | Conflict | Conflict |  | [schema](#8a1488b0-2d2c-42a0-a57a-6560aaf3ec76-409-schema) |
| [500](#8a1488b0-2d2c-42a0-a57a-6560aaf3ec76-500) | Internal Server Error | Internal Server Error |  | [schema](#8a1488b0-2d2c-42a0-a57a-6560aaf3ec76-500-schema) |

#### Responses


##### <span id="8a1488b0-2d2c-42a0-a57a-6560aaf3ec76-201"></span> 201 - Created
Status: Created

###### <span id="8a1488b0-2d2c-42a0-a57a-6560aaf3ec76-201-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="8a1488b0-2d2c-42a0-a57a-6560aaf3ec76-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="8a1488b0-2d2c-42a0-a57a-6560aaf3ec76-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="8a1488b0-2d2c-42a0-a57a-6560aaf3ec76-409"></span> 409 - Conflict
Status: Conflict

###### <span id="8a1488b0-2d2c-42a0-a57a-6560aaf3ec76-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="8a1488b0-2d2c-42a0-a57a-6560aaf3ec76-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="8a1488b0-2d2c-42a0-a57a-6560aaf3ec76-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="a7979074-e16c-4aec-86e0-e5a154bbfc51"></span> Update a user (*a7979074-e16c-4aec-86e0-e5a154bbfc51*)

```
PUT /users/{user_id}
```

Update a user

#### Consumes
  * application/json

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| user_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The user ID in UUID format |
| body | `body` | [ModelUpdateUserRequest](#model-update-user-request) | `models.ModelUpdateUserRequest` | | ✓ | | User update request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#a7979074-e16c-4aec-86e0-e5a154bbfc51-200) | OK | OK |  | [schema](#a7979074-e16c-4aec-86e0-e5a154bbfc51-200-schema) |
| [400](#a7979074-e16c-4aec-86e0-e5a154bbfc51-400) | Bad Request | Bad Request |  | [schema](#a7979074-e16c-4aec-86e0-e5a154bbfc51-400-schema) |
| [409](#a7979074-e16c-4aec-86e0-e5a154bbfc51-409) | Conflict | Conflict |  | [schema](#a7979074-e16c-4aec-86e0-e5a154bbfc51-409-schema) |
| [500](#a7979074-e16c-4aec-86e0-e5a154bbfc51-500) | Internal Server Error | Internal Server Error |  | [schema](#a7979074-e16c-4aec-86e0-e5a154bbfc51-500-schema) |

#### Responses


##### <span id="a7979074-e16c-4aec-86e0-e5a154bbfc51-200"></span> 200 - OK
Status: OK

###### <span id="a7979074-e16c-4aec-86e0-e5a154bbfc51-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="a7979074-e16c-4aec-86e0-e5a154bbfc51-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="a7979074-e16c-4aec-86e0-e5a154bbfc51-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="a7979074-e16c-4aec-86e0-e5a154bbfc51-409"></span> 409 - Conflict
Status: Conflict

###### <span id="a7979074-e16c-4aec-86e0-e5a154bbfc51-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="a7979074-e16c-4aec-86e0-e5a154bbfc51-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="a7979074-e16c-4aec-86e0-e5a154bbfc51-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="b51b8ab6-4bb4-4b37-af5c-9825ba7e71e5"></span> List users (*b51b8ab6-4bb4-4b37-af5c-9825ba7e71e5*)

```
GET /users
```

List users with pagination and filtering

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| fields | `query` | string (formatted string) | `string` |  |  |  | Fields to return. Example: id,first_name,last_name |
| filter | `query` | string (formatted string) | `string` |  |  |  | Filter field. Example: id=1 AND first_name='John' |
| limit | `query` | int (formatted integer) | `int64` |  |  |  | Limit |
| next_token | `query` | string (formatted string) | `string` |  |  |  | Next cursor |
| prev_token | `query` | string (formatted string) | `string` |  |  |  | Previous cursor |
| sort | `query` | string (formatted string) | `string` |  |  |  | Comma-separated list of fields to sort by. Example: first_name ASC, created_at DESC |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#b51b8ab6-4bb4-4b37-af5c-9825ba7e71e5-200) | OK | OK |  | [schema](#b51b8ab6-4bb4-4b37-af5c-9825ba7e71e5-200-schema) |
| [400](#b51b8ab6-4bb4-4b37-af5c-9825ba7e71e5-400) | Bad Request | Bad Request |  | [schema](#b51b8ab6-4bb4-4b37-af5c-9825ba7e71e5-400-schema) |
| [500](#b51b8ab6-4bb4-4b37-af5c-9825ba7e71e5-500) | Internal Server Error | Internal Server Error |  | [schema](#b51b8ab6-4bb4-4b37-af5c-9825ba7e71e5-500-schema) |

#### Responses


##### <span id="b51b8ab6-4bb4-4b37-af5c-9825ba7e71e5-200"></span> 200 - OK
Status: OK

###### <span id="b51b8ab6-4bb4-4b37-af5c-9825ba7e71e5-200-schema"></span> Schema
   
  

[ModelListUsersResponse](#model-list-users-response)

##### <span id="b51b8ab6-4bb4-4b37-af5c-9825ba7e71e5-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="b51b8ab6-4bb4-4b37-af5c-9825ba7e71e5-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="b51b8ab6-4bb4-4b37-af5c-9825ba7e71e5-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="b51b8ab6-4bb4-4b37-af5c-9825ba7e71e5-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="b823ba3c-3b83-4eaa-bdf7-ce1b05237f23"></span> Get a user by ID (*b823ba3c-3b83-4eaa-bdf7-ce1b05237f23*)

```
GET /users/{user_id}
```

Get a user by ID

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| user_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The user ID in UUID format |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#b823ba3c-3b83-4eaa-bdf7-ce1b05237f23-200) | OK | OK |  | [schema](#b823ba3c-3b83-4eaa-bdf7-ce1b05237f23-200-schema) |
| [400](#b823ba3c-3b83-4eaa-bdf7-ce1b05237f23-400) | Bad Request | Bad Request |  | [schema](#b823ba3c-3b83-4eaa-bdf7-ce1b05237f23-400-schema) |
| [404](#b823ba3c-3b83-4eaa-bdf7-ce1b05237f23-404) | Not Found | Not Found |  | [schema](#b823ba3c-3b83-4eaa-bdf7-ce1b05237f23-404-schema) |
| [500](#b823ba3c-3b83-4eaa-bdf7-ce1b05237f23-500) | Internal Server Error | Internal Server Error |  | [schema](#b823ba3c-3b83-4eaa-bdf7-ce1b05237f23-500-schema) |

#### Responses


##### <span id="b823ba3c-3b83-4eaa-bdf7-ce1b05237f23-200"></span> 200 - OK
Status: OK

###### <span id="b823ba3c-3b83-4eaa-bdf7-ce1b05237f23-200-schema"></span> Schema
   
  

[ModelUser](#model-user)

##### <span id="b823ba3c-3b83-4eaa-bdf7-ce1b05237f23-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="b823ba3c-3b83-4eaa-bdf7-ce1b05237f23-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="b823ba3c-3b83-4eaa-bdf7-ce1b05237f23-404"></span> 404 - Not Found
Status: Not Found

###### <span id="b823ba3c-3b83-4eaa-bdf7-ce1b05237f23-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="b823ba3c-3b83-4eaa-bdf7-ce1b05237f23-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="b823ba3c-3b83-4eaa-bdf7-ce1b05237f23-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="d85b4a3f-b032-4dd1-b3ab-bc9a00f95eb5"></span> Get the version of the service (*d85b4a3f-b032-4dd1-b3ab-bc9a00f95eb5*)

```
GET /version
```

Get the version of the service

#### Produces
  * application/json

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#d85b4a3f-b032-4dd1-b3ab-bc9a00f95eb5-200) | OK | OK |  | [schema](#d85b4a3f-b032-4dd1-b3ab-bc9a00f95eb5-200-schema) |
| [500](#d85b4a3f-b032-4dd1-b3ab-bc9a00f95eb5-500) | Internal Server Error | Internal Server Error |  | [schema](#d85b4a3f-b032-4dd1-b3ab-bc9a00f95eb5-500-schema) |

#### Responses


##### <span id="d85b4a3f-b032-4dd1-b3ab-bc9a00f95eb5-200"></span> 200 - OK
Status: OK

###### <span id="d85b4a3f-b032-4dd1-b3ab-bc9a00f95eb5-200-schema"></span> Schema
   
  

[ModelVersion](#model-version)

##### <span id="d85b4a3f-b032-4dd1-b3ab-bc9a00f95eb5-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="d85b4a3f-b032-4dd1-b3ab-bc9a00f95eb5-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

## Models

### <span id="model-check"></span> model.Check


> Health check of the service
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | [interface{}](#interface)| `interface{}` |  | |  |  |
| kind | string (formatted string)| `string` |  | |  | `database` |
| name | string (formatted string)| `string` |  | |  | `database` |
| status | boolean (formatted boolean)| `bool` |  | |  | `true` |



### <span id="model-create-user-request"></span> model.CreateUserRequest


> CreateUserRequest represents the input for the CreateUser method
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| email | email (formatted string)| `strfmt.Email` |  | |  | `my@email.com` |
| first_name | string (formatted string)| `string` |  | |  | `John` |
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `550e8400-e29b-41d4-a716-446655440000` |
| last_name | string (formatted string)| `string` |  | |  | `Doe` |
| password | string (formatted string)| `string` |  | |  | `ThisIs4Passw0rd` |



### <span id="model-http-message"></span> model.HTTPMessage


> HTTPMessage represents a message to be sent to the client though the HTTP REST API.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| message | string (formatted string)| `string` |  | |  | `Hello, World!` |
| method | string (formatted string)| `string` |  | |  | `GET` |
| path | string (formatted string)| `string` |  | |  | `/api/v1/hello` |
| status_code | int32 (formatted integer)| `int32` |  | |  | `200` |
| timestamp | date-time (formatted string)| `strfmt.DateTime` |  | |  | `2021-01-01T00:00:00Z` |



### <span id="model-health"></span> model.Health


> Health check of the service
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| checks | [][ModelCheck](#model-check)| `[]*ModelCheck` |  | |  |  |
| status | boolean (formatted boolean)| `bool` |  | |  | `true` |



### <span id="model-list-users-response"></span> model.ListUsersResponse


> ListUsersResponse represents a list of users
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| items | [][ModelUser](#model-user)| `[]*ModelUser` |  | |  |  |
| paginator | [ModelPaginator](#model-paginator)| `ModelPaginator` |  | |  |  |



### <span id="model-paginator"></span> model.Paginator


> Paginator represents a paginator
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| limit | int (formatted integer)| `int64` |  | |  | `10` |
| next_page | string (formatted string)| `string` |  | |  | `http://localhost:8080/users?next_token=ZmZmZmZmZmYtZmZmZi0tZmZmZmZmZmY=\u0026limit=10` |
| next_token | string (formatted string)| `string` |  | |  | `ZmZmZmZmZmYtZmZmZi0tZmZmZmZmZmY=` |
| prev_page | string (formatted string)| `string` |  | |  | `http://localhost:8080/users?prev_token=ZmZmZmZmZmYtZmZmZi0tZmZmZmZmZmY=\u0026limit=10` |
| prev_token | string (formatted string)| `string` |  | |  | `ZmZmZmZmZmYtZmZmZi0tZmZmZmZmZmY=` |
| size | int (formatted integer)| `int64` |  | |  | `10` |



### <span id="model-update-user-request"></span> model.UpdateUserRequest


> UpdateUserRequest represents the input for the UpdateUser method
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| disabled | boolean (formatted boolean)| `bool` |  | |  | `false` |
| email | email (formatted string)| `strfmt.Email` |  | |  | `my@email.com` |
| first_name | string (formatted string)| `string` |  | |  | `John` |
| last_name | string (formatted string)| `string` |  | |  | `Doe` |
| password | string (formatted string)| `string` |  | |  | `ThisIs4Passw0rd` |



### <span id="model-user"></span> model.User


> User represents a user entity
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| created_at | date-time (formatted string)| `strfmt.DateTime` |  | |  | `2021-01-01T00:00:00Z` |
| disabled | boolean (formatted boolean)| `bool` |  | |  | `false` |
| email | email (formatted string)| `strfmt.Email` |  | |  | `my@email.com` |
| first_name | string (formatted string)| `string` |  | |  | `John` |
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `550e8400-e29b-41d4-a716-446655440000` |
| last_name | string (formatted string)| `string` |  | |  | `Doe` |
| updated_at | date-time (formatted string)| `strfmt.DateTime` |  | |  | `2021-01-01T00:00:00Z` |



### <span id="model-version"></span> model.Version


> Version is the struct that holds the version information.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| build_date | string (formatted string)| `string` |  | |  | `2021-01-01T00:00:00Z` |
| git_branch | string (formatted string)| `string` |  | |  | `main` |
| git_commit | string (formatted string)| `string` |  | |  | `abcdef123456` |
| go_version | string (formatted string)| `string` |  | |  | `go1.24.1` |
| go_version_arch | string (formatted string)| `string` |  | |  | `amd64` |
| go_version_os | string (formatted string)| `string` |  | |  | `linux` |
| version | string (formatted string)| `string` |  | |  | `1.0.0` |


