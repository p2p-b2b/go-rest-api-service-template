-- +goose Up
-- +goose StatementBegin

-----------------------------------------------------------------------------------------
-- table roles
-----------------------------------------------------------------------------------------
INSERT INTO roles (id, name, description, system, auto_assign) VALUES
-- Administrator
('019791d2-adef-74be-b82a-079b56877764', 'Administrator', 'Administrator role.',                         TRUE, FALSE),
-- AuthenticatedUser
('019791d2-adef-758d-b043-55ea5be663a0', 'AuthenticatedUser', 'Authenticated user role. Allow do login', TRUE, TRUE);

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
('019791cc-06c7-7d8c-81e4-914dd89098e8', 'Login user'                           , 'Authenticate user credentials and return JWT access and refresh tokens'      , 'POST'  , '/auth/login'                                                     , TRUE    ),
('019791cc-06c7-7e4c-961c-b1bf5f40d633', 'Logout user'                          , 'Logout user and invalidate session tokens'                                   , 'DELETE', '/auth/logout'                                                    , TRUE    ),
('019791cc-06c7-7e50-a5b6-bd1c82d5c031', 'Refresh access token'                 , 'Generate new access token using valid refresh token'                         , 'POST'  , '/auth/refresh'                                                   , TRUE    ),
('019791cc-06c7-7e38-ac82-6919df575ff7', 'Register user'                        , 'Create a new user account and send email verification'                       , 'POST'  , '/auth/register'                                                  , TRUE    ),
('019791cc-06c7-7e48-b046-99f072088c50', 'Resend verification'                  , 'Resend account verification email to user'                                   , 'POST'  , '/auth/verify'                                                    , TRUE    ),
('019791cc-06c7-7e40-bcc3-149cd13bee44', 'Verify user'                          , 'Verify user account using JWT verification token'                            , 'GET'   , '/auth/verify/{jwt}'                                              , TRUE    ),
('0197adda-c4f1-7816-9feb-d2541affe41b', 'List features'                        , 'Retrieve paginated list of all features in the system'                       , 'GET'   , '/features'                                                       , TRUE    ),
('0197adda-c4f1-7806-9b2e-6a2a75ba6c84', 'Create feature'                       , 'Create a new feature with specified configuration'                           , 'POST'  , '/features'                                                       , TRUE    ),
('0197adda-c4f1-7812-98e9-2e3f9a77f05f', 'Delete feature'                       , 'Remove a feature permanently from the system'                                , 'DELETE', '/features/{feature_id}'                                          , TRUE    ),
('0197adda-c4f1-780e-b9f0-7d448ae83a23', 'Update feature'                       , 'Modify an existing feature by its ID'                                        , 'PUT'   , '/features/{feature_id}'                                          , TRUE    ),
('0197adda-c4f1-77d0-986f-21ac3fdbcafe', 'Get feature'                          , 'Retrieve a specific feature by its unique identifier'                        , 'GET'   , '/features/{feature_id}'                                          , TRUE    ),
('0197a873-45a2-72ee-8efb-9e04e917e0b1', 'List licenses'                        , 'Retrieve paginated list of all licenses in the system'                       , 'GET'   , '/licenses'                                                       , TRUE    ),
('01979837-c712-7802-8347-8b39255f0c48', 'List payment processor types'         , 'Retrieve paginated list of all payment processor types in the system'        , 'GET'   , '/payment_processor_types'                                        , TRUE    ),
('01979837-c712-7806-88a8-97fddb9e1f77', 'Get payment processor type'           , 'Retrieve a specific payment processor type by its identifier'                , 'GET'   , '/payment_processor_types/{payment_processor_type_id}'            , TRUE    ),
('0197984d-8f8d-7cb2-9834-1d3d291ba21e', 'List payment processors'              , 'Retrieve paginated list of all payment processors in the system'             , 'GET'   , '/payment_processors'                                             , TRUE    ),
('019791cc-06c7-7e63-8ec9-de5b38235dbf', 'Create policy'                        , 'Create a new policy with specified permissions'                              , 'POST'  , '/policies'                                                       , TRUE    ),
('019791cc-06c7-7e73-96aa-7e0383caae0d', 'List policies'                        , 'Retrieve paginated list of all policies in the system'                       , 'GET'   , '/policies'                                                       , TRUE    ),
('019791cc-06c7-7e67-9e1b-49a34edfe07c', 'Update policy'                        , 'Modify an existing policy by its ID'                                         , 'PUT'   , '/policies/{policy_id}'                                           , TRUE    ),
('019791cc-06c7-7e5b-b363-1ef381f1e832', 'Get policy'                           , 'Retrieve a specific policy by its unique identifier'                         , 'GET'   , '/policies/{policy_id}'                                           , TRUE    ),
('019791cc-06c7-7e6b-b308-a2b2cbc2aaa1', 'Delete policy'                        , 'Remove a policy permanently from the system'                                 , 'DELETE', '/policies/{policy_id}'                                           , TRUE    ),
('019791cc-06c7-7ecd-93da-d21bac8fd613', 'List roles by policy'                 , 'Retrieve paginated list of roles associated with a specific policy'          , 'GET'   , '/policies/{policy_id}/roles'                                     , TRUE    ),
('019791cc-06c7-7e7b-bdfc-381b015c44e7', 'Unlink roles from policy'             , 'Remove role associations from a specific policy'                             , 'DELETE', '/policies/{policy_id}/roles'                                     , TRUE    ),
('019791cc-06c7-7e77-a2c3-4ed693a2bcdd', 'Link roles to policy'                 , 'Associate multiple roles with a specific policy for authorization'           , 'POST'  , '/policies/{policy_id}/roles'                                     , TRUE    ),
('01979db7-f53f-73b5-a499-d6390831c94c', 'List products'                        , 'Retrieve paginated list of all products in the system'                       , 'GET'   , '/products'                                                       , TRUE    ),
('0197a873-45a2-72da-939a-d2c00f22cf5c', 'Create license'                       , 'Create a new license with specified configuration'                           , 'POST'  , '/products/{product_id}/licenses'                                 , TRUE    ),
('0197a873-45a2-72ea-b0d0-e797febd2d82', 'List licenses by product'             , 'Retrieve paginated list of licenses for a specific product'                  , 'GET'   , '/products/{product_id}/licenses'                                 , TRUE    ),
('0197a873-45a2-7249-9bf7-ecf630638056', 'Get license'                          , 'Retrieve a specific license by its unique identifier'                        , 'GET'   , '/products/{product_id}/licenses/{license_id}'                    , TRUE    ),
('0197a873-45a2-72de-a4d8-c0348e46b752', 'Update license'                       , 'Modify an existing license by its ID'                                        , 'PUT'   , '/products/{product_id}/licenses/{license_id}'                    , TRUE    ),
('0197a873-45a2-72e6-9193-7550fde8d4e4', 'Delete license'                       , 'Remove a license permanently from the system'                                , 'DELETE', '/products/{product_id}/licenses/{license_id}'                    , TRUE    ),
('019797e6-138a-7d00-98db-740f21794f11', 'Create project'                       , 'Create a new project with specified configuration'                           , 'POST'  , '/projects'                                                       , TRUE    ),
('019797e6-138a-7cf1-b7bb-fa9c5e168c49', 'List projects'                        , 'Retrieve paginated list of all projects in the system'                       , 'GET'   , '/projects'                                                       , TRUE    ),
('019797e6-138a-7d04-8db3-1d4755b25db3', 'Get project'                          , 'Retrieve a specific project by its unique identifier'                        , 'GET'   , '/projects/{project_id}'                                          , TRUE    ),
('019797e6-138a-7cf8-9887-e4c44ad0ae19', 'Update project'                       , 'Modify an existing project by its ID'                                        , 'PUT'   , '/projects/{project_id}'                                          , TRUE    ),
('019797e6-138a-7cf4-8694-e4611baded39', 'Delete project'                       , 'Remove a project permanently from the system'                                , 'DELETE', '/projects/{project_id}'                                          , TRUE    ),
('01979cbe-daab-79d3-b81c-dd1f4d0296cf', 'List payment processors by project'   , 'Retrieve paginated list of payment processors for a specific project'        , 'GET'   , '/projects/{project_id}/payment_processors'                       , TRUE    ),
('0197984d-8f8d-7cae-a107-4d0288eebfdf', 'Create payment processor'             , 'Create a new payment processor with specified configuration'                 , 'POST'  , '/projects/{project_id}/payment_processors'                       , TRUE    ),
('0197984d-8f8d-7caa-89a9-41bc53dca6bf', 'Update payment processor'             , 'Modify an existing payment processor by its ID'                              , 'PUT'   , '/projects/{project_id}/payment_processors/{payment_processor_id}', TRUE    ),
('0197984d-8f8d-7cb5-9940-b00cd6d1babb', 'Get payment processor'                , 'Retrieve a specific payment processor by its identifier'                     , 'GET'   , '/projects/{project_id}/payment_processors/{payment_processor_id}', TRUE    ),
('0197984d-8f8d-7ca6-9684-7c3232eb5e95', 'Delete payment processor'             , 'Remove a payment processor permanently from the system'                      , 'DELETE', '/projects/{project_id}/payment_processors/{payment_processor_id}', TRUE    ),
('01979db7-f53f-73a5-b916-297c6db5b714', 'Create product'                       , 'Create a new product with specified configuration'                           , 'POST'  , '/projects/{project_id}/products'                                 , TRUE    ),
('01979db7-f53f-73b1-993f-15f77e72c8cc', 'List products by project'             , 'Retrieve paginated list of products for a specific project'                  , 'GET'   , '/projects/{project_id}/products'                                 , TRUE    ),
('01979db7-f53f-73ad-a84e-49bbdfe9e5c9', 'Delete product'                       , 'Remove a product permanently from the system'                                , 'DELETE', '/projects/{project_id}/products/{product_id}'                    , TRUE    ),
('01979db7-f53f-73a1-aab2-74802b79be51', 'Get product'                          , 'Retrieve a specific product by its unique identifier'                        , 'GET'   , '/projects/{project_id}/products/{product_id}'                    , TRUE    ),
('01979db7-f53f-73a9-bd58-e1cd5d7df436', 'Update product'                       , 'Modify an existing product by its ID'                                        , 'PUT'   , '/projects/{project_id}/products/{product_id}'                    , TRUE    ),
('01979db7-f53f-73b9-818f-cdd1848f15d0', 'Unlink product from payment processor', 'Remove the association between a product and a payment processor'            , 'DELETE', '/projects/{project_id}/products/{product_id}/payment_processor'  , TRUE    ),
('01979db7-f53f-73bd-b6c0-4541a48549c2', 'Link product to payment processor'    , 'Associate a product with a payment processor to enable billing and invoicing', 'POST'  , '/projects/{project_id}/products/{product_id}/payment_processor'  , TRUE    ),
('019791cc-06c7-7e8e-8d7e-cd3f9296e0fd', 'List resources'                       , 'Retrieve paginated list of all resources in the system'                      , 'GET'   , '/resources'                                                      , TRUE    ),
('019791cc-06c7-7e92-9152-cb35902f79c4', 'Match resources'                      , 'Find resources that match specific action and resource policy patterns'      , 'GET'   , '/resources/matches'                                              , TRUE    ),
('019791cc-06c7-7e86-ad42-b777bfcc9e40', 'Get resource'                         , 'Retrieve a specific resource by its identifier'                              , 'GET'   , '/resources/{resource_id}'                                        , TRUE    ),
('019791cc-06c7-7e9e-87bf-dcedfa5aefa7', 'Create role'                          , 'Create a new role with specified permissions and access levels'              , 'POST'  , '/roles'                                                          , TRUE    ),
('019791cc-06c7-7ead-968a-2a457714a7ee', 'List roles'                           , 'Retrieve paginated list of all roles in the system'                          , 'GET'   , '/roles'                                                          , TRUE    ),
('019791cc-06c7-7ea2-8f0e-7b9f7cbc203a', 'Update role'                          , 'Modify an existing role by its ID'                                           , 'PUT'   , '/roles/{role_id}'                                                , TRUE    ),
('019791cc-06c7-7ea6-9423-184c13540c26', 'Delete role'                          , 'Remove a role permanently from the system'                                   , 'DELETE', '/roles/{role_id}'                                                , TRUE    ),
('019791cc-06c7-7e96-a284-ad37f86475bd', 'Get role'                             , 'Retrieve a specific role by its unique identifier'                           , 'GET'   , '/roles/{role_id}'                                                , TRUE    ),
('019791cc-06c7-7e82-967d-e13c399f5018', 'List policies by role'                , 'Retrieve paginated list of policies associated with a specific role'         , 'GET'   , '/roles/{role_id}/policies'                                       , TRUE    ),
('019791cc-06c7-7ebd-b750-8c93b165d503', 'Link policies to role'                , 'Associate multiple policies with a specific role for authorization'          , 'POST'  , '/roles/{role_id}/policies'                                       , TRUE    ),
('019791cc-06c7-7ec1-aca1-291132927db6', 'Unlink policies from role'            , 'Remove policy associations from a specific role'                             , 'DELETE', '/roles/{role_id}/policies'                                       , TRUE    ),
('019791cc-06c7-7eb1-b10c-4fbcd5943885', 'Link users to role'                   , 'Associate multiple users with a specific role for authorization'             , 'POST'  , '/roles/{role_id}/users'                                          , TRUE    ),
('019791cc-06c7-7eb5-9b74-ba394be221b4', 'Unlink users from role'               , 'Remove user associations from a specific role'                               , 'DELETE', '/roles/{role_id}/users'                                          , TRUE    ),
('019791cc-06c7-7efb-99c2-b25af11e600c', 'List users by role'                   , 'Retrieve paginated list of users associated with a specific role'            , 'GET'   , '/roles/{role_id}/users'                                          , TRUE    ),
('019791cc-06c7-7ed4-8f8b-2297e4565de3', 'Create user'                          , 'Create a new user account with specified configuration'                      , 'POST'  , '/users'                                                          , TRUE    ),
('019791cc-06c7-7ee4-8f2b-ea43720d520b', 'List users'                           , 'Retrieve paginated list of all users in the system'                          , 'GET'   , '/users'                                                          , TRUE    ),
('019791cc-06c7-7ee0-85b7-45450ad476eb', 'Delete user'                          , 'Remove a user account permanently from the system'                           , 'DELETE', '/users/{user_id}'                                                , TRUE    ),
('019791cc-06c7-7edc-94d6-843e3f99e96f', 'Update user'                          , 'Modify an existing user account by its ID'                                   , 'PUT'   , '/users/{user_id}'                                                , TRUE    ),
('019791cc-06c7-7ed0-9140-556f721c5749', 'Get user'                             , 'Retrieve a specific user account by its unique identifier'                   , 'GET'   , '/users/{user_id}'                                                , TRUE    ),
('019791cc-06c7-7ef4-afa5-81125e9dcde9', 'Get user authorization'               , 'Retrieve user authorization permissions and roles for access control'        , 'GET'   , '/users/{user_id}/authz'                                          , TRUE    ),
('019791cc-06c7-7ef0-9394-d4ac3f52e94c', 'Unlink roles from user'               , 'Remove role associations from a user within a specific project'              , 'DELETE', '/users/{user_id}/roles'                                          , TRUE    ),
('019791cc-06c7-7eec-83f3-bcaed0c4d46f', 'Link roles to user'                   , 'Associate multiple roles with a user within a specific project'              , 'POST'  , '/users/{user_id}/roles'                                          , TRUE    ),
('019791cc-06c7-7ec5-87fe-096d6d2760a9', 'List roles by user'                   , 'Retrieve paginated list of roles assigned to a specific user'                , 'GET'   , '/users/{user_id}/roles'                                          , TRUE    );

