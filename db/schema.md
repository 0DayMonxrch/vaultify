# Database Schema

### `users`
```sql
id          UUID PRIMARY KEY DEFAULT gen_random_uuid()
email       TEXT UNIQUE NOT NULL
password_hash TEXT NOT NULL        -- Argon2id hash
created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

### `projects`
```sql
id          UUID PRIMARY KEY DEFAULT gen_random_uuid()
name        TEXT NOT NULL
slug        TEXT UNIQUE NOT NULL   -- URL-safe, used in CLI: "my-app"
kek_salt    BYTEA NOT NULL         -- 32 random bytes, generated once on create
created_by  UUID REFERENCES users
created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

### `project_members`
```sql
project_id  UUID REFERENCES projects ON DELETE CASCADE
user_id     UUID REFERENCES users ON DELETE CASCADE
role        TEXT NOT NULL CHECK (role IN ('owner', 'member'))
PRIMARY KEY (project_id, user_id)
```

### `secrets`
```sql
id              UUID PRIMARY KEY DEFAULT gen_random_uuid()
project_id      UUID REFERENCES projects ON DELETE CASCADE
key_name        TEXT NOT NULL          -- e.g. DATABASE_URL, STRIPE_KEY
environment     TEXT NOT NULL DEFAULT 'production'
encrypted_value TEXT NOT NULL          -- AES-256-GCM ciphertext, base64
nonce           TEXT NOT NULL          -- 12-byte GCM nonce, base64
created_by      UUID REFERENCES users
updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
UNIQUE (project_id, key_name, environment)
```

### `audit_log`
```sql
id          BIGSERIAL PRIMARY KEY
user_id     UUID                   -- no FK, preserved if user is deleted
project_id  UUID                   -- no FK, preserved if project is deleted
action      TEXT NOT NULL          -- see action list below
key_name    TEXT                   -- which secret was touched, if any
ip_address  INET
created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

Audit actions: `SECRET_READ`, `SECRET_WRITE`, `SECRET_DELETE`, `AUTH_LOGIN`,
`AUTH_FAILED`, `MEMBER_INVITE`, `MEMBER_REMOVE`, `TOKEN_CREATE`, `TOKEN_REVOKE`.

### `api_tokens`
```sql
id           UUID PRIMARY KEY DEFAULT gen_random_uuid()
user_id      UUID REFERENCES users ON DELETE CASCADE
project_id   UUID REFERENCES projects ON DELETE CASCADE
name         TEXT NOT NULL
token_hash   TEXT NOT NULL          -- SHA-256 of raw token
token_prefix TEXT NOT NULL          -- first 8 chars, for display
role         TEXT NOT NULL CHECK (role IN ('read', 'write'))
last_used_at TIMESTAMPTZ
expires_at   TIMESTAMPTZ            -- NULL = no expiry
revoked      BOOLEAN NOT NULL DEFAULT false
created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

---