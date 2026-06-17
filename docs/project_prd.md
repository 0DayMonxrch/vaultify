# Vaultify - PRD
**What it is:** A self-hosted secrets manager where teams store API keys, database
URLs, and environment variables encrypted at rest, and a CLI tool that injects them
directly into a running process without ever writing them to disk.

**Why it stands out:** The encryption is real (AES-256-GCM + Argon2id), the CLI
subprocess injection is a non-obvious systems technique, and the audit log demonstrates
security thinking. Together these make it a backend portfolio project that has actual
talking points — not a to-do app with a different name.

**Stack:** `Go`, `PostgreSQL`, `Redis`, `React`, `Docker`

## The Core Feature (Understand This First)

The reason Vaultify exists: developers share secrets badly.

A `.env` file gets committed to git. An API key gets sent over Slack. A database
password lives in a Google Doc. None of these have access control, versioning, or
an audit trail.

Vaultify stores secrets encrypted at rest. Even if someone dumps your PostgreSQL
database, they get ciphertext — nothing readable without the master key. The CLI
injects secrets into a subprocess so they never touch disk on the developer's machine.

---

## What We Are Building

A web application with three parts:

1. **API server** - Go binary. Handles auth, project management, secret CRUD with
   encryption/decryption, audit logging, and API token management.
2. **React dashboard** - SPA embedded in the Go binary. Create projects, manage
   secrets, view audit log, manage API tokens.
3. **CLI tool** - Separate Go binary. Authenticates with an API token, fetches
   secrets, and starts a subprocess with those secrets injected as environment
   variables.


## What You Are NOT Building

These are deliberate cuts, not omissions. Know them — you will be asked.

| Skipped Feature | Why |
|---|---|
| TOTP / 2FA | Adds 1 week of work, not core to the secrets story |
| Secret version history | Adds schema complexity, teaches no new concept |
| Three-tier RBAC | Owner + Member is enough to demonstrate the pattern |
| Team / org management | Out of scope for a solo portfolio project |
| Kubernetes deploy | Fly.io covers the deployment story |
| Custom domain support | Irrelevant to the core feature |
| Webhook notifications | Out of scope |

---


## Tech Stack

| Layer | Choice | Version |
|---|---|---|
| Language | Go | 1.26 |
| Router | chi | v5 |
| DB Driver | pgx + pgxpool | v5 |
| SQL Codegen | sqlc | v2 |
| Migrations | golang-migrate | v4 |
| Redis | go-redis | v9 |
| JWT | golang-jwt/jwt | v5 |
| Encryption | golang.org/x/crypto (Argon2id) + stdlib crypto/aes | latest |
| TOTP | — | not in scope |
| Logging | zerolog | v1 |
| Testing | testify | v1 |
| CLI framework | Cobra | v2 |
| QR codes | — | not in scope |
| Frontend | React + Vite | 19 + 6 |
| UI components | shadcn/ui + Tailwind | v4 |
| Server state | TanStack Query | v5 |
| Forms | react-hook-form + Zod | latest |
| Database | PostgreSQL | 17 |
| Cache | Redis | 7 |
| Deploy | Fly.io | — |



## Encryption Design

Every project in Vaultify has a random 32-byte salt stored in the database. When you
read or write a secret, the server derives an encryption key from two things:

```
project_key = Argon2id(MASTER_KEY, project.salt)
```

`MASTER_KEY` is an environment variable — never stored anywhere. `project.salt` is
stored in the database. Neither one alone is enough to decrypt anything.

The secret value is then encrypted with that derived key:

```
ciphertext, nonce = AES-256-GCM.Encrypt(plaintext, project_key)
```

Store `ciphertext` and `nonce` in the database. Store nothing else.

To decrypt: derive the same project key, call `AES-256-GCM.Decrypt(ciphertext, nonce, project_key)`. Zero the plaintext from memory after use.

---

## Database Schema

Six tables

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

## API Endpoints

### Auth
| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/auth/register` | None | Create account |
| POST | `/auth/login` | None | Returns JWT access token + refresh token cookie |
| POST | `/auth/refresh` | Cookie | Rotate refresh token, return new access token |
| DELETE | `/auth/logout` | JWT | Revoke refresh token |
| GET | `/auth/me` | JWT | Return current user info |

### Projects

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/projects` | JWT | List projects the user is a member of |
| POST | `/projects` | JWT | Create project. Creator becomes Owner. |
| GET | `/projects/:id` | JWT | Get project details and member list |
| PATCH | `/projects/:id` | JWT, Owner | Rename project |
| DELETE | `/projects/:id` | JWT, Owner | Delete project and all secrets |
| POST | `/projects/:id/members` | JWT, Owner | Invite user by email, assign role |
| DELETE | `/projects/:id/members/:userId` | JWT, Owner | Remove member |

