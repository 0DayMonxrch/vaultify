# API Token Specification

### The "Shown Once" Guarantee & UX Architecture

- **The Security Consequence:** Storing the raw API token in cleartext or via reversible encryption means a complete compromise of the PostgreSQL database exposes every active token. This allows attackers to instantly hijack CLI integrations across all projects without detection, rendering breach containment impossible.

- **The UX Implementation:** Because the raw token cannot be reconstructed from the SHA-256 hash, the `/tokens` list UI relies entirely on the `token_prefix` (e.g., `vt_a1b2c3d4`) and metadata. The UI displays a masked view (e.g., `vt_a1b2c3d4************************`) alongside the token's name, creation timestamp, role scope, and `last_used_at` telemetry to give the user identification capability without leaking secrets.

---

### SHA-256 over Argon2id

- **The Math Behind the Input:** Argon2id is a slow, memory-hard hashing function designed to defend low-entropy, guessable inputs (like human-generated passwords) against offline brute-force attacks. Vaultify's API tokens are generated via crypto/rand, yielding 32 bytes of true randomness.

- **Why SHA-256 Wins Here:** An input with 256 bits of full cryptographic entropy is mathematically impossible to brute-force or precompute via rainbow tables, making Argon2id’s defensive work factors completely redundant. Therefore, a single iteration of SHA-256 is the correct primitive because it allows the system to compute hashes instantly ($O(1)$ CPU usage) and leverage native B-Tree database indexing for high-throughput lookups without introducing an intentional denial-of-service vector on the server.

---

### The Timing Attack Surface Evaluation

- **Where It Surfaces:** The potential timing attack surface sits during the database lookup step: SELECT * FROM api_tokens WHERE token_hash = $1. Database B-Tree indexes string-match from left-to-right, meaning a query that fails on the first character rejects marginally faster than a query that matches a long prefix before failing.

- **Architectural Position:** Mitigation is completely unnecessary. A string-comparison timing attack requires microsecond-level precision across thousands of identical requests to filter out environmental variance.

- **Why It Doesn't Matter in Practice:** In Vaultify, the incoming token is hashed with SHA-256 before hitting the database lookup. Because hashes are uniform and non-linear, an attacker cannot guess a raw token that yields an "almost matching" hash. Furthermore, network jitter, connection pooling overhead, database lock contention, and our middleware's dynamic PostgreSQL project/role checks inject massive non-deterministic noise that completely flattens any measurable cryptographic timing signals.

---

### Token Lifecycle

Every API token progresses through four strict states, undergoing unidirectional cryptographic transformations along the way:
```plaintext
[1. Generation] ──> crypto/rand (32 bytes) ──> Base64 Encode ──> vt_ + 44-char string
                          │
                          ▼ (Transform: SHA-256 Hash)
[2. Storage]    ──> hex.EncodeToString() ──> Store in `api_tokens.token_hash`
                          │
                          ▼ (Incoming Request Match)
[3. Validation] ──> Hash Request Token ──> Index-Scan matching `token_hash` & `revoked = false`
                          │
                          ▼ (Admin Delete Request)
[4. Revoking]   ──> UPDATE api_tokens SET revoked = true WHERE id = $1

```

- **Generation:** The server reads 32 bytes of cryptographically secure randomness from the operating system's entropy pool (crypto/rand). This binary payload is transformed into an alpha-numeric string using standard URL-Safe Base64 encoding.

- **Storage:** The raw string is displayed to the user exactly once. The server immediately hashes the raw string using `crypto/sha256`, transforming it into a fixed-length 64-character hex string which is written to the database's `token_hash` column. The cleartext token is instantly dropped from the server's memory.

- **Validation:** When a client provides the token, the server runs the identical `sha256.Sum256()` hashing step on the incoming string and uses the resulting hex representation to perform an optimized index lookup.

- **Revocation:** The token record is permanently neutralized in the database. Future request lookups immediately fail to match active criteria.

---

### Token Format

Vaultify enforces a rigid, easily identifiable token structure to simplify security filtering, logging, and client handling:

`vt_  •  8a7f9b2c  •  A1b2C3d4E5f6G7h8I9j0K1l2M3n4O5p6Q7r8S9t0U1v`

