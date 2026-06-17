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

	limit := 50
	page := 1
	pageStr := r.URL.Query().Get("page")
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	offset := (page - 1) * limit

	logs, err := h.svc.ListProjectLogsWithEmail(r.Context(), projectID, int32(limit), int32(offset))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	total, _ := h.svc.CountProjectLogs(r.Context(), projectID)
	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}

	type AuditEvent struct {
		ID            string `json:"id"`
		ProjectID     string `json:"project_id"`
		UserEmail     string `json:"user_email"`
		Action        string `json:"action"`
		TargetKeyName string `json:"target_key_name"`
		IpAddress     string `json:"ip_address"`
		CreatedAt     string `json:"created_at"`
	}

	var events []AuditEvent
	for _, l := range logs {
		ip := ""
		if l.IpAddress != nil {
			ip = l.IpAddress.String()
		}
		events = append(events, AuditEvent{
			ID:            strconv.FormatInt(l.ID, 10),
			ProjectID:     uuid.UUID(l.ProjectID.Bytes).String(),
			UserEmail:     l.UserEmail,
			Action:        l.Action,
			TargetKeyName: l.KeyName.String,
			IpAddress:     ip,
			CreatedAt:     l.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		})
	}

	if events == nil {
		events = []AuditEvent{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":        events,
		"total":       total,
		"page":        page,
		"per_page":    limit,
		"total_pages": totalPages,
	})
}
