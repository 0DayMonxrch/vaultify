package projects

import (
	"crypto/rand"
	"encoding/json"
	"log"
	"net/http"
	"net/netip"
	"strconv"

	"github.com/0DayMonxrch/vaultify/internal/ctxkey"
	"github.com/0DayMonxrch/vaultify/internal/db"
	"github.com/0DayMonxrch/vaultify/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Handlers struct {
	queries *db.Queries
}

func NewHandlers(queries *db.Queries) *Handlers {
	return &Handlers{queries: queries}
}

func (h *Handlers) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware) {
	r.Route("/projects", func(r chi.Router) {
		r.Use(authMiddleware.Authenticator)

		// Basic project CRUD
		r.Post("/", h.CreateProject)
		r.Get("/", h.ListProjects)

		// Project-specific routes
		r.Route("/{id}", func(r chi.Router) {
			r.Use(authMiddleware.ContextEnricher)

			r.With(authMiddleware.RequireMember).Get("/", h.GetProject)
			r.With(authMiddleware.RequireMember).Get("/members", h.GetMembers)
			// Project settings (Owner only)
			r.With(authMiddleware.RequireOwner).Patch("/", h.UpdateProject)
			r.With(authMiddleware.RequireOwner).Delete("/", h.DeleteProject)

			// Member management (Owner only)
			r.With(authMiddleware.RequireOwner).Post("/members", h.AddMember)
			r.With(authMiddleware.RequireOwner).Delete("/members/{userId}", h.RemoveMember)
		})
	})
}

