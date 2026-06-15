package audit

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/0DayMonxrch/vaultify/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handlers struct {
	svc *AuditService
}

func NewHandlers(svc *AuditService) *Handlers {
	return &Handlers{svc: svc}
}

func (h *Handlers) RegisterRoutes(r chi.Router, authMw *middleware.AuthMiddleware) {
	r.Route("/projects/{id}/audit", func(r chi.Router) {
		r.Use(authMw.Authenticator)
		r.Use(authMw.ContextEnricher)
		r.Use(authMw.RequireMember)

		r.Get("/", h.HandleListAuditLogs)
	})
}

func (h *Handlers) HandleListAuditLogs(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = o
	}

	logs, err := h.svc.ListProjectLogs(r.Context(), projectID, int32(limit), int32(offset))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if logs == nil {
		w.Write([]byte("[]"))
		return
	}
	json.NewEncoder(w).Encode(logs)
}
