package secrets

import (
	"encoding/json"
	"net/http"

	"github.com/0DayMonxrch/vaultify/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handlers struct {
	svc *SecretService
}

func NewHandlers(svc *SecretService) *Handlers {
	return &Handlers{svc: svc}
}

func (h *Handlers) RegisterRoutes(r chi.Router, authMw *middleware.AuthMiddleware) {
	r.Route("/projects/{id}/secrets", func(r chi.Router) {
		r.Use(authMw.Authenticator)
		r.Use(authMw.ContextEnricher)
		r.Use(authMw.RequireMember)

		r.Get("/", h.HandleListSecrets)
		r.Post("/", h.HandleCreateSecret)

		r.Route("/{secretId}", func(r chi.Router) {
			r.Get("/", h.HandleGetSecret)
			r.Put("/", h.HandleUpdateSecret)

			// Delete requires Owner
			r.With(authMw.RequireOwner).Delete("/", h.HandleDeleteSecret)
		})
	})
}

type secretRequest struct {
	KeyName     string `json:"key_name,omitempty"`
	Environment string `json:"environment,omitempty"`
	Value       string `json:"value"`
}

func (h *Handlers) HandleCreateSecret(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	var req secretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	plaintext := []byte(req.Value)
	defer clear(plaintext)

	if req.Environment == "" {
		req.Environment = "production"
	}

	secret, err := h.svc.CreateSecret(r.Context(), projectID, req.KeyName, req.Environment, plaintext)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(secret)
}

func (h *Handlers) HandleListSecrets(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	env := r.URL.Query().Get("env")

	secrets, err := h.svc.ListSecrets(r.Context(), projectID, env)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if secrets == nil {
		w.Write([]byte("[]"))
		return
	}
	json.NewEncoder(w).Encode(secrets)
}

func (h *Handlers) HandleGetSecret(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	secretIDStr := chi.URLParam(r, "secretId")
	secretID, err := uuid.Parse(secretIDStr)
	if err != nil {
		http.Error(w, "invalid secret id", http.StatusBadRequest)
		return
	}

	plaintext, err := h.svc.GetSecret(r.Context(), projectID, secretID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer clear(plaintext)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"value": string(plaintext),
	})
}

func (h *Handlers) HandleUpdateSecret(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	secretIDStr := chi.URLParam(r, "secretId")
	secretID, err := uuid.Parse(secretIDStr)
	if err != nil {
		http.Error(w, "invalid secret id", http.StatusBadRequest)
		return
	}

	var req secretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	plaintext := []byte(req.Value)
	defer clear(plaintext)

	secret, err := h.svc.UpdateSecret(r.Context(), projectID, secretID, plaintext)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(secret)
}

func (h *Handlers) HandleDeleteSecret(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	secretIDStr := chi.URLParam(r, "secretId")
	secretID, err := uuid.Parse(secretIDStr)
	if err != nil {
		http.Error(w, "invalid secret id", http.StatusBadRequest)
		return
	}

	err = h.svc.DeleteSecret(r.Context(), projectID, secretID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
