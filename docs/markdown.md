


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
| POST | /auth/login | [0198042a f9c5 7547 a6e7 567af5db26cd](#0198042a-f9c5-7547-a6e7-567af5db26cd) | Login user |
| POST | /auth/register | [0198042a f9c5 75c8 9231 ad5fc9e7b32e](#0198042a-f9c5-75c8-9231-ad5fc9e7b32e) | Register user |
| GET | /auth/verify/{jwt} | [0198042a f9c5 75cc 9dd2 e3ff9f6c1e3a](#0198042a-f9c5-75cc-9dd2-e3ff9f6c1e3a) | Verify user |
| POST | /auth/verify | [0198042a f9c5 75d0 8c20 fea31b65587f](#0198042a-f9c5-75d0-8c20-fea31b65587f) | Resend verification |
| DELETE | /auth/logout | [0198042a f9c5 75d4 afa6 fe658744c80f](#0198042a-f9c5-75d4-afa6-fe658744c80f) | Logout user |
| POST | /auth/refresh | [0198042a f9c5 75d8 aa7b 37524ea4f124](#0198042a-f9c5-75d8-aa7b-37524ea4f124) | Refresh access token |
  


###  health

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /health/status | [0198042a f9c5 76be ba9e 8186a69f48c4](#0198042a-f9c5-76be-ba9e-8186a69f48c4) | Check health |
  


###  policies

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /policies/{policy_id} | [0198042a f9c5 76c2 96f2 d16b0674bcd9](#0198042a-f9c5-76c2-96f2-d16b0674bcd9) | Get policy |
| POST | /policies | [0198042a f9c5 76c6 9a07 0c8948640ac2](#0198042a-f9c5-76c6-9a07-0c8948640ac2) | Create policy |
| PUT | /policies/{policy_id} | [0198042a f9c5 76ca b40d b1de1d359c22](#0198042a-f9c5-76ca-b40d-b1de1d359c22) | Update policy |
| DELETE | /policies/{policy_id} | [0198042a f9c5 76ce b208 2f58f7ccd177](#0198042a-f9c5-76ce-b208-2f58f7ccd177) | Delete policy |
| GET | /policies | [0198042a f9c5 76d2 a491 9cc989c1d59c](#0198042a-f9c5-76d2-a491-9cc989c1d59c) | List policies |
| POST | /policies/{policy_id}/roles | [0198042a f9c5 76d6 b1f3 0bfb57a9197f](#0198042a-f9c5-76d6-b1f3-0bfb57a9197f) | Link roles to policy |
| DELETE | /policies/{policy_id}/roles | [0198042a f9c5 76d9 8019 babd51a0c340](#0198042a-f9c5-76d9-8019-babd51a0c340) | Unlink roles from policy |
| GET | /roles/{role_id}/policies | [0198042a f9c5 76dd 8fa8 98df6be12d44](#0198042a-f9c5-76dd-8fa8-98df6be12d44) | List policies by role |
  


###  products

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /projects/{project_id}/products/{product_id} | [0198042a f9c5 7603 99b1 7c20ee58542b](#0198042a-f9c5-7603-99b1-7c20ee58542b) | Get product |
| POST | /projects/{project_id}/products | [0198042a f9c5 7606 8aab 1c2db5b81a89](#0198042a-f9c5-7606-8aab-1c2db5b81a89) | Create product |
| PUT | /projects/{project_id}/products/{product_id} | [0198042a f9c5 7607 b75a 532912a6f35d](#0198042a-f9c5-7607-b75a-532912a6f35d) | Update product |
| DELETE | /projects/{project_id}/products/{product_id} | [0198042a f9c5 760a 99c8 1f68d597d300](#0198042a-f9c5-760a-99c8-1f68d597d300) | Delete product |
| GET | /projects/{project_id}/products | [0198042a f9c5 760e 9d2f 94cce8243e5a](#0198042a-f9c5-760e-9d2f-94cce8243e5a) | List products by project |
| GET | /products | [0198042a f9c5 7612 a055 58177eca0772](#0198042a-f9c5-7612-a055-58177eca0772) | List products |
| POST | /projects/{project_id}/products/{product_id}/payment_processor | [0198042a f9c5 7616 8c3b e4f19d83a033](#0198042a-f9c5-7616-8c3b-e4f19d83a033) | Link product to payment processor |
| DELETE | /projects/{project_id}/products/{product_id}/payment_processor | [0198042a f9c5 761a bd02 da039b52bea2](#0198042a-f9c5-761a-bd02-da039b52bea2) | Unlink product from payment processor |
  


###  projects

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /projects/{project_id} | [0198042a f9c5 761e b1c2 66a3f8ab30d6](#0198042a-f9c5-761e-b1c2-66a3f8ab30d6) | Get project |
| POST | /projects | [0198042a f9c5 7622 9142 88fbaa727659](#0198042a-f9c5-7622-9142-88fbaa727659) | Create project |
| PUT | /projects/{project_id} | [0198042a f9c5 7626 be9f 996a2898ef07](#0198042a-f9c5-7626-be9f-996a2898ef07) | Update project |
| DELETE | /projects/{project_id} | [0198042a f9c5 762a 8033 649a1526901d](#0198042a-f9c5-762a-8033-649a1526901d) | Delete project |
| GET | /projects | [0198042a f9c5 76a7 a480 fbcb978b8501](#0198042a-f9c5-76a7-a480-fbcb978b8501) | List projects |
  


###  resources

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /resources/{resource_id} | [0198042a f9c5 76b2 b8b1 bc0223a0f18d](#0198042a-f9c5-76b2-b8b1-bc0223a0f18d) | Get resource |
| GET | /resources | [0198042a f9c5 76b6 bd55 f34dff7b0632](#0198042a-f9c5-76b6-bd55-f34dff7b0632) | List resources |
| GET | /resources/matches | [0198042a f9c5 76ba bc87 6e9e32988407](#0198042a-f9c5-76ba-bc87-6e9e32988407) | Match resources |
  


###  roles

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /roles/{role_id} | [0198042a f9c5 76e1 a650 772c826f079e](#0198042a-f9c5-76e1-a650-772c826f079e) | Get role |
| POST | /roles | [0198042a f9c5 76e5 8fe5 b93a07311c47](#0198042a-f9c5-76e5-8fe5-b93a07311c47) | Create role |
| PUT | /roles/{role_id} | [0198042a f9c5 76e9 922d 2411530cd8f8](#0198042a-f9c5-76e9-922d-2411530cd8f8) | Update role |
| DELETE | /roles/{role_id} | [0198042a f9c5 76ed 99a5 84923071fa6b](#0198042a-f9c5-76ed-99a5-84923071fa6b) | Delete role |
| GET | /roles | [0198042a f9c5 76f1 9cf8 37e45b647fc0](#0198042a-f9c5-76f1-9cf8-37e45b647fc0) | List roles |
| POST | /roles/{role_id}/users | [0198042a f9c5 76f5 8ff6 b4479bdaa6b6](#0198042a-f9c5-76f5-8ff6-b4479bdaa6b6) | Link users to role |
| DELETE | /roles/{role_id}/users | [0198042a f9c5 76f9 9394 170db55f62f4](#0198042a-f9c5-76f9-9394-170db55f62f4) | Unlink users from role |
| POST | /roles/{role_id}/policies | [0198042a f9c5 76fd 8012 5c9a2957e289](#0198042a-f9c5-76fd-8012-5c9a2957e289) | Link policies to role |
| DELETE | /roles/{role_id}/policies | [0198042a f9c5 7700 9e40 e64f7b8c947c](#0198042a-f9c5-7700-9e40-e64f7b8c947c) | Unlink policies from role |
| GET | /users/{user_id}/roles | [0198042a f9c5 7704 b73b 55e2ec093586](#0198042a-f9c5-7704-b73b-55e2ec093586) | List roles by user |
| GET | /policies/{policy_id}/roles | [0198042a f9c5 7704 b73b 55e2ec093587](#0198042a-f9c5-7704-b73b-55e2ec093587) | List roles by policy |
  


###  users

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /users/{user_id} | [0198042a f9c5 75df b843 b92a4d5c590e](#0198042a-f9c5-75df-b843-b92a4d5c590e) | Get user |
| POST | /users | [0198042a f9c5 75e3 acf6 6901bb33ae65](#0198042a-f9c5-75e3-acf6-6901bb33ae65) | Create user |
| PUT | /users/{user_id} | [0198042a f9c5 75e7 8cb9 231bee55c64e](#0198042a-f9c5-75e7-8cb9-231bee55c64e) | Update user |
| DELETE | /users/{user_id} | [0198042a f9c5 75eb b683 6c1847af7108](#0198042a-f9c5-75eb-b683-6c1847af7108) | Delete user |
| GET | /users | [0198042a f9c5 75ef 8ea1 29ecbbe01a2e](#0198042a-f9c5-75ef-8ea1-29ecbbe01a2e) | List users |
| POST | /users/{user_id}/roles | [0198042a f9c5 75f3 985f d30e67bb3688](#0198042a-f9c5-75f3-985f-d30e67bb3688) | Link roles to user |
| DELETE | /users/{user_id}/roles | [0198042a f9c5 75f7 b802 343518ee3788](#0198042a-f9c5-75f7-b802-343518ee3788) | Unlink roles from user |
| GET | /users/{user_id}/authz | [0198042a f9c5 75fb b324 ec962beb2277](#0198042a-f9c5-75fb-b324-ec962beb2277) | Get user authorization |
| GET | /roles/{role_id}/users | [0198042a f9c5 75ff bbfc 224bf4342886](#0198042a-f9c5-75ff-bbfc-224bf4342886) | List users by role |
  


###  version

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /version | [0198042a f9c5 7704 b73b 55e2ec093588](#0198042a-f9c5-7704-b73b-55e2ec093588) | Get version |
  


## Paths

### <span id="0198042a-f9c5-7547-a6e7-567af5db26cd"></span> Login user (*0198042a-f9c5-7547-a6e7-567af5db26cd*)

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
| [200](#0198042a-f9c5-7547-a6e7-567af5db26cd-200) | OK | OK |  | [schema](#0198042a-f9c5-7547-a6e7-567af5db26cd-200-schema) |
| [400](#0198042a-f9c5-7547-a6e7-567af5db26cd-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-7547-a6e7-567af5db26cd-400-schema) |
| [401](#0198042a-f9c5-7547-a6e7-567af5db26cd-401) | Unauthorized | Unauthorized |  | [schema](#0198042a-f9c5-7547-a6e7-567af5db26cd-401-schema) |
| [500](#0198042a-f9c5-7547-a6e7-567af5db26cd-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-7547-a6e7-567af5db26cd-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-7547-a6e7-567af5db26cd-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-7547-a6e7-567af5db26cd-200-schema"></span> Schema
   
  

[ModelLoginUserResponse](#model-login-user-response)

##### <span id="0198042a-f9c5-7547-a6e7-567af5db26cd-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-7547-a6e7-567af5db26cd-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-7547-a6e7-567af5db26cd-401"></span> 401 - Unauthorized
Status: Unauthorized

###### <span id="0198042a-f9c5-7547-a6e7-567af5db26cd-401-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-7547-a6e7-567af5db26cd-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-7547-a6e7-567af5db26cd-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-75c8-9231-ad5fc9e7b32e"></span> Register user (*0198042a-f9c5-75c8-9231-ad5fc9e7b32e*)

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
| [201](#0198042a-f9c5-75c8-9231-ad5fc9e7b32e-201) | Created | Created |  | [schema](#0198042a-f9c5-75c8-9231-ad5fc9e7b32e-201-schema) |
| [400](#0198042a-f9c5-75c8-9231-ad5fc9e7b32e-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-75c8-9231-ad5fc9e7b32e-400-schema) |
| [500](#0198042a-f9c5-75c8-9231-ad5fc9e7b32e-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-75c8-9231-ad5fc9e7b32e-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-75c8-9231-ad5fc9e7b32e-201"></span> 201 - Created
Status: Created

###### <span id="0198042a-f9c5-75c8-9231-ad5fc9e7b32e-201-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75c8-9231-ad5fc9e7b32e-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-75c8-9231-ad5fc9e7b32e-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75c8-9231-ad5fc9e7b32e-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-75c8-9231-ad5fc9e7b32e-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-75cc-9dd2-e3ff9f6c1e3a"></span> Verify user (*0198042a-f9c5-75cc-9dd2-e3ff9f6c1e3a*)

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
| [200](#0198042a-f9c5-75cc-9dd2-e3ff9f6c1e3a-200) | OK | OK |  | [schema](#0198042a-f9c5-75cc-9dd2-e3ff9f6c1e3a-200-schema) |
| [400](#0198042a-f9c5-75cc-9dd2-e3ff9f6c1e3a-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-75cc-9dd2-e3ff9f6c1e3a-400-schema) |
| [401](#0198042a-f9c5-75cc-9dd2-e3ff9f6c1e3a-401) | Unauthorized | Unauthorized |  | [schema](#0198042a-f9c5-75cc-9dd2-e3ff9f6c1e3a-401-schema) |
| [404](#0198042a-f9c5-75cc-9dd2-e3ff9f6c1e3a-404) | Not Found | Not Found |  | [schema](#0198042a-f9c5-75cc-9dd2-e3ff9f6c1e3a-404-schema) |
| [500](#0198042a-f9c5-75cc-9dd2-e3ff9f6c1e3a-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-75cc-9dd2-e3ff9f6c1e3a-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-75cc-9dd2-e3ff9f6c1e3a-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-75cc-9dd2-e3ff9f6c1e3a-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75cc-9dd2-e3ff9f6c1e3a-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-75cc-9dd2-e3ff9f6c1e3a-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75cc-9dd2-e3ff9f6c1e3a-401"></span> 401 - Unauthorized
Status: Unauthorized

###### <span id="0198042a-f9c5-75cc-9dd2-e3ff9f6c1e3a-401-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75cc-9dd2-e3ff9f6c1e3a-404"></span> 404 - Not Found
Status: Not Found

###### <span id="0198042a-f9c5-75cc-9dd2-e3ff9f6c1e3a-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75cc-9dd2-e3ff9f6c1e3a-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-75cc-9dd2-e3ff9f6c1e3a-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-75d0-8c20-fea31b65587f"></span> Resend verification (*0198042a-f9c5-75d0-8c20-fea31b65587f*)

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
| [200](#0198042a-f9c5-75d0-8c20-fea31b65587f-200) | OK | OK |  | [schema](#0198042a-f9c5-75d0-8c20-fea31b65587f-200-schema) |
| [400](#0198042a-f9c5-75d0-8c20-fea31b65587f-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-75d0-8c20-fea31b65587f-400-schema) |
| [401](#0198042a-f9c5-75d0-8c20-fea31b65587f-401) | Unauthorized | Unauthorized |  | [schema](#0198042a-f9c5-75d0-8c20-fea31b65587f-401-schema) |
| [404](#0198042a-f9c5-75d0-8c20-fea31b65587f-404) | Not Found | Not Found |  | [schema](#0198042a-f9c5-75d0-8c20-fea31b65587f-404-schema) |
| [500](#0198042a-f9c5-75d0-8c20-fea31b65587f-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-75d0-8c20-fea31b65587f-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-75d0-8c20-fea31b65587f-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-75d0-8c20-fea31b65587f-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75d0-8c20-fea31b65587f-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-75d0-8c20-fea31b65587f-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75d0-8c20-fea31b65587f-401"></span> 401 - Unauthorized
Status: Unauthorized

###### <span id="0198042a-f9c5-75d0-8c20-fea31b65587f-401-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75d0-8c20-fea31b65587f-404"></span> 404 - Not Found
Status: Not Found

###### <span id="0198042a-f9c5-75d0-8c20-fea31b65587f-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75d0-8c20-fea31b65587f-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-75d0-8c20-fea31b65587f-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-75d4-afa6-fe658744c80f"></span> Logout user (*0198042a-f9c5-75d4-afa6-fe658744c80f*)

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
| [200](#0198042a-f9c5-75d4-afa6-fe658744c80f-200) | OK | OK |  | [schema](#0198042a-f9c5-75d4-afa6-fe658744c80f-200-schema) |
| [401](#0198042a-f9c5-75d4-afa6-fe658744c80f-401) | Unauthorized | Unauthorized |  | [schema](#0198042a-f9c5-75d4-afa6-fe658744c80f-401-schema) |
| [500](#0198042a-f9c5-75d4-afa6-fe658744c80f-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-75d4-afa6-fe658744c80f-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-75d4-afa6-fe658744c80f-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-75d4-afa6-fe658744c80f-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75d4-afa6-fe658744c80f-401"></span> 401 - Unauthorized
Status: Unauthorized

###### <span id="0198042a-f9c5-75d4-afa6-fe658744c80f-401-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75d4-afa6-fe658744c80f-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-75d4-afa6-fe658744c80f-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-75d8-aa7b-37524ea4f124"></span> Refresh access token (*0198042a-f9c5-75d8-aa7b-37524ea4f124*)

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
| [200](#0198042a-f9c5-75d8-aa7b-37524ea4f124-200) | OK | OK |  | [schema](#0198042a-f9c5-75d8-aa7b-37524ea4f124-200-schema) |
| [400](#0198042a-f9c5-75d8-aa7b-37524ea4f124-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-75d8-aa7b-37524ea4f124-400-schema) |
| [401](#0198042a-f9c5-75d8-aa7b-37524ea4f124-401) | Unauthorized | Unauthorized |  | [schema](#0198042a-f9c5-75d8-aa7b-37524ea4f124-401-schema) |
| [404](#0198042a-f9c5-75d8-aa7b-37524ea4f124-404) | Not Found | Not Found |  | [schema](#0198042a-f9c5-75d8-aa7b-37524ea4f124-404-schema) |
| [500](#0198042a-f9c5-75d8-aa7b-37524ea4f124-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-75d8-aa7b-37524ea4f124-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-75d8-aa7b-37524ea4f124-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-75d8-aa7b-37524ea4f124-200-schema"></span> Schema
   
  

[ModelRefreshTokenResponse](#model-refresh-token-response)

##### <span id="0198042a-f9c5-75d8-aa7b-37524ea4f124-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-75d8-aa7b-37524ea4f124-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75d8-aa7b-37524ea4f124-401"></span> 401 - Unauthorized
Status: Unauthorized

###### <span id="0198042a-f9c5-75d8-aa7b-37524ea4f124-401-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75d8-aa7b-37524ea4f124-404"></span> 404 - Not Found
Status: Not Found

###### <span id="0198042a-f9c5-75d8-aa7b-37524ea4f124-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75d8-aa7b-37524ea4f124-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-75d8-aa7b-37524ea4f124-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-75df-b843-b92a4d5c590e"></span> Get user (*0198042a-f9c5-75df-b843-b92a4d5c590e*)

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
| [200](#0198042a-f9c5-75df-b843-b92a4d5c590e-200) | OK | OK |  | [schema](#0198042a-f9c5-75df-b843-b92a4d5c590e-200-schema) |
| [400](#0198042a-f9c5-75df-b843-b92a4d5c590e-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-75df-b843-b92a4d5c590e-400-schema) |
| [404](#0198042a-f9c5-75df-b843-b92a4d5c590e-404) | Not Found | Not Found |  | [schema](#0198042a-f9c5-75df-b843-b92a4d5c590e-404-schema) |
| [500](#0198042a-f9c5-75df-b843-b92a4d5c590e-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-75df-b843-b92a4d5c590e-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-75df-b843-b92a4d5c590e-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-75df-b843-b92a4d5c590e-200-schema"></span> Schema
   
  

[ModelUser](#model-user)

##### <span id="0198042a-f9c5-75df-b843-b92a4d5c590e-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-75df-b843-b92a4d5c590e-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75df-b843-b92a4d5c590e-404"></span> 404 - Not Found
Status: Not Found

###### <span id="0198042a-f9c5-75df-b843-b92a4d5c590e-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75df-b843-b92a4d5c590e-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-75df-b843-b92a4d5c590e-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-75e3-acf6-6901bb33ae65"></span> Create user (*0198042a-f9c5-75e3-acf6-6901bb33ae65*)

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
| [201](#0198042a-f9c5-75e3-acf6-6901bb33ae65-201) | Created | Created |  | [schema](#0198042a-f9c5-75e3-acf6-6901bb33ae65-201-schema) |
| [400](#0198042a-f9c5-75e3-acf6-6901bb33ae65-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-75e3-acf6-6901bb33ae65-400-schema) |
| [409](#0198042a-f9c5-75e3-acf6-6901bb33ae65-409) | Conflict | Conflict |  | [schema](#0198042a-f9c5-75e3-acf6-6901bb33ae65-409-schema) |
| [500](#0198042a-f9c5-75e3-acf6-6901bb33ae65-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-75e3-acf6-6901bb33ae65-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-75e3-acf6-6901bb33ae65-201"></span> 201 - Created
Status: Created

###### <span id="0198042a-f9c5-75e3-acf6-6901bb33ae65-201-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75e3-acf6-6901bb33ae65-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-75e3-acf6-6901bb33ae65-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75e3-acf6-6901bb33ae65-409"></span> 409 - Conflict
Status: Conflict

###### <span id="0198042a-f9c5-75e3-acf6-6901bb33ae65-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75e3-acf6-6901bb33ae65-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-75e3-acf6-6901bb33ae65-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-75e7-8cb9-231bee55c64e"></span> Update user (*0198042a-f9c5-75e7-8cb9-231bee55c64e*)

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
| [200](#0198042a-f9c5-75e7-8cb9-231bee55c64e-200) | OK | OK |  | [schema](#0198042a-f9c5-75e7-8cb9-231bee55c64e-200-schema) |
| [400](#0198042a-f9c5-75e7-8cb9-231bee55c64e-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-75e7-8cb9-231bee55c64e-400-schema) |
| [409](#0198042a-f9c5-75e7-8cb9-231bee55c64e-409) | Conflict | Conflict |  | [schema](#0198042a-f9c5-75e7-8cb9-231bee55c64e-409-schema) |
| [500](#0198042a-f9c5-75e7-8cb9-231bee55c64e-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-75e7-8cb9-231bee55c64e-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-75e7-8cb9-231bee55c64e-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-75e7-8cb9-231bee55c64e-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75e7-8cb9-231bee55c64e-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-75e7-8cb9-231bee55c64e-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75e7-8cb9-231bee55c64e-409"></span> 409 - Conflict
Status: Conflict

###### <span id="0198042a-f9c5-75e7-8cb9-231bee55c64e-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75e7-8cb9-231bee55c64e-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-75e7-8cb9-231bee55c64e-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-75eb-b683-6c1847af7108"></span> Delete user (*0198042a-f9c5-75eb-b683-6c1847af7108*)

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
| [200](#0198042a-f9c5-75eb-b683-6c1847af7108-200) | OK | OK |  | [schema](#0198042a-f9c5-75eb-b683-6c1847af7108-200-schema) |
| [400](#0198042a-f9c5-75eb-b683-6c1847af7108-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-75eb-b683-6c1847af7108-400-schema) |
| [500](#0198042a-f9c5-75eb-b683-6c1847af7108-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-75eb-b683-6c1847af7108-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-75eb-b683-6c1847af7108-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-75eb-b683-6c1847af7108-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75eb-b683-6c1847af7108-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-75eb-b683-6c1847af7108-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75eb-b683-6c1847af7108-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-75eb-b683-6c1847af7108-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-75ef-8ea1-29ecbbe01a2e"></span> List users (*0198042a-f9c5-75ef-8ea1-29ecbbe01a2e*)

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
| [200](#0198042a-f9c5-75ef-8ea1-29ecbbe01a2e-200) | OK | OK |  | [schema](#0198042a-f9c5-75ef-8ea1-29ecbbe01a2e-200-schema) |
| [400](#0198042a-f9c5-75ef-8ea1-29ecbbe01a2e-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-75ef-8ea1-29ecbbe01a2e-400-schema) |
| [500](#0198042a-f9c5-75ef-8ea1-29ecbbe01a2e-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-75ef-8ea1-29ecbbe01a2e-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-75ef-8ea1-29ecbbe01a2e-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-75ef-8ea1-29ecbbe01a2e-200-schema"></span> Schema
   
  

[ModelListUsersResponse](#model-list-users-response)

##### <span id="0198042a-f9c5-75ef-8ea1-29ecbbe01a2e-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-75ef-8ea1-29ecbbe01a2e-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75ef-8ea1-29ecbbe01a2e-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-75ef-8ea1-29ecbbe01a2e-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-75f3-985f-d30e67bb3688"></span> Link roles to user (*0198042a-f9c5-75f3-985f-d30e67bb3688*)

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
| [200](#0198042a-f9c5-75f3-985f-d30e67bb3688-200) | OK | OK |  | [schema](#0198042a-f9c5-75f3-985f-d30e67bb3688-200-schema) |
| [400](#0198042a-f9c5-75f3-985f-d30e67bb3688-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-75f3-985f-d30e67bb3688-400-schema) |
| [409](#0198042a-f9c5-75f3-985f-d30e67bb3688-409) | Conflict | Conflict |  | [schema](#0198042a-f9c5-75f3-985f-d30e67bb3688-409-schema) |
| [500](#0198042a-f9c5-75f3-985f-d30e67bb3688-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-75f3-985f-d30e67bb3688-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-75f3-985f-d30e67bb3688-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-75f3-985f-d30e67bb3688-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75f3-985f-d30e67bb3688-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-75f3-985f-d30e67bb3688-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75f3-985f-d30e67bb3688-409"></span> 409 - Conflict
Status: Conflict

###### <span id="0198042a-f9c5-75f3-985f-d30e67bb3688-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75f3-985f-d30e67bb3688-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-75f3-985f-d30e67bb3688-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-75f7-b802-343518ee3788"></span> Unlink roles from user (*0198042a-f9c5-75f7-b802-343518ee3788*)

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
| [200](#0198042a-f9c5-75f7-b802-343518ee3788-200) | OK | OK |  | [schema](#0198042a-f9c5-75f7-b802-343518ee3788-200-schema) |
| [400](#0198042a-f9c5-75f7-b802-343518ee3788-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-75f7-b802-343518ee3788-400-schema) |
| [409](#0198042a-f9c5-75f7-b802-343518ee3788-409) | Conflict | Conflict |  | [schema](#0198042a-f9c5-75f7-b802-343518ee3788-409-schema) |
| [500](#0198042a-f9c5-75f7-b802-343518ee3788-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-75f7-b802-343518ee3788-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-75f7-b802-343518ee3788-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-75f7-b802-343518ee3788-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75f7-b802-343518ee3788-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-75f7-b802-343518ee3788-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75f7-b802-343518ee3788-409"></span> 409 - Conflict
Status: Conflict

###### <span id="0198042a-f9c5-75f7-b802-343518ee3788-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75f7-b802-343518ee3788-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-75f7-b802-343518ee3788-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-75fb-b324-ec962beb2277"></span> Get user authorization (*0198042a-f9c5-75fb-b324-ec962beb2277*)

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
| [200](#0198042a-f9c5-75fb-b324-ec962beb2277-200) | OK | OK |  | [schema](#0198042a-f9c5-75fb-b324-ec962beb2277-200-schema) |
| [400](#0198042a-f9c5-75fb-b324-ec962beb2277-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-75fb-b324-ec962beb2277-400-schema) |
| [404](#0198042a-f9c5-75fb-b324-ec962beb2277-404) | Not Found | Not Found |  | [schema](#0198042a-f9c5-75fb-b324-ec962beb2277-404-schema) |
| [500](#0198042a-f9c5-75fb-b324-ec962beb2277-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-75fb-b324-ec962beb2277-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-75fb-b324-ec962beb2277-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-75fb-b324-ec962beb2277-200-schema"></span> Schema
   
  

any

##### <span id="0198042a-f9c5-75fb-b324-ec962beb2277-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-75fb-b324-ec962beb2277-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75fb-b324-ec962beb2277-404"></span> 404 - Not Found
Status: Not Found

###### <span id="0198042a-f9c5-75fb-b324-ec962beb2277-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75fb-b324-ec962beb2277-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-75fb-b324-ec962beb2277-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-75ff-bbfc-224bf4342886"></span> List users by role (*0198042a-f9c5-75ff-bbfc-224bf4342886*)

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
| [200](#0198042a-f9c5-75ff-bbfc-224bf4342886-200) | OK | OK |  | [schema](#0198042a-f9c5-75ff-bbfc-224bf4342886-200-schema) |
| [400](#0198042a-f9c5-75ff-bbfc-224bf4342886-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-75ff-bbfc-224bf4342886-400-schema) |
| [500](#0198042a-f9c5-75ff-bbfc-224bf4342886-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-75ff-bbfc-224bf4342886-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-75ff-bbfc-224bf4342886-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-75ff-bbfc-224bf4342886-200-schema"></span> Schema
   
  

[ModelListUsersResponse](#model-list-users-response)

##### <span id="0198042a-f9c5-75ff-bbfc-224bf4342886-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-75ff-bbfc-224bf4342886-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-75ff-bbfc-224bf4342886-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-75ff-bbfc-224bf4342886-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-7603-99b1-7c20ee58542b"></span> Get product (*0198042a-f9c5-7603-99b1-7c20ee58542b*)

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
| [200](#0198042a-f9c5-7603-99b1-7c20ee58542b-200) | OK | OK |  | [schema](#0198042a-f9c5-7603-99b1-7c20ee58542b-200-schema) |
| [400](#0198042a-f9c5-7603-99b1-7c20ee58542b-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-7603-99b1-7c20ee58542b-400-schema) |
| [404](#0198042a-f9c5-7603-99b1-7c20ee58542b-404) | Not Found | Not Found |  | [schema](#0198042a-f9c5-7603-99b1-7c20ee58542b-404-schema) |
| [500](#0198042a-f9c5-7603-99b1-7c20ee58542b-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-7603-99b1-7c20ee58542b-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-7603-99b1-7c20ee58542b-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-7603-99b1-7c20ee58542b-200-schema"></span> Schema
   
  

[ModelProduct](#model-product)

##### <span id="0198042a-f9c5-7603-99b1-7c20ee58542b-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-7603-99b1-7c20ee58542b-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-7603-99b1-7c20ee58542b-404"></span> 404 - Not Found
Status: Not Found

###### <span id="0198042a-f9c5-7603-99b1-7c20ee58542b-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-7603-99b1-7c20ee58542b-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-7603-99b1-7c20ee58542b-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-7606-8aab-1c2db5b81a89"></span> Create product (*0198042a-f9c5-7606-8aab-1c2db5b81a89*)

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
| [201](#0198042a-f9c5-7606-8aab-1c2db5b81a89-201) | Created | Created |  | [schema](#0198042a-f9c5-7606-8aab-1c2db5b81a89-201-schema) |
| [400](#0198042a-f9c5-7606-8aab-1c2db5b81a89-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-7606-8aab-1c2db5b81a89-400-schema) |
| [409](#0198042a-f9c5-7606-8aab-1c2db5b81a89-409) | Conflict | Conflict |  | [schema](#0198042a-f9c5-7606-8aab-1c2db5b81a89-409-schema) |
| [500](#0198042a-f9c5-7606-8aab-1c2db5b81a89-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-7606-8aab-1c2db5b81a89-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-7606-8aab-1c2db5b81a89-201"></span> 201 - Created
Status: Created

###### <span id="0198042a-f9c5-7606-8aab-1c2db5b81a89-201-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-7606-8aab-1c2db5b81a89-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-7606-8aab-1c2db5b81a89-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-7606-8aab-1c2db5b81a89-409"></span> 409 - Conflict
Status: Conflict

###### <span id="0198042a-f9c5-7606-8aab-1c2db5b81a89-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-7606-8aab-1c2db5b81a89-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-7606-8aab-1c2db5b81a89-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-7607-b75a-532912a6f35d"></span> Update product (*0198042a-f9c5-7607-b75a-532912a6f35d*)

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
| [200](#0198042a-f9c5-7607-b75a-532912a6f35d-200) | OK | OK |  | [schema](#0198042a-f9c5-7607-b75a-532912a6f35d-200-schema) |
| [400](#0198042a-f9c5-7607-b75a-532912a6f35d-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-7607-b75a-532912a6f35d-400-schema) |
| [409](#0198042a-f9c5-7607-b75a-532912a6f35d-409) | Conflict | Conflict |  | [schema](#0198042a-f9c5-7607-b75a-532912a6f35d-409-schema) |
| [500](#0198042a-f9c5-7607-b75a-532912a6f35d-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-7607-b75a-532912a6f35d-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-7607-b75a-532912a6f35d-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-7607-b75a-532912a6f35d-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-7607-b75a-532912a6f35d-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-7607-b75a-532912a6f35d-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-7607-b75a-532912a6f35d-409"></span> 409 - Conflict
Status: Conflict

###### <span id="0198042a-f9c5-7607-b75a-532912a6f35d-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-7607-b75a-532912a6f35d-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-7607-b75a-532912a6f35d-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-760a-99c8-1f68d597d300"></span> Delete product (*0198042a-f9c5-760a-99c8-1f68d597d300*)

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
| [200](#0198042a-f9c5-760a-99c8-1f68d597d300-200) | OK | OK |  | [schema](#0198042a-f9c5-760a-99c8-1f68d597d300-200-schema) |
| [400](#0198042a-f9c5-760a-99c8-1f68d597d300-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-760a-99c8-1f68d597d300-400-schema) |
| [500](#0198042a-f9c5-760a-99c8-1f68d597d300-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-760a-99c8-1f68d597d300-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-760a-99c8-1f68d597d300-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-760a-99c8-1f68d597d300-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-760a-99c8-1f68d597d300-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-760a-99c8-1f68d597d300-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-760a-99c8-1f68d597d300-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-760a-99c8-1f68d597d300-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-760e-9d2f-94cce8243e5a"></span> List products by project (*0198042a-f9c5-760e-9d2f-94cce8243e5a*)

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
| [200](#0198042a-f9c5-760e-9d2f-94cce8243e5a-200) | OK | OK |  | [schema](#0198042a-f9c5-760e-9d2f-94cce8243e5a-200-schema) |
| [400](#0198042a-f9c5-760e-9d2f-94cce8243e5a-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-760e-9d2f-94cce8243e5a-400-schema) |
| [500](#0198042a-f9c5-760e-9d2f-94cce8243e5a-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-760e-9d2f-94cce8243e5a-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-760e-9d2f-94cce8243e5a-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-760e-9d2f-94cce8243e5a-200-schema"></span> Schema
   
  

[ModelListProductsResponse](#model-list-products-response)

##### <span id="0198042a-f9c5-760e-9d2f-94cce8243e5a-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-760e-9d2f-94cce8243e5a-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-760e-9d2f-94cce8243e5a-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-760e-9d2f-94cce8243e5a-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-7612-a055-58177eca0772"></span> List products (*0198042a-f9c5-7612-a055-58177eca0772*)

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
| [200](#0198042a-f9c5-7612-a055-58177eca0772-200) | OK | OK |  | [schema](#0198042a-f9c5-7612-a055-58177eca0772-200-schema) |
| [400](#0198042a-f9c5-7612-a055-58177eca0772-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-7612-a055-58177eca0772-400-schema) |
| [500](#0198042a-f9c5-7612-a055-58177eca0772-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-7612-a055-58177eca0772-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-7612-a055-58177eca0772-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-7612-a055-58177eca0772-200-schema"></span> Schema
   
  

[ModelListProductsResponse](#model-list-products-response)

##### <span id="0198042a-f9c5-7612-a055-58177eca0772-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-7612-a055-58177eca0772-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-7612-a055-58177eca0772-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-7612-a055-58177eca0772-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-7616-8c3b-e4f19d83a033"></span> Link product to payment processor (*0198042a-f9c5-7616-8c3b-e4f19d83a033*)

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
| [200](#0198042a-f9c5-7616-8c3b-e4f19d83a033-200) | OK | OK |  | [schema](#0198042a-f9c5-7616-8c3b-e4f19d83a033-200-schema) |
| [400](#0198042a-f9c5-7616-8c3b-e4f19d83a033-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-7616-8c3b-e4f19d83a033-400-schema) |
| [500](#0198042a-f9c5-7616-8c3b-e4f19d83a033-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-7616-8c3b-e4f19d83a033-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-7616-8c3b-e4f19d83a033-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-7616-8c3b-e4f19d83a033-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-7616-8c3b-e4f19d83a033-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-7616-8c3b-e4f19d83a033-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-7616-8c3b-e4f19d83a033-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-7616-8c3b-e4f19d83a033-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-761a-bd02-da039b52bea2"></span> Unlink product from payment processor (*0198042a-f9c5-761a-bd02-da039b52bea2*)

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
| [200](#0198042a-f9c5-761a-bd02-da039b52bea2-200) | OK | OK |  | [schema](#0198042a-f9c5-761a-bd02-da039b52bea2-200-schema) |
| [400](#0198042a-f9c5-761a-bd02-da039b52bea2-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-761a-bd02-da039b52bea2-400-schema) |
| [500](#0198042a-f9c5-761a-bd02-da039b52bea2-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-761a-bd02-da039b52bea2-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-761a-bd02-da039b52bea2-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-761a-bd02-da039b52bea2-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-761a-bd02-da039b52bea2-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-761a-bd02-da039b52bea2-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-761a-bd02-da039b52bea2-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-761a-bd02-da039b52bea2-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-761e-b1c2-66a3f8ab30d6"></span> Get project (*0198042a-f9c5-761e-b1c2-66a3f8ab30d6*)

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
| [200](#0198042a-f9c5-761e-b1c2-66a3f8ab30d6-200) | OK | OK |  | [schema](#0198042a-f9c5-761e-b1c2-66a3f8ab30d6-200-schema) |
| [400](#0198042a-f9c5-761e-b1c2-66a3f8ab30d6-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-761e-b1c2-66a3f8ab30d6-400-schema) |
| [404](#0198042a-f9c5-761e-b1c2-66a3f8ab30d6-404) | Not Found | Not Found |  | [schema](#0198042a-f9c5-761e-b1c2-66a3f8ab30d6-404-schema) |
| [500](#0198042a-f9c5-761e-b1c2-66a3f8ab30d6-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-761e-b1c2-66a3f8ab30d6-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-761e-b1c2-66a3f8ab30d6-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-761e-b1c2-66a3f8ab30d6-200-schema"></span> Schema
   
  

[ModelProject](#model-project)

##### <span id="0198042a-f9c5-761e-b1c2-66a3f8ab30d6-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-761e-b1c2-66a3f8ab30d6-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-761e-b1c2-66a3f8ab30d6-404"></span> 404 - Not Found
Status: Not Found

###### <span id="0198042a-f9c5-761e-b1c2-66a3f8ab30d6-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-761e-b1c2-66a3f8ab30d6-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-761e-b1c2-66a3f8ab30d6-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-7622-9142-88fbaa727659"></span> Create project (*0198042a-f9c5-7622-9142-88fbaa727659*)

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
| [201](#0198042a-f9c5-7622-9142-88fbaa727659-201) | Created | Created |  | [schema](#0198042a-f9c5-7622-9142-88fbaa727659-201-schema) |
| [400](#0198042a-f9c5-7622-9142-88fbaa727659-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-7622-9142-88fbaa727659-400-schema) |
| [409](#0198042a-f9c5-7622-9142-88fbaa727659-409) | Conflict | Conflict |  | [schema](#0198042a-f9c5-7622-9142-88fbaa727659-409-schema) |
| [500](#0198042a-f9c5-7622-9142-88fbaa727659-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-7622-9142-88fbaa727659-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-7622-9142-88fbaa727659-201"></span> 201 - Created
Status: Created

###### <span id="0198042a-f9c5-7622-9142-88fbaa727659-201-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-7622-9142-88fbaa727659-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-7622-9142-88fbaa727659-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-7622-9142-88fbaa727659-409"></span> 409 - Conflict
Status: Conflict

###### <span id="0198042a-f9c5-7622-9142-88fbaa727659-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-7622-9142-88fbaa727659-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-7622-9142-88fbaa727659-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-7626-be9f-996a2898ef07"></span> Update project (*0198042a-f9c5-7626-be9f-996a2898ef07*)

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
| [200](#0198042a-f9c5-7626-be9f-996a2898ef07-200) | OK | OK |  | [schema](#0198042a-f9c5-7626-be9f-996a2898ef07-200-schema) |
| [400](#0198042a-f9c5-7626-be9f-996a2898ef07-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-7626-be9f-996a2898ef07-400-schema) |
| [409](#0198042a-f9c5-7626-be9f-996a2898ef07-409) | Conflict | Conflict |  | [schema](#0198042a-f9c5-7626-be9f-996a2898ef07-409-schema) |
| [500](#0198042a-f9c5-7626-be9f-996a2898ef07-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-7626-be9f-996a2898ef07-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-7626-be9f-996a2898ef07-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-7626-be9f-996a2898ef07-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-7626-be9f-996a2898ef07-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-7626-be9f-996a2898ef07-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-7626-be9f-996a2898ef07-409"></span> 409 - Conflict
Status: Conflict

###### <span id="0198042a-f9c5-7626-be9f-996a2898ef07-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-7626-be9f-996a2898ef07-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-7626-be9f-996a2898ef07-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-762a-8033-649a1526901d"></span> Delete project (*0198042a-f9c5-762a-8033-649a1526901d*)

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
| [200](#0198042a-f9c5-762a-8033-649a1526901d-200) | OK | OK |  | [schema](#0198042a-f9c5-762a-8033-649a1526901d-200-schema) |
| [400](#0198042a-f9c5-762a-8033-649a1526901d-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-762a-8033-649a1526901d-400-schema) |
| [500](#0198042a-f9c5-762a-8033-649a1526901d-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-762a-8033-649a1526901d-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-762a-8033-649a1526901d-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-762a-8033-649a1526901d-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-762a-8033-649a1526901d-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-762a-8033-649a1526901d-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-762a-8033-649a1526901d-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-762a-8033-649a1526901d-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-76a7-a480-fbcb978b8501"></span> List projects (*0198042a-f9c5-76a7-a480-fbcb978b8501*)

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
| [200](#0198042a-f9c5-76a7-a480-fbcb978b8501-200) | OK | OK |  | [schema](#0198042a-f9c5-76a7-a480-fbcb978b8501-200-schema) |
| [400](#0198042a-f9c5-76a7-a480-fbcb978b8501-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-76a7-a480-fbcb978b8501-400-schema) |
| [500](#0198042a-f9c5-76a7-a480-fbcb978b8501-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-76a7-a480-fbcb978b8501-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-76a7-a480-fbcb978b8501-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-76a7-a480-fbcb978b8501-200-schema"></span> Schema
   
  

[ModelListProjectsResponse](#model-list-projects-response)

##### <span id="0198042a-f9c5-76a7-a480-fbcb978b8501-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-76a7-a480-fbcb978b8501-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76a7-a480-fbcb978b8501-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-76a7-a480-fbcb978b8501-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-76b2-b8b1-bc0223a0f18d"></span> Get resource (*0198042a-f9c5-76b2-b8b1-bc0223a0f18d*)

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
| [200](#0198042a-f9c5-76b2-b8b1-bc0223a0f18d-200) | OK | OK |  | [schema](#0198042a-f9c5-76b2-b8b1-bc0223a0f18d-200-schema) |
| [400](#0198042a-f9c5-76b2-b8b1-bc0223a0f18d-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-76b2-b8b1-bc0223a0f18d-400-schema) |
| [404](#0198042a-f9c5-76b2-b8b1-bc0223a0f18d-404) | Not Found | Not Found |  | [schema](#0198042a-f9c5-76b2-b8b1-bc0223a0f18d-404-schema) |
| [500](#0198042a-f9c5-76b2-b8b1-bc0223a0f18d-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-76b2-b8b1-bc0223a0f18d-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-76b2-b8b1-bc0223a0f18d-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-76b2-b8b1-bc0223a0f18d-200-schema"></span> Schema
   
  

[ModelResource](#model-resource)

##### <span id="0198042a-f9c5-76b2-b8b1-bc0223a0f18d-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-76b2-b8b1-bc0223a0f18d-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76b2-b8b1-bc0223a0f18d-404"></span> 404 - Not Found
Status: Not Found

###### <span id="0198042a-f9c5-76b2-b8b1-bc0223a0f18d-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76b2-b8b1-bc0223a0f18d-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-76b2-b8b1-bc0223a0f18d-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-76b6-bd55-f34dff7b0632"></span> List resources (*0198042a-f9c5-76b6-bd55-f34dff7b0632*)

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
| [200](#0198042a-f9c5-76b6-bd55-f34dff7b0632-200) | OK | OK |  | [schema](#0198042a-f9c5-76b6-bd55-f34dff7b0632-200-schema) |
| [400](#0198042a-f9c5-76b6-bd55-f34dff7b0632-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-76b6-bd55-f34dff7b0632-400-schema) |
| [500](#0198042a-f9c5-76b6-bd55-f34dff7b0632-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-76b6-bd55-f34dff7b0632-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-76b6-bd55-f34dff7b0632-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-76b6-bd55-f34dff7b0632-200-schema"></span> Schema
   
  

[ModelListResourcesResponse](#model-list-resources-response)

##### <span id="0198042a-f9c5-76b6-bd55-f34dff7b0632-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-76b6-bd55-f34dff7b0632-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76b6-bd55-f34dff7b0632-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-76b6-bd55-f34dff7b0632-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-76ba-bc87-6e9e32988407"></span> Match resources (*0198042a-f9c5-76ba-bc87-6e9e32988407*)

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
| [200](#0198042a-f9c5-76ba-bc87-6e9e32988407-200) | OK | OK |  | [schema](#0198042a-f9c5-76ba-bc87-6e9e32988407-200-schema) |
| [400](#0198042a-f9c5-76ba-bc87-6e9e32988407-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-76ba-bc87-6e9e32988407-400-schema) |
| [500](#0198042a-f9c5-76ba-bc87-6e9e32988407-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-76ba-bc87-6e9e32988407-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-76ba-bc87-6e9e32988407-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-76ba-bc87-6e9e32988407-200-schema"></span> Schema
   
  

[ModelListResourcesResponse](#model-list-resources-response)

##### <span id="0198042a-f9c5-76ba-bc87-6e9e32988407-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-76ba-bc87-6e9e32988407-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76ba-bc87-6e9e32988407-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-76ba-bc87-6e9e32988407-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-76be-ba9e-8186a69f48c4"></span> Check health (*0198042a-f9c5-76be-ba9e-8186a69f48c4*)

```
GET /health/status
```

Check service health status including database connectivity and system metrics

#### Produces
  * application/json

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#0198042a-f9c5-76be-ba9e-8186a69f48c4-200) | OK | OK |  | [schema](#0198042a-f9c5-76be-ba9e-8186a69f48c4-200-schema) |
| [500](#0198042a-f9c5-76be-ba9e-8186a69f48c4-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-76be-ba9e-8186a69f48c4-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-76be-ba9e-8186a69f48c4-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-76be-ba9e-8186a69f48c4-200-schema"></span> Schema
   
  

[ModelHealth](#model-health)

##### <span id="0198042a-f9c5-76be-ba9e-8186a69f48c4-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-76be-ba9e-8186a69f48c4-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-76c2-96f2-d16b0674bcd9"></span> Get policy (*0198042a-f9c5-76c2-96f2-d16b0674bcd9*)

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
| [200](#0198042a-f9c5-76c2-96f2-d16b0674bcd9-200) | OK | OK |  | [schema](#0198042a-f9c5-76c2-96f2-d16b0674bcd9-200-schema) |
| [400](#0198042a-f9c5-76c2-96f2-d16b0674bcd9-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-76c2-96f2-d16b0674bcd9-400-schema) |
| [404](#0198042a-f9c5-76c2-96f2-d16b0674bcd9-404) | Not Found | Not Found |  | [schema](#0198042a-f9c5-76c2-96f2-d16b0674bcd9-404-schema) |
| [500](#0198042a-f9c5-76c2-96f2-d16b0674bcd9-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-76c2-96f2-d16b0674bcd9-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-76c2-96f2-d16b0674bcd9-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-76c2-96f2-d16b0674bcd9-200-schema"></span> Schema
   
  

[ModelPolicy](#model-policy)

##### <span id="0198042a-f9c5-76c2-96f2-d16b0674bcd9-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-76c2-96f2-d16b0674bcd9-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76c2-96f2-d16b0674bcd9-404"></span> 404 - Not Found
Status: Not Found

###### <span id="0198042a-f9c5-76c2-96f2-d16b0674bcd9-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76c2-96f2-d16b0674bcd9-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-76c2-96f2-d16b0674bcd9-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-76c6-9a07-0c8948640ac2"></span> Create policy (*0198042a-f9c5-76c6-9a07-0c8948640ac2*)

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
| [201](#0198042a-f9c5-76c6-9a07-0c8948640ac2-201) | Created | Created |  | [schema](#0198042a-f9c5-76c6-9a07-0c8948640ac2-201-schema) |
| [400](#0198042a-f9c5-76c6-9a07-0c8948640ac2-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-76c6-9a07-0c8948640ac2-400-schema) |
| [409](#0198042a-f9c5-76c6-9a07-0c8948640ac2-409) | Conflict | Conflict |  | [schema](#0198042a-f9c5-76c6-9a07-0c8948640ac2-409-schema) |
| [500](#0198042a-f9c5-76c6-9a07-0c8948640ac2-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-76c6-9a07-0c8948640ac2-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-76c6-9a07-0c8948640ac2-201"></span> 201 - Created
Status: Created

###### <span id="0198042a-f9c5-76c6-9a07-0c8948640ac2-201-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76c6-9a07-0c8948640ac2-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-76c6-9a07-0c8948640ac2-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76c6-9a07-0c8948640ac2-409"></span> 409 - Conflict
Status: Conflict

###### <span id="0198042a-f9c5-76c6-9a07-0c8948640ac2-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76c6-9a07-0c8948640ac2-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-76c6-9a07-0c8948640ac2-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-76ca-b40d-b1de1d359c22"></span> Update policy (*0198042a-f9c5-76ca-b40d-b1de1d359c22*)

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
| [200](#0198042a-f9c5-76ca-b40d-b1de1d359c22-200) | OK | OK |  | [schema](#0198042a-f9c5-76ca-b40d-b1de1d359c22-200-schema) |
| [400](#0198042a-f9c5-76ca-b40d-b1de1d359c22-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-76ca-b40d-b1de1d359c22-400-schema) |
| [404](#0198042a-f9c5-76ca-b40d-b1de1d359c22-404) | Not Found | Not Found |  | [schema](#0198042a-f9c5-76ca-b40d-b1de1d359c22-404-schema) |
| [500](#0198042a-f9c5-76ca-b40d-b1de1d359c22-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-76ca-b40d-b1de1d359c22-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-76ca-b40d-b1de1d359c22-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-76ca-b40d-b1de1d359c22-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76ca-b40d-b1de1d359c22-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-76ca-b40d-b1de1d359c22-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76ca-b40d-b1de1d359c22-404"></span> 404 - Not Found
Status: Not Found

###### <span id="0198042a-f9c5-76ca-b40d-b1de1d359c22-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76ca-b40d-b1de1d359c22-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-76ca-b40d-b1de1d359c22-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-76ce-b208-2f58f7ccd177"></span> Delete policy (*0198042a-f9c5-76ce-b208-2f58f7ccd177*)

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
| [200](#0198042a-f9c5-76ce-b208-2f58f7ccd177-200) | OK | OK |  | [schema](#0198042a-f9c5-76ce-b208-2f58f7ccd177-200-schema) |
| [400](#0198042a-f9c5-76ce-b208-2f58f7ccd177-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-76ce-b208-2f58f7ccd177-400-schema) |
| [404](#0198042a-f9c5-76ce-b208-2f58f7ccd177-404) | Not Found | Not Found |  | [schema](#0198042a-f9c5-76ce-b208-2f58f7ccd177-404-schema) |
| [500](#0198042a-f9c5-76ce-b208-2f58f7ccd177-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-76ce-b208-2f58f7ccd177-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-76ce-b208-2f58f7ccd177-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-76ce-b208-2f58f7ccd177-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76ce-b208-2f58f7ccd177-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-76ce-b208-2f58f7ccd177-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76ce-b208-2f58f7ccd177-404"></span> 404 - Not Found
Status: Not Found

###### <span id="0198042a-f9c5-76ce-b208-2f58f7ccd177-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76ce-b208-2f58f7ccd177-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-76ce-b208-2f58f7ccd177-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-76d2-a491-9cc989c1d59c"></span> List policies (*0198042a-f9c5-76d2-a491-9cc989c1d59c*)

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
| [200](#0198042a-f9c5-76d2-a491-9cc989c1d59c-200) | OK | OK |  | [schema](#0198042a-f9c5-76d2-a491-9cc989c1d59c-200-schema) |
| [400](#0198042a-f9c5-76d2-a491-9cc989c1d59c-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-76d2-a491-9cc989c1d59c-400-schema) |
| [500](#0198042a-f9c5-76d2-a491-9cc989c1d59c-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-76d2-a491-9cc989c1d59c-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-76d2-a491-9cc989c1d59c-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-76d2-a491-9cc989c1d59c-200-schema"></span> Schema
   
  

[ModelListPoliciesResponse](#model-list-policies-response)

##### <span id="0198042a-f9c5-76d2-a491-9cc989c1d59c-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-76d2-a491-9cc989c1d59c-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76d2-a491-9cc989c1d59c-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-76d2-a491-9cc989c1d59c-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-76d6-b1f3-0bfb57a9197f"></span> Link roles to policy (*0198042a-f9c5-76d6-b1f3-0bfb57a9197f*)

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
| [200](#0198042a-f9c5-76d6-b1f3-0bfb57a9197f-200) | OK | OK |  | [schema](#0198042a-f9c5-76d6-b1f3-0bfb57a9197f-200-schema) |
| [400](#0198042a-f9c5-76d6-b1f3-0bfb57a9197f-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-76d6-b1f3-0bfb57a9197f-400-schema) |
| [404](#0198042a-f9c5-76d6-b1f3-0bfb57a9197f-404) | Not Found | Not Found |  | [schema](#0198042a-f9c5-76d6-b1f3-0bfb57a9197f-404-schema) |
| [500](#0198042a-f9c5-76d6-b1f3-0bfb57a9197f-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-76d6-b1f3-0bfb57a9197f-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-76d6-b1f3-0bfb57a9197f-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-76d6-b1f3-0bfb57a9197f-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76d6-b1f3-0bfb57a9197f-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-76d6-b1f3-0bfb57a9197f-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76d6-b1f3-0bfb57a9197f-404"></span> 404 - Not Found
Status: Not Found

###### <span id="0198042a-f9c5-76d6-b1f3-0bfb57a9197f-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76d6-b1f3-0bfb57a9197f-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-76d6-b1f3-0bfb57a9197f-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-76d9-8019-babd51a0c340"></span> Unlink roles from policy (*0198042a-f9c5-76d9-8019-babd51a0c340*)

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
| [200](#0198042a-f9c5-76d9-8019-babd51a0c340-200) | OK | OK |  | [schema](#0198042a-f9c5-76d9-8019-babd51a0c340-200-schema) |
| [400](#0198042a-f9c5-76d9-8019-babd51a0c340-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-76d9-8019-babd51a0c340-400-schema) |
| [404](#0198042a-f9c5-76d9-8019-babd51a0c340-404) | Not Found | Not Found |  | [schema](#0198042a-f9c5-76d9-8019-babd51a0c340-404-schema) |
| [500](#0198042a-f9c5-76d9-8019-babd51a0c340-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-76d9-8019-babd51a0c340-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-76d9-8019-babd51a0c340-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-76d9-8019-babd51a0c340-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76d9-8019-babd51a0c340-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-76d9-8019-babd51a0c340-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76d9-8019-babd51a0c340-404"></span> 404 - Not Found
Status: Not Found

###### <span id="0198042a-f9c5-76d9-8019-babd51a0c340-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76d9-8019-babd51a0c340-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-76d9-8019-babd51a0c340-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-76dd-8fa8-98df6be12d44"></span> List policies by role (*0198042a-f9c5-76dd-8fa8-98df6be12d44*)

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
| [200](#0198042a-f9c5-76dd-8fa8-98df6be12d44-200) | OK | OK |  | [schema](#0198042a-f9c5-76dd-8fa8-98df6be12d44-200-schema) |
| [400](#0198042a-f9c5-76dd-8fa8-98df6be12d44-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-76dd-8fa8-98df6be12d44-400-schema) |
| [500](#0198042a-f9c5-76dd-8fa8-98df6be12d44-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-76dd-8fa8-98df6be12d44-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-76dd-8fa8-98df6be12d44-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-76dd-8fa8-98df6be12d44-200-schema"></span> Schema
   
  

[ModelListPoliciesResponse](#model-list-policies-response)

##### <span id="0198042a-f9c5-76dd-8fa8-98df6be12d44-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-76dd-8fa8-98df6be12d44-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76dd-8fa8-98df6be12d44-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-76dd-8fa8-98df6be12d44-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-76e1-a650-772c826f079e"></span> Get role (*0198042a-f9c5-76e1-a650-772c826f079e*)

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
| [200](#0198042a-f9c5-76e1-a650-772c826f079e-200) | OK | OK |  | [schema](#0198042a-f9c5-76e1-a650-772c826f079e-200-schema) |
| [400](#0198042a-f9c5-76e1-a650-772c826f079e-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-76e1-a650-772c826f079e-400-schema) |
| [404](#0198042a-f9c5-76e1-a650-772c826f079e-404) | Not Found | Not Found |  | [schema](#0198042a-f9c5-76e1-a650-772c826f079e-404-schema) |
| [500](#0198042a-f9c5-76e1-a650-772c826f079e-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-76e1-a650-772c826f079e-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-76e1-a650-772c826f079e-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-76e1-a650-772c826f079e-200-schema"></span> Schema
   
  

[ModelRole](#model-role)

##### <span id="0198042a-f9c5-76e1-a650-772c826f079e-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-76e1-a650-772c826f079e-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76e1-a650-772c826f079e-404"></span> 404 - Not Found
Status: Not Found

###### <span id="0198042a-f9c5-76e1-a650-772c826f079e-404-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76e1-a650-772c826f079e-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-76e1-a650-772c826f079e-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-76e5-8fe5-b93a07311c47"></span> Create role (*0198042a-f9c5-76e5-8fe5-b93a07311c47*)

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
| [201](#0198042a-f9c5-76e5-8fe5-b93a07311c47-201) | Created | Created |  | [schema](#0198042a-f9c5-76e5-8fe5-b93a07311c47-201-schema) |
| [400](#0198042a-f9c5-76e5-8fe5-b93a07311c47-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-76e5-8fe5-b93a07311c47-400-schema) |
| [409](#0198042a-f9c5-76e5-8fe5-b93a07311c47-409) | Conflict | Conflict |  | [schema](#0198042a-f9c5-76e5-8fe5-b93a07311c47-409-schema) |
| [500](#0198042a-f9c5-76e5-8fe5-b93a07311c47-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-76e5-8fe5-b93a07311c47-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-76e5-8fe5-b93a07311c47-201"></span> 201 - Created
Status: Created

###### <span id="0198042a-f9c5-76e5-8fe5-b93a07311c47-201-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76e5-8fe5-b93a07311c47-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-76e5-8fe5-b93a07311c47-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76e5-8fe5-b93a07311c47-409"></span> 409 - Conflict
Status: Conflict

###### <span id="0198042a-f9c5-76e5-8fe5-b93a07311c47-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76e5-8fe5-b93a07311c47-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-76e5-8fe5-b93a07311c47-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-76e9-922d-2411530cd8f8"></span> Update role (*0198042a-f9c5-76e9-922d-2411530cd8f8*)

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
| [200](#0198042a-f9c5-76e9-922d-2411530cd8f8-200) | OK | OK |  | [schema](#0198042a-f9c5-76e9-922d-2411530cd8f8-200-schema) |
| [400](#0198042a-f9c5-76e9-922d-2411530cd8f8-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-76e9-922d-2411530cd8f8-400-schema) |
| [409](#0198042a-f9c5-76e9-922d-2411530cd8f8-409) | Conflict | Conflict |  | [schema](#0198042a-f9c5-76e9-922d-2411530cd8f8-409-schema) |
| [500](#0198042a-f9c5-76e9-922d-2411530cd8f8-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-76e9-922d-2411530cd8f8-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-76e9-922d-2411530cd8f8-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-76e9-922d-2411530cd8f8-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76e9-922d-2411530cd8f8-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-76e9-922d-2411530cd8f8-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76e9-922d-2411530cd8f8-409"></span> 409 - Conflict
Status: Conflict

###### <span id="0198042a-f9c5-76e9-922d-2411530cd8f8-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76e9-922d-2411530cd8f8-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-76e9-922d-2411530cd8f8-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-76ed-99a5-84923071fa6b"></span> Delete role (*0198042a-f9c5-76ed-99a5-84923071fa6b*)

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
| [200](#0198042a-f9c5-76ed-99a5-84923071fa6b-200) | OK | OK |  | [schema](#0198042a-f9c5-76ed-99a5-84923071fa6b-200-schema) |
| [400](#0198042a-f9c5-76ed-99a5-84923071fa6b-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-76ed-99a5-84923071fa6b-400-schema) |
| [500](#0198042a-f9c5-76ed-99a5-84923071fa6b-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-76ed-99a5-84923071fa6b-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-76ed-99a5-84923071fa6b-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-76ed-99a5-84923071fa6b-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76ed-99a5-84923071fa6b-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-76ed-99a5-84923071fa6b-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76ed-99a5-84923071fa6b-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-76ed-99a5-84923071fa6b-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-76f1-9cf8-37e45b647fc0"></span> List roles (*0198042a-f9c5-76f1-9cf8-37e45b647fc0*)

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
| [200](#0198042a-f9c5-76f1-9cf8-37e45b647fc0-200) | OK | OK |  | [schema](#0198042a-f9c5-76f1-9cf8-37e45b647fc0-200-schema) |
| [400](#0198042a-f9c5-76f1-9cf8-37e45b647fc0-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-76f1-9cf8-37e45b647fc0-400-schema) |
| [500](#0198042a-f9c5-76f1-9cf8-37e45b647fc0-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-76f1-9cf8-37e45b647fc0-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-76f1-9cf8-37e45b647fc0-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-76f1-9cf8-37e45b647fc0-200-schema"></span> Schema
   
  

[ModelListRolesResponse](#model-list-roles-response)

##### <span id="0198042a-f9c5-76f1-9cf8-37e45b647fc0-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-76f1-9cf8-37e45b647fc0-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76f1-9cf8-37e45b647fc0-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-76f1-9cf8-37e45b647fc0-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-76f5-8ff6-b4479bdaa6b6"></span> Link users to role (*0198042a-f9c5-76f5-8ff6-b4479bdaa6b6*)

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
| [200](#0198042a-f9c5-76f5-8ff6-b4479bdaa6b6-200) | OK | OK |  | [schema](#0198042a-f9c5-76f5-8ff6-b4479bdaa6b6-200-schema) |
| [400](#0198042a-f9c5-76f5-8ff6-b4479bdaa6b6-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-76f5-8ff6-b4479bdaa6b6-400-schema) |
| [409](#0198042a-f9c5-76f5-8ff6-b4479bdaa6b6-409) | Conflict | Conflict |  | [schema](#0198042a-f9c5-76f5-8ff6-b4479bdaa6b6-409-schema) |
| [500](#0198042a-f9c5-76f5-8ff6-b4479bdaa6b6-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-76f5-8ff6-b4479bdaa6b6-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-76f5-8ff6-b4479bdaa6b6-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-76f5-8ff6-b4479bdaa6b6-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76f5-8ff6-b4479bdaa6b6-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-76f5-8ff6-b4479bdaa6b6-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76f5-8ff6-b4479bdaa6b6-409"></span> 409 - Conflict
Status: Conflict

###### <span id="0198042a-f9c5-76f5-8ff6-b4479bdaa6b6-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76f5-8ff6-b4479bdaa6b6-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-76f5-8ff6-b4479bdaa6b6-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-76f9-9394-170db55f62f4"></span> Unlink users from role (*0198042a-f9c5-76f9-9394-170db55f62f4*)

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
| [200](#0198042a-f9c5-76f9-9394-170db55f62f4-200) | OK | OK |  | [schema](#0198042a-f9c5-76f9-9394-170db55f62f4-200-schema) |
| [400](#0198042a-f9c5-76f9-9394-170db55f62f4-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-76f9-9394-170db55f62f4-400-schema) |
| [409](#0198042a-f9c5-76f9-9394-170db55f62f4-409) | Conflict | Conflict |  | [schema](#0198042a-f9c5-76f9-9394-170db55f62f4-409-schema) |
| [500](#0198042a-f9c5-76f9-9394-170db55f62f4-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-76f9-9394-170db55f62f4-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-76f9-9394-170db55f62f4-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-76f9-9394-170db55f62f4-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76f9-9394-170db55f62f4-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-76f9-9394-170db55f62f4-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76f9-9394-170db55f62f4-409"></span> 409 - Conflict
Status: Conflict

###### <span id="0198042a-f9c5-76f9-9394-170db55f62f4-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76f9-9394-170db55f62f4-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-76f9-9394-170db55f62f4-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-76fd-8012-5c9a2957e289"></span> Link policies to role (*0198042a-f9c5-76fd-8012-5c9a2957e289*)

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
| [200](#0198042a-f9c5-76fd-8012-5c9a2957e289-200) | OK | OK |  | [schema](#0198042a-f9c5-76fd-8012-5c9a2957e289-200-schema) |
| [400](#0198042a-f9c5-76fd-8012-5c9a2957e289-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-76fd-8012-5c9a2957e289-400-schema) |
| [409](#0198042a-f9c5-76fd-8012-5c9a2957e289-409) | Conflict | Conflict |  | [schema](#0198042a-f9c5-76fd-8012-5c9a2957e289-409-schema) |
| [500](#0198042a-f9c5-76fd-8012-5c9a2957e289-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-76fd-8012-5c9a2957e289-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-76fd-8012-5c9a2957e289-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-76fd-8012-5c9a2957e289-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76fd-8012-5c9a2957e289-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-76fd-8012-5c9a2957e289-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76fd-8012-5c9a2957e289-409"></span> 409 - Conflict
Status: Conflict

###### <span id="0198042a-f9c5-76fd-8012-5c9a2957e289-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-76fd-8012-5c9a2957e289-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-76fd-8012-5c9a2957e289-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-7700-9e40-e64f7b8c947c"></span> Unlink policies from role (*0198042a-f9c5-7700-9e40-e64f7b8c947c*)

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
| [200](#0198042a-f9c5-7700-9e40-e64f7b8c947c-200) | OK | OK |  | [schema](#0198042a-f9c5-7700-9e40-e64f7b8c947c-200-schema) |
| [400](#0198042a-f9c5-7700-9e40-e64f7b8c947c-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-7700-9e40-e64f7b8c947c-400-schema) |
| [409](#0198042a-f9c5-7700-9e40-e64f7b8c947c-409) | Conflict | Conflict |  | [schema](#0198042a-f9c5-7700-9e40-e64f7b8c947c-409-schema) |
| [500](#0198042a-f9c5-7700-9e40-e64f7b8c947c-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-7700-9e40-e64f7b8c947c-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-7700-9e40-e64f7b8c947c-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-7700-9e40-e64f7b8c947c-200-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-7700-9e40-e64f7b8c947c-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-7700-9e40-e64f7b8c947c-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-7700-9e40-e64f7b8c947c-409"></span> 409 - Conflict
Status: Conflict

###### <span id="0198042a-f9c5-7700-9e40-e64f7b8c947c-409-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-7700-9e40-e64f7b8c947c-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-7700-9e40-e64f7b8c947c-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-7704-b73b-55e2ec093586"></span> List roles by user (*0198042a-f9c5-7704-b73b-55e2ec093586*)

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
| [200](#0198042a-f9c5-7704-b73b-55e2ec093586-200) | OK | OK |  | [schema](#0198042a-f9c5-7704-b73b-55e2ec093586-200-schema) |
| [400](#0198042a-f9c5-7704-b73b-55e2ec093586-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-7704-b73b-55e2ec093586-400-schema) |
| [500](#0198042a-f9c5-7704-b73b-55e2ec093586-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-7704-b73b-55e2ec093586-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-7704-b73b-55e2ec093586-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-7704-b73b-55e2ec093586-200-schema"></span> Schema
   
  

[ModelListRolesResponse](#model-list-roles-response)

##### <span id="0198042a-f9c5-7704-b73b-55e2ec093586-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-7704-b73b-55e2ec093586-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-7704-b73b-55e2ec093586-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-7704-b73b-55e2ec093586-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-7704-b73b-55e2ec093587"></span> List roles by policy (*0198042a-f9c5-7704-b73b-55e2ec093587*)

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
| [200](#0198042a-f9c5-7704-b73b-55e2ec093587-200) | OK | OK |  | [schema](#0198042a-f9c5-7704-b73b-55e2ec093587-200-schema) |
| [400](#0198042a-f9c5-7704-b73b-55e2ec093587-400) | Bad Request | Bad Request |  | [schema](#0198042a-f9c5-7704-b73b-55e2ec093587-400-schema) |
| [500](#0198042a-f9c5-7704-b73b-55e2ec093587-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-7704-b73b-55e2ec093587-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-7704-b73b-55e2ec093587-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-7704-b73b-55e2ec093587-200-schema"></span> Schema
   
  

[ModelListRolesResponse](#model-list-roles-response)

##### <span id="0198042a-f9c5-7704-b73b-55e2ec093587-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="0198042a-f9c5-7704-b73b-55e2ec093587-400-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

##### <span id="0198042a-f9c5-7704-b73b-55e2ec093587-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-7704-b73b-55e2ec093587-500-schema"></span> Schema
   
  

[ModelHTTPMessage](#model-http-message)

### <span id="0198042a-f9c5-7704-b73b-55e2ec093588"></span> Get version (*0198042a-f9c5-7704-b73b-55e2ec093588*)

```
GET /version
```

Retrieve the current version and build information of the service

#### Produces
  * application/json

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#0198042a-f9c5-7704-b73b-55e2ec093588-200) | OK | OK |  | [schema](#0198042a-f9c5-7704-b73b-55e2ec093588-200-schema) |
| [500](#0198042a-f9c5-7704-b73b-55e2ec093588-500) | Internal Server Error | Internal Server Error |  | [schema](#0198042a-f9c5-7704-b73b-55e2ec093588-500-schema) |

#### Responses


##### <span id="0198042a-f9c5-7704-b73b-55e2ec093588-200"></span> 200 - OK
Status: OK

###### <span id="0198042a-f9c5-7704-b73b-55e2ec093588-200-schema"></span> Schema
   
  

[ModelVersion](#model-version)

##### <span id="0198042a-f9c5-7704-b73b-55e2ec093588-500"></span> 500 - Internal Server Error
Status: Internal Server Error

###### <span id="0198042a-f9c5-7704-b73b-55e2ec093588-500-schema"></span> Schema
   
  

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
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01980434-b7ff-7a9e-b343-668d79691032` |
| name | string (formatted string)| `string` | ✓ | |  | `List Policies for project` |



### <span id="model-create-product-request"></span> model.CreateProductRequest


> CreateProductRequest represents the input for the CreateProduct method.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| description | string (formatted string)| `string` |  | |  | `This is a product` |
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01980434-b7ff-7ac1-b7b0-13de306cc1cb` |
| name | string (formatted string)| `string` |  | |  | `New product name` |



### <span id="model-create-project-request"></span> model.CreateProjectRequest


> CreateProjectRequest represents the inputs necessary to create a new project.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| description | string (formatted string)| `string` | ✓ | |  | `This is a new project` |
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01980434-b7ff-7aa6-a131-a7c3590a1ce1` |
| name | string (formatted string)| `string` | ✓ | |  | `New project name` |



### <span id="model-create-role-request"></span> model.CreateRoleRequest


> CreateRoleRequest represents the input for the CreateRole method.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| description | string (formatted string)| `string` | ✓ | |  | `This is a role` |
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01980434-b7ff-7aba-a3ef-1b38309c9a1f` |
| name | string (formatted string)| `string` | ✓ | |  | `New role name` |



### <span id="model-create-user-request"></span> model.CreateUserRequest


> Create user request.
  





**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| email | email (formatted string)| `strfmt.Email` | ✓ | |  | `my@email.com` |
| first_name | string (formatted string)| `string` | ✓ | |  | `John` |
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01980434-b7ff-7ab2-b903-524ba1d47616` |
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
| role_ids | []uuid (formatted string)| `[]strfmt.UUID` | ✓ | |  | `["01980434-b7ff-7a96-b0c8-dbabed881cf5"]` |



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
| user_id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01980434-b7ff-7a54-a71f-34868a34e51e` |



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
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01980434-b7ff-7a93-b5b4-ca4c73283131` |
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
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01980434-b7ff-7abe-a45d-7311bc7011f5` |
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
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01980434-b7ff-7aa2-bfc2-d862a423985c` |
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
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01980434-b7ff-7a8b-b8e9-144341357314` |
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
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01980434-b7ff-7aaa-a09c-d46077eff792` |
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
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01980434-b7ff-7ab6-8c97-3e2f8905173a` |
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
| role_ids | []uuid (formatted string)| `[]strfmt.UUID` | ✓ | |  | `["01980434-b7ff-7a96-b0c8-dbabed881cf5"]` |



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
| id | uuid (formatted string)| `strfmt.UUID` |  | |  | `01980434-b7ff-7aae-95c6-051c9895119c` |
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


