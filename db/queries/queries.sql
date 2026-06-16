-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserById :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: CreateProject :one
INSERT INTO projects (name, slug, kek_salt, created_by)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetProjectById :one
SELECT * FROM projects
WHERE id = $1 LIMIT 1;

-- name: AddProjectMember :exec
INSERT INTO project_members (project_id, user_id, role)
VALUES ($1, $2, $3);

-- name: GetProjectMember :one
SELECT * FROM project_members
WHERE project_id = $1 AND user_id = $2 LIMIT 1;

-- name: GetProjectsForUser :many
SELECT p.* FROM projects p
JOIN project_members pm ON p.id = pm.project_id
WHERE pm.user_id = $1;

-- name: RemoveProjectMember :exec
DELETE FROM project_members
WHERE project_id = $1 AND user_id = $2;

-- name: UpdateProject :one
UPDATE projects
SET name = COALESCE(NULLIF($2::text, ''), name),
    slug = COALESCE(NULLIF($3::text, ''), slug)
WHERE id = $1
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects
WHERE id = $1;

-- name: CreateSecret :one
INSERT INTO secrets (project_id, key_name, environment, encrypted_value, nonce, created_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetSecretByID :one
SELECT * FROM secrets
WHERE id = $1 LIMIT 1;

-- name: ListSecretsByProject :many
SELECT id, key_name, environment, updated_at, created_at
FROM secrets
WHERE project_id = $1;

-- name: UpdateSecret :one
UPDATE secrets
SET encrypted_value = $2,
    nonce = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteSecret :exec
DELETE FROM secrets
WHERE id = $1;

-- name: InsertAuditLog :exec
INSERT INTO audit_log (user_id, project_id, action, key_name, ip_address)
VALUES ($1, $2, $3, $4, $5);

-- name: ListAuditLogsByProject :many
SELECT * FROM audit_log
WHERE project_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
