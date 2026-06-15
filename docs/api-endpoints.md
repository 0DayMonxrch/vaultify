# API Endpoints

### Auth

| Method | Path             | Auth   | Description                                     |
| ------ | ---------------- | ------ | ----------------------------------------------- |
| POST   | `/auth/register` | None   | Create account                                  |
| POST   | `/auth/login`    | None   | Returns JWT access token + refresh token cookie |
| POST   | `/auth/refresh`  | Cookie | Rotate refresh token, return new access token   |
| DELETE | `/auth/logout`   | JWT    | Revoke refresh token                            |
| GET    | `/auth/me`       | JWT    | Return current user info                        |

### Projects

| Method | Path                            | Auth       | Description                            |
| ------ | ------------------------------- | ---------- | -------------------------------------- |
| GET    | `/projects`                     | JWT        | List projects the user is a member of  |
| POST   | `/projects`                     | JWT        | Create project. Creator becomes Owner. |
| GET    | `/projects/:id`                 | JWT        | Get project details and member list    |
| PATCH  | `/projects/:id`                 | JWT, Owner | Rename project                         |
| DELETE | `/projects/:id`                 | JWT, Owner | Delete project and all secrets         |
| POST   | `/projects/:id/members`         | JWT, Owner | Invite user by email, assign role      |
| DELETE | `/projects/:id/members/:userId` | JWT, Owner | Remove member                          |

### Secrets

| Method | Path                              | Auth                 | Description                                            |
| ------ | --------------------------------- | -------------------- | ------------------------------------------------------ |
| GET    | `/projects/:id/secrets`           | JWT or Token         | List secret key names (not values). Filter by `?env=`. |
| POST   | `/projects/:id/secrets`           | JWT or Token (write) | Create a secret. Encrypt before storing.               |
| GET    | `/projects/:id/secrets/:secretId` | JWT or Token         | Decrypt and return value. Write audit log.             |
| PUT    | `/projects/:id/secrets/:secretId` | JWT or Token (write) | Update a secret's value. Re-encrypt.                   |
| DELETE | `/projects/:id/secrets/:secretId` | JWT, Owner           | Delete secret. Write audit log.                        |

### Audit & Tokens

| Method | Path                  | Auth | Description                                          |
| ------ | --------------------- | ---- | ---------------------------------------------------- |
| GET    | `/projects/:id/audit` | JWT  | Paginated audit log for the project                  |
| GET    | `/tokens`             | JWT  | List the user's API tokens (prefix, role, last used) |
| POST   | `/tokens`             | JWT  | Create API token. Returns raw token exactly once.    |
| DELETE | `/tokens/:id`         | JWT  | Revoke a token immediately                           |

### Utility

| Method | Path       | Auth | Description                                            |
| ------ | ---------- | ---- | ------------------------------------------------------ |
| GET    | `/healthz` | None | Returns `{"status":"ok","postgres":"ok","redis":"ok"}` |

---

## RBAC — Two Roles Only

| Action                          | Owner | Member |
| ------------------------------- | ----- | ------ |
| Read secret values              | YES   | YES    |
| Create / update secrets         | YES   | YES    |
| Delete secrets                  | YES   | NO     |
| View audit log                  | YES   | YES    |
| Invite / remove members         | YES   | NO     |
| Delete project                  | YES   | NO     |
| Create API tokens (read scope)  | YES   | YES    |
| Create API tokens (write scope) | YES   | NO     |

Enforce in middleware. Every mutating request checks the role from `project_members`
before doing anything. Never check roles inside handlers — put it in middleware.

---
