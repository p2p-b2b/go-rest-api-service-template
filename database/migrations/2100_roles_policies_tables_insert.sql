-- +goose Up
-- +goose StatementBegin

-----------------------------------------------------------------------------------------
-- table roles
-----------------------------------------------------------------------------------------
INSERT INTO roles (id, name, description, system, auto_assign) VALUES
-- Administrator
('019791d2-adef-74be-b82a-079b56877764', 'Administrator', 'Administrator role.',                         TRUE, FALSE),
-- AuthenticatedUser
('019791d2-adef-758d-b043-55ea5be663a0', 'AuthenticatedUser', 'Authenticated user role. Allow do login', TRUE, TRUE),
-- ProjectAdmin
('01980464-8a12-7b13-9404-495b6634614f', 'ProjectAdmin', 'Project administrator role. Allow manage projects.', TRUE, FALSE);

-----------------------------------------------------------------------------------------
-- table resources
-----------------------------------------------------------------------------------------
INSERT INTO resources (id, name, description, action, resource, system) VALUES
-- full access
('019791d2-adef-7595-be12-21850e11b3ce', 'Allow all', 'Allow all actions on all resources', '*', '*', TRUE),
-- all paths only GET
('019791d2-adef-759d-8e4a-75282f75ffa1', 'Read only', 'Allow all GET action on all resources', 'GET', '*', TRUE),
-- all paths only POST
('019791d2-adef-75a5-990f-370991c94c0b', 'Create only', 'Allow all POST action on all resources', 'POST', '*', TRUE),
-- all paths only DELETE
('019791d2-adef-75ad-9c19-70872109cf61', 'Delete only', 'Allow all DELETE action on all resources', 'DELETE', '*', TRUE),
-- all paths only PUT
('019791d2-adef-75b1-8ffa-88d79ae18c09', 'Update only', 'Allow all PUT action on all resources',  'PUT', '*', TRUE),
-- all paths only PATCH
('019791d2-adef-75b8-876a-4fd32d28057a', 'Partial update only', 'Allow all PATCH action on all resources', 'PATCH', '*', TRUE),
-- all paths only OPTIONS
('019791d2-adef-75c0-ac6c-10e282b95a56', 'Options only', 'Allow all OPTIONS action on all resources', 'OPTIONS', '*', TRUE),
-- all paths only HEAD
('019791d2-adef-75c8-9f6c-0e3023d6668c', 'Head only', 'Allow all HEAD action on all resources', 'HEAD', '*', TRUE),
-- all paths only TRACE
('019791d2-adef-75cc-a2da-a7d60b6bf567', 'Trace only', 'Allow all TRACE action on all resources', 'TRACE', '*', TRUE),
-- automatic generate with the program apiendpoints
('0198042a-f9c5-7547-a6e7-567af5db26cd', 'Login user'                           , 'Authenticate user credentials and return JWT access and refresh tokens'      , 'POST'  , '/auth/login'                                                   , TRUE    ),
('0198042a-f9c5-75d4-afa6-fe658744c80f', 'Logout user'                          , 'Logout user and invalidate session tokens'                                   , 'DELETE', '/auth/logout'                                                  , TRUE    ),
('0198042a-f9c5-75d8-aa7b-37524ea4f124', 'Refresh access token'                 , 'Generate new access token using valid refresh token'                         , 'POST'  , '/auth/refresh'                                                 , TRUE    ),
('0198042a-f9c5-75c8-9231-ad5fc9e7b32e', 'Register user'                        , 'Create a new user account and send email verification'                       , 'POST'  , '/auth/register'                                                , TRUE    ),
('0198042a-f9c5-75d0-8c20-fea31b65587f', 'Resend verification'                  , 'Resend account verification email to user'                                   , 'POST'  , '/auth/verify'                                                  , TRUE    ),
('0198042a-f9c5-75cc-9dd2-e3ff9f6c1e3a', 'Verify user'                          , 'Verify user account using JWT verification token'                            , 'GET'   , '/auth/verify/{jwt}'                                            , TRUE    ),
('0198042a-f9c5-76c6-9a07-0c8948640ac2', 'Create policy'                        , 'Create a new policy with specified permissions'                              , 'POST'  , '/policies'                                                     , TRUE    ),
('0198042a-f9c5-76d2-a491-9cc989c1d59c', 'List policies'                        , 'Retrieve paginated list of all policies in the system'                       , 'GET'   , '/policies'                                                     , TRUE    ),
('0198042a-f9c5-76ca-b40d-b1de1d359c22', 'Update policy'                        , 'Modify an existing policy by its ID'                                         , 'PUT'   , '/policies/{policy_id}'                                         , TRUE    ),
('0198042a-f9c5-76ce-b208-2f58f7ccd177', 'Delete policy'                        , 'Remove a policy permanently from the system'                                 , 'DELETE', '/policies/{policy_id}'                                         , TRUE    ),
('0198042a-f9c5-76c2-96f2-d16b0674bcd9', 'Get policy'                           , 'Retrieve a specific policy by its unique identifier'                         , 'GET'   , '/policies/{policy_id}'                                         , TRUE    ),
('0198042a-f9c5-7704-b73b-55e2ec093587', 'List roles by policy'                 , 'Retrieve paginated list of roles associated with a specific policy'          , 'GET'   , '/policies/{policy_id}/roles'                                   , TRUE    ),
('0198042a-f9c5-76d9-8019-babd51a0c340', 'Unlink roles from policy'             , 'Remove role associations from a specific policy'                             , 'DELETE', '/policies/{policy_id}/roles'                                   , TRUE    ),
('0198042a-f9c5-76d6-b1f3-0bfb57a9197f', 'Link roles to policy'                 , 'Associate multiple roles with a specific policy for authorization'           , 'POST'  , '/policies/{policy_id}/roles'                                   , TRUE    ),
('0198042a-f9c5-7612-a055-58177eca0772', 'List products'                        , 'Retrieve paginated list of all products in the system'                       , 'GET'   , '/products'                                                     , TRUE    ),
('0198042a-f9c5-7622-9142-88fbaa727659', 'Create project'                       , 'Create a new project with specified configuration'                           , 'POST'  , '/projects'                                                     , TRUE    ),
('0198042a-f9c5-76a7-a480-fbcb978b8501', 'List projects'                        , 'Retrieve paginated list of all projects in the system'                       , 'GET'   , '/projects'                                                     , TRUE    ),
('0198042a-f9c5-7626-be9f-996a2898ef07', 'Update project'                       , 'Modify an existing project by its ID'                                        , 'PUT'   , '/projects/{project_id}'                                        , TRUE    ),
('0198042a-f9c5-761e-b1c2-66a3f8ab30d6', 'Get project'                          , 'Retrieve a specific project by its unique identifier'                        , 'GET'   , '/projects/{project_id}'                                        , TRUE    ),
('0198042a-f9c5-762a-8033-649a1526901d', 'Delete project'                       , 'Remove a project permanently from the system'                                , 'DELETE', '/projects/{project_id}'                                        , TRUE    ),
('0198042a-f9c5-7606-8aab-1c2db5b81a89', 'Create product'                       , 'Create a new product with specified configuration'                           , 'POST'  , '/projects/{project_id}/products'                               , TRUE    ),
('0198042a-f9c5-760e-9d2f-94cce8243e5a', 'List products by project'             , 'Retrieve paginated list of products for a specific project'                  , 'GET'   , '/projects/{project_id}/products'                               , TRUE    ),
('0198042a-f9c5-760a-99c8-1f68d597d300', 'Delete product'                       , 'Remove a product permanently from the system'                                , 'DELETE', '/projects/{project_id}/products/{product_id}'                  , TRUE    ),
('0198042a-f9c5-7603-99b1-7c20ee58542b', 'Get product'                          , 'Retrieve a specific product by its unique identifier'                        , 'GET'   , '/projects/{project_id}/products/{product_id}'                  , TRUE    ),
('0198042a-f9c5-7607-b75a-532912a6f35d', 'Update product'                       , 'Modify an existing product by its ID'                                        , 'PUT'   , '/projects/{project_id}/products/{product_id}'                  , TRUE    ),
('0198042a-f9c5-761a-bd02-da039b52bea2', 'Unlink product from payment processor', 'Remove the association between a product and a payment processor'            , 'DELETE', '/projects/{project_id}/products/{product_id}/payment_processor', TRUE    ),
('0198042a-f9c5-7616-8c3b-e4f19d83a033', 'Link product to payment processor'    , 'Associate a product with a payment processor to enable billing and invoicing', 'POST'  , '/projects/{project_id}/products/{product_id}/payment_processor', TRUE    ),
('0198042a-f9c5-76b6-bd55-f34dff7b0632', 'List resources'                       , 'Retrieve paginated list of all resources in the system'                      , 'GET'   , '/resources'                                                    , TRUE    ),
('0198042a-f9c5-76ba-bc87-6e9e32988407', 'Match resources'                      , 'Find resources that match specific action and resource policy patterns'      , 'GET'   , '/resources/matches'                                            , TRUE    ),
('0198042a-f9c5-76b2-b8b1-bc0223a0f18d', 'Get resource'                         , 'Retrieve a specific resource by its identifier'                              , 'GET'   , '/resources/{resource_id}'                                      , TRUE    ),
('0198042a-f9c5-76f1-9cf8-37e45b647fc0', 'List roles'                           , 'Retrieve paginated list of all roles in the system'                          , 'GET'   , '/roles'                                                        , TRUE    ),
('0198042a-f9c5-76e5-8fe5-b93a07311c47', 'Create role'                          , 'Create a new role with specified permissions and access levels'              , 'POST'  , '/roles'                                                        , TRUE    ),
('0198042a-f9c5-76e9-922d-2411530cd8f8', 'Update role'                          , 'Modify an existing role by its ID'                                           , 'PUT'   , '/roles/{role_id}'                                              , TRUE    ),
('0198042a-f9c5-76e1-a650-772c826f079e', 'Get role'                             , 'Retrieve a specific role by its unique identifier'                           , 'GET'   , '/roles/{role_id}'                                              , TRUE    ),
('0198042a-f9c5-76ed-99a5-84923071fa6b', 'Delete role'                          , 'Remove a role permanently from the system'                                   , 'DELETE', '/roles/{role_id}'                                              , TRUE    ),
('0198042a-f9c5-76fd-8012-5c9a2957e289', 'Link policies to role'                , 'Associate multiple policies with a specific role for authorization'          , 'POST'  , '/roles/{role_id}/policies'                                     , TRUE    ),
('0198042a-f9c5-76dd-8fa8-98df6be12d44', 'List policies by role'                , 'Retrieve paginated list of policies associated with a specific role'         , 'GET'   , '/roles/{role_id}/policies'                                     , TRUE    ),
('0198042a-f9c5-7700-9e40-e64f7b8c947c', 'Unlink policies from role'            , 'Remove policy associations from a specific role'                             , 'DELETE', '/roles/{role_id}/policies'                                     , TRUE    ),
('0198042a-f9c5-76f9-9394-170db55f62f4', 'Unlink users from role'               , 'Remove user associations from a specific role'                               , 'DELETE', '/roles/{role_id}/users'                                        , TRUE    ),
('0198042a-f9c5-76f5-8ff6-b4479bdaa6b6', 'Link users to role'                   , 'Associate multiple users with a specific role for authorization'             , 'POST'  , '/roles/{role_id}/users'                                        , TRUE    ),
('0198042a-f9c5-75ff-bbfc-224bf4342886', 'List users by role'                   , 'Retrieve paginated list of users associated with a specific role'            , 'GET'   , '/roles/{role_id}/users'                                        , TRUE    ),
('0198042a-f9c5-75ef-8ea1-29ecbbe01a2e', 'List users'                           , 'Retrieve paginated list of all users in the system'                          , 'GET'   , '/users'                                                        , TRUE    ),
('0198042a-f9c5-75e3-acf6-6901bb33ae65', 'Create user'                          , 'Create a new user account with specified configuration'                      , 'POST'  , '/users'                                                        , TRUE    ),
('0198042a-f9c5-75df-b843-b92a4d5c590e', 'Get user'                             , 'Retrieve a specific user account by its unique identifier'                   , 'GET'   , '/users/{user_id}'                                              , TRUE    ),
('0198042a-f9c5-75e7-8cb9-231bee55c64e', 'Update user'                          , 'Modify an existing user account by its ID'                                   , 'PUT'   , '/users/{user_id}'                                              , TRUE    ),
('0198042a-f9c5-75eb-b683-6c1847af7108', 'Delete user'                          , 'Remove a user account permanently from the system'                           , 'DELETE', '/users/{user_id}'                                              , TRUE    ),
('0198042a-f9c5-75fb-b324-ec962beb2277', 'Get user authorization'               , 'Retrieve user authorization permissions and roles for access control'        , 'GET'   , '/users/{user_id}/authz'                                        , TRUE    ),
('0198042a-f9c5-7704-b73b-55e2ec093586', 'List roles by user'                   , 'Retrieve paginated list of roles assigned to a specific user'                , 'GET'   , '/users/{user_id}/roles'                                        , TRUE    ),
('0198042a-f9c5-75f3-985f-d30e67bb3688', 'Link roles to user'                   , 'Associate multiple roles with a user within a specific project'              , 'POST'  , '/users/{user_id}/roles'                                        , TRUE    ),
('0198042a-f9c5-75f7-b802-343518ee3788', 'Unlink roles from user'               , 'Remove role associations from a user within a specific project'              , 'DELETE', '/users/{user_id}/roles'                                        , TRUE    );