### Secrets

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/projects/:id/secrets` | JWT or Token | List secret key names (not values). Filter by `?env=`. |
| POST | `/projects/:id/secrets` | JWT or Token (write) | Create a secret. Encrypt before storing. |
| GET | `/projects/:id/secrets/:secretId` | JWT or Token | Decrypt and return value. Write audit log. |
| PUT | `/projects/:id/secrets/:secretId` | JWT or Token (write) | Update a secret's value. Re-encrypt. |
| DELETE | `/projects/:id/secrets/:secretId` | JWT, Owner | Delete secret. Write audit log. |

### Audit & Tokens

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/projects/:id/audit` | JWT | Paginated audit log for the project |
| GET | `/tokens` | JWT | List the user's API tokens (prefix, role, last used) |
| POST | `/tokens` | JWT | Create API token. Returns raw token exactly once. |
| DELETE | `/tokens/:id` | JWT | Revoke a token immediately |

### Utility

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/healthz` | None | Returns `{"status":"ok","postgres":"ok","redis":"ok"}` |

---

## RBAC — Two Roles Only

| Action | Owner | Member |
|---|---|---|
| Read secret values | YES | YES |
| Create / update secrets | YES | YES |
| Delete secrets | YES | NO |
| View audit log | YES | YES |
| Invite / remove members | YES | NO |
| Delete project | YES | NO |
| Create API tokens (read scope) | YES | YES |
| Create API tokens (write scope) | YES | NO |

## CLI Tool Design

The CLI is a separate binary (`cmd/vaultify/main.go`). It communicates with the API
over HTTPS. It has no shared code with the server — it is a client.

### Commands

```
vaultify login --token vt_xxxx --host https://your-vaultify.fly.dev
vaultify logout
vaultify projects list
vaultify secrets list --project my-app --env production
vaultify secrets get --project my-app DATABASE_URL
vaultify run --project my-app --env production -- node server.js
```

### The `run` Command — How It Works

This is the reason the CLI exists. When you run:

```bash
vaultify run --project my-app --env production -- ./bin/server
```

The CLI:
1. Reads the stored API token from `~/.vaultify/config` (file created by `vaultify login`)
2. Calls `GET /projects/my-app/secrets?env=production` → receives `[{key: "DATABASE_URL", value: "postgres://..."}, ...]`
3. Builds a `[]string` slice of `KEY=VALUE` pairs in memory
4. Calls `exec.Command("./bin/server")` with `proc.Env = append(os.Environ(), secrets...)`
5. The subprocess starts with secrets as environment variables
6. Secrets never touch disk. They exist in memory from API response to subprocess env.
7. After `proc.Start()`, zero the secrets slice.

```go
// The core of vaultify run
proc := exec.CommandContext(ctx, args[0], args[1:]...)
proc.Stdin = os.Stdin
proc.Stdout = os.Stdout
proc.Stderr = os.Stderr
proc.Env = append(os.Environ(), secretEnvVars...)
return proc.Run()
```

### Config File

Stored at `~/.vaultify/config` with `0600` permissions (only the owner can read it).

```toml
host = "https://your-vaultify.fly.dev"
token = "vt_xxxxxxxxxxxxxxxxxxxxx"
default_project = "my-app"
```

---

## Makefile Commands

```makefile
make dev      # docker-compose up (postgres + redis)
make test     # go test ./...
make lint     # golangci-lint run
make migrate  # run pending migrations
make sqlc     # sqlc generate
make build    # go build ./cmd/api and ./cmd/vaultify
```

---

## Environment Variables

| Variable | Required | Notes |
|---|---|---|
| `MASTER_KEY` | YES | Hex string, 64 chars. `openssl rand -hex 32`. Never stored. If lost, all secrets are unrecoverable. |
| `DATABASE_URL` | YES | `postgres://user:pass@host:5432/vaultify` |
| `REDIS_URL` | YES | `redis://:pass@host:6379/0` |
| `JWT_SECRET` | YES | Min 32 bytes. `openssl rand -base64 32` |
| `APP_URL` | YES | `https://your-vaultify.fly.dev` |
| `PORT` | No | Default `8080` |
| `ENV` | No | `development` or `production` |

---

## What to Put in Your README

The README is a recruiting asset, not a feature list. Write these sections:

**1. What it is** (2 sentences max)

> Vaultify is a self-hosted secrets manager built in Go. It stores environment
> variables and API keys encrypted at rest using AES-256-GCM, and provides a CLI
> tool that injects secrets into a subprocess without writing them to disk.

**2. Architecture** — simple diagram showing: CLI → API → Postgres (encrypted at rest),
Redis (sessions + rate limits). One paragraph explaining the three components.

**3. How the encryption works** — explain the Argon2id + AES-256-GCM model in plain
English. Why Argon2id? What does the salt do? What happens if someone dumps the DB?
This is your main talking point.

**4. How the CLI injection works** — explain `exec.Cmd.Env`, why secrets never touch
disk, what `os.Environ()` is.

**5. Local setup** — three commands to get it running.

**6. Live demo** — your Fly.io URL.

---

## What You Will Have Built by the End

A working, deployed, self-hostable application that:
- Encrypts secrets at rest with industry-standard cryptography
- Has a CLI tool that uses a non-obvious OS-level technique (subprocess env injection)
- Has proper JWT auth with refresh token rotation
- Has RBAC enforced in middleware, not in handlers
- Has an audit log of every secret access
- Is deployed at a public URL
- Has a README that explains the engineering decisions

That is enough to carry a 20-minute technical interview conversation about
cryptography, auth, RBAC, Go systems programming, and deployment — without
overstating what you built.