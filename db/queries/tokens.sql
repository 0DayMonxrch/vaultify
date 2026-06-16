-- name: CreateToken :one
INSERT INTO api_tokens (user_id, project_id, name, token_hash, token_prefix, role, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListUserTokens :many
SELECT * FROM api_tokens
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetTokenByHash :one
SELECT * FROM api_tokens
WHERE token_hash = $1 
  AND revoked = false
  AND (expires_at IS NULL OR expires_at > NOW())
LIMIT 1;

-- name: UpdateTokenLastUsed :exec
UPDATE api_tokens
SET last_used_at = NOW()
WHERE id = $1;

-- name: SoftRevokeToken :exec
UPDATE api_tokens
SET revoked = true
WHERE id = $1 AND user_id = $2;