- **The vt_ Prefix:** Stands for "Vaultify Token". This hardcoded explicit prefix acts as a signature pattern. It enables automated Secret Scanning tools (like GitHub Advanced Security or Trufflehog) to effortlessly detect accidentally leaked tokens in source code repositories before they are committed.

- **The Suffix Generation:** Composed of a display suffix and a random payload. The server extracts the first 8 characters of the Base64-encoded random payload to serve as the immutable token_prefix for UX display and admin tracking. The remaining segment contains the rest of the high-entropy base64 string, ensuring that the token is long enough to prevent brute-force attacks.

- **The Final String Example:** `vt_8a7f9b2cA1b2C3d4E5f6G7h8I9j0K1l2M3n4O5p6Q7r8S9t0U1v`. This format packs exactly 256 bits of underlying entropy into a single, compact, copy-pasteable configuration string.

---

### Token Validation Middleware

When the Authenticator middleware catches an X-Vaultify-Token header, it halts standard JWT evaluation and executes this linear validation checklist:

- **Step 1:** Format Structural Check. Verify the incoming string starts with vt_ and matches the expected character length. If malformed, stop and return a 401 Unauthorized response.

- **Step 2:** Cryptographic Hashing. Compute the SHA-256 hash of the entire incoming raw token string inside the Go runtime to convert it into a lookup-ready hex string.

- **Step 3:** Database Index Search. Execute a precise lookup query: `SELECT user_id, project_id, role, revoked, expires_at FROM api_tokens WHERE token_hash = $1;`
If the row is missing (no hash match), abort immediately and return a 401 Unauthorized to prevent user enumeration.

- **Step 4:** Status Evaluation. Check the revoked boolean column. If revoked == true, abort immediately and return a 401 Unauthorized.

- **Step 5:** Expiration Verification. Check the expires_at timestamp. If the current system time is past expires_at, abort immediately and return a 401 Unauthorized.

- **Step 6:** Context Injection & Telemetry. Inject the validated user_id, project_id, and token role context directly into the request's Go `context.Context` payload. Concurrently dispatch an asynchronous background query to update the token's `last_used_at` timestamp and log a SECRET_READ or SECRET_WRITE marker to the audit_log. Pass the request to the downstream RBAC middleware.

---

### Token Revocation

- **Database Execution:** Calling DELETE `/tokens/:id` does not execute an SQL DELETE structural row wipe. Instead, the endpoint triggers an explicit soft-revocation mutation:
`UPDATE api_tokens SET revoked = true WHERE id = $1 AND user_id = $2;`

- **Why Soft-Revoking is Mandatory:** Hard-deleting rows would destroy foreign key integrity, severing the audit trail link on historic logs that track which token fetched specific keys. Flipping the revoked flag to true instantly breaks the active lookup validation path while keeping data intact for compliance auditing.

- **In-Flight Request Treatment:** Because Vaultify API tokens are validated dynamically via database lookups on every incoming request, revocation takes effect immediately. Any in-flight HTTP request that has already passed the middleware checks will finish executing, but any subsequent connection made even milliseconds after the database update drops will be rejected instantly at the middleware layer with a 401 Unauthorized response.

---

### Token Expiry

- **What NULL Signifies:** A value of `NULL` in the `expires_at` column represents a token with no expiration date. This configuration is common for long-running infrastructure integrations, automation daemons, or CI/CD pipelines that require persistent access without regular human intervention.

- **How Expiry is Checked:** Expiry validation is handled inline during the single database lookup step. The middleware relies on a SQL condition like `AND (expires_at IS NULL OR expires_at > NOW())` to filter active tokens. This approach avoids the need for separate conditional time-parsing logic within the application code.

- **Why We Choose DB-Side Expiry Over Redis TTL:** While Redis excels at managing temporary user sessions with automated TTL eviction, API tokens require long-term persistence and strict data integrity.

- **The Trade-off Justification:** If tokens relied purely on Redis TTL evictions, a Redis restart, cache flush, or memory exhaustion event could accidentally erase token data, disrupting critical background systems. Using PostgreSQL ensures that your token records remain durable and crash-resilient. Furthermore, because token expiration metadata is tracked natively alongside the token status, the system can display clear historical context on the UI dashboard (e.g., distinguishing between "Expired" vs. "Revoked") even years after a token becomes inactive.

---

