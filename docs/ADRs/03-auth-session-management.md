# Authentication & Session Management Spec

### 1. Token Lifecycle

- **Register/Login:** User registers, then logs in via `/auth/login`; the server issues a short-lived JWT Access Token in the JSON response body and sets a long-lived Refresh Token in a secure, HttpOnly cookie while saving the session state in Redis.

- **Access & Expiry:** The client includes the JWT in the `Authorization: Bearer` header for rapid API authorization; once the JWT expires, resource servers reject it.

- **Refresh & Logout:** The client hits `/auth/login` or `/auth/refresh` with the cookie, the server validates it against Redis, deletes the old session, and issues a fresh pair; on `/auth/logout`, the server deletes the session key from Redis and clears the client's cookie.

- **Login vs. Refresh**: The user targets `/auth/login` solely with their raw credentials (email and password) via the request body to initiate a brand-new session.

- **The Token Refresh Process**: The client hits the `/auth/refresh` endpoint exclusively, where the browser automatically appends the HttpOnly cookie containing the current refresh token; conflating these two endpoints breaks standard routing separation and session lifecycle tracking.

---

### 2. Why Refresh Token in `httpOnly` Cookie

- **XSS Protection:** Storing long-lived tokens in JavaScript-accessible storage (like `localStorage` or response bodies) makes them highly vulnerable to theft via Cross-Site Scripting (XSS) attacks.

- **Browser Isolation:** An `HttpOnly` cookie ensures that browser scripts cannot read or extract the token, restricting its transmission strictly to automated, encrypted browser network requests.

- **Scoped Security:** Combined with `Secure` and `SameSite=Strict/Lax` attributes, it dramatically mitigates both token exfiltration and Cross-Site Request Forgery (CSRF) vectors.

---

### 3. Redis Configs (Refresh Tokens)

- **Data Structure**: A simple String structure is used where the key tracks the token/session identifier, and the value stores metadata like user_id and device_fingerprint as a JSON string.

- **Key Schema**: Follows an idiomatic pattern like `session:{user_id}:{refresh_token_uuid}` to allow straightforward lookups and target specific user sessions.

- **TTL**: The Redis key is configured with an explicit Time-To-Live (TTL) that matches the exact lifespan of the refresh token (7 days), ensuring expired sessions self-destruct automatically to optimize memory.

---

### 4. Logout Functionality

- **Server-Side Eviction**: The API server processes the request, extracts the session identifier, and issues a DEL command to wipe the corresponding session key from Redis instantly.

- **Client-Side Clearing**: The server sends an Set-Cookie header in the HTTP response with an expired date (e.g., Max-Age=0) to force the browser to purge the cookie.

- **Instant Authorization Block**: Any subsequent attempts to use that specific refresh token will fail at the Redis validation step, effectively terminating the long-term session.

---

### 5. Logout From All Devices Design

- **Meaning**: It means invalidating every single active session and refresh token associated with a specific user profile across all browsers and devices simultaneously.

- **Implementation**: Execute a Redis `SCAN` operation or query a set tracking user-to-session mappings matching session:{user_id}:*, and execute a pipeline execution to delete all matching keys at once.

- **Alternative Pattern**: Alternatively, maintain an incrementing token_version integer in the PostgreSQL users table, include this version inside the JWT claims, and bump the version number to render all outstanding access tokens and sessions invalid.

### Scaling Session Revocation
- **The Choice:** We reject the SCAN session:{user_id}:* approach for production. While acceptable for tiny self-hosted instances, SCAN scales linearly (O(N)) with the total keyspace, making it a critical performance bottleneck under heavy traffic.

- **The Design Decision:** Vaultify will maintain a Redis Set per user at the key user:sessions:{user_id}. When a login occurs, the new session UUID is added to this set.

- **The Execution:** During a "logout from all devices" event, the server calls SMEMBERS user:sessions:{user_id} to retrieve all active session IDs in O(1) time, passes them directly to a pipelined `DEL` command, and then drops the set.

---

### 6. JWT Access Token

- **Claims**: Carries standard JSON fields including `sub` (User UUID), `exp` (Expiration Unix Timestamp), `iat` (Issued At)

- **TTL & Reason**: 10 mins

- **Client Storage**: Kept purely in short-lived application memory (React state hook), ensuring it is destroyed the moment the user closes or refreshes the browser tab.

- **Project Membership and Role Evaluation**: will be handled dynamically in the API gateway/router middleware on a per-request basis.

---

### 7. Current Tradeoffs

- **The Window**: If an active JWT access token is intercepted or stolen, the attacker can freely make authorized API requests until that specific token naturally reaches its expiration timestamp (up to 10 minutes).

- **Why it is accepted**: Checking a centralized database or blocklist for every single API call defeats the performance benefits of stateless JWTs; the short TTL acts as an engineered engineering trade-off balancing high-throughput scalability with acceptable security risks.

--- 

### Refresh Token Rotation & Replay Attacks

- **Normal Rotation**: On every single hit to `/auth/refresh`, the server immediately deletes the presented refresh token UUID from Redis before issuing a brand-new access token and a newly generated refresh token cookie

- **Replay Attack Defense**: To catch token theft, each session family tracks a version or reuse flag. If an old, already-deleted refresh token is sent to `/auth/refresh`, the server detects a replay attack (indicating an attacker intercepted a token earlier).

- **The Mitigation Action**: The system immediately triggers an alert, invalidates the entire user session set (user:sessions:{user_id}), purges all active tokens for that user from Redis, and forces a hard re-authentication across all devices.

--- 


