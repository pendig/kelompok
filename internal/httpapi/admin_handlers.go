package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pendig/kelompok/internal/impact"
	"github.com/pendig/kelompok/internal/members"
	"github.com/pendig/kelompok/internal/organizations"
	"github.com/pendig/kelompok/internal/posts"
)

func (s *Server) handleCreateOrganizationClaim(w http.ResponseWriter, r *http.Request) {
	var input organizations.ClaimInput
	if !decodeJSONBody(w, r, &input) {
		return
	}

	item, err := s.organizations.CreateClaim(r.Context(), r.PathValue("slug"), input)
	if errors.Is(err, organizations.ErrNotFound) {
		writeError(w, http.StatusNotFound, "organization_not_found", "Organization not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "claim_create_failed", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusCreated, response{Data: item, Message: "ok"})
}

func (s *Server) handleListAdminOrganizations(w http.ResponseWriter, r *http.Request) {
	limit := limitFromRequest(r, 50, 100)
	items, err := s.organizations.ListPublic(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "admin_organizations_list_failed", "Failed to list organizations", nil)
		return
	}

	writeJSON(w, http.StatusOK, response{
		Data:    items,
		Meta:    map[string]any{"count": len(items), "limit": limit},
		Message: "ok",
	})
}

func (s *Server) handleCreateAdminOrganization(w http.ResponseWriter, r *http.Request) {
	var input organizations.AdminInput
	if !decodeJSONBody(w, r, &input) {
		return
	}

	item, err := s.organizations.Create(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "organization_create_failed", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusCreated, response{Data: item, Message: "ok"})
}

func (s *Server) handleGetAdminOrganization(w http.ResponseWriter, r *http.Request) {
	item, err := s.organizations.FindBySlug(r.Context(), r.PathValue("slug"))
	if errors.Is(err, organizations.ErrNotFound) {
		writeError(w, http.StatusNotFound, "organization_not_found", "Organization not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "organization_lookup_failed", "Failed to load organization", nil)
		return
	}

	writeJSON(w, http.StatusOK, response{Data: item, Message: "ok"})
}

func (s *Server) handleUpdateAdminOrganization(w http.ResponseWriter, r *http.Request) {
	var input organizations.AdminInput
	if !decodeJSONBody(w, r, &input) {
		return
	}

	item, err := s.organizations.UpdateBySlug(r.Context(), r.PathValue("slug"), input)
	if errors.Is(err, organizations.ErrNotFound) {
		writeError(w, http.StatusNotFound, "organization_not_found", "Organization not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "organization_update_failed", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, response{Data: item, Message: "ok"})
}

func (s *Server) handleListOrganizationClaims(w http.ResponseWriter, r *http.Request) {
	if !s.ensureOrganization(w, r, r.PathValue("slug")) {
		return
	}

	limit := limitFromRequest(r, 20, 100)
	items, err := s.organizations.ListClaimsByOrganizationSlug(r.Context(), r.PathValue("slug"), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "claims_list_failed", "Failed to list claim requests", nil)
		return
	}

	writeJSON(w, http.StatusOK, response{
		Data:    items,
		Meta:    map[string]any{"count": len(items), "limit": limit},
		Message: "ok",
	})
}

func (s *Server) handleListOrganizationMembers(w http.ResponseWriter, r *http.Request) {
	limit := limitFromRequest(r, 20, 100)
	items, err := s.members.ListByOrganizationSlug(r.Context(), r.PathValue("slug"), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "members_list_failed", "Failed to list organization members", nil)
		return
	}

	writeJSON(w, http.StatusOK, response{
		Data:    items,
		Meta:    map[string]any{"count": len(items), "limit": limit},
		Message: "ok",
	})
}

func (s *Server) handleCreateOrganizationMember(w http.ResponseWriter, r *http.Request) {
	var input members.Input
	if !decodeJSONBody(w, r, &input) {
		return
	}

	item, err := s.members.Create(r.Context(), r.PathValue("slug"), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "member_create_failed", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusCreated, response{Data: item, Message: "ok"})
}

func (s *Server) handleUpdateAdminMember(w http.ResponseWriter, r *http.Request) {
	var input members.Input
	if !decodeJSONBody(w, r, &input) {
		return
	}

	item, err := s.members.UpdateByID(r.Context(), r.PathValue("id"), input)
	if errors.Is(err, members.ErrNotFound) {
		writeError(w, http.StatusNotFound, "member_not_found", "Member not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "member_update_failed", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, response{Data: item, Message: "ok"})
}

func (s *Server) handleDeleteAdminMember(w http.ResponseWriter, r *http.Request) {
	if err := s.members.DeleteByID(r.Context(), r.PathValue("id")); errors.Is(err, members.ErrNotFound) {
		writeError(w, http.StatusNotFound, "member_not_found", "Member not found", nil)
		return
	} else if err != nil {
		writeError(w, http.StatusBadRequest, "member_delete_failed", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, response{Data: map[string]string{"id": r.PathValue("id")}, Message: "ok"})
}

func (s *Server) handleListAdminPosts(w http.ResponseWriter, r *http.Request) {
	limit := limitFromRequest(r, 50, 100)
	items, err := s.posts.ListAdmin(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "admin_posts_list_failed", "Failed to list posts", nil)
		return
	}

	writeJSON(w, http.StatusOK, response{
		Data:    items,
		Meta:    map[string]any{"count": len(items), "limit": limit},
		Message: "ok",
	})
}

func (s *Server) handleCreateAdminPost(w http.ResponseWriter, r *http.Request) {
	var input posts.AdminInput
	if !decodeJSONBody(w, r, &input) {
		return
	}

	item, err := s.posts.Create(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "post_create_failed", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusCreated, response{Data: item, Message: "ok"})
}

func (s *Server) handleUpdateAdminPost(w http.ResponseWriter, r *http.Request) {
	var input posts.AdminInput
	if !decodeJSONBody(w, r, &input) {
		return
	}

	item, err := s.posts.UpdateByID(r.Context(), r.PathValue("id"), input)
	if errors.Is(err, posts.ErrNotFound) {
		writeError(w, http.StatusNotFound, "post_not_found", "Post not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "post_update_failed", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, response{Data: item, Message: "ok"})
}

func (s *Server) handlePublishAdminPost(w http.ResponseWriter, r *http.Request) {
	item, err := s.posts.PublishByID(r.Context(), r.PathValue("id"))
	if errors.Is(err, posts.ErrNotFound) {
		writeError(w, http.StatusNotFound, "post_not_found", "Post not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "post_publish_failed", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, response{Data: item, Message: "ok"})
}

func (s *Server) handleArchiveAdminPost(w http.ResponseWriter, r *http.Request) {
	item, err := s.posts.ArchiveByID(r.Context(), r.PathValue("id"))
	if errors.Is(err, posts.ErrNotFound) {
		writeError(w, http.StatusNotFound, "post_not_found", "Post not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "post_archive_failed", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, response{Data: item, Message: "ok"})
}

func (s *Server) handleListAdminImpactReports(w http.ResponseWriter, r *http.Request) {
	limit := limitFromRequest(r, 50, 100)
	items, err := s.impact.ListAdmin(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "admin_impact_reports_list_failed", "Failed to list impact reports", nil)
		return
	}

	writeJSON(w, http.StatusOK, response{
		Data:    items,
		Meta:    map[string]any{"count": len(items), "limit": limit},
		Message: "ok",
	})
}

func (s *Server) handleCreateAdminImpactReport(w http.ResponseWriter, r *http.Request) {
	var input impact.AdminInput
	if !decodeJSONBody(w, r, &input) {
		return
	}

	item, err := s.impact.Create(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "impact_report_create_failed", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusCreated, response{Data: item, Message: "ok"})
}

func (s *Server) handleUpdateAdminImpactReport(w http.ResponseWriter, r *http.Request) {
	var input impact.AdminInput
	if !decodeJSONBody(w, r, &input) {
		return
	}

	item, err := s.impact.UpdateByID(r.Context(), r.PathValue("id"), input)
	if errors.Is(err, impact.ErrNotFound) {
		writeError(w, http.StatusNotFound, "impact_report_not_found", "Impact report not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "impact_report_update_failed", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, response{Data: item, Message: "ok"})
}

func (s *Server) handlePublishAdminImpactReport(w http.ResponseWriter, r *http.Request) {
	item, err := s.impact.PublishByID(r.Context(), r.PathValue("id"))
	if errors.Is(err, impact.ErrNotFound) {
		writeError(w, http.StatusNotFound, "impact_report_not_found", "Impact report not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "impact_report_publish_failed", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, response{Data: item, Message: "ok"})
}

func (s *Server) handleArchiveAdminImpactReport(w http.ResponseWriter, r *http.Request) {
	item, err := s.impact.ArchiveByID(r.Context(), r.PathValue("id"))
	if errors.Is(err, impact.ErrNotFound) {
		writeError(w, http.StatusNotFound, "impact_report_not_found", "Impact report not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "impact_report_archive_failed", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, response{Data: item, Message: "ok"})
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON", map[string]string{
			"error": err.Error(),
		})
		return false
	}
	return true
}