func (h *Handlers) CreateProject(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value(ctxkey.UserID).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusInternalServerError)
		return
	}

	// Generate salt for the Key Encryption Key (ADR 002)
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		http.Error(w, "failed to generate crypto salt", http.StatusInternalServerError)
		return
	}

	// Note: We should ideally use a transaction here, but for Foundations we just do it sequentially.
	project, err := h.queries.CreateProject(r.Context(), db.CreateProjectParams{
		Name:      req.Name,
		Slug:      req.Slug,
		KekSalt:   salt,
		CreatedBy: pgtype.UUID{Bytes: userUUID, Valid: true},
	})
	if err != nil {
		http.Error(w, "failed to create project: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Add creator as owner
	err = h.queries.AddProjectMember(r.Context(), db.AddProjectMemberParams{
		ProjectID: project.ID,
		UserID:    pgtype.UUID{Bytes: userUUID, Valid: true},
		Role:      "owner",
	})
	if err != nil {
		http.Error(w, "failed to set project owner", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(project)
}

func (h *Handlers) ListProjects(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value(ctxkey.UserID).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusInternalServerError)
		return
	}

	projects, err := h.queries.GetProjectsForUser(r.Context(), pgtype.UUID{Bytes: userUUID, Valid: true})
	if err != nil {
		http.Error(w, "failed to list projects", http.StatusInternalServerError)
		return
	}

	// If the user has no projects, return an empty array instead of null
	if projects == nil {
		projects = []db.Project{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func (h *Handlers) GetProject(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "id")

	var projectUUID pgtype.UUID
	if err := projectUUID.Scan(projectIDStr); err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	project, err := h.queries.GetProjectById(r.Context(), projectUUID)
	if err != nil {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

func (h *Handlers) GetMembers(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "id")

	var projectUUID pgtype.UUID
	if err := projectUUID.Scan(projectIDStr); err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	members, err := h.queries.ListProjectMembers(r.Context(), projectUUID)
	if err != nil {
		http.Error(w, "failed to fetch members", http.StatusInternalServerError)
		return
	}

	if members == nil {
		members = []db.ListProjectMembersRow{}
	}

	type MemberResponse struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
		Role   string `json:"role"`
	}

	res := make([]MemberResponse, 0)
	for _, m := range members {
		res = append(res, MemberResponse{
			UserID: uuid.UUID(m.UserID.Bytes).String(),
			Email:  m.Email,
			Role:   m.Role,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *Handlers) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "id")
	var projectUUID pgtype.UUID
	if err := projectUUID.Scan(projectIDStr); err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	pageStr := r.URL.Query().Get("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit := 50
	offset := (page - 1) * limit

	logs, err := h.queries.ListAuditLogsWithEmail(r.Context(), projectUUID, int32(limit), int32(offset))
	if err != nil {
		http.Error(w, "failed to fetch audit logs", http.StatusInternalServerError)
		return
	}

	total, _ := h.queries.CountAuditLogs(r.Context(), projectUUID)
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

func (h *Handlers) AddMember(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "id")

	var projectUUID pgtype.UUID
	if err := projectUUID.Scan(projectIDStr); err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Role != "owner" && req.Role != "member" {
		http.Error(w, "invalid role", http.StatusBadRequest)
		return
	}

	user, err := h.queries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	err = h.queries.AddProjectMember(r.Context(), db.AddProjectMemberParams{
		ProjectID: projectUUID,
		UserID:    user.ID,
		Role:      req.Role,
	})
	if err != nil {
		http.Error(w, "failed to add member", http.StatusInternalServerError)
		return
	}

	currentUserIDStr, _ := r.Context().Value(ctxkey.UserID).(string)
	currentUserUUID, _ := uuid.Parse(currentUserIDStr)
	var ipAddr *netip.Addr
	if addrPort, err := netip.ParseAddrPort(r.RemoteAddr); err == nil {
		addr := addrPort.Addr()
		ipAddr = &addr
	}
	err = h.queries.InsertAuditLog(r.Context(), db.InsertAuditLogParams{
		UserID:    pgtype.UUID{Bytes: currentUserUUID, Valid: true},
		ProjectID: projectUUID,
		Action:    "MEMBER_CREATE",
		KeyName:   pgtype.Text{String: req.Email, Valid: true},
		IpAddress: ipAddr,
	})
	if err != nil {
		log.Printf("failed to insert audit log for MEMBER_CREATE: %v", err)
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handlers) RemoveMember(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "id")
	targetUserIDStr := chi.URLParam(r, "userId")

	var projectUUID pgtype.UUID
	if err := projectUUID.Scan(projectIDStr); err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	var targetUserUUID pgtype.UUID
	if err := targetUserUUID.Scan(targetUserIDStr); err != nil {
		http.Error(w, "invalid target user id", http.StatusBadRequest)
		return
	}

	// Prevent removing oneself (optional safety feature, but good practice)
	currentUserIDStr, _ := r.Context().Value(ctxkey.UserID).(string)
	if currentUserIDStr == targetUserIDStr {
		http.Error(w, "cannot remove yourself", http.StatusBadRequest)
		return
	}

	err := h.queries.RemoveProjectMember(r.Context(), db.RemoveProjectMemberParams{
		ProjectID: projectUUID,
		UserID:    targetUserUUID,
	})
	if err != nil {
		http.Error(w, "failed to remove member", http.StatusInternalServerError)
		return
	}

	currentUserUUID, _ := uuid.Parse(currentUserIDStr)
	var ipAddr *netip.Addr
	if addrPort, err := netip.ParseAddrPort(r.RemoteAddr); err == nil {
		addr := addrPort.Addr()
		ipAddr = &addr
	}
	err = h.queries.InsertAuditLog(r.Context(), db.InsertAuditLogParams{
		UserID:    pgtype.UUID{Bytes: currentUserUUID, Valid: true},
		ProjectID: projectUUID,
		Action:    "MEMBER_DELETE",
		KeyName:   pgtype.Text{String: targetUserIDStr, Valid: true},
		IpAddress: ipAddr,
	})
	if err != nil {
		log.Printf("failed to insert audit log for MEMBER_DELETE: %v", err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) UpdateProject(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "id")

	var projectUUID pgtype.UUID
	if err := projectUUID.Scan(projectIDStr); err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// db.UpdateProjectParams comes from sqlc
	project, err := h.queries.UpdateProject(r.Context(), db.UpdateProjectParams{
		ID:      projectUUID,
		Column2: req.Name,
		Column3: req.Slug,
	})
	if err != nil {
		http.Error(w, "failed to update project: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

func (h *Handlers) DeleteProject(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "id")

	var projectUUID pgtype.UUID
	if err := projectUUID.Scan(projectIDStr); err != nil {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}

	if err := h.queries.DeleteProject(r.Context(), projectUUID); err != nil {
		http.Error(w, "failed to delete project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
