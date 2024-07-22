-- +goose Up
-- +goose StatementBegin

-- insert users
-- users with created_at and updated_at
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('483688e9-b8af-42c7-bba4-fb4a29cb7887', 'user', '1',  'user.1@mail.com',  '2024-01-01 01:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('ea1c33d1-76de-4cb2-96b9-844ebbf39cdd', 'user', '2',  'user.2@mail.com',  '2024-01-02 02:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('9dae34f7-0b1a-4ff3-a07f-50aabe80b899', 'user', '3',  'user.3@mail.com',  '2024-01-03 03:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('6898ad86-d299-4aeb-862c-70020a4884b7', 'user', '4',  'user.4@mail.com',  '2024-01-04 04:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('1e34a1a6-49df-4ad3-b8d8-3f92f374659b', 'user', '5',  'user.5@mail.com',  '2024-01-05 05:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('02c2e7d0-34c2-40ef-a904-b66fe8ca9580', 'user', '6',  'user.6@mail.com',  '2024-01-06 06:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('218c8792-1bcc-4975-9a4d-662fc4c3fcf0', 'user', '7',  'user.7@mail.com',  '2024-01-07 07:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('17f88acd-5bc9-46f6-b81b-43de9b3b4385', 'user', '8',  'user.8@mail.com',  '2024-01-08 08:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('117b9606-9be5-44c9-b457-752192f0c25e', 'user', '9',  'user.9@mail.com',  '2024-01-09 09:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('9ae16eec-d5f6-40f3-99a4-10faa2083a80', 'user', '10', 'user.10@mail.com', '2024-01-10 10:00:00.000000+00', '2024-05-11 09:14:56.030582+00');

INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('78634c53-6111-48e7-9c77-448f62746f3a', 'user', '11', 'user.11@mail.com', '2024-02-01 01:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('1863d095-3f3a-4df5-9a37-3c8566f38424', 'user', '12', 'user.12@mail.com', '2024-02-02 02:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('ec916281-1cd4-455d-bc84-a1f7c40e407d', 'user', '13', 'user.13@mail.com', '2024-02-03 03:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('2e707625-ef3a-43f6-b1f6-afc1eb3a2186', 'user', '14', 'user.14@mail.com', '2024-02-04 04:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('f5cacfb5-dcf5-41ce-9347-b69274bb91bc', 'user', '15', 'user.15@mail.com', '2024-02-05 05:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('9d560296-e72a-4d47-a4de-eeff749f0e29', 'user', '16', 'user.16@mail.com', '2024-02-06 06:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('642cb103-5eb4-4ac3-a723-028903fbad5b', 'user', '17', 'user.17@mail.com', '2024-02-07 07:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('8c702c57-e6d8-4b7f-b220-6b8f6c336916', 'user', '18', 'user.18@mail.com', '2024-02-08 08:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('8d200236-be28-4796-b886-342d762f0ec2', 'user', '19', 'user.19@mail.com', '2024-02-09 09:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('8abf5a9c-ab2a-4f55-b633-bf1215268810', 'user', '20', 'user.20@mail.com', '2024-02-10 10:00:00.000000+00', '2024-05-11 09:14:56.030582+00');

INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('9625f602-b91a-4618-9f0b-4261837cc039', 'user', '21', 'user.21@mail.com', '2024-03-01 01:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('621389f1-5666-4ca9-af33-464d618f08f7', 'user', '22', 'user.22@mail.com', '2024-03-02 02:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('ae29d1ad-65c4-49d2-81c5-0237b2dd482a', 'user', '23', 'user.23@mail.com', '2024-03-03 03:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('7e045188-0586-4080-9494-f6c3237492a6', 'user', '24', 'user.24@mail.com', '2024-03-04 04:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('8aee6242-037c-4d13-ae67-8baeb3556da2', 'user', '25', 'user.25@mail.com', '2024-03-05 05:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('97094949-2c91-446a-b50e-5bcb27b5bbef', 'user', '26', 'user.26@mail.com', '2024-03-06 06:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('085d16e2-0200-47f9-8bdc-732dd12677be', 'user', '27', 'user.27@mail.com', '2024-03-07 07:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('c3b11505-9606-4046-b1f2-7a2a5cf6df58', 'user', '28', 'user.28@mail.com', '2024-03-08 08:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('bfdfc84d-33a1-4c5a-b61d-7bd35a5819cd', 'user', '29', 'user.29@mail.com', '2024-03-09 09:00:00.000000+00', '2024-05-11 09:14:56.030582+00');
INSERT INTO public.users (id, first_name, last_name, email, created_at, updated_at) VALUES ('6f7c13c8-9c6a-432f-a5f6-80a0a1bd29eb', 'user', '30', 'user.30@mail.com', '2024-03-10 10:00:00.000000+00', '2024-05-11 09:14:56.030582+00');

-- users without created_at and updated_at
INSERT INTO public.users (id, first_name, last_name, email) VALUES
('ec728934-bc76-40e6-80c8-20a6f9535468', 'user', '31', 'user.31@mail.com'),
('3afb6be1-4125-401f-a482-2fd58f0bfc08', 'user', '32', 'user.32@mail.com'),
('7cb3f482-af6f-45d8-b956-ee7f3edbabf4', 'user', '33', 'user.33@mail.com'),
('fb8812e6-45fd-42a1-b623-484ae0cae2fb', 'user', '34', 'user.34@mail.com'),
('7b404583-062f-409d-ba4c-6cc8f40e67ec', 'user', '35', 'user.35@mail.com'),
('a2273def-aa32-4ceb-b7a5-62355197e313', 'user', '36', 'user.36@mail.com'),
('a0256b99-82f3-42a2-a49b-12ea51869101', 'user', '37', 'user.37@mail.com'),
('b6593db8-e488-4dcb-965b-555d7cb2e203', 'user', '38', 'user.38@mail.com'),
('1c8768ae-beb6-412a-9dd9-a6b8f93ac721', 'user', '39', 'user.39@mail.com'),
('bf2bf061-599f-424a-a646-0185e70275d0', 'user', '40', 'user.40@mail.com'),
('3081c8a5-59f2-4249-afb9-ca393ce10f94', 'user', '41', 'user.41@mail.com'),
('1e480e0e-7216-4a71-bcb5-8d2986d7d94a', 'user', '42', 'user.42@mail.com'),
('cbff082f-5824-434e-8309-e592ee6462e2', 'user', '43', 'user.43@mail.com'),
('bc4eac22-c3ae-4616-ae90-536e877b1ae3', 'user', '44', 'user.44@mail.com'),
('5b7dfe58-40d8-4a04-90df-4050712d001b', 'user', '45', 'user.45@mail.com'),
('af676299-94e4-4845-ba8d-6ac445d0faa8', 'user', '46', 'user.46@mail.com'),
('47864f2f-a8d9-4840-8cae-2034c9b64fa4', 'user', '47', 'user.47@mail.com'),
('d4d227ca-0ef6-4984-abae-f2bf1a6fc4c3', 'user', '48', 'user.48@mail.com'),
('cb85b9f9-d22f-4f63-9a47-6b1cae5ff4b7', 'user', '49', 'user.49@mail.com'),
('8de5743a-f2e6-4acd-9921-ef5416841bac', 'user', '50', 'user.50@mail.com'),
('0f683a93-00d9-42eb-b515-e0535e85e33d', 'user', '51', 'user.51@mail.com');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- delete users
DELETE FROM users;

-- +goose StatementEnd
