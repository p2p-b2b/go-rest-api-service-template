-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
--
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
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
--
DELETE FROM users;
-- +goose StatementEnd
