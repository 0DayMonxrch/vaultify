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
