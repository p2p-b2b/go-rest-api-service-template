


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

###  auth

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| POST | /auth/login | [019791cc 06c7 7d8c 81e4 914dd89098e8](#019791cc-06c7-7d8c-81e4-914dd89098e8) | Login user |
| POST | /auth/register | [019791cc 06c7 7e38 ac82 6919df575ff7](#019791cc-06c7-7e38-ac82-6919df575ff7) | Register user |
| GET | /auth/verify/{jwt} | [019791cc 06c7 7e40 bcc3 149cd13bee44](#019791cc-06c7-7e40-bcc3-149cd13bee44) | Verify user |
| POST | /auth/verify | [019791cc 06c7 7e48 b046 99f072088c50](#019791cc-06c7-7e48-b046-99f072088c50) | Resend verification |
| DELETE | /auth/logout | [019791cc 06c7 7e4c 961c b1bf5f40d633](#019791cc-06c7-7e4c-961c-b1bf5f40d633) | Logout user |
| POST | /auth/refresh | [019791cc 06c7 7e50 a5b6 bd1c82d5c031](#019791cc-06c7-7e50-a5b6-bd1c82d5c031) | Refresh access token |
  


###  health

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /health/status | [019791cc 06c7 7e57 8be6 a6f650ad5431](#019791cc-06c7-7e57-8be6-a6f650ad5431) | Check health |
  


###  policies

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /policies/{policy_id} | [019791cc 06c7 7e5b b363 1ef381f1e832](#019791cc-06c7-7e5b-b363-1ef381f1e832) | Get policy |
| POST | /policies | [019791cc 06c7 7e63 8ec9 de5b38235dbf](#019791cc-06c7-7e63-8ec9-de5b38235dbf) | Create policy |
| PUT | /policies/{policy_id} | [019791cc 06c7 7e67 9e1b 49a34edfe07c](#019791cc-06c7-7e67-9e1b-49a34edfe07c) | Update policy |
| DELETE | /policies/{policy_id} | [019791cc 06c7 7e6b b308 a2b2cbc2aaa1](#019791cc-06c7-7e6b-b308-a2b2cbc2aaa1) | Delete policy |
| GET | /policies | [019791cc 06c7 7e73 96aa 7e0383caae0d](#019791cc-06c7-7e73-96aa-7e0383caae0d) | List policies |
| POST | /policies/{policy_id}/roles | [019791cc 06c7 7e77 a2c3 4ed693a2bcdd](#019791cc-06c7-7e77-a2c3-4ed693a2bcdd) | Link roles to policy |
| DELETE | /policies/{policy_id}/roles | [019791cc 06c7 7e7b bdfc 381b015c44e7](#019791cc-06c7-7e7b-bdfc-381b015c44e7) | Unlink roles from policy |
| GET | /roles/{role_id}/policies | [019791cc 06c7 7e82 967d e13c399f5018](#019791cc-06c7-7e82-967d-e13c399f5018) | List policies by role |
  


###  products

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /projects/{project_id}/products/{product_id} | [01979db7 f53f 73a1 aab2 74802b79be51](#01979db7-f53f-73a1-aab2-74802b79be51) | Get product |
| POST | /projects/{project_id}/products | [01979db7 f53f 73a5 b916 297c6db5b714](#01979db7-f53f-73a5-b916-297c6db5b714) | Create product |
| PUT | /projects/{project_id}/products/{product_id} | [01979db7 f53f 73a9 bd58 e1cd5d7df436](#01979db7-f53f-73a9-bd58-e1cd5d7df436) | Update product |
| DELETE | /projects/{project_id}/products/{product_id} | [01979db7 f53f 73ad a84e 49bbdfe9e5c9](#01979db7-f53f-73ad-a84e-49bbdfe9e5c9) | Delete product |
| GET | /projects/{project_id}/products | [01979db7 f53f 73b1 993f 15f77e72c8cc](#01979db7-f53f-73b1-993f-15f77e72c8cc) | List products by project |
| GET | /products | [01979db7 f53f 73b5 a499 d6390831c94c](#01979db7-f53f-73b5-a499-d6390831c94c) | List products |
| DELETE | /projects/{project_id}/products/{product_id}/payment_processor | [01979db7 f53f 73b9 818f cdd1848f15d0](#01979db7-f53f-73b9-818f-cdd1848f15d0) | Unlink product from payment processor |
| POST | /projects/{project_id}/products/{product_id}/payment_processor | [01979db7 f53f 73bd b6c0 4541a48549c2](#01979db7-f53f-73bd-b6c0-4541a48549c2) | Link product to payment processor |
  


###  projects

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /projects | [019797e6 138a 7cf1 b7bb fa9c5e168c49](#019797e6-138a-7cf1-b7bb-fa9c5e168c49) | List projects |
| DELETE | /projects/{project_id} | [019797e6 138a 7cf4 8694 e4611baded39](#019797e6-138a-7cf4-8694-e4611baded39) | Delete project |
| PUT | /projects/{project_id} | [019797e6 138a 7cf8 9887 e4c44ad0ae19](#019797e6-138a-7cf8-9887-e4c44ad0ae19) | Update project |
| POST | /projects | [019797e6 138a 7d00 98db 740f21794f11](#019797e6-138a-7d00-98db-740f21794f11) | Create project |
| GET | /projects/{project_id} | [019797e6 138a 7d04 8db3 1d4755b25db3](#019797e6-138a-7d04-8db3-1d4755b25db3) | Get project |
  


###  resources

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /resources/{resource_id} | [019791cc 06c7 7e86 ad42 b777bfcc9e40](#019791cc-06c7-7e86-ad42-b777bfcc9e40) | Get resource |
| GET | /resources | [019791cc 06c7 7e8e 8d7e cd3f9296e0fd](#019791cc-06c7-7e8e-8d7e-cd3f9296e0fd) | List resources |
| GET | /resources/matches | [019791cc 06c7 7e92 9152 cb35902f79c4](#019791cc-06c7-7e92-9152-cb35902f79c4) | Match resources |
  


###  roles

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /roles/{role_id} | [019791cc 06c7 7e96 a284 ad37f86475bd](#019791cc-06c7-7e96-a284-ad37f86475bd) | Get role |
| POST | /roles | [019791cc 06c7 7e9e 87bf dcedfa5aefa7](#019791cc-06c7-7e9e-87bf-dcedfa5aefa7) | Create role |
| PUT | /roles/{role_id} | [019791cc 06c7 7ea2 8f0e 7b9f7cbc203a](#019791cc-06c7-7ea2-8f0e-7b9f7cbc203a) | Update role |
| DELETE | /roles/{role_id} | [019791cc 06c7 7ea6 9423 184c13540c26](#019791cc-06c7-7ea6-9423-184c13540c26) | Delete role |
| GET | /roles | [019791cc 06c7 7ead 968a 2a457714a7ee](#019791cc-06c7-7ead-968a-2a457714a7ee) | List roles |
| POST | /roles/{role_id}/users | [019791cc 06c7 7eb1 b10c 4fbcd5943885](#019791cc-06c7-7eb1-b10c-4fbcd5943885) | Link users to role |
| DELETE | /roles/{role_id}/users | [019791cc 06c7 7eb5 9b74 ba394be221b4](#019791cc-06c7-7eb5-9b74-ba394be221b4) | Unlink users from role |
| POST | /roles/{role_id}/policies | [019791cc 06c7 7ebd b750 8c93b165d503](#019791cc-06c7-7ebd-b750-8c93b165d503) | Link policies to role |
| DELETE | /roles/{role_id}/policies | [019791cc 06c7 7ec1 aca1 291132927db6](#019791cc-06c7-7ec1-aca1-291132927db6) | Unlink policies from role |
| GET | /users/{user_id}/roles | [019791cc 06c7 7ec5 87fe 096d6d2760a9](#019791cc-06c7-7ec5-87fe-096d6d2760a9) | List roles by user |
| GET | /policies/{policy_id}/roles | [019791cc 06c7 7ecd 93da d21bac8fd613](#019791cc-06c7-7ecd-93da-d21bac8fd613) | List roles by policy |
  


###  users

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /users/{user_id} | [019791cc 06c7 7ed0 9140 556f721c5749](#019791cc-06c7-7ed0-9140-556f721c5749) | Get user |
| POST | /users | [019791cc 06c7 7ed4 8f8b 2297e4565de3](#019791cc-06c7-7ed4-8f8b-2297e4565de3) | Create user |
| PUT | /users/{user_id} | [019791cc 06c7 7edc 94d6 843e3f99e96f](#019791cc-06c7-7edc-94d6-843e3f99e96f) | Update user |
| DELETE | /users/{user_id} | [019791cc 06c7 7ee0 85b7 45450ad476eb](#019791cc-06c7-7ee0-85b7-45450ad476eb) | Delete user |
| GET | /users | [019791cc 06c7 7ee4 8f2b ea43720d520b](#019791cc-06c7-7ee4-8f2b-ea43720d520b) | List users |
| POST | /users/{user_id}/roles | [019791cc 06c7 7eec 83f3 bcaed0c4d46f](#019791cc-06c7-7eec-83f3-bcaed0c4d46f) | Link roles to user |
| DELETE | /users/{user_id}/roles | [019791cc 06c7 7ef0 9394 d4ac3f52e94c](#019791cc-06c7-7ef0-9394-d4ac3f52e94c) | Unlink roles from user |
| GET | /users/{user_id}/authz | [019791cc 06c7 7ef4 afa5 81125e9dcde9](#019791cc-06c7-7ef4-afa5-81125e9dcde9) | Get user authorization |
| GET | /roles/{role_id}/users | [019791cc 06c7 7efb 99c2 b25af11e600c](#019791cc-06c7-7efb-99c2-b25af11e600c) | List users by role |
  


###  version

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /version | [019791cc 06c7 7eff a1df fbc2ad0b27c9](#019791cc-06c7-7eff-a1df-fbc2ad0b27c9) | Get version |
  


## Paths

### <span id="019791cc-06c7-7d8c-81e4-914dd89098e8"></span> Login user (*019791cc-06c7-7d8c-81e4-914dd89098e8*)

```
POST /auth/login
```

Authenticate user credentials and return JWT access and refresh tokens

#### Consumes
  * application/json

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| body | `body` | [ModelLoginUserRequest](#model-login-user-request) | `models.ModelLoginUserRequest` | | ✓ | | The information of the user to login |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7d8c-81e4-914dd89098e8-200) | OK | OK |  | [schema](#019791cc-06c7-7d8c-81e4-914dd89098e8-200-schema) |
| [400](#019791cc-06c7-7d8c-81e4-914dd89098e8-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7d8c-81e4-914dd89098e8-400-schema) |
| [401](#019791cc-06c7-7d8c-81e4-914dd89098e8-401) | Unauthorized | Unauthorized |  | [schema](#019791cc-06c7-7d8c-81e4-914dd89098e8-401-schema) |
| [500](#019791cc-06c7-7d8c-81e4-914dd89098e8-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7d8c-81e4-914dd89098e8-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7d8c-81e4-914dd89098e8-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7d8c-81e4-914dd89098e8-200-schema"></span> Schema
   
  

[ModelLoginUserResponse](#model-login-user-response)

##### <span id="019791cc-06c7-7d8c-81e4-914dd89098e8-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7d8c-81e4-914dd89098e8-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7d8c-81e4-914dd89098e8-401"></span> 401 - Unauthorized
Status: Unauthorized

###### <span id="019791cc-06c7-7d8c-81e4-914dd89098e8-401-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7d8c-81e4-914dd89098e8-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7d8c-81e4-914dd89098e8-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7e38-ac82-6919df575ff7"></span> Register user (*019791cc-06c7-7e38-ac82-6919df575ff7*)

```
POST /auth/register
```

Create a new user account and send email verification

#### Consumes
  * application/json

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| body | `body` | [ModelRegisterUserRequest](#model-register-user-request) | `models.ModelRegisterUserRequest` | | ✓ | | The information of the user to register |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [201](#019791cc-06c7-7e38-ac82-6919df575ff7-201) | Created | Created |  | [schema](#019791cc-06c7-7e38-ac82-6919df575ff7-201-schema) |
| [400](#019791cc-06c7-7e38-ac82-6919df575ff7-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7e38-ac82-6919df575ff7-400-schema) |
| [500](#019791cc-06c7-7e38-ac82-6919df575ff7-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7e38-ac82-6919df575ff7-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7e38-ac82-6919df575ff7-201"></span> 201 - Created
Status: Created

###### <span id="019791cc-06c7-7e38-ac82-6919df575ff7-201-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e38-ac82-6919df575ff7-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7e38-ac82-6919df575ff7-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e38-ac82-6919df575ff7-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7e38-ac82-6919df575ff7-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7e40-bcc3-149cd13bee44"></span> Verify user (*019791cc-06c7-7e40-bcc3-149cd13bee44*)

```
GET /auth/verify/{jwt}
```

Verify user account using JWT verification token

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| jwt | `path` | jwt (formatted string) | `string` |  | ✓ |  | The JWT token to use |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7e40-bcc3-149cd13bee44-200) | OK | OK |  | [schema](#019791cc-06c7-7e40-bcc3-149cd13bee44-200-schema) |
| [400](#019791cc-06c7-7e40-bcc3-149cd13bee44-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7e40-bcc3-149cd13bee44-400-schema) |
| [401](#019791cc-06c7-7e40-bcc3-149cd13bee44-401) | Unauthorized | Unauthorized |  | [schema](#019791cc-06c7-7e40-bcc3-149cd13bee44-401-schema) |
| [404](#019791cc-06c7-7e40-bcc3-149cd13bee44-404) | Not Found | Not Found |  | [schema](#019791cc-06c7-7e40-bcc3-149cd13bee44-404-schema) |
| [500](#019791cc-06c7-7e40-bcc3-149cd13bee44-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7e40-bcc3-149cd13bee44-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7e40-bcc3-149cd13bee44-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7e40-bcc3-149cd13bee44-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e40-bcc3-149cd13bee44-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7e40-bcc3-149cd13bee44-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e40-bcc3-149cd13bee44-401"></span> 401 - Unauthorized
Status: Unauthorized

###### <span id="019791cc-06c7-7e40-bcc3-149cd13bee44-401-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e40-bcc3-149cd13bee44-404"></span> 404 - Not Found
Status: Not Found

###### <span id="019791cc-06c7-7e40-bcc3-149cd13bee44-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e40-bcc3-149cd13bee44-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7e40-bcc3-149cd13bee44-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7e48-b046-99f072088c50"></span> Resend verification (*019791cc-06c7-7e48-b046-99f072088c50*)

```
POST /auth/verify
```

Resend account verification email to user

#### Consumes
  * application/json

#### Produces
  * application/json

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| body | `body` | [ModelReVerifyUserRequest](#model-re-verify-user-request) | `models.ModelReVerifyUserRequest` | | ✓ | | The email of the user to re-verify |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7e48-b046-99f072088c50-200) | OK | OK |  | [schema](#019791cc-06c7-7e48-b046-99f072088c50-200-schema) |
| [400](#019791cc-06c7-7e48-b046-99f072088c50-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7e48-b046-99f072088c50-400-schema) |
| [401](#019791cc-06c7-7e48-b046-99f072088c50-401) | Unauthorized | Unauthorized |  | [schema](#019791cc-06c7-7e48-b046-99f072088c50-401-schema) |
| [404](#019791cc-06c7-7e48-b046-99f072088c50-404) | Not Found | Not Found |  | [schema](#019791cc-06c7-7e48-b046-99f072088c50-404-schema) |
| [500](#019791cc-06c7-7e48-b046-99f072088c50-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7e48-b046-99f072088c50-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7e48-b046-99f072088c50-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7e48-b046-99f072088c50-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e48-b046-99f072088c50-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7e48-b046-99f072088c50-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e48-b046-99f072088c50-401"></span> 401 - Unauthorized
Status: Unauthorized

###### <span id="019791cc-06c7-7e48-b046-99f072088c50-401-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e48-b046-99f072088c50-404"></span> 404 - Not Found
Status: Not Found

###### <span id="019791cc-06c7-7e48-b046-99f072088c50-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e48-b046-99f072088c50-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7e48-b046-99f072088c50-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7e4c-961c-b1bf5f40d633"></span> Logout user (*019791cc-06c7-7e4c-961c-b1bf5f40d633*)

```
DELETE /auth/logout
```

Logout user and invalidate session tokens

#### Consumes
  * application/json

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7e4c-961c-b1bf5f40d633-200) | OK | OK |  | [schema](#019791cc-06c7-7e4c-961c-b1bf5f40d633-200-schema) |
| [401](#019791cc-06c7-7e4c-961c-b1bf5f40d633-401) | Unauthorized | Unauthorized |  | [schema](#019791cc-06c7-7e4c-961c-b1bf5f40d633-401-schema) |
| [500](#019791cc-06c7-7e4c-961c-b1bf5f40d633-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7e4c-961c-b1bf5f40d633-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7e4c-961c-b1bf5f40d633-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7e4c-961c-b1bf5f40d633-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e4c-961c-b1bf5f40d633-401"></span> 401 - Unauthorized
Status: Unauthorized

###### <span id="019791cc-06c7-7e4c-961c-b1bf5f40d633-401-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e4c-961c-b1bf5f40d633-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7e4c-961c-b1bf5f40d633-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7e50-a5b6-bd1c82d5c031"></span> Refresh access token (*019791cc-06c7-7e50-a5b6-bd1c82d5c031*)

```
POST /auth/refresh
```

Generate new access token using valid refresh token

#### Consumes
  * application/json

#### Produces
  * application/json

#### Security Requirements
  * RefreshToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| body | `body` | [ModelRefreshTokenRequest](#model-refresh-token-request) | `models.ModelRefreshTokenRequest` | | ✓ | | The refresh token to use |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7e50-a5b6-bd1c82d5c031-200) | OK | OK |  | [schema](#019791cc-06c7-7e50-a5b6-bd1c82d5c031-200-schema) |
| [400](#019791cc-06c7-7e50-a5b6-bd1c82d5c031-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7e50-a5b6-bd1c82d5c031-400-schema) |
| [401](#019791cc-06c7-7e50-a5b6-bd1c82d5c031-401) | Unauthorized | Unauthorized |  | [schema](#019791cc-06c7-7e50-a5b6-bd1c82d5c031-401-schema) |
| [404](#019791cc-06c7-7e50-a5b6-bd1c82d5c031-404) | Not Found | Not Found |  | [schema](#019791cc-06c7-7e50-a5b6-bd1c82d5c031-404-schema) |
| [500](#019791cc-06c7-7e50-a5b6-bd1c82d5c031-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7e50-a5b6-bd1c82d5c031-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7e50-a5b6-bd1c82d5c031-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7e50-a5b6-bd1c82d5c031-200-schema"></span> Schema
   
  

[ModelRefreshTokenResponse](#model-refresh-token-response)

##### <span id="019791cc-06c7-7e50-a5b6-bd1c82d5c031-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7e50-a5b6-bd1c82d5c031-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e50-a5b6-bd1c82d5c031-401"></span> 401 - Unauthorized
Status: Unauthorized

###### <span id="019791cc-06c7-7e50-a5b6-bd1c82d5c031-401-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e50-a5b6-bd1c82d5c031-404"></span> 404 - Not Found
Status: Not Found

###### <span id="019791cc-06c7-7e50-a5b6-bd1c82d5c031-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e50-a5b6-bd1c82d5c031-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7e50-a5b6-bd1c82d5c031-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7e57-8be6-a6f650ad5431"></span> Check health (*019791cc-06c7-7e57-8be6-a6f650ad5431*)

```
GET /health/status
```

Check service health status including database connectivity and system metrics

#### Produces
  * application/json

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7e57-8be6-a6f650ad5431-200) | OK | OK |  | [schema](#019791cc-06c7-7e57-8be6-a6f650ad5431-200-schema) |
| [500](#019791cc-06c7-7e57-8be6-a6f650ad5431-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7e57-8be6-a6f650ad5431-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7e57-8be6-a6f650ad5431-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7e57-8be6-a6f650ad5431-200-schema"></span> Schema
   
  

[ModelHealth](#model-health)

##### <span id="019791cc-06c7-7e57-8be6-a6f650ad5431-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7e57-8be6-a6f650ad5431-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7e5b-b363-1ef381f1e832"></span> Get policy (*019791cc-06c7-7e5b-b363-1ef381f1e832*)

```
GET /policies/{policy_id}
```

Retrieve a specific policy by its unique identifier

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| policy_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The policy id in UUID format |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7e5b-b363-1ef381f1e832-200) | OK | OK |  | [schema](#019791cc-06c7-7e5b-b363-1ef381f1e832-200-schema) |
| [400](#019791cc-06c7-7e5b-b363-1ef381f1e832-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7e5b-b363-1ef381f1e832-400-schema) |
| [404](#019791cc-06c7-7e5b-b363-1ef381f1e832-404) | Not Found | Not Found |  | [schema](#019791cc-06c7-7e5b-b363-1ef381f1e832-404-schema) |
| [500](#019791cc-06c7-7e5b-b363-1ef381f1e832-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7e5b-b363-1ef381f1e832-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7e5b-b363-1ef381f1e832-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7e5b-b363-1ef381f1e832-200-schema"></span> Schema
   
  

[ModelPolicy](#model-policy)

##### <span id="019791cc-06c7-7e5b-b363-1ef381f1e832-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7e5b-b363-1ef381f1e832-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e5b-b363-1ef381f1e832-404"></span> 404 - Not Found
Status: Not Found

###### <span id="019791cc-06c7-7e5b-b363-1ef381f1e832-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e5b-b363-1ef381f1e832-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7e5b-b363-1ef381f1e832-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7e63-8ec9-de5b38235dbf"></span> Create policy (*019791cc-06c7-7e63-8ec9-de5b38235dbf*)

```
POST /policies
```

Create a new policy with specified permissions

#### Consumes
  * application/json

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| body | `body` | [ModelCreatePolicyRequest](#model-create-policy-request) | `models.ModelCreatePolicyRequest` | | ✓ | | Create policy Request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [201](#019791cc-06c7-7e63-8ec9-de5b38235dbf-201) | Created | Created |  | [schema](#019791cc-06c7-7e63-8ec9-de5b38235dbf-201-schema) |
| [400](#019791cc-06c7-7e63-8ec9-de5b38235dbf-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7e63-8ec9-de5b38235dbf-400-schema) |
| [409](#019791cc-06c7-7e63-8ec9-de5b38235dbf-409) | Conflict | Conflict |  | [schema](#019791cc-06c7-7e63-8ec9-de5b38235dbf-409-schema) |
| [500](#019791cc-06c7-7e63-8ec9-de5b38235dbf-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7e63-8ec9-de5b38235dbf-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7e63-8ec9-de5b38235dbf-201"></span> 201 - Created
Status: Created

###### <span id="019791cc-06c7-7e63-8ec9-de5b38235dbf-201-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e63-8ec9-de5b38235dbf-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7e63-8ec9-de5b38235dbf-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e63-8ec9-de5b38235dbf-409"></span> 409 - Conflict
Status: Conflict

###### <span id="019791cc-06c7-7e63-8ec9-de5b38235dbf-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e63-8ec9-de5b38235dbf-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7e63-8ec9-de5b38235dbf-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7e67-9e1b-49a34edfe07c"></span> Update policy (*019791cc-06c7-7e67-9e1b-49a34edfe07c*)

```
PUT /policies/{policy_id}
```

Modify an existing policy by its ID

#### Consumes
  * application/json

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| policy_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The policy id in UUID format |
| body | `body` | [ModelUpdatePolicyRequest](#model-update-policy-request) | `models.ModelUpdatePolicyRequest` | | ✓ | | Update policy Request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7e67-9e1b-49a34edfe07c-200) | OK | OK |  | [schema](#019791cc-06c7-7e67-9e1b-49a34edfe07c-200-schema) |
| [400](#019791cc-06c7-7e67-9e1b-49a34edfe07c-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7e67-9e1b-49a34edfe07c-400-schema) |
| [404](#019791cc-06c7-7e67-9e1b-49a34edfe07c-404) | Not Found | Not Found |  | [schema](#019791cc-06c7-7e67-9e1b-49a34edfe07c-404-schema) |
| [500](#019791cc-06c7-7e67-9e1b-49a34edfe07c-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7e67-9e1b-49a34edfe07c-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7e67-9e1b-49a34edfe07c-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7e67-9e1b-49a34edfe07c-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e67-9e1b-49a34edfe07c-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7e67-9e1b-49a34edfe07c-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e67-9e1b-49a34edfe07c-404"></span> 404 - Not Found
Status: Not Found

###### <span id="019791cc-06c7-7e67-9e1b-49a34edfe07c-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e67-9e1b-49a34edfe07c-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7e67-9e1b-49a34edfe07c-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7e6b-b308-a2b2cbc2aaa1"></span> Delete policy (*019791cc-06c7-7e6b-b308-a2b2cbc2aaa1*)

```
DELETE /policies/{policy_id}
```

Remove a policy permanently from the system

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| policy_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The policy id in UUID format |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7e6b-b308-a2b2cbc2aaa1-200) | OK | OK |  | [schema](#019791cc-06c7-7e6b-b308-a2b2cbc2aaa1-200-schema) |
| [400](#019791cc-06c7-7e6b-b308-a2b2cbc2aaa1-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7e6b-b308-a2b2cbc2aaa1-400-schema) |
| [404](#019791cc-06c7-7e6b-b308-a2b2cbc2aaa1-404) | Not Found | Not Found |  | [schema](#019791cc-06c7-7e6b-b308-a2b2cbc2aaa1-404-schema) |
| [500](#019791cc-06c7-7e6b-b308-a2b2cbc2aaa1-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7e6b-b308-a2b2cbc2aaa1-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7e6b-b308-a2b2cbc2aaa1-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7e6b-b308-a2b2cbc2aaa1-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e6b-b308-a2b2cbc2aaa1-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7e6b-b308-a2b2cbc2aaa1-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e6b-b308-a2b2cbc2aaa1-404"></span> 404 - Not Found
Status: Not Found

###### <span id="019791cc-06c7-7e6b-b308-a2b2cbc2aaa1-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e6b-b308-a2b2cbc2aaa1-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7e6b-b308-a2b2cbc2aaa1-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7e73-96aa-7e0383caae0d"></span> List policies (*019791cc-06c7-7e73-96aa-7e0383caae0d*)

```
GET /policies
```

Retrieve paginated list of all policies in the system

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

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
| [200](#019791cc-06c7-7e73-96aa-7e0383caae0d-200) | OK | OK |  | [schema](#019791cc-06c7-7e73-96aa-7e0383caae0d-200-schema) |
| [400](#019791cc-06c7-7e73-96aa-7e0383caae0d-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7e73-96aa-7e0383caae0d-400-schema) |
| [500](#019791cc-06c7-7e73-96aa-7e0383caae0d-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7e73-96aa-7e0383caae0d-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7e73-96aa-7e0383caae0d-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7e73-96aa-7e0383caae0d-200-schema"></span> Schema
   
  

[ModelListPoliciesResponse](#model-list-policies-response)

##### <span id="019791cc-06c7-7e73-96aa-7e0383caae0d-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7e73-96aa-7e0383caae0d-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e73-96aa-7e0383caae0d-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7e73-96aa-7e0383caae0d-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7e77-a2c3-4ed693a2bcdd"></span> Link roles to policy (*019791cc-06c7-7e77-a2c3-4ed693a2bcdd*)

```
POST /policies/{policy_id}/roles
```

Associate multiple roles with a specific policy for authorization

#### Consumes
  * application/json

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| policy_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The policy id in UUID format |
| body | `body` | [ModelLinkRolesToPolicyRequest](#model-link-roles-to-policy-request) | `models.ModelLinkRolesToPolicyRequest` | | ✓ | | Link policy to roles Request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7e77-a2c3-4ed693a2bcdd-200) | OK | OK |  | [schema](#019791cc-06c7-7e77-a2c3-4ed693a2bcdd-200-schema) |
| [400](#019791cc-06c7-7e77-a2c3-4ed693a2bcdd-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7e77-a2c3-4ed693a2bcdd-400-schema) |
| [404](#019791cc-06c7-7e77-a2c3-4ed693a2bcdd-404) | Not Found | Not Found |  | [schema](#019791cc-06c7-7e77-a2c3-4ed693a2bcdd-404-schema) |
| [500](#019791cc-06c7-7e77-a2c3-4ed693a2bcdd-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7e77-a2c3-4ed693a2bcdd-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7e77-a2c3-4ed693a2bcdd-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7e77-a2c3-4ed693a2bcdd-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e77-a2c3-4ed693a2bcdd-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7e77-a2c3-4ed693a2bcdd-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e77-a2c3-4ed693a2bcdd-404"></span> 404 - Not Found
Status: Not Found

###### <span id="019791cc-06c7-7e77-a2c3-4ed693a2bcdd-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e77-a2c3-4ed693a2bcdd-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7e77-a2c3-4ed693a2bcdd-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7e7b-bdfc-381b015c44e7"></span> Unlink roles from policy (*019791cc-06c7-7e7b-bdfc-381b015c44e7*)

```
DELETE /policies/{policy_id}/roles
```

Remove role associations from a specific policy

#### Consumes
  * application/json

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| policy_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The policy id in UUID format |
| body | `body` | [ModelUnlinkRolesFromPolicyRequest](#model-unlink-roles-from-policy-request) | `models.ModelUnlinkRolesFromPolicyRequest` | | ✓ | | Unlink policy from roles Request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7e7b-bdfc-381b015c44e7-200) | OK | OK |  | [schema](#019791cc-06c7-7e7b-bdfc-381b015c44e7-200-schema) |
| [400](#019791cc-06c7-7e7b-bdfc-381b015c44e7-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7e7b-bdfc-381b015c44e7-400-schema) |
| [404](#019791cc-06c7-7e7b-bdfc-381b015c44e7-404) | Not Found | Not Found |  | [schema](#019791cc-06c7-7e7b-bdfc-381b015c44e7-404-schema) |
| [500](#019791cc-06c7-7e7b-bdfc-381b015c44e7-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7e7b-bdfc-381b015c44e7-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7e7b-bdfc-381b015c44e7-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7e7b-bdfc-381b015c44e7-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e7b-bdfc-381b015c44e7-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7e7b-bdfc-381b015c44e7-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e7b-bdfc-381b015c44e7-404"></span> 404 - Not Found
Status: Not Found

###### <span id="019791cc-06c7-7e7b-bdfc-381b015c44e7-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e7b-bdfc-381b015c44e7-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7e7b-bdfc-381b015c44e7-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7e82-967d-e13c399f5018"></span> List policies by role (*019791cc-06c7-7e82-967d-e13c399f5018*)

```
GET /roles/{role_id}/policies
```

Retrieve paginated list of policies associated with a specific role

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| role_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The role id in UUID format |
| fields | `query` | string (formatted string) | `string` |  |  |  | Fields to return. Example: id,first_name,last_name |
| filter | `query` | string (formatted string) | `string` |  |  |  | Filter field. Example: id=1 AND first_name='John' |
| limit | `query` | int (formatted integer) | `int64` |  |  |  | Limit |
| next_token | `query` | string (formatted string) | `string` |  |  |  | Next cursor |
| prev_token | `query` | string (formatted string) | `string` |  |  |  | Previous cursor |
| sort | `query` | string (formatted string) | `string` |  |  |  | Comma-separated list of fields to sort by. Example: first_name ASC, created_at DESC |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7e82-967d-e13c399f5018-200) | OK | OK |  | [schema](#019791cc-06c7-7e82-967d-e13c399f5018-200-schema) |
| [400](#019791cc-06c7-7e82-967d-e13c399f5018-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7e82-967d-e13c399f5018-400-schema) |
| [500](#019791cc-06c7-7e82-967d-e13c399f5018-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7e82-967d-e13c399f5018-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7e82-967d-e13c399f5018-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7e82-967d-e13c399f5018-200-schema"></span> Schema
   
  

[ModelListPoliciesResponse](#model-list-policies-response)

##### <span id="019791cc-06c7-7e82-967d-e13c399f5018-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7e82-967d-e13c399f5018-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e82-967d-e13c399f5018-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7e82-967d-e13c399f5018-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7e86-ad42-b777bfcc9e40"></span> Get resource (*019791cc-06c7-7e86-ad42-b777bfcc9e40*)

```
GET /resources/{resource_id}
```

Retrieve a specific resource by its identifier

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| resource_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The permission id in UUID format |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7e86-ad42-b777bfcc9e40-200) | OK | OK |  | [schema](#019791cc-06c7-7e86-ad42-b777bfcc9e40-200-schema) |
| [400](#019791cc-06c7-7e86-ad42-b777bfcc9e40-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7e86-ad42-b777bfcc9e40-400-schema) |
| [404](#019791cc-06c7-7e86-ad42-b777bfcc9e40-404) | Not Found | Not Found |  | [schema](#019791cc-06c7-7e86-ad42-b777bfcc9e40-404-schema) |
| [500](#019791cc-06c7-7e86-ad42-b777bfcc9e40-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7e86-ad42-b777bfcc9e40-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7e86-ad42-b777bfcc9e40-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7e86-ad42-b777bfcc9e40-200-schema"></span> Schema
   
  

[ModelResource](#model-resource)

##### <span id="019791cc-06c7-7e86-ad42-b777bfcc9e40-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7e86-ad42-b777bfcc9e40-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e86-ad42-b777bfcc9e40-404"></span> 404 - Not Found
Status: Not Found

###### <span id="019791cc-06c7-7e86-ad42-b777bfcc9e40-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e86-ad42-b777bfcc9e40-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7e86-ad42-b777bfcc9e40-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7e8e-8d7e-cd3f9296e0fd"></span> List resources (*019791cc-06c7-7e8e-8d7e-cd3f9296e0fd*)

```
GET /resources
```

Retrieve paginated list of all resources in the system

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

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
| [200](#019791cc-06c7-7e8e-8d7e-cd3f9296e0fd-200) | OK | OK |  | [schema](#019791cc-06c7-7e8e-8d7e-cd3f9296e0fd-200-schema) |
| [400](#019791cc-06c7-7e8e-8d7e-cd3f9296e0fd-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7e8e-8d7e-cd3f9296e0fd-400-schema) |
| [500](#019791cc-06c7-7e8e-8d7e-cd3f9296e0fd-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7e8e-8d7e-cd3f9296e0fd-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7e8e-8d7e-cd3f9296e0fd-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7e8e-8d7e-cd3f9296e0fd-200-schema"></span> Schema
   
  

[ModelListResourcesResponse](#model-list-resources-response)

##### <span id="019791cc-06c7-7e8e-8d7e-cd3f9296e0fd-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7e8e-8d7e-cd3f9296e0fd-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e8e-8d7e-cd3f9296e0fd-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7e8e-8d7e-cd3f9296e0fd-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7e92-9152-cb35902f79c4"></span> Match resources (*019791cc-06c7-7e92-9152-cb35902f79c4*)

```
GET /resources/matches
```

Find resources that match specific action and resource policy patterns

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| action | `query` | string (formatted string) | `string` |  | ✓ |  | Action to filter by |
| fields | `query` | string (formatted string) | `string` |  |  |  | Fields to return. Example: id,first_name,last_name |
| limit | `query` | int (formatted integer) | `int64` |  |  |  | Limit |
| next_token | `query` | string (formatted string) | `string` |  |  |  | Next cursor |
| prev_token | `query` | string (formatted string) | `string` |  |  |  | Previous cursor |
| resource | `query` | string (formatted string) | `string` |  | ✓ |  | Resource to filter by |
| sort | `query` | string (formatted string) | `string` |  |  |  | Comma-separated list of fields to sort by. Example: first_name ASC, created_at DESC |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7e92-9152-cb35902f79c4-200) | OK | OK |  | [schema](#019791cc-06c7-7e92-9152-cb35902f79c4-200-schema) |
| [400](#019791cc-06c7-7e92-9152-cb35902f79c4-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7e92-9152-cb35902f79c4-400-schema) |
| [500](#019791cc-06c7-7e92-9152-cb35902f79c4-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7e92-9152-cb35902f79c4-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7e92-9152-cb35902f79c4-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7e92-9152-cb35902f79c4-200-schema"></span> Schema
   
  

[ModelListResourcesResponse](#model-list-resources-response)

##### <span id="019791cc-06c7-7e92-9152-cb35902f79c4-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7e92-9152-cb35902f79c4-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e92-9152-cb35902f79c4-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7e92-9152-cb35902f79c4-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7e96-a284-ad37f86475bd"></span> Get role (*019791cc-06c7-7e96-a284-ad37f86475bd*)

```
GET /roles/{role_id}
```

Retrieve a specific role by its unique identifier

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| role_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The role id in UUID format |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7e96-a284-ad37f86475bd-200) | OK | OK |  | [schema](#019791cc-06c7-7e96-a284-ad37f86475bd-200-schema) |
| [400](#019791cc-06c7-7e96-a284-ad37f86475bd-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7e96-a284-ad37f86475bd-400-schema) |
| [404](#019791cc-06c7-7e96-a284-ad37f86475bd-404) | Not Found | Not Found |  | [schema](#019791cc-06c7-7e96-a284-ad37f86475bd-404-schema) |
| [500](#019791cc-06c7-7e96-a284-ad37f86475bd-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7e96-a284-ad37f86475bd-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7e96-a284-ad37f86475bd-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7e96-a284-ad37f86475bd-200-schema"></span> Schema
   
  

[ModelRole](#model-role)

##### <span id="019791cc-06c7-7e96-a284-ad37f86475bd-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7e96-a284-ad37f86475bd-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e96-a284-ad37f86475bd-404"></span> 404 - Not Found
Status: Not Found

###### <span id="019791cc-06c7-7e96-a284-ad37f86475bd-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e96-a284-ad37f86475bd-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7e96-a284-ad37f86475bd-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7e9e-87bf-dcedfa5aefa7"></span> Create role (*019791cc-06c7-7e9e-87bf-dcedfa5aefa7*)

```
POST /roles
```

Create a new role with specified permissions and access levels

#### Consumes
  * application/json

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| body | `body` | [ModelCreateRoleRequest](#model-create-role-request) | `models.ModelCreateRoleRequest` | | ✓ | | Create role request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [201](#019791cc-06c7-7e9e-87bf-dcedfa5aefa7-201) | Created | Created |  | [schema](#019791cc-06c7-7e9e-87bf-dcedfa5aefa7-201-schema) |
| [400](#019791cc-06c7-7e9e-87bf-dcedfa5aefa7-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7e9e-87bf-dcedfa5aefa7-400-schema) |
| [409](#019791cc-06c7-7e9e-87bf-dcedfa5aefa7-409) | Conflict | Conflict |  | [schema](#019791cc-06c7-7e9e-87bf-dcedfa5aefa7-409-schema) |
| [500](#019791cc-06c7-7e9e-87bf-dcedfa5aefa7-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7e9e-87bf-dcedfa5aefa7-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7e9e-87bf-dcedfa5aefa7-201"></span> 201 - Created
Status: Created

###### <span id="019791cc-06c7-7e9e-87bf-dcedfa5aefa7-201-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e9e-87bf-dcedfa5aefa7-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7e9e-87bf-dcedfa5aefa7-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e9e-87bf-dcedfa5aefa7-409"></span> 409 - Conflict
Status: Conflict

###### <span id="019791cc-06c7-7e9e-87bf-dcedfa5aefa7-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7e9e-87bf-dcedfa5aefa7-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7e9e-87bf-dcedfa5aefa7-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7ea2-8f0e-7b9f7cbc203a"></span> Update role (*019791cc-06c7-7ea2-8f0e-7b9f7cbc203a*)

```
PUT /roles/{role_id}
```

Modify an existing role by its ID

#### Consumes
  * application/json

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| role_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The model id in UUID format |
| body | `body` | [ModelUpdateRoleRequest](#model-update-role-request) | `models.ModelUpdateRoleRequest` | | ✓ | | Update role request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7ea2-8f0e-7b9f7cbc203a-200) | OK | OK |  | [schema](#019791cc-06c7-7ea2-8f0e-7b9f7cbc203a-200-schema) |
| [400](#019791cc-06c7-7ea2-8f0e-7b9f7cbc203a-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7ea2-8f0e-7b9f7cbc203a-400-schema) |
| [409](#019791cc-06c7-7ea2-8f0e-7b9f7cbc203a-409) | Conflict | Conflict |  | [schema](#019791cc-06c7-7ea2-8f0e-7b9f7cbc203a-409-schema) |
| [500](#019791cc-06c7-7ea2-8f0e-7b9f7cbc203a-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7ea2-8f0e-7b9f7cbc203a-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7ea2-8f0e-7b9f7cbc203a-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7ea2-8f0e-7b9f7cbc203a-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ea2-8f0e-7b9f7cbc203a-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7ea2-8f0e-7b9f7cbc203a-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ea2-8f0e-7b9f7cbc203a-409"></span> 409 - Conflict
Status: Conflict

###### <span id="019791cc-06c7-7ea2-8f0e-7b9f7cbc203a-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ea2-8f0e-7b9f7cbc203a-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7ea2-8f0e-7b9f7cbc203a-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7ea6-9423-184c13540c26"></span> Delete role (*019791cc-06c7-7ea6-9423-184c13540c26*)

```
DELETE /roles/{role_id}
```

Remove a role permanently from the system

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| role_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The role id in UUID format |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7ea6-9423-184c13540c26-200) | OK | OK |  | [schema](#019791cc-06c7-7ea6-9423-184c13540c26-200-schema) |
| [400](#019791cc-06c7-7ea6-9423-184c13540c26-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7ea6-9423-184c13540c26-400-schema) |
| [500](#019791cc-06c7-7ea6-9423-184c13540c26-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7ea6-9423-184c13540c26-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7ea6-9423-184c13540c26-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7ea6-9423-184c13540c26-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ea6-9423-184c13540c26-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7ea6-9423-184c13540c26-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ea6-9423-184c13540c26-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7ea6-9423-184c13540c26-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7ead-968a-2a457714a7ee"></span> List roles (*019791cc-06c7-7ead-968a-2a457714a7ee*)

```
GET /roles
```

Retrieve paginated list of all roles in the system

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

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
| [200](#019791cc-06c7-7ead-968a-2a457714a7ee-200) | OK | OK |  | [schema](#019791cc-06c7-7ead-968a-2a457714a7ee-200-schema) |
| [400](#019791cc-06c7-7ead-968a-2a457714a7ee-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7ead-968a-2a457714a7ee-400-schema) |
| [500](#019791cc-06c7-7ead-968a-2a457714a7ee-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7ead-968a-2a457714a7ee-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7ead-968a-2a457714a7ee-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7ead-968a-2a457714a7ee-200-schema"></span> Schema
   
  

[ModelListRolesResponse](#model-list-roles-response)

##### <span id="019791cc-06c7-7ead-968a-2a457714a7ee-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7ead-968a-2a457714a7ee-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ead-968a-2a457714a7ee-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7ead-968a-2a457714a7ee-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7eb1-b10c-4fbcd5943885"></span> Link users to role (*019791cc-06c7-7eb1-b10c-4fbcd5943885*)

```
POST /roles/{role_id}/users
```

Associate multiple users with a specific role for authorization

#### Consumes
  * application/json

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| role_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The role id in UUID format |
| body | `body` | [ModelLinkUsersToRoleRequest](#model-link-users-to-role-request) | `models.ModelLinkUsersToRoleRequest` | | ✓ | | Link users to role request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7eb1-b10c-4fbcd5943885-200) | OK | OK |  | [schema](#019791cc-06c7-7eb1-b10c-4fbcd5943885-200-schema) |
| [400](#019791cc-06c7-7eb1-b10c-4fbcd5943885-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7eb1-b10c-4fbcd5943885-400-schema) |
| [409](#019791cc-06c7-7eb1-b10c-4fbcd5943885-409) | Conflict | Conflict |  | [schema](#019791cc-06c7-7eb1-b10c-4fbcd5943885-409-schema) |
| [500](#019791cc-06c7-7eb1-b10c-4fbcd5943885-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7eb1-b10c-4fbcd5943885-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7eb1-b10c-4fbcd5943885-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7eb1-b10c-4fbcd5943885-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7eb1-b10c-4fbcd5943885-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7eb1-b10c-4fbcd5943885-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7eb1-b10c-4fbcd5943885-409"></span> 409 - Conflict
Status: Conflict

###### <span id="019791cc-06c7-7eb1-b10c-4fbcd5943885-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7eb1-b10c-4fbcd5943885-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7eb1-b10c-4fbcd5943885-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7eb5-9b74-ba394be221b4"></span> Unlink users from role (*019791cc-06c7-7eb5-9b74-ba394be221b4*)

```
DELETE /roles/{role_id}/users
```

Remove user associations from a specific role

#### Consumes
  * application/json

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| role_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The Embeddings Role ID in UUID format |
| body | `body` | [ModelUnlinkUsersFromRoleRequest](#model-unlink-users-from-role-request) | `models.ModelUnlinkUsersFromRoleRequest` | | ✓ | | UnLink users from role request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7eb5-9b74-ba394be221b4-200) | OK | OK |  | [schema](#019791cc-06c7-7eb5-9b74-ba394be221b4-200-schema) |
| [400](#019791cc-06c7-7eb5-9b74-ba394be221b4-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7eb5-9b74-ba394be221b4-400-schema) |
| [409](#019791cc-06c7-7eb5-9b74-ba394be221b4-409) | Conflict | Conflict |  | [schema](#019791cc-06c7-7eb5-9b74-ba394be221b4-409-schema) |
| [500](#019791cc-06c7-7eb5-9b74-ba394be221b4-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7eb5-9b74-ba394be221b4-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7eb5-9b74-ba394be221b4-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7eb5-9b74-ba394be221b4-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7eb5-9b74-ba394be221b4-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7eb5-9b74-ba394be221b4-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7eb5-9b74-ba394be221b4-409"></span> 409 - Conflict
Status: Conflict

###### <span id="019791cc-06c7-7eb5-9b74-ba394be221b4-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7eb5-9b74-ba394be221b4-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7eb5-9b74-ba394be221b4-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7ebd-b750-8c93b165d503"></span> Link policies to role (*019791cc-06c7-7ebd-b750-8c93b165d503*)

```
POST /roles/{role_id}/policies
```

Associate multiple policies with a specific role for authorization

#### Consumes
  * application/json

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| role_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The role id in UUID format |
| body | `body` | [ModelLinkPoliciesToRoleRequest](#model-link-policies-to-role-request) | `models.ModelLinkPoliciesToRoleRequest` | | ✓ | | Link policies to role request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7ebd-b750-8c93b165d503-200) | OK | OK |  | [schema](#019791cc-06c7-7ebd-b750-8c93b165d503-200-schema) |
| [400](#019791cc-06c7-7ebd-b750-8c93b165d503-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7ebd-b750-8c93b165d503-400-schema) |
| [409](#019791cc-06c7-7ebd-b750-8c93b165d503-409) | Conflict | Conflict |  | [schema](#019791cc-06c7-7ebd-b750-8c93b165d503-409-schema) |
| [500](#019791cc-06c7-7ebd-b750-8c93b165d503-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7ebd-b750-8c93b165d503-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7ebd-b750-8c93b165d503-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7ebd-b750-8c93b165d503-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ebd-b750-8c93b165d503-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7ebd-b750-8c93b165d503-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ebd-b750-8c93b165d503-409"></span> 409 - Conflict
Status: Conflict

###### <span id="019791cc-06c7-7ebd-b750-8c93b165d503-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ebd-b750-8c93b165d503-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7ebd-b750-8c93b165d503-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7ec1-aca1-291132927db6"></span> Unlink policies from role (*019791cc-06c7-7ec1-aca1-291132927db6*)

```
DELETE /roles/{role_id}/policies
```

Remove policy associations from a specific role

#### Consumes
  * application/json

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| role_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The role id in UUID format |
| body | `body` | [ModelUnlinkPoliciesFromRoleRequest](#model-unlink-policies-from-role-request) | `models.ModelUnlinkPoliciesFromRoleRequest` | | ✓ | | UnLink policies from role request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7ec1-aca1-291132927db6-200) | OK | OK |  | [schema](#019791cc-06c7-7ec1-aca1-291132927db6-200-schema) |
| [400](#019791cc-06c7-7ec1-aca1-291132927db6-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7ec1-aca1-291132927db6-400-schema) |
| [409](#019791cc-06c7-7ec1-aca1-291132927db6-409) | Conflict | Conflict |  | [schema](#019791cc-06c7-7ec1-aca1-291132927db6-409-schema) |
| [500](#019791cc-06c7-7ec1-aca1-291132927db6-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7ec1-aca1-291132927db6-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7ec1-aca1-291132927db6-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7ec1-aca1-291132927db6-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ec1-aca1-291132927db6-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7ec1-aca1-291132927db6-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ec1-aca1-291132927db6-409"></span> 409 - Conflict
Status: Conflict

###### <span id="019791cc-06c7-7ec1-aca1-291132927db6-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ec1-aca1-291132927db6-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7ec1-aca1-291132927db6-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7ec5-87fe-096d6d2760a9"></span> List roles by user (*019791cc-06c7-7ec5-87fe-096d6d2760a9*)

```
GET /users/{user_id}/roles
```

Retrieve paginated list of roles assigned to a specific user

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| user_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The user id in UUID format |
| fields | `query` | string (formatted string) | `string` |  |  |  | Fields to return. Example: id,first_name,last_name |
| filter | `query` | string (formatted string) | `string` |  |  |  | Filter field. Example: id=1 AND first_name='John' |
| limit | `query` | int (formatted integer) | `int64` |  |  |  | Limit |
| next_token | `query` | string (formatted string) | `string` |  |  |  | Next cursor |
| prev_token | `query` | string (formatted string) | `string` |  |  |  | Previous cursor |
| sort | `query` | string (formatted string) | `string` |  |  |  | Comma-separated list of fields to sort by. Example: first_name ASC, created_at DESC |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7ec5-87fe-096d6d2760a9-200) | OK | OK |  | [schema](#019791cc-06c7-7ec5-87fe-096d6d2760a9-200-schema) |
| [400](#019791cc-06c7-7ec5-87fe-096d6d2760a9-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7ec5-87fe-096d6d2760a9-400-schema) |
| [500](#019791cc-06c7-7ec5-87fe-096d6d2760a9-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7ec5-87fe-096d6d2760a9-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7ec5-87fe-096d6d2760a9-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7ec5-87fe-096d6d2760a9-200-schema"></span> Schema
   
  

[ModelListRolesResponse](#model-list-roles-response)

##### <span id="019791cc-06c7-7ec5-87fe-096d6d2760a9-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7ec5-87fe-096d6d2760a9-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ec5-87fe-096d6d2760a9-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7ec5-87fe-096d6d2760a9-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7ecd-93da-d21bac8fd613"></span> List roles by policy (*019791cc-06c7-7ecd-93da-d21bac8fd613*)

```
GET /policies/{policy_id}/roles
```

Retrieve paginated list of roles associated with a specific policy

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| policy_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The policy id in UUID format |
| fields | `query` | string (formatted string) | `string` |  |  |  | Fields to return. Example: id,first_name,last_name |
| filter | `query` | string (formatted string) | `string` |  |  |  | Filter field. Example: id=1 AND first_name='John' |
| limit | `query` | int (formatted integer) | `int64` |  |  |  | Limit |
| next_token | `query` | string (formatted string) | `string` |  |  |  | Next cursor |
| prev_token | `query` | string (formatted string) | `string` |  |  |  | Previous cursor |
| sort | `query` | string (formatted string) | `string` |  |  |  | Comma-separated list of fields to sort by. Example: first_name ASC, created_at DESC |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7ecd-93da-d21bac8fd613-200) | OK | OK |  | [schema](#019791cc-06c7-7ecd-93da-d21bac8fd613-200-schema) |
| [400](#019791cc-06c7-7ecd-93da-d21bac8fd613-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7ecd-93da-d21bac8fd613-400-schema) |
| [500](#019791cc-06c7-7ecd-93da-d21bac8fd613-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7ecd-93da-d21bac8fd613-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7ecd-93da-d21bac8fd613-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7ecd-93da-d21bac8fd613-200-schema"></span> Schema
   
  

[ModelListRolesResponse](#model-list-roles-response)

##### <span id="019791cc-06c7-7ecd-93da-d21bac8fd613-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7ecd-93da-d21bac8fd613-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ecd-93da-d21bac8fd613-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7ecd-93da-d21bac8fd613-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7ed0-9140-556f721c5749"></span> Get user (*019791cc-06c7-7ed0-9140-556f721c5749*)

```
GET /users/{user_id}
```

Retrieve a specific user account by its unique identifier

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| user_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The user ID in UUID format |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7ed0-9140-556f721c5749-200) | OK | OK |  | [schema](#019791cc-06c7-7ed0-9140-556f721c5749-200-schema) |
| [400](#019791cc-06c7-7ed0-9140-556f721c5749-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7ed0-9140-556f721c5749-400-schema) |
| [404](#019791cc-06c7-7ed0-9140-556f721c5749-404) | Not Found | Not Found |  | [schema](#019791cc-06c7-7ed0-9140-556f721c5749-404-schema) |
| [500](#019791cc-06c7-7ed0-9140-556f721c5749-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7ed0-9140-556f721c5749-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7ed0-9140-556f721c5749-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7ed0-9140-556f721c5749-200-schema"></span> Schema
   
  

[ModelUser](#model-user)

##### <span id="019791cc-06c7-7ed0-9140-556f721c5749-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7ed0-9140-556f721c5749-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ed0-9140-556f721c5749-404"></span> 404 - Not Found
Status: Not Found

###### <span id="019791cc-06c7-7ed0-9140-556f721c5749-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ed0-9140-556f721c5749-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7ed0-9140-556f721c5749-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7ed4-8f8b-2297e4565de3"></span> Create user (*019791cc-06c7-7ed4-8f8b-2297e4565de3*)

```
POST /users
```

Create a new user account with specified configuration

#### Consumes
  * application/json

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| body | `body` | [ModelCreateUserRequest](#model-create-user-request) | `models.ModelCreateUserRequest` | | ✓ | | Create user request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [201](#019791cc-06c7-7ed4-8f8b-2297e4565de3-201) | Created | Created |  | [schema](#019791cc-06c7-7ed4-8f8b-2297e4565de3-201-schema) |
| [400](#019791cc-06c7-7ed4-8f8b-2297e4565de3-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7ed4-8f8b-2297e4565de3-400-schema) |
| [409](#019791cc-06c7-7ed4-8f8b-2297e4565de3-409) | Conflict | Conflict |  | [schema](#019791cc-06c7-7ed4-8f8b-2297e4565de3-409-schema) |
| [500](#019791cc-06c7-7ed4-8f8b-2297e4565de3-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7ed4-8f8b-2297e4565de3-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7ed4-8f8b-2297e4565de3-201"></span> 201 - Created
Status: Created

###### <span id="019791cc-06c7-7ed4-8f8b-2297e4565de3-201-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ed4-8f8b-2297e4565de3-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7ed4-8f8b-2297e4565de3-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ed4-8f8b-2297e4565de3-409"></span> 409 - Conflict
Status: Conflict

###### <span id="019791cc-06c7-7ed4-8f8b-2297e4565de3-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ed4-8f8b-2297e4565de3-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7ed4-8f8b-2297e4565de3-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7edc-94d6-843e3f99e96f"></span> Update user (*019791cc-06c7-7edc-94d6-843e3f99e96f*)

```
PUT /users/{user_id}
```

Modify an existing user account by its ID

#### Consumes
  * application/json

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| user_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The user ID in UUID format |
| body | `body` | [ModelUpdateUserRequest](#model-update-user-request) | `models.ModelUpdateUserRequest` | | ✓ | | Update user request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7edc-94d6-843e3f99e96f-200) | OK | OK |  | [schema](#019791cc-06c7-7edc-94d6-843e3f99e96f-200-schema) |
| [400](#019791cc-06c7-7edc-94d6-843e3f99e96f-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7edc-94d6-843e3f99e96f-400-schema) |
| [409](#019791cc-06c7-7edc-94d6-843e3f99e96f-409) | Conflict | Conflict |  | [schema](#019791cc-06c7-7edc-94d6-843e3f99e96f-409-schema) |
| [500](#019791cc-06c7-7edc-94d6-843e3f99e96f-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7edc-94d6-843e3f99e96f-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7edc-94d6-843e3f99e96f-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7edc-94d6-843e3f99e96f-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7edc-94d6-843e3f99e96f-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7edc-94d6-843e3f99e96f-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7edc-94d6-843e3f99e96f-409"></span> 409 - Conflict
Status: Conflict

###### <span id="019791cc-06c7-7edc-94d6-843e3f99e96f-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7edc-94d6-843e3f99e96f-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7edc-94d6-843e3f99e96f-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7ee0-85b7-45450ad476eb"></span> Delete user (*019791cc-06c7-7ee0-85b7-45450ad476eb*)

```
DELETE /users/{user_id}
```

Remove a user account permanently from the system

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| user_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The user ID in UUID format |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7ee0-85b7-45450ad476eb-200) | OK | OK |  | [schema](#019791cc-06c7-7ee0-85b7-45450ad476eb-200-schema) |
| [400](#019791cc-06c7-7ee0-85b7-45450ad476eb-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7ee0-85b7-45450ad476eb-400-schema) |
| [500](#019791cc-06c7-7ee0-85b7-45450ad476eb-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7ee0-85b7-45450ad476eb-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7ee0-85b7-45450ad476eb-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7ee0-85b7-45450ad476eb-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ee0-85b7-45450ad476eb-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7ee0-85b7-45450ad476eb-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ee0-85b7-45450ad476eb-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7ee0-85b7-45450ad476eb-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7ee4-8f2b-ea43720d520b"></span> List users (*019791cc-06c7-7ee4-8f2b-ea43720d520b*)

```
GET /users
```

Retrieve paginated list of all users in the system

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

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
| [200](#019791cc-06c7-7ee4-8f2b-ea43720d520b-200) | OK | OK |  | [schema](#019791cc-06c7-7ee4-8f2b-ea43720d520b-200-schema) |
| [400](#019791cc-06c7-7ee4-8f2b-ea43720d520b-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7ee4-8f2b-ea43720d520b-400-schema) |
| [500](#019791cc-06c7-7ee4-8f2b-ea43720d520b-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7ee4-8f2b-ea43720d520b-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7ee4-8f2b-ea43720d520b-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7ee4-8f2b-ea43720d520b-200-schema"></span> Schema
   
  

[ModelListUsersResponse](#model-list-users-response)

##### <span id="019791cc-06c7-7ee4-8f2b-ea43720d520b-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7ee4-8f2b-ea43720d520b-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ee4-8f2b-ea43720d520b-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7ee4-8f2b-ea43720d520b-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7eec-83f3-bcaed0c4d46f"></span> Link roles to user (*019791cc-06c7-7eec-83f3-bcaed0c4d46f*)

```
POST /users/{user_id}/roles
```

Associate multiple roles with a user within a specific project

#### Consumes
  * application/json

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| user_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The user ID in UUID format |
| user | `body` | [ModelLinkRolesToUserRequest](#model-link-roles-to-user-request) | `models.ModelLinkRolesToUserRequest` | | ✓ | | Link Roles Request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7eec-83f3-bcaed0c4d46f-200) | OK | OK |  | [schema](#019791cc-06c7-7eec-83f3-bcaed0c4d46f-200-schema) |
| [400](#019791cc-06c7-7eec-83f3-bcaed0c4d46f-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7eec-83f3-bcaed0c4d46f-400-schema) |
| [409](#019791cc-06c7-7eec-83f3-bcaed0c4d46f-409) | Conflict | Conflict |  | [schema](#019791cc-06c7-7eec-83f3-bcaed0c4d46f-409-schema) |
| [500](#019791cc-06c7-7eec-83f3-bcaed0c4d46f-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7eec-83f3-bcaed0c4d46f-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7eec-83f3-bcaed0c4d46f-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7eec-83f3-bcaed0c4d46f-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7eec-83f3-bcaed0c4d46f-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7eec-83f3-bcaed0c4d46f-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7eec-83f3-bcaed0c4d46f-409"></span> 409 - Conflict
Status: Conflict

###### <span id="019791cc-06c7-7eec-83f3-bcaed0c4d46f-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7eec-83f3-bcaed0c4d46f-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7eec-83f3-bcaed0c4d46f-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7ef0-9394-d4ac3f52e94c"></span> Unlink roles from user (*019791cc-06c7-7ef0-9394-d4ac3f52e94c*)

```
DELETE /users/{user_id}/roles
```

Remove role associations from a user within a specific project

#### Consumes
  * application/json

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| user_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The user ID in UUID format |
| body | `body` | [ModelUnlinkRolesFromUserRequest](#model-unlink-roles-from-user-request) | `models.ModelUnlinkRolesFromUserRequest` | | ✓ | | UnLink Roles Request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7ef0-9394-d4ac3f52e94c-200) | OK | OK |  | [schema](#019791cc-06c7-7ef0-9394-d4ac3f52e94c-200-schema) |
| [400](#019791cc-06c7-7ef0-9394-d4ac3f52e94c-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7ef0-9394-d4ac3f52e94c-400-schema) |
| [409](#019791cc-06c7-7ef0-9394-d4ac3f52e94c-409) | Conflict | Conflict |  | [schema](#019791cc-06c7-7ef0-9394-d4ac3f52e94c-409-schema) |
| [500](#019791cc-06c7-7ef0-9394-d4ac3f52e94c-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7ef0-9394-d4ac3f52e94c-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7ef0-9394-d4ac3f52e94c-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7ef0-9394-d4ac3f52e94c-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ef0-9394-d4ac3f52e94c-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7ef0-9394-d4ac3f52e94c-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ef0-9394-d4ac3f52e94c-409"></span> 409 - Conflict
Status: Conflict

###### <span id="019791cc-06c7-7ef0-9394-d4ac3f52e94c-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ef0-9394-d4ac3f52e94c-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7ef0-9394-d4ac3f52e94c-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7ef4-afa5-81125e9dcde9"></span> Get user authorization (*019791cc-06c7-7ef4-afa5-81125e9dcde9*)

```
GET /users/{user_id}/authz
```

Retrieve user authorization permissions and roles for access control

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| user_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The user ID in UUID format |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7ef4-afa5-81125e9dcde9-200) | OK | OK |  | [schema](#019791cc-06c7-7ef4-afa5-81125e9dcde9-200-schema) |
| [400](#019791cc-06c7-7ef4-afa5-81125e9dcde9-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7ef4-afa5-81125e9dcde9-400-schema) |
| [404](#019791cc-06c7-7ef4-afa5-81125e9dcde9-404) | Not Found | Not Found |  | [schema](#019791cc-06c7-7ef4-afa5-81125e9dcde9-404-schema) |
| [500](#019791cc-06c7-7ef4-afa5-81125e9dcde9-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7ef4-afa5-81125e9dcde9-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7ef4-afa5-81125e9dcde9-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7ef4-afa5-81125e9dcde9-200-schema"></span> Schema
   
  

any

##### <span id="019791cc-06c7-7ef4-afa5-81125e9dcde9-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7ef4-afa5-81125e9dcde9-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ef4-afa5-81125e9dcde9-404"></span> 404 - Not Found
Status: Not Found

###### <span id="019791cc-06c7-7ef4-afa5-81125e9dcde9-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7ef4-afa5-81125e9dcde9-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7ef4-afa5-81125e9dcde9-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7efb-99c2-b25af11e600c"></span> List users by role (*019791cc-06c7-7efb-99c2-b25af11e600c*)

```
GET /roles/{role_id}/users
```

Retrieve paginated list of users associated with a specific role

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| role_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The role id in UUID format |
| fields | `query` | string (formatted string) | `string` |  |  |  | Fields to return. Example: id,first_name,last_name |
| filter | `query` | string (formatted string) | `string` |  |  |  | Filter field. Example: id=1 AND first_name='John' |
| limit | `query` | int (formatted integer) | `int64` |  |  |  | Limit |
| next_token | `query` | string (formatted string) | `string` |  |  |  | Next cursor |
| prev_token | `query` | string (formatted string) | `string` |  |  |  | Previous cursor |
| sort | `query` | string (formatted string) | `string` |  |  |  | Comma-separated list of fields to sort by. Example: first_name ASC, created_at DESC |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7efb-99c2-b25af11e600c-200) | OK | OK |  | [schema](#019791cc-06c7-7efb-99c2-b25af11e600c-200-schema) |
| [400](#019791cc-06c7-7efb-99c2-b25af11e600c-400) | Bad Request | Bad Request |  | [schema](#019791cc-06c7-7efb-99c2-b25af11e600c-400-schema) |
| [500](#019791cc-06c7-7efb-99c2-b25af11e600c-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7efb-99c2-b25af11e600c-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7efb-99c2-b25af11e600c-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7efb-99c2-b25af11e600c-200-schema"></span> Schema
   
  

[ModelListUsersResponse](#model-list-users-response)

##### <span id="019791cc-06c7-7efb-99c2-b25af11e600c-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019791cc-06c7-7efb-99c2-b25af11e600c-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019791cc-06c7-7efb-99c2-b25af11e600c-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7efb-99c2-b25af11e600c-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019791cc-06c7-7eff-a1df-fbc2ad0b27c9"></span> Get version (*019791cc-06c7-7eff-a1df-fbc2ad0b27c9*)

```
GET /version
```

Retrieve the current version and build information of the service

#### Produces
  * application/json

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019791cc-06c7-7eff-a1df-fbc2ad0b27c9-200) | OK | OK |  | [schema](#019791cc-06c7-7eff-a1df-fbc2ad0b27c9-200-schema) |
| [500](#019791cc-06c7-7eff-a1df-fbc2ad0b27c9-500) | Internal Server Error | Internal Server Error |  | [schema](#019791cc-06c7-7eff-a1df-fbc2ad0b27c9-500-schema) |

#### Responses


##### <span id="019791cc-06c7-7eff-a1df-fbc2ad0b27c9-200"></span> 200 - OK
Status: OK

###### <span id="019791cc-06c7-7eff-a1df-fbc2ad0b27c9-200-schema"></span> Schema
   
  

[ModelVersion](#model-version)

##### <span id="019791cc-06c7-7eff-a1df-fbc2ad0b27c9-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019791cc-06c7-7eff-a1df-fbc2ad0b27c9-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019797e6-138a-7cf1-b7bb-fa9c5e168c49"></span> List projects (*019797e6-138a-7cf1-b7bb-fa9c5e168c49*)

```
GET /projects
```

Retrieve paginated list of all projects in the system

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

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
| [200](#019797e6-138a-7cf1-b7bb-fa9c5e168c49-200) | OK | OK |  | [schema](#019797e6-138a-7cf1-b7bb-fa9c5e168c49-200-schema) |
| [400](#019797e6-138a-7cf1-b7bb-fa9c5e168c49-400) | Bad Request | Bad Request |  | [schema](#019797e6-138a-7cf1-b7bb-fa9c5e168c49-400-schema) |
| [500](#019797e6-138a-7cf1-b7bb-fa9c5e168c49-500) | Internal Server Error | Internal Server Error |  | [schema](#019797e6-138a-7cf1-b7bb-fa9c5e168c49-500-schema) |

#### Responses


##### <span id="019797e6-138a-7cf1-b7bb-fa9c5e168c49-200"></span> 200 - OK
Status: OK

###### <span id="019797e6-138a-7cf1-b7bb-fa9c5e168c49-200-schema"></span> Schema
   
  

[ModelListProjectsResponse](#model-list-projects-response)

##### <span id="019797e6-138a-7cf1-b7bb-fa9c5e168c49-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019797e6-138a-7cf1-b7bb-fa9c5e168c49-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019797e6-138a-7cf1-b7bb-fa9c5e168c49-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019797e6-138a-7cf1-b7bb-fa9c5e168c49-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019797e6-138a-7cf4-8694-e4611baded39"></span> Delete project (*019797e6-138a-7cf4-8694-e4611baded39*)

```
DELETE /projects/{project_id}
```

Remove a project permanently from the system

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| project_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The project id in UUID format |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019797e6-138a-7cf4-8694-e4611baded39-200) | OK | OK |  | [schema](#019797e6-138a-7cf4-8694-e4611baded39-200-schema) |
| [400](#019797e6-138a-7cf4-8694-e4611baded39-400) | Bad Request | Bad Request |  | [schema](#019797e6-138a-7cf4-8694-e4611baded39-400-schema) |
| [500](#019797e6-138a-7cf4-8694-e4611baded39-500) | Internal Server Error | Internal Server Error |  | [schema](#019797e6-138a-7cf4-8694-e4611baded39-500-schema) |

#### Responses


##### <span id="019797e6-138a-7cf4-8694-e4611baded39-200"></span> 200 - OK
Status: OK

###### <span id="019797e6-138a-7cf4-8694-e4611baded39-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019797e6-138a-7cf4-8694-e4611baded39-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019797e6-138a-7cf4-8694-e4611baded39-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019797e6-138a-7cf4-8694-e4611baded39-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019797e6-138a-7cf4-8694-e4611baded39-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019797e6-138a-7cf8-9887-e4c44ad0ae19"></span> Update project (*019797e6-138a-7cf8-9887-e4c44ad0ae19*)

```
PUT /projects/{project_id}
```

Modify an existing project by its ID

#### Consumes
  * application/json

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| project_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The project id in UUID format |
| body | `body` | [ModelUpdateProjectRequest](#model-update-project-request) | `models.ModelUpdateProjectRequest` | | ✓ | | Update Project Request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019797e6-138a-7cf8-9887-e4c44ad0ae19-200) | OK | OK |  | [schema](#019797e6-138a-7cf8-9887-e4c44ad0ae19-200-schema) |
| [400](#019797e6-138a-7cf8-9887-e4c44ad0ae19-400) | Bad Request | Bad Request |  | [schema](#019797e6-138a-7cf8-9887-e4c44ad0ae19-400-schema) |
| [409](#019797e6-138a-7cf8-9887-e4c44ad0ae19-409) | Conflict | Conflict |  | [schema](#019797e6-138a-7cf8-9887-e4c44ad0ae19-409-schema) |
| [500](#019797e6-138a-7cf8-9887-e4c44ad0ae19-500) | Internal Server Error | Internal Server Error |  | [schema](#019797e6-138a-7cf8-9887-e4c44ad0ae19-500-schema) |

#### Responses


##### <span id="019797e6-138a-7cf8-9887-e4c44ad0ae19-200"></span> 200 - OK
Status: OK

###### <span id="019797e6-138a-7cf8-9887-e4c44ad0ae19-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019797e6-138a-7cf8-9887-e4c44ad0ae19-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019797e6-138a-7cf8-9887-e4c44ad0ae19-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019797e6-138a-7cf8-9887-e4c44ad0ae19-409"></span> 409 - Conflict
Status: Conflict

###### <span id="019797e6-138a-7cf8-9887-e4c44ad0ae19-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019797e6-138a-7cf8-9887-e4c44ad0ae19-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019797e6-138a-7cf8-9887-e4c44ad0ae19-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019797e6-138a-7d00-98db-740f21794f11"></span> Create project (*019797e6-138a-7d00-98db-740f21794f11*)

```
POST /projects
```

Create a new project with specified configuration

#### Consumes
  * application/json

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| body | `body` | [ModelCreateProjectRequest](#model-create-project-request) | `models.ModelCreateProjectRequest` | | ✓ | | Create Project Request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [201](#019797e6-138a-7d00-98db-740f21794f11-201) | Created | Created |  | [schema](#019797e6-138a-7d00-98db-740f21794f11-201-schema) |
| [400](#019797e6-138a-7d00-98db-740f21794f11-400) | Bad Request | Bad Request |  | [schema](#019797e6-138a-7d00-98db-740f21794f11-400-schema) |
| [409](#019797e6-138a-7d00-98db-740f21794f11-409) | Conflict | Conflict |  | [schema](#019797e6-138a-7d00-98db-740f21794f11-409-schema) |
| [500](#019797e6-138a-7d00-98db-740f21794f11-500) | Internal Server Error | Internal Server Error |  | [schema](#019797e6-138a-7d00-98db-740f21794f11-500-schema) |

#### Responses


##### <span id="019797e6-138a-7d00-98db-740f21794f11-201"></span> 201 - Created
Status: Created

###### <span id="019797e6-138a-7d00-98db-740f21794f11-201-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019797e6-138a-7d00-98db-740f21794f11-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019797e6-138a-7d00-98db-740f21794f11-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019797e6-138a-7d00-98db-740f21794f11-409"></span> 409 - Conflict
Status: Conflict

###### <span id="019797e6-138a-7d00-98db-740f21794f11-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019797e6-138a-7d00-98db-740f21794f11-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019797e6-138a-7d00-98db-740f21794f11-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="019797e6-138a-7d04-8db3-1d4755b25db3"></span> Get project (*019797e6-138a-7d04-8db3-1d4755b25db3*)

```
GET /projects/{project_id}
```

Retrieve a specific project by its unique identifier

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| project_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The project id in UUID format |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#019797e6-138a-7d04-8db3-1d4755b25db3-200) | OK | OK |  | [schema](#019797e6-138a-7d04-8db3-1d4755b25db3-200-schema) |
| [400](#019797e6-138a-7d04-8db3-1d4755b25db3-400) | Bad Request | Bad Request |  | [schema](#019797e6-138a-7d04-8db3-1d4755b25db3-400-schema) |
| [404](#019797e6-138a-7d04-8db3-1d4755b25db3-404) | Not Found | Not Found |  | [schema](#019797e6-138a-7d04-8db3-1d4755b25db3-404-schema) |
| [500](#019797e6-138a-7d04-8db3-1d4755b25db3-500) | Internal Server Error | Internal Server Error |  | [schema](#019797e6-138a-7d04-8db3-1d4755b25db3-500-schema) |

#### Responses


##### <span id="019797e6-138a-7d04-8db3-1d4755b25db3-200"></span> 200 - OK
Status: OK

###### <span id="019797e6-138a-7d04-8db3-1d4755b25db3-200-schema"></span> Schema
   
  

[ModelProject](#model-project)

##### <span id="019797e6-138a-7d04-8db3-1d4755b25db3-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="019797e6-138a-7d04-8db3-1d4755b25db3-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019797e6-138a-7d04-8db3-1d4755b25db3-404"></span> 404 - Not Found
Status: Not Found

###### <span id="019797e6-138a-7d04-8db3-1d4755b25db3-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="019797e6-138a-7d04-8db3-1d4755b25db3-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="019797e6-138a-7d04-8db3-1d4755b25db3-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="01979db7-f53f-73a1-aab2-74802b79be51"></span> Get product (*01979db7-f53f-73a1-aab2-74802b79be51*)

```
GET /projects/{project_id}/products/{product_id}
```

Retrieve a specific product by its unique identifier

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| product_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The product id in UUID format |
| project_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The project id in UUID format |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#01979db7-f53f-73a1-aab2-74802b79be51-200) | OK | OK |  | [schema](#01979db7-f53f-73a1-aab2-74802b79be51-200-schema) |
| [400](#01979db7-f53f-73a1-aab2-74802b79be51-400) | Bad Request | Bad Request |  | [schema](#01979db7-f53f-73a1-aab2-74802b79be51-400-schema) |
| [404](#01979db7-f53f-73a1-aab2-74802b79be51-404) | Not Found | Not Found |  | [schema](#01979db7-f53f-73a1-aab2-74802b79be51-404-schema) |
| [500](#01979db7-f53f-73a1-aab2-74802b79be51-500) | Internal Server Error | Internal Server Error |  | [schema](#01979db7-f53f-73a1-aab2-74802b79be51-500-schema) |

#### Responses


##### <span id="01979db7-f53f-73a1-aab2-74802b79be51-200"></span> 200 - OK
Status: OK

###### <span id="01979db7-f53f-73a1-aab2-74802b79be51-200-schema"></span> Schema
   
  

[ModelProduct](#model-product)

##### <span id="01979db7-f53f-73a1-aab2-74802b79be51-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="01979db7-f53f-73a1-aab2-74802b79be51-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="01979db7-f53f-73a1-aab2-74802b79be51-404"></span> 404 - Not Found
Status: Not Found

###### <span id="01979db7-f53f-73a1-aab2-74802b79be51-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="01979db7-f53f-73a1-aab2-74802b79be51-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="01979db7-f53f-73a1-aab2-74802b79be51-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="01979db7-f53f-73a5-b916-297c6db5b714"></span> Create product (*01979db7-f53f-73a5-b916-297c6db5b714*)

```
POST /projects/{project_id}/products
```

Create a new product with specified configuration

#### Consumes
  * application/json

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| project_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The project id in UUID format |
| body | `body` | [ModelCreateProductRequest](#model-create-product-request) | `models.ModelCreateProductRequest` | | ✓ | | Create product request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [201](#01979db7-f53f-73a5-b916-297c6db5b714-201) | Created | Created |  | [schema](#01979db7-f53f-73a5-b916-297c6db5b714-201-schema) |
| [400](#01979db7-f53f-73a5-b916-297c6db5b714-400) | Bad Request | Bad Request |  | [schema](#01979db7-f53f-73a5-b916-297c6db5b714-400-schema) |
| [409](#01979db7-f53f-73a5-b916-297c6db5b714-409) | Conflict | Conflict |  | [schema](#01979db7-f53f-73a5-b916-297c6db5b714-409-schema) |
| [500](#01979db7-f53f-73a5-b916-297c6db5b714-500) | Internal Server Error | Internal Server Error |  | [schema](#01979db7-f53f-73a5-b916-297c6db5b714-500-schema) |

#### Responses


##### <span id="01979db7-f53f-73a5-b916-297c6db5b714-201"></span> 201 - Created
Status: Created

###### <span id="01979db7-f53f-73a5-b916-297c6db5b714-201-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="01979db7-f53f-73a5-b916-297c6db5b714-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="01979db7-f53f-73a5-b916-297c6db5b714-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="01979db7-f53f-73a5-b916-297c6db5b714-409"></span> 409 - Conflict
Status: Conflict

###### <span id="01979db7-f53f-73a5-b916-297c6db5b714-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="01979db7-f53f-73a5-b916-297c6db5b714-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="01979db7-f53f-73a5-b916-297c6db5b714-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="01979db7-f53f-73a9-bd58-e1cd5d7df436"></span> Update product (*01979db7-f53f-73a9-bd58-e1cd5d7df436*)

```
PUT /projects/{project_id}/products/{product_id}
```

Modify an existing product by its ID

#### Consumes
  * application/json

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| product_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The model id in UUID format |
| project_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The project id in UUID format |
| body | `body` | [ModelUpdateProductRequest](#model-update-product-request) | `models.ModelUpdateProductRequest` | | ✓ | | Update product request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#01979db7-f53f-73a9-bd58-e1cd5d7df436-200) | OK | OK |  | [schema](#01979db7-f53f-73a9-bd58-e1cd5d7df436-200-schema) |
| [400](#01979db7-f53f-73a9-bd58-e1cd5d7df436-400) | Bad Request | Bad Request |  | [schema](#01979db7-f53f-73a9-bd58-e1cd5d7df436-400-schema) |
| [409](#01979db7-f53f-73a9-bd58-e1cd5d7df436-409) | Conflict | Conflict |  | [schema](#01979db7-f53f-73a9-bd58-e1cd5d7df436-409-schema) |
| [500](#01979db7-f53f-73a9-bd58-e1cd5d7df436-500) | Internal Server Error | Internal Server Error |  | [schema](#01979db7-f53f-73a9-bd58-e1cd5d7df436-500-schema) |

#### Responses


##### <span id="01979db7-f53f-73a9-bd58-e1cd5d7df436-200"></span> 200 - OK
Status: OK

###### <span id="01979db7-f53f-73a9-bd58-e1cd5d7df436-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="01979db7-f53f-73a9-bd58-e1cd5d7df436-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="01979db7-f53f-73a9-bd58-e1cd5d7df436-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="01979db7-f53f-73a9-bd58-e1cd5d7df436-409"></span> 409 - Conflict
Status: Conflict

###### <span id="01979db7-f53f-73a9-bd58-e1cd5d7df436-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="01979db7-f53f-73a9-bd58-e1cd5d7df436-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="01979db7-f53f-73a9-bd58-e1cd5d7df436-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="01979db7-f53f-73ad-a84e-49bbdfe9e5c9"></span> Delete product (*01979db7-f53f-73ad-a84e-49bbdfe9e5c9*)

```
DELETE /projects/{project_id}/products/{product_id}
```

Remove a product permanently from the system

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| product_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The product id in UUID format |
| project_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The project id in UUID format |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#01979db7-f53f-73ad-a84e-49bbdfe9e5c9-200) | OK | OK |  | [schema](#01979db7-f53f-73ad-a84e-49bbdfe9e5c9-200-schema) |
| [400](#01979db7-f53f-73ad-a84e-49bbdfe9e5c9-400) | Bad Request | Bad Request |  | [schema](#01979db7-f53f-73ad-a84e-49bbdfe9e5c9-400-schema) |
| [500](#01979db7-f53f-73ad-a84e-49bbdfe9e5c9-500) | Internal Server Error | Internal Server Error |  | [schema](#01979db7-f53f-73ad-a84e-49bbdfe9e5c9-500-schema) |

#### Responses


##### <span id="01979db7-f53f-73ad-a84e-49bbdfe9e5c9-200"></span> 200 - OK
Status: OK

###### <span id="01979db7-f53f-73ad-a84e-49bbdfe9e5c9-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="01979db7-f53f-73ad-a84e-49bbdfe9e5c9-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="01979db7-f53f-73ad-a84e-49bbdfe9e5c9-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="01979db7-f53f-73ad-a84e-49bbdfe9e5c9-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="01979db7-f53f-73ad-a84e-49bbdfe9e5c9-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="01979db7-f53f-73b1-993f-15f77e72c8cc"></span> List products by project (*01979db7-f53f-73b1-993f-15f77e72c8cc*)

```
GET /projects/{project_id}/products
```

Retrieve paginated list of products for a specific project

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| project_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The project id in UUID format |
| fields | `query` | string (formatted string) | `string` |  |  |  | Fields to return. Example: id,first_name,last_name |
| filter | `query` | string (formatted string) | `string` |  |  |  | Filter field. Example: id=1 AND first_name='John' |
| limit | `query` | int (formatted integer) | `int64` |  |  |  | Limit |
| next_token | `query` | string (formatted string) | `string` |  |  |  | Next cursor |
| prev_token | `query` | string (formatted string) | `string` |  |  |  | Previous cursor |
| sort | `query` | string (formatted string) | `string` |  |  |  | Comma-separated list of fields to sort by. Example: first_name ASC, created_at DESC |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#01979db7-f53f-73b1-993f-15f77e72c8cc-200) | OK | OK |  | [schema](#01979db7-f53f-73b1-993f-15f77e72c8cc-200-schema) |
| [400](#01979db7-f53f-73b1-993f-15f77e72c8cc-400) | Bad Request | Bad Request |  | [schema](#01979db7-f53f-73b1-993f-15f77e72c8cc-400-schema) |
| [500](#01979db7-f53f-73b1-993f-15f77e72c8cc-500) | Internal Server Error | Internal Server Error |  | [schema](#01979db7-f53f-73b1-993f-15f77e72c8cc-500-schema) |

#### Responses


##### <span id="01979db7-f53f-73b1-993f-15f77e72c8cc-200"></span> 200 - OK
Status: OK

###### <span id="01979db7-f53f-73b1-993f-15f77e72c8cc-200-schema"></span> Schema
   
  

[ModelListProductsResponse](#model-list-products-response)

##### <span id="01979db7-f53f-73b1-993f-15f77e72c8cc-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="01979db7-f53f-73b1-993f-15f77e72c8cc-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="01979db7-f53f-73b1-993f-15f77e72c8cc-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="01979db7-f53f-73b1-993f-15f77e72c8cc-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="01979db7-f53f-73b5-a499-d6390831c94c"></span> List products (*01979db7-f53f-73b5-a499-d6390831c94c*)

```
GET /products
```

Retrieve paginated list of all products in the system

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

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
| [200](#01979db7-f53f-73b5-a499-d6390831c94c-200) | OK | OK |  | [schema](#01979db7-f53f-73b5-a499-d6390831c94c-200-schema) |
| [400](#01979db7-f53f-73b5-a499-d6390831c94c-400) | Bad Request | Bad Request |  | [schema](#01979db7-f53f-73b5-a499-d6390831c94c-400-schema) |
| [500](#01979db7-f53f-73b5-a499-d6390831c94c-500) | Internal Server Error | Internal Server Error |  | [schema](#01979db7-f53f-73b5-a499-d6390831c94c-500-schema) |

#### Responses


##### <span id="01979db7-f53f-73b5-a499-d6390831c94c-200"></span> 200 - OK
Status: OK

###### <span id="01979db7-f53f-73b5-a499-d6390831c94c-200-schema"></span> Schema
   
  

[ModelListProductsResponse](#model-list-products-response)

##### <span id="01979db7-f53f-73b5-a499-d6390831c94c-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="01979db7-f53f-73b5-a499-d6390831c94c-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="01979db7-f53f-73b5-a499-d6390831c94c-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="01979db7-f53f-73b5-a499-d6390831c94c-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="01979db7-f53f-73b9-818f-cdd1848f15d0"></span> Unlink product from payment processor (*01979db7-f53f-73b9-818f-cdd1848f15d0*)

```
DELETE /projects/{project_id}/products/{product_id}/payment_processor
```

Remove the association between a product and a payment processor

#### Consumes
  * application/json

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| product_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The product id in UUID format |
| project_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The project id in UUID format |
| body | `body` | [ModelUnlinkProductFromPaymentProcessorRequest](#model-unlink-product-from-payment-processor-request) | `models.ModelUnlinkProductFromPaymentProcessorRequest` | | ✓ | | Unlink product from payment processor request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#01979db7-f53f-73b9-818f-cdd1848f15d0-200) | OK | OK |  | [schema](#01979db7-f53f-73b9-818f-cdd1848f15d0-200-schema) |
| [400](#01979db7-f53f-73b9-818f-cdd1848f15d0-400) | Bad Request | Bad Request |  | [schema](#01979db7-f53f-73b9-818f-cdd1848f15d0-400-schema) |
| [500](#01979db7-f53f-73b9-818f-cdd1848f15d0-500) | Internal Server Error | Internal Server Error |  | [schema](#01979db7-f53f-73b9-818f-cdd1848f15d0-500-schema) |

#### Responses


##### <span id="01979db7-f53f-73b9-818f-cdd1848f15d0-200"></span> 200 - OK
Status: OK

###### <span id="01979db7-f53f-73b9-818f-cdd1848f15d0-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="01979db7-f53f-73b9-818f-cdd1848f15d0-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="01979db7-f53f-73b9-818f-cdd1848f15d0-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="01979db7-f53f-73b9-818f-cdd1848f15d0-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="01979db7-f53f-73b9-818f-cdd1848f15d0-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="01979db7-f53f-73bd-b6c0-4541a48549c2"></span> Link product to payment processor (*01979db7-f53f-73bd-b6c0-4541a48549c2*)

```
POST /projects/{project_id}/products/{product_id}/payment_processor
```

Associate a product with a payment processor to enable billing and invoicing

#### Consumes
  * application/json

#### Produces
  * application/json

#### Security Requirements
  * AccessToken

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| product_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The product id in UUID format |
| project_id | `path` | uuid (formatted string) | `strfmt.UUID` |  | ✓ |  | The project id in UUID format |
| body | `body` | [ModelLinkProductToPaymentProcessorRequest](#model-link-product-to-payment-processor-request) | `models.ModelLinkProductToPaymentProcessorRequest` | | ✓ | | Link product to payment processor request |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#01979db7-f53f-73bd-b6c0-4541a48549c2-200) | OK | OK |  | [schema](#01979db7-f53f-73bd-b6c0-4541a48549c2-200-schema) |
| [400](#01979db7-f53f-73bd-b6c0-4541a48549c2-400) | Bad Request | Bad Request |  | [schema](#01979db7-f53f-73bd-b6c0-4541a48549c2-400-schema) |
| [500](#01979db7-f53f-73bd-b6c0-4541a48549c2-500) | Internal Server Error | Internal Server Error |  | [schema](#01979db7-f53f-73bd-b6c0-4541a48549c2-500-schema) |

#### Responses


##### <span id="01979db7-f53f-73bd-b6c0-4541a48549c2-200"></span> 200 - OK
Status: OK

###### <span id="01979db7-f53f-73bd-b6c0-4541a48549c2-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="01979db7-f53f-73bd-b6c0-4541a48549c2-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="01979db7-f53f-73bd-b6c0-4541a48549c2-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="01979db7-f53f-73bd-b6c0-4541a48549c2-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="01979db7-f53f-73bd-b6c0-4541a48549c2-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

## Models

### <span id="model-check"></span> model.Check


> Health check of the service.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| data | map of any | `map[string]interface{}` |  | |  |  |
| kind | string (formatted string)| `string` |  | |  | `database` |
| name | string (formatted string)| `string` |  | |  | `database` |
| status | string (formatted boolean)| `string` |  | | Health status of a service. | `true` |



### <span id="model-create-policy-request"></span> model.CreatePolicyRequest


> Create a policy.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| allowed_action | string (formatted string)| `string` | ✓ | |  | `GET` |
| allowed_resource | string (formatted string)| `string` | ✓ | |  | `/projects/39a4707f-536e-433f-8597-6fc0d53a724f/tokens` |
| description | string (formatted string)| `string` |  | |  | `This allows to list all the policies of a specific project` |
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01979cde-6d91-7733-8381-eaef585fad97` |
| name | string (formatted string)| `string` | ✓ | |  | `List Policies for project` |



### <span id="model-create-product-request"></span> model.CreateProductRequest


> CreateProductRequest represents the input for the CreateProduct method.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| description | string (formatted string)| `string` |  | |  | `This is a product` |
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01979cde-6d91-7756-bd58-75b685419eb3` |
| name | string (formatted string)| `string` |  | |  | `New product name` |



### <span id="model-create-project-request"></span> model.CreateProjectRequest


> CreateProjectRequest represents the inputs necessary to create a new project.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| description | string (formatted string)| `string` | ✓ | |  | `This is a new project` |
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01979cde-6d91-7720-aaa5-d45582f94ac4` |
| name | string (formatted string)| `string` | ✓ | |  | `New project name` |



### <span id="model-create-role-request"></span> model.CreateRoleRequest


> CreateRoleRequest represents the input for the CreateRole method.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| description | string (formatted string)| `string` | ✓ | |  | `This is a role` |
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01979cde-6d91-7766-a529-ebe2a773d447` |
| name | string (formatted string)| `string` | ✓ | |  | `New role name` |



### <span id="model-create-user-request"></span> model.CreateUserRequest


> Create user request.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| email | email (formatted string)| `strfmt.Email` | ✓ | |  | `my@email.com` |
| first_name | string (formatted string)| `string` | ✓ | |  | `John` |
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01979cde-6d91-775e-b049-9627e2c6f848` |
| last_name | string (formatted string)| `string` | ✓ | |  | `Doe` |
| password | string (formatted string)| `string` | ✓ | |  | `ThisIs4Passw0rd` |



### <span id="model-http-message"></span> model.HTTPMessage


> HTTPMessage represents a message to be sent to the client trough HTTP REST API.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| message | string (formatted string)| `string` |  | |  | `success` |
| method | string (formatted string)| `string` |  | |  | `GET` |
| path | string (formatted string)| `string` |  | |  | `/api/v1/users` |
| status_code | int32 (formatted integer)| `int32` |  | |  | `200` |
| timestamp | date-time (formatted string)| `strfmt.DateTime` |  | |  | `2021-07-01T00:00:00Z` |



### <span id="model-health"></span> model.Health


> Health check of the service.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| checks | [][ModelCheck](#model-check)| `[]*ModelCheck` |  | |  |  |
| status | string (formatted boolean)| `string` |  | | Health status of a service. | `true` |



### <span id="model-link-policies-to-role-request"></span> model.LinkPoliciesToRoleRequest


> LinkPoliciesToRoleRequest input values for linking policies to a role.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| policy_ids | []uuid (formatted string)| `[]strfmt.UUID` | ✓ | |  |  |



### <span id="model-link-product-to-payment-processor-request"></span> model.LinkProductToPaymentProcessorRequest


> LinkProductToPaymentProcessorRequest represents the input for linking a product to a payment processor.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| payment_processor_id | string| `string` |  | |  |  |
| payment_processor_product_id | string| `string` |  | |  |  |



### <span id="model-link-roles-to-policy-request"></span> model.LinkRolesToPolicyRequest


> Link roles to a policy.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| role_ids | []uuid (formatted string)| `[]strfmt.UUID` | ✓ | |  | `["01979cde-6d91-772b-ac5b-c7a2aa7512f5"]` |



### <span id="model-link-roles-to-user-request"></span> model.LinkRolesToUserRequest


> Link roles request.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| role_ids | []uuid (formatted string)| `[]strfmt.UUID` | ✓ | |  |  |



### <span id="model-link-users-to-role-request"></span> model.LinkUsersToRoleRequest


> LinkUsersToRoleRequest input values for linking users to a role.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| user_ids | []uuid (formatted string)| `[]strfmt.UUID` | ✓ | |  |  |



### <span id="model-list-policies-response"></span> model.ListPoliciesResponse


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| items | [][ModelPolicy](#model-policy)| `[]*ModelPolicy` |  | |  |  |
| paginator | [ModelPaginator](#model-paginator)| `ModelPaginator` |  | |  |  |



### <span id="model-list-products-response"></span> model.ListProductsResponse


> ListProductResponse represents a list of users.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| items | [][ModelProduct](#model-product)| `[]*ModelProduct` |  | |  |  |
| paginator | [ModelPaginator](#model-paginator)| `ModelPaginator` |  | |  |  |



### <span id="model-list-projects-response"></span> model.ListProjectsResponse


> ListProjectsResponse represents a list of users.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| items | [][ModelProject](#model-project)| `[]*ModelProject` |  | |  |  |
| paginator | [ModelPaginator](#model-paginator)| `ModelPaginator` |  | |  |  |



### <span id="model-list-resources-response"></span> model.ListResourcesResponse


> ListResourcesResponse represents a list of users.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| items | [][ModelResource](#model-resource)| `[]*ModelResource` |  | |  |  |
| paginator | [ModelPaginator](#model-paginator)| `ModelPaginator` |  | |  |  |



### <span id="model-list-roles-response"></span> model.ListRolesResponse


> ListRoleResponse represents a list of users.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| items | [][ModelRole](#model-role)| `[]*ModelRole` |  | |  |  |
| paginator | [ModelPaginator](#model-paginator)| `ModelPaginator` |  | |  |  |



### <span id="model-list-users-response"></span> model.ListUsersResponse


> List of users.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| items | [][ModelUser](#model-user)| `[]*ModelUser` |  | |  |  |
| paginator | [ModelPaginator](#model-paginator)| `ModelPaginator` |  | |  |  |



### <span id="model-login-user-request"></span> model.LoginUserRequest


> LoginUserRequest is the request struct for the LoginUser handler.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| email | email (formatted string)| `strfmt.Email` | ✓ | |  | `admin@qu3ry.me` |
| password | string (formatted string)| `string` | ✓ | |  | `ThisIsApassw0rd.,` |



### <span id="model-login-user-response"></span> model.LoginUserResponse


> LoginUserResponse is the response when a user logs in.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| access_token | string (formatted string)| `string` |  | |  |  |
| permissions | map of any | `map[string]interface{}` |  | |  |  |
| refresh_token | string (formatted string)| `string` |  | |  |  |
| token_type | string (formatted string)| `string` |  | |  | `Bearer` |
| user_id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01979cde-6d91-773b-a03c-3779be3b55b3` |



### <span id="model-paginator"></span> model.Paginator


> Paginator represents a paginator.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| limit | int (formatted integer)| `int64` |  | |  | `10` |
| next_page | string (formatted string)| `string` |  | |  | `http://localhost:8080/users?next_token=ZmZmZmZmZmYtZmZmZi0tZmZmZmZmZmY=\u0026limit=10` |
| next_token | string (formatted string)| `string` |  | |  | `ZmZmZmZmZmYtZmZmZi0tZmZmZmZmZmY=` |
| prev_page | string (formatted string)| `string` |  | |  | `http://localhost:8080/users?prev_token=ZmZmZmZmZmYtZmZmZi0tZmZmZmZmZmY=\u0026limit=10` |
| prev_token | string (formatted string)| `string` |  | |  | `ZmZmZmZmZmYtZmZmZi0tZmZmZmZmZmY=` |
| size | int (formatted integer)| `int64` |  | |  | `10` |



### <span id="model-policy"></span> model.Policy


> Policy represents a role.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| allowed_action | string (formatted string)| `string` |  | |  | `GET` |
| allowed_resource | string (formatted string)| `string` |  | |  | `/projects/39a4707f-536e-433f-8597-6fc0d53a724f/tokens` |
| created_at | date-time (formatted string)| `strfmt.DateTime` |  | |  | `2021-01-01T00:00:00Z` |
| description | string (formatted string)| `string` |  | |  | `This is a role` |
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01979cde-6d91-7728-bfbb-ec90edcd6767` |
| name | string (formatted string)| `string` |  | |  | `Policy Name` |
| resource | [ModelResource](#model-resource)| `ModelResource` |  | |  |  |
| system | boolean (formatted boolean)| `bool` |  | |  | `false` |
| updated_at | date-time (formatted string)| `strfmt.DateTime` |  | |  | `2021-01-01T00:00:00Z` |



### <span id="model-product"></span> model.Product


> Product represents a product.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| created_at | date-time (formatted string)| `strfmt.DateTime` |  | |  | `2021-01-01T00:00:00Z` |
| description | string (formatted string)| `string` |  | |  | `This is a product` |
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01979cde-6d91-7753-9b69-975f70601c14` |
| name | string (formatted string)| `string` |  | |  | `Product Name` |
| project | [ModelProject](#model-project)| `ModelProject` |  | |  |  |
| updated_at | date-time (formatted string)| `strfmt.DateTime` |  | |  | `2021-01-01T00:00:00Z` |



### <span id="model-project"></span> model.Project


> Project represents a project.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| created_at | date-time (formatted string)| `strfmt.DateTime` |  | |  | `2021-01-01T00:00:00Z` |
| description | string (formatted string)| `string` |  | |  | `This is a project` |
| disabled | boolean (formatted boolean)| `bool` |  | |  | `false` |
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01979cde-6d91-769f-8d3e-b04eeb538a83` |
| name | string (formatted string)| `string` |  | |  | `John` |
| system | boolean (formatted boolean)| `bool` |  | |  | `false` |
| updated_at | date-time (formatted string)| `strfmt.DateTime` |  | |  | `2021-01-01T00:00:00Z` |



### <span id="model-re-verify-user-request"></span> model.ReVerifyUserRequest


> ReVerifyUserRequest is the request struct for the ReVerifyUser handler.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| email | email (formatted string)| `strfmt.Email` | ✓ | |  | `user@mail.com` |



### <span id="model-refresh-token-request"></span> model.RefreshTokenRequest


> RefreshTokenRequest is the request struct for the RefreshToken handler.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| refresh_token | string (formatted string)| `string` |  | |  |  |



### <span id="model-refresh-token-response"></span> model.RefreshTokenResponse


> RefreshTokenResponse is the response when a user refreshes their token.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| access_token | string (formatted string)| `string` |  | |  |  |
| token_type | string (formatted string)| `string` |  | |  | `Bearer` |



### <span id="model-register-user-request"></span> model.RegisterUserRequest


> RegisterUserRequest is the request struct for the RegisterUser handler.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| email | email (formatted string)| `strfmt.Email` | ✓ | |  | `john.doe@email.com` |
| first_name | string (formatted string)| `string` | ✓ | |  | `John` |
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01979cde-6d91-773f-a630-7eed93217f35` |
| last_name | string (formatted string)| `string` | ✓ | |  | `Doe` |
| password | string (formatted string)| `string` | ✓ | |  | `ThisIsApassw0rd.,` |



### <span id="model-resource"></span> model.Resource


> Resource represents a permission.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| action | string (formatted string)| `string` |  | |  | `GET` |
| created_at | date-time (formatted string)| `strfmt.DateTime` |  | |  | `2021-01-01T00:00:00Z` |
| description | string (formatted string)| `string` |  | |  | `Allows reading of users` |
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01979cde-6d91-7743-a481-bfe67f26a3c2` |
| name | string (formatted string)| `string` |  | |  | `Read Users` |
| resource | string (formatted string)| `string` |  | |  | `users` |
| system | bool (formatted boolean)| `bool` |  | |  | `false` |
| updated_at | date-time (formatted string)| `strfmt.DateTime` |  | |  | `2021-01-01T00:00:00Z` |



### <span id="model-role"></span> model.Role


> Role represents a role.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| auto_assign | boolean (formatted boolean)| `bool` |  | |  | `false` |
| created_at | date-time (formatted string)| `strfmt.DateTime` |  | |  | `2021-01-01T00:00:00Z` |
| description | string (formatted string)| `string` |  | |  | `This is a role` |
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01979cde-6d91-7762-aae3-871d5e5ee975` |
| name | string (formatted string)| `string` |  | |  | `Role Name` |
| system | boolean (formatted boolean)| `bool` |  | |  | `false` |
| updated_at | date-time (formatted string)| `strfmt.DateTime` |  | |  | `2021-01-01T00:00:00Z` |



### <span id="model-unlink-policies-from-role-request"></span> model.UnlinkPoliciesFromRoleRequest


> LinkPoliciesToRoleRequest input values for linking policies to a role.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| policy_ids | []uuid (formatted string)| `[]strfmt.UUID` | ✓ | |  |  |



### <span id="model-unlink-product-from-payment-processor-request"></span> model.UnlinkProductFromPaymentProcessorRequest


> LinkProductToPaymentProcessorRequest represents the input for linking a product to a payment processor.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| payment_processor_id | string| `string` |  | |  |  |
| payment_processor_product_id | string| `string` |  | |  |  |



### <span id="model-unlink-roles-from-policy-request"></span> model.UnlinkRolesFromPolicyRequest


> Link roles to a policy.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| role_ids | []uuid (formatted string)| `[]strfmt.UUID` | ✓ | |  | `["01979cde-6d91-772b-ac5b-c7a2aa7512f5"]` |



### <span id="model-unlink-roles-from-user-request"></span> model.UnlinkRolesFromUserRequest


> Link roles request.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| role_ids | []uuid (formatted string)| `[]strfmt.UUID` | ✓ | |  |  |



### <span id="model-unlink-users-from-role-request"></span> model.UnlinkUsersFromRoleRequest


> LinkUsersToRoleRequest input values for linking users to a role.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| user_ids | []uuid (formatted string)| `[]strfmt.UUID` | ✓ | |  |  |



### <span id="model-update-policy-request"></span> model.UpdatePolicyRequest


> Update a policy.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| allowed_action | string (formatted string)| `string` |  | |  | `GET` |
| allowed_resource | string (formatted string)| `string` |  | |  | `/projects/39a4707f-536e-433f-8597-6fc0d53a724f/tokens` |
| description | string (formatted string)| `string` |  | |  | `This is a role` |
| name | string (formatted string)| `string` |  | |  | `Policy Name` |



### <span id="model-update-product-request"></span> model.UpdateProductRequest


> UpdateProductRequest represents the input for the UpdateProduct method.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| description | string (formatted string)| `string` |  | |  | `This is a product` |
| name | string (formatted string)| `string` |  | |  | `Modified product name` |



### <span id="model-update-project-request"></span> model.UpdateProjectRequest


> UpdateProjectRequest represents the inputs necessary to update a project.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| description | string (formatted string)| `string` |  | |  | `This is a new project data` |
| disabled | boolean (formatted boolean)| `bool` |  | |  | `false` |
| name | string (formatted string)| `string` |  | |  | `New project name` |



### <span id="model-update-role-request"></span> model.UpdateRoleRequest


> UpdateRoleRequest represents the input for the UpdateRole method.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| description | string (formatted string)| `string` |  | |  | `This is a role` |
| name | string (formatted string)| `string` |  | |  | `Modified role name` |



### <span id="model-update-user-request"></span> model.UpdateUserRequest


> Update user request.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| disabled | boolean (formatted boolean)| `bool` |  | |  | `false` |
| email | email (formatted string)| `strfmt.Email` |  | |  | `my@email.com` |
| first_name | string (formatted string)| `string` |  | |  | `John` |
| last_name | string (formatted string)| `string` |  | |  | `Doe` |
| password | string (formatted string)| `string` |  | |  | `ThisIs4Passw0rd` |



### <span id="model-user"></span> model.User


> User represents a user entity.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| created_at | date-time (formatted string)| `strfmt.DateTime` |  | |  | `2021-01-01T00:00:00Z` |
| disabled | boolean (formatted boolean)| `bool` |  | |  | `false` |
| email | email (formatted string)| `strfmt.Email` |  | |  | `my@email.com` |
| first_name | string (formatted string)| `string` |  | |  | `John` |
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01979cde-6d91-775a-ad3f-1a97b23ee649` |
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