-----------------------------------------------------------------------------------------
-- table policies
-----------------------------------------------------------------------------------------
INSERT INTO policies (id, resources_id, name, description, allowed_action, allowed_resource, system) VALUES
-- full access
('01979221-694f-7ba0-8930-8e7b9e147c2e', '019791d2-adef-7595-be12-21850e11b3ce', 'Full Access', 'Allow all actions on all resources', '*', '*', TRUE),
-- Allow logout
('01979221-694f-7b7d-ab32-5c301e0e1745', '0198042a-f9c5-75d4-afa6-fe658744c80f', 'Allow logout', 'Allow make logout', 'DELETE', '/auth/logout', TRUE),
-- Allow refresh token
('01979221-694f-7b69-a3a8-c8fce0f43afc', '0198042a-f9c5-75d8-aa7b-37524ea4f124', 'Allow refresh token', 'Allow refresh token', 'POST', '/auth/refresh', TRUE),
-- Allow create project
('01980464-8a12-7b17-af6f-bc536bc6d71c', '0198042a-f9c5-7622-9142-88fbaa727659', 'Allow create project', 'Allow create project', 'POST', '/projects', TRUE),
-- Allow list projects
('019804ef-e875-76be-bddf-b50533fd7f67', '0198042a-f9c5-76a7-a480-fbcb978b8501', 'Allow list projects', 'Allow list projects', 'GET', '/projects', TRUE);