-----------------------------------------------------------------------------------------
-- table policies
-----------------------------------------------------------------------------------------
INSERT INTO policies (id, resources_id, name, description, allowed_action, allowed_resource, system) VALUES
-- full access
('01979221-694f-7ba0-8930-8e7b9e147c2e', '019791d2-adef-7595-be12-21850e11b3ce', 'Full Access', 'Allow all actions on all resources', '*', '*', TRUE),
-- Allow logout
('01979221-694f-7b7d-ab32-5c301e0e1745', '019791cc-06c7-7e4c-961c-b1bf5f40d633', 'Allow logout', 'Allow make logout', 'DELETE', '/auth/logout', TRUE),
-- Allow refresh token
('01979221-694f-7b69-a3a8-c8fce0f43afc', '019791cc-06c7-7e50-a5b6-bd1c82d5c031', 'Allow refresh token', 'Allow refresh token', 'POST', '/auth/refresh', TRUE);

-----------------------------------------------------------------------------------------
-- table roles_policies
-----------------------------------------------------------------------------------------
INSERT INTO roles_policies (roles_id, policies_id) VALUES
-- Owner
('019791d2-adef-74be-b82a-079b56877764', '01979221-694f-7ba0-8930-8e7b9e147c2e'),
-- AuthenticatedUser
('019791d2-adef-758d-b043-55ea5be663a0', '01979221-694f-7b7d-ab32-5c301e0e1745'),
('019791d2-adef-758d-b043-55ea5be663a0', '01979221-694f-7b69-a3a8-c8fce0f43afc');

-----------------------------------------------------------------------------------------
-- table users_roles
-----------------------------------------------------------------------------------------
INSERT INTO users_roles (users_id, roles_id) VALUES
-- Administrator
('019791d2-adef-76d2-a865-5b19e5073e60', '019791d2-adef-74be-b82a-079b56877764');

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
