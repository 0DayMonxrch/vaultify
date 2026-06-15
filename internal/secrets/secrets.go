package secrets

import (
	"context"
	"encoding/base64"
	"errors"

	"github.com/0DayMonxrch/vaultify/internal/crypto"
	"github.com/0DayMonxrch/vaultify/internal/ctxkey"
	"github.com/0DayMonxrch/vaultify/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrProjectNotFound = errors.New("secrets: project not found")
	ErrSecretNotFound  = errors.New("secrets: secret not found")
	ErrUnauthorized    = errors.New("secrets: unauthorized")
)

type AuditLogger interface {
	WriteLog(ctx context.Context, projectID uuid.UUID, action string, keyName *string) error
}

type SecretService struct {
	queries   *db.Queries
	audit     AuditLogger
	masterKey []byte
}

func NewSecretService(queries *db.Queries, auditLogger AuditLogger, masterKey []byte) *SecretService {
	return &SecretService{
		queries:   queries,
		audit:     auditLogger,
		masterKey: masterKey,
	}
}

func (s *SecretService) CreateSecret(ctx context.Context, projectID uuid.UUID, keyName, env string, plaintext []byte) (db.Secret, error) {
	userIDStr, ok := ctx.Value(ctxkey.UserID).(string)
	if !ok || userIDStr == "" {
		return db.Secret{}, ErrUnauthorized
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return db.Secret{}, ErrUnauthorized
	}

	project, err := s.queries.GetProjectById(ctx, pgtype.UUID{Bytes: projectID, Valid: true})
	if err != nil {
		return db.Secret{}, ErrProjectNotFound
	}

	key, err := crypto.DeriveKey(s.masterKey, project.KekSalt)
	if err != nil {
		return db.Secret{}, err
	}
	defer clear(key)

	ciphertext, nonce, err := crypto.Encrypt(plaintext, key)
	if err != nil {
		return db.Secret{}, err
	}
	// Memory Safety (Critical): zero out plaintext immediately after Encrypt
	clear(plaintext)

	encValBase64 := base64.StdEncoding.EncodeToString(ciphertext)
	nonceBase64 := base64.StdEncoding.EncodeToString(nonce)

	secret, err := s.queries.CreateSecret(ctx, db.CreateSecretParams{
		ProjectID:      pgtype.UUID{Bytes: projectID, Valid: true},
		KeyName:        keyName,
		Environment:    env,
		EncryptedValue: encValBase64,
		Nonce:          nonceBase64,
		CreatedBy:      pgtype.UUID{Bytes: userID, Valid: true},
	})
	if err != nil {
		return db.Secret{}, err
	}

	_ = s.audit.WriteLog(ctx, projectID, "SECRET_WRITE", &keyName)

	return secret, nil
}

func (s *SecretService) GetSecret(ctx context.Context, projectID, secretID uuid.UUID) ([]byte, error) {
	secret, err := s.queries.GetSecretByID(ctx, pgtype.UUID{Bytes: secretID, Valid: true})
	if err != nil {
		return nil, ErrSecretNotFound
	}

	if secret.ProjectID.Bytes != projectID {
		return nil, ErrSecretNotFound
	}

	project, err := s.queries.GetProjectById(ctx, pgtype.UUID{Bytes: projectID, Valid: true})
	if err != nil {
		return nil, ErrProjectNotFound
	}

	key, err := crypto.DeriveKey(s.masterKey, project.KekSalt)
	if err != nil {
		return nil, err
	}
	defer clear(key)

	ciphertext, err := base64.StdEncoding.DecodeString(secret.EncryptedValue)
	if err != nil {
		return nil, err
	}

	nonce, err := base64.StdEncoding.DecodeString(secret.Nonce)
	if err != nil {
		return nil, err
	}

	plaintext, err := crypto.Decrypt(ciphertext, nonce, key)
	if err != nil {
		return nil, err
	}

	_ = s.audit.WriteLog(ctx, projectID, "SECRET_READ", &secret.KeyName)

	return plaintext, nil
}

func (s *SecretService) ListSecrets(ctx context.Context, projectID uuid.UUID, env string) ([]db.ListSecretsByProjectRow, error) {
	secrets, err := s.queries.ListSecretsByProject(ctx, pgtype.UUID{Bytes: projectID, Valid: true})
	if err != nil {
		return nil, err
	}

	if env == "" {
		return secrets, nil
	}

	var filtered []db.ListSecretsByProjectRow
	for _, sec := range secrets {
		if sec.Environment == env {
			filtered = append(filtered, sec)
		}
	}

	return filtered, nil
}

func (s *SecretService) UpdateSecret(ctx context.Context, projectID, secretID uuid.UUID, plaintext []byte) (db.Secret, error) {
	secret, err := s.queries.GetSecretByID(ctx, pgtype.UUID{Bytes: secretID, Valid: true})
	if err != nil || secret.ProjectID.Bytes != projectID {
		return db.Secret{}, ErrSecretNotFound
	}

	project, err := s.queries.GetProjectById(ctx, pgtype.UUID{Bytes: projectID, Valid: true})
	if err != nil {
		return db.Secret{}, ErrProjectNotFound
	}

	key, err := crypto.DeriveKey(s.masterKey, project.KekSalt)
	if err != nil {
		return db.Secret{}, err
	}
	defer clear(key)

	ciphertext, nonce, err := crypto.Encrypt(plaintext, key)
	if err != nil {
		return db.Secret{}, err
	}
	// Memory Safety (Critical): zero out plaintext immediately after Encrypt
	clear(plaintext)

	encValBase64 := base64.StdEncoding.EncodeToString(ciphertext)
	nonceBase64 := base64.StdEncoding.EncodeToString(nonce)

	updatedSecret, err := s.queries.UpdateSecret(ctx, db.UpdateSecretParams{
		ID:             pgtype.UUID{Bytes: secretID, Valid: true},
		EncryptedValue: encValBase64,
		Nonce:          nonceBase64,
	})
	if err != nil {
		return db.Secret{}, err
	}

	_ = s.audit.WriteLog(ctx, projectID, "SECRET_WRITE", &updatedSecret.KeyName)

	return updatedSecret, nil
}

func (s *SecretService) DeleteSecret(ctx context.Context, projectID, secretID uuid.UUID) error {
	secret, err := s.queries.GetSecretByID(ctx, pgtype.UUID{Bytes: secretID, Valid: true})
	if err != nil || secret.ProjectID.Bytes != projectID {
		return ErrSecretNotFound
	}

	err = s.queries.DeleteSecret(ctx, pgtype.UUID{Bytes: secretID, Valid: true})
	if err != nil {
		return err
	}

	_ = s.audit.WriteLog(ctx, projectID, "SECRET_DELETE", &secret.KeyName)

	return nil
}