-----------------------------------------------------------------------------------------
-- table roles_policies
-----------------------------------------------------------------------------------------
INSERT INTO roles_policies (roles_id, policies_id) VALUES
-- Owner
('019791d2-adef-74be-b82a-079b56877764', '01979221-694f-7ba0-8930-8e7b9e147c2e'),
-- AuthenticatedUser
('019791d2-adef-758d-b043-55ea5be663a0', '01979221-694f-7b7d-ab32-5c301e0e1745'),
('019791d2-adef-758d-b043-55ea5be663a0', '01979221-694f-7b69-a3a8-c8fce0f43afc'),
-- ProjectAdmin
('01980464-8a12-7b13-9404-495b6634614f', '01980464-8a12-7b17-af6f-bc536bc6d71c'),
('01980464-8a12-7b13-9404-495b6634614f', '019804ef-e875-76be-bddf-b50533fd7f67');

-----------------------------------------------------------------------------------------
-- table users_roles
-----------------------------------------------------------------------------------------
INSERT INTO users_roles (users_id, roles_id) VALUES
-- Administrator
('019791d2-adef-76d2-a865-5b19e5073e60', '019791d2-adef-74be-b82a-079b56877764'),
-- User
('01980464-8a12-7b1b-8e3b-8d065c7a08c2', '019791d2-adef-758d-b043-55ea5be663a0'),
('01980464-8a12-7b1b-8e3b-8d065c7a08c2', '01980464-8a12-7b13-9404-495b6634614f');

-- +goose StatementEnd
--
-- +goose Down
-- +goose StatementBegin

-- delete all roles_policies
DELETE FROM roles_policies;

-- delete all policies
DELETE FROM policies;

-- delete all roles
DELETE FROM roles;

-- delete all resources
DELETE FROM resources;

-- +goose StatementEnd
