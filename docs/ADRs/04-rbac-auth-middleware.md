# RBAC & Auth Middleware Spec

### 1. Authentication vs Authorization

- **Authentication (AuthN):** Establishes identity ("Who are you?"). It verifies the cryptographic signature of a JWT or matches an API token hash against the database.

- **Authorization (AuthZ):** Establishes permissions ("What are you allowed to do?"). It checks if the identified user has sufficient access rights to perform a specific action on a specific project.

- **Why Order is Matters:** You cannot evaluate what someone is allowed to do until you know who they are. Reversing or combining the order breaks the security model because authorization logic relies entirely on the authenticated context (e.g., user_id) injected by the authentication layer.

---

### 2. Middleware Request Chain (Protected Request)

For a protected endpoint like POST /projects/:id/secrets, the request must traverse this exact sequential chain:

```text
[Incoming Request]
       │
       ▼
 1. Recoverer / Logger   --> Catches panics, logs requests (zerolog)
       │
       ▼
 2. Authenticator        --> Extracts JWT or API Token, validates, injects Context (User/Token metadata)
       │
       ▼
 3. Context Enricher     --> Extracts :id from URL, queries DB for user's project membership/role
       │
       ▼
 4. RBAC Gatekeeper      --> Evaluates User Role vs. Token Scope; blocks or allows execution
       │
       ▼
[Target Handler]
```

---

### 3. Parallel Auth Paths Converging into a Single AuthZ Layer

- **Parallel Paths**: The Authenticator middleware checks two distinct headers. If an Authorization: Bearer <JWT> header exists, it routes to the JWT validation path. If a custom X-Vaultify-Token: <raw_token> header exists, it routes to the API token database lookup path.

- **The Convergence Point**: Regardless of the path taken, both extract a unified structural context containing the `user_id` and the `project_id`.

- **The Result**: The downstream Context Enricher and RBAC Gatekeeper middlewares do not care how the client authenticated; they only consume the standardized identity context to enforce identical project role policies.

- **Note:** If both present, the JWT path takes precedence, and the API token header is completely ignored.

---

### 4. Effective Permission Matrix (Token Scope + Project Role)

The effective permission is the strict logical `AND` intersection of the User's Project Role and the API Token Scope

| Project Role | Token Scope | Permitted Actions                                                                                        |
| ------------ | ----------- | -------------------------------------------------------------------------------------------------------- |
| Owner        | `read`      | "Read secret values, view audit logs"                                                                    |
| Owner        | `write`     | "Read/write/delete secrets, view audit logs, manage members/tokens"                                      |
| Member       | `read`      | "Read secret values, view audit logs"                                                                    |
| Member       | `write`     | "Read/write secrets, view audit logs, create read-only tokens (NO Delete Secrets, NO Member Management)" |

---

### 5. Why RBAC in Middlware not Handlers

- **The Risk of Handler Checks:** Placing access logic inside handlers forces manual re-implementation on every endpoint, introducing human error during code duplication.

- **The Attack Scenario:** If Vaultify adds PUT `/projects/:id/secrets/:secretId/rollback` and the developer omits the `if user.Role != "owner"` condition, a basic member can perform owner-only structural actions.

- **The Middleware Defense:** By binding the route declaratively in the router config (`r.With(VerifyOwner).Put("/rollback", HandleRollback)`), the handler is structurally decoupled from security policies, enforcing a fail-closed architecture by default.

---

### 6. 401 Unauthorized vs. 403 Forbidden

- **401 Unauthorized (Unauthenticated):** Returned when the request fails the Authentication step. This occurs if the JWT signature is invalid, expired, missing, or if the API token does not exist/is revoked. It means: "We do not know who you are; please log in again."

- **403 Forbidden (Unauthorized):** Returned when the request passes authentication but fails the Authorization step. This occurs if a valid user tries to access a project they don't belong to, or if a project member tries to call an owner-only endpoint. It means: "We know exactly who you are, but you do not have permission to touch this resource."

---

