-- +goose Up
-- +goose StatementBegin

---------------------------------------------------------------------------------------------------
-- table for users
---------------------------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY NOT NULL UNIQUE DEFAULT uuidv7(),
    first_name VARCHAR(25) NOT NULL,
    last_name VARCHAR(25) NOT NULL,
    -- RFC 5321's cap, the same one ValidateEmail enforces. At 50 an address the
    -- validator had accepted was refused by the column, which the API could only
    -- report as a 500.
    email VARCHAR(254) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    -- disabled says whether the account may sign in; verified says whether the
    -- address has been proven. They used to be one flag: a registration was
    -- disabled until the verification link was followed, and the service could
    -- not tell that account from one an administrator had switched off. Password
    -- recovery therefore refused both, silently. Registration sets verified FALSE
    -- and disabled TRUE; following the link sets verified TRUE, and disabled
    -- FALSE unless an administrator disabled the account (disabled_by_admin);
    -- an update touches disabled and never verified. An account an
    -- administrator creates is stored verified: the administrator vouches for
    -- the address, and no link is mailed for it.
    disabled BOOLEAN NOT NULL DEFAULT TRUE,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    -- disabled_by_admin says whose doing a disabled account is. "Disabled until
    -- the link is followed" and "disabled by an administrator" were the same
    -- two values, verified FALSE and disabled TRUE, and following the link
    -- enables every account with them. It is set for an account stored
    -- disabled by POST /users, and follows a CHANGE of disabled made through
    -- PUT /users/{user_id}, in that statement; the verification enables only
    -- a row that does not have it. Not who disabled the account by name: "a
    -- disable that a link must respect".
    --
    -- Added to this file after it first shipped, as the rule is before a
    -- first deploy. goose does not see an edit, so a database created from
    -- an earlier version starts and then fails every write to this table:
    -- recreate it, or
    --   ALTER TABLE users ADD COLUMN disabled_by_admin BOOLEAN NOT NULL DEFAULT FALSE;
    -- and set it by hand on the unverified, disabled rows that an
    -- administrator disabled, which that statement leaves unmarked.
    disabled_by_admin BOOLEAN NOT NULL DEFAULT FALSE,
    -- disabled, admin and local_account are NOT NULL: a NULL was read three
    -- ways. With `disabled` NULL and the flag, the sign-in and the account's
    -- standing took the account for enabled, while the count of enabled
    -- administrator accounts (`disabled IS FALSE`) left it out; the
    -- projects' view, which splits on `admin = TRUE` and `admin = FALSE`,
    -- gave an account with `admin` NULL no project at all. A flag is set or
    -- it is not.
    --
    -- Made NOT NULL in this file after it first shipped, as the rule is
    -- before a first deploy. goose does not see an edit, so a database
    -- created from an earlier version keeps the nullable columns: recreate
    -- it, or
    --   UPDATE users SET disabled = COALESCE(disabled, TRUE),
    --                    admin = COALESCE(admin, FALSE),
    --                    local_account = COALESCE(local_account, TRUE);
    --   ALTER TABLE users ALTER COLUMN disabled SET NOT NULL,
    --                     ALTER COLUMN admin SET NOT NULL,
    --                     ALTER COLUMN local_account SET NOT NULL;
    -- which gives a value never set the column's default.
    admin BOOLEAN NOT NULL DEFAULT FALSE,
    local_account BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- An account an administrator locked is a disabled one. Every statement
    -- of the service keeps the two together; this holds a hand-written one
    -- to it, because an enabled account that is marked is read back as
    -- "locked" by a client and cannot be locked through the API (the value
    -- is no change). "IS TRUE", because disabled is nullable here and a
    -- check that is NULL passes: a marked row whose disabled is NULL signs
    -- in. Added to this file as the column was: a database created from an
    -- earlier version does not have it until it is recreated, or
    --   ALTER TABLE users ADD CONSTRAINT "users_locked_is_disabled" CHECK (NOT disabled_by_admin OR disabled IS TRUE);
    CONSTRAINT "users_locked_is_disabled" CHECK (NOT disabled_by_admin OR disabled IS TRUE),

    -- email is unique
    CONSTRAINT "users_email" UNIQUE (email),

    -- serial_id is used for pagination
    serial_id BIGSERIAL NOT NULL UNIQUE
);

-- indexes for users
CREATE INDEX "idx_users_pagination" ON users (serial_id, id);

-- +goose StatementEnd

-- The account migration 00003 seeds keeps its `admin` flag, whoever asks.
-- The flag is granted and taken back through the API by an account that has
-- it, and the service keeps the last one that can sign in; this row is the
-- one a deployment starts from and the one the two bootstrap links are
-- guarded for (00007), so it is held here, for a hand-written statement too.
-- It can still be disabled, by an administrator account, while another can
-- sign in.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_protect_bootstrap_administrator()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.id = '019822af-b448-73fb-89a1-447e8f8d1cde'
       AND OLD.admin IS TRUE AND NEW.admin IS NOT TRUE THEN
        RAISE EXCEPTION 'The bootstrap administrator keeps the admin flag.';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER tr_protect_bootstrap_administrator
BEFORE UPDATE OF admin ON users
FOR EACH ROW EXECUTE FUNCTION fn_protect_bootstrap_administrator();
--

-- +goose Down
-- +goose StatementBegin

-- CASCADE because these tables reference one another and the drop order would otherwise
-- have to be maintained by hand; dropping a table takes its indexes and triggers with it.
-- The trigger FUNCTIONS are schema-level objects and survive the table, so they are named.

DROP TABLE IF EXISTS users CASCADE;
DROP FUNCTION IF EXISTS fn_protect_bootstrap_administrator();

-- +goose StatementEnd
