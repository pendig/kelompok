package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

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
	if item, ok := principalFromContext(r); ok {
		if !item.AdminKey && item.User.Role != "superadmin" {
			writeError(w, http.StatusForbidden, "admin_org_scope_required", "User sessions must use organization-scoped endpoints", nil)
			return
		}
		if item.AdminKey && s.adminScopeConfigured() {
			writeError(w, http.StatusForbidden, "admin_org_scope_required", "Scoped admin keys must use organization-scoped endpoints", nil)
			return
		}
	}

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
	if !s.ensureAdminOrganizationSlugForRequest(w, r, input.Slug) {
		return
	}

	item, err := s.organizations.Create(r.Context(), input)
	if errors.Is(err, organizations.ErrSlugTaken) {
		writeError(w, http.StatusConflict, "organization_slug_taken", "Organization slug is already used", nil)
		return
	}
	if writeOrganizationValidationError(w, err) {
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "organization_create_failed", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusCreated, response{Data: item, Message: "ok"})
}

func (s *Server) handleGetAdminOrganization(w http.ResponseWriter, r *http.Request) {
	if !s.ensureAdminOrganizationSlugForRequest(w, r, r.PathValue("slug")) {
		return
	}

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
	if !s.ensureAdminOrganizationSlugForRequest(w, r, r.PathValue("slug")) {
		return
	}
	if input.Slug != "" && !s.ensureAdminOrganizationSlugForRequest(w, r, input.Slug) {
		return
	}

	item, err := s.organizations.UpdateBySlug(r.Context(), r.PathValue("slug"), input)
	if errors.Is(err, organizations.ErrNotFound) {
		writeError(w, http.StatusNotFound, "organization_not_found", "Organization not found", nil)
		return
	}
	if errors.Is(err, organizations.ErrSlugTaken) {
		writeError(w, http.StatusConflict, "organization_slug_taken", "Organization slug is already used", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "organization_update_failed", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, response{Data: item, Message: "ok"})
}

func (s *Server) handleListOrganizationRelationships(w http.ResponseWriter, r *http.Request) {
	if !s.ensureAdminOrganizationSlugForRequest(w, r, r.PathValue("slug")) {
		return
	}
	if !s.ensureOrganization(w, r, r.PathValue("slug")) {
		return
	}

	limit := limitFromRequest(r, 50, 100)
	items, err := s.organizations.ListRelationshipsByOrganizationSlug(r.Context(), r.PathValue("slug"), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "relationships_list_failed", "Failed to list organization relationships", nil)
		return
	}

	writeJSON(w, http.StatusOK, response{
		Data:    items,
		Meta:    map[string]any{"count": len(items), "limit": limit},
		Message: "ok",
	})
}

func (s *Server) handleCreateOrganizationRelationship(w http.ResponseWriter, r *http.Request) {
	var input organizations.RelationshipInput
	if !decodeJSONBody(w, r, &input) {
		return
	}
	if !s.ensureAdminAnyOrganizationSlugForRequest(w, r, input.ParentOrganizationSlug, input.ChildOrganizationSlug) {
		return
	}

	item, err := s.organizations.CreateRelationship(r.Context(), input, relationshipAuditActorFromRequest(r))
	if errors.Is(err, organizations.ErrRelationshipOrganizationNotFound) {
		writeError(w, http.StatusNotFound, "relationship_organization_not_found", "Parent or child organization not found", nil)
		return
	}
	if errors.Is(err, organizations.ErrRelationshipDuplicate) {
		writeError(w, http.StatusConflict, "relationship_duplicate", "Organization relationship already exists", nil)
		return
	}
	if errors.Is(err, organizations.ErrRelationshipSelfLink) {
		writeError(w, http.StatusBadRequest, "relationship_self_link", "Organization cannot be related to itself", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "relationship_create_failed", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusCreated, response{Data: item, Message: "ok"})
}

func (s *Server) handleCreateRelatedOrganization(w http.ResponseWriter, r *http.Request) {
	parentSlug := r.PathValue("slug")
	if !s.ensureAdminOrganizationSlugForRequest(w, r, parentSlug) {
		return
	}

	var input organizations.RelatedOrganizationInput
	if !decodeJSONBody(w, r, &input) {
		return
	}

	item, err := s.organizations.CreateRelatedOrganization(r.Context(), parentSlug, input, relationshipAuditActorFromRequest(r))
	if errors.Is(err, organizations.ErrNotFound) {
		writeError(w, http.StatusNotFound, "organization_not_found", "Parent organization not found", nil)
		return
	}
	if errors.Is(err, organizations.ErrRelationshipDuplicate) {
		writeError(w, http.StatusConflict, "relationship_duplicate", "Organization relationship already exists", nil)
		return
	}
	if errors.Is(err, organizations.ErrRelationshipSelfLink) {
		writeError(w, http.StatusBadRequest, "relationship_self_link", "Organization cannot be related to itself", nil)
		return
	}
	if errors.Is(err, organizations.ErrSlugTaken) {
		writeError(w, http.StatusConflict, "organization_slug_taken", "Organization slug is already used", nil)
		return
	}
	if writeOrganizationValidationError(w, err) {
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "related_organization_create_failed", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusCreated, response{Data: item, Message: "ok"})
}

func (s *Server) handleUpdateOrganizationRelationship(w http.ResponseWriter, r *http.Request) {
	existing, err := s.organizations.FindRelationshipByID(r.Context(), r.PathValue("id"))
	if errors.Is(err, organizations.ErrRelationshipNotFound) {
		writeError(w, http.StatusNotFound, "relationship_not_found", "Organization relationship not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "relationship_lookup_failed", err.Error(), nil)
		return
	}
	if !s.ensureAdminAnyOrganizationSlugForRequest(w, r, existing.Parent.Slug, existing.Child.Slug) {
		return
	}

	input, ok := decodeRelationshipPatchBody(w, r)
	if !ok {
		return
	}
	parentSlug := strings.TrimSpace(input.ParentOrganizationSlug)
	if parentSlug == "" {
		parentSlug = existing.Parent.Slug
	}
	childSlug := strings.TrimSpace(input.ChildOrganizationSlug)
	if childSlug == "" {
		childSlug = existing.Child.Slug
	}
	if !s.ensureAdminAnyOrganizationSlugForRequest(w, r, parentSlug, childSlug) {
		return
	}

	item, err := s.organizations.UpdateRelationshipByID(r.Context(), r.PathValue("id"), input, relationshipAuditActorFromRequest(r))
	if errors.Is(err, organizations.ErrRelationshipOrganizationNotFound) {
		writeError(w, http.StatusNotFound, "relationship_organization_not_found", "Parent or child organization not found", nil)
		return
	}
	if errors.Is(err, organizations.ErrRelationshipDuplicate) {
		writeError(w, http.StatusConflict, "relationship_duplicate", "Organization relationship already exists", nil)
		return
	}
	if errors.Is(err, organizations.ErrRelationshipSelfLink) {
		writeError(w, http.StatusBadRequest, "relationship_self_link", "Organization cannot be related to itself", nil)
		return
	}
	if errors.Is(err, organizations.ErrRelationshipNotFound) {
		writeError(w, http.StatusNotFound, "relationship_not_found", "Organization relationship not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "relationship_update_failed", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, response{Data: item, Message: "ok"})
}

func (s *Server) handleDeleteOrganizationRelationship(w http.ResponseWriter, r *http.Request) {
	existing, err := s.organizations.FindRelationshipByID(r.Context(), r.PathValue("id"))
	if errors.Is(err, organizations.ErrRelationshipNotFound) {
		writeError(w, http.StatusNotFound, "relationship_not_found", "Organization relationship not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "relationship_lookup_failed", err.Error(), nil)
		return
	}
	if !s.ensureAdminAnyOrganizationSlugForRequest(w, r, existing.Parent.Slug, existing.Child.Slug) {
		return
	}

	item, err := s.organizations.DeleteRelationshipByID(r.Context(), r.PathValue("id"), relationshipAuditActorFromRequest(r))
	if errors.Is(err, organizations.ErrRelationshipNotFound) {
		writeError(w, http.StatusNotFound, "relationship_not_found", "Organization relationship not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "relationship_delete_failed", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, response{Data: item, Message: "ok"})
}

func (s *Server) handleListOrganizationClaims(w http.ResponseWriter, r *http.Request) {
	if !s.ensureAdminOrganizationSlugForRequest(w, r, r.PathValue("slug")) {
		return
	}
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

func (s *Server) handleListDelegatedOrganizationClaims(w http.ResponseWriter, r *http.Request) {
	if !s.ensureAdminOrganizationSlugForRequest(w, r, r.PathValue("slug")) {
		return
	}
	if !s.ensureOrganization(w, r, r.PathValue("slug")) {
		return
	}

	limit := limitFromRequest(r, 20, 100)
	items, err := s.organizations.ListDelegatedClaimsByReviewerOrganizationSlug(r.Context(), r.PathValue("slug"), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "delegated_claims_list_failed", "Failed to list delegated claim requests", nil)
		return
	}

	writeJSON(w, http.StatusOK, response{
		Data:    items,
		Meta:    map[string]any{"count": len(items), "limit": limit},
		Message: "ok",
	})
}

func (s *Server) handleApproveOrganizationClaim(w http.ResponseWriter, r *http.Request) {
	s.handleReviewOrganizationClaim(w, r, "approve")
}

func (s *Server) handleRejectOrganizationClaim(w http.ResponseWriter, r *http.Request) {
	s.handleReviewOrganizationClaim(w, r, "reject")
}

func (s *Server) handleReviewOrganizationClaim(w http.ResponseWriter, r *http.Request, decision string) {
	claim, organizationSlug, err := s.organizations.FindClaimByID(r.Context(), r.PathValue("id"))
	if errors.Is(err, organizations.ErrClaimNotFound) {
		writeError(w, http.StatusNotFound, "claim_not_found", "Claim request not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "claim_lookup_failed", err.Error(), nil)
		return
	}
	if !s.ensureAdminCanReviewClaimForRequest(w, r, organizationSlug) {
		return
	}

	item, principal := claim, principal{}
	if current, ok := principalFromContext(r); ok {
		principal = current
	}
	reviewerUserID := principal.User.ID
	if decision == "approve" {
		item, err = s.organizations.ApproveClaim(r.Context(), claim.ID, reviewerUserID)
	} else {
		item, err = s.organizations.RejectClaim(r.Context(), claim.ID, reviewerUserID)
	}
	if errors.Is(err, organizations.ErrClaimNotFound) {
		writeError(w, http.StatusNotFound, "claim_not_found", "Claim request not found", nil)
		return
	}
	if errors.Is(err, organizations.ErrClaimNotPending) {
		writeError(w, http.StatusConflict, "claim_not_pending", "Only pending claim requests can be reviewed", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "claim_review_failed", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, response{Data: item, Message: "ok"})
}

func (s *Server) handleListOrganizationAuditLogs(w http.ResponseWriter, r *http.Request) {
	if !s.ensureAdminOrganizationSlugForRequest(w, r, r.PathValue("slug")) {
		return
	}
	if !s.ensureOrganization(w, r, r.PathValue("slug")) {
		return
	}

	limit := limitFromRequest(r, 30, 100)
	items, err := s.audit.ListByOrganizationSlug(r.Context(), r.PathValue("slug"), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "audit_logs_list_failed", "Failed to list audit logs", nil)
		return
	}

	writeJSON(w, http.StatusOK, response{
		Data:    items,
		Meta:    map[string]any{"count": len(items), "limit": limit},
		Message: "ok",
	})
}

func (s *Server) handleListOrganizationMembers(w http.ResponseWriter, r *http.Request) {
	if !s.ensureAdminOrganizationSlugForRequest(w, r, r.PathValue("slug")) {
		return
	}
	if !s.ensureOrganization(w, r, r.PathValue("slug")) {
		return
	}

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
	if !s.ensureAdminOrganizationSlugForRequest(w, r, r.PathValue("slug")) {
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

	existing, err := s.members.FindByID(r.Context(), r.PathValue("id"))
	if errors.Is(err, members.ErrNotFound) {
		writeError(w, http.StatusNotFound, "member_not_found", "Member not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "member_lookup_failed", err.Error(), nil)
		return
	}
	if !s.ensureAdminOrganizationSlugForRequest(w, r, existing.OrganizationSlug) {
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
	existing, err := s.members.FindByID(r.Context(), r.PathValue("id"))
	if errors.Is(err, members.ErrNotFound) {
		writeError(w, http.StatusNotFound, "member_not_found", "Member not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "member_lookup_failed", err.Error(), nil)
		return
	}
	if !s.ensureAdminOrganizationSlugForRequest(w, r, existing.OrganizationSlug) {
		return
	}

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
	organizationSlug := r.URL.Query().Get("organization_slug")
	if !s.ensureAdminListScope(w, r, organizationSlug) {
		return
	}

	var (
		items []posts.Post
		err   error
	)
	if organizationSlug != "" {
		items, err = s.posts.ListAdminByOrganizationSlug(r.Context(), organizationSlug, limit)
	} else {
		items, err = s.posts.ListAdmin(r.Context(), limit)
	}
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
	if !s.ensureAdminOrganizationSlugForRequest(w, r, input.OrganizationSlug) {
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

	existing, err := s.posts.FindByID(r.Context(), r.PathValue("id"))
	if errors.Is(err, posts.ErrNotFound) {
		writeError(w, http.StatusNotFound, "post_not_found", "Post not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "post_lookup_failed", err.Error(), nil)
		return
	}
	if !s.ensureAdminOrganizationSlugForRequest(w, r, existing.OrganizationSlug) || !s.ensureAdminOrganizationSlugForRequest(w, r, input.OrganizationSlug) {
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
	existing, err := s.posts.FindByID(r.Context(), r.PathValue("id"))
	if errors.Is(err, posts.ErrNotFound) {
		writeError(w, http.StatusNotFound, "post_not_found", "Post not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "post_lookup_failed", err.Error(), nil)
		return
	}
	if !s.ensureAdminOrganizationSlugForRequest(w, r, existing.OrganizationSlug) {
		return
	}

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
	existing, err := s.posts.FindByID(r.Context(), r.PathValue("id"))
	if errors.Is(err, posts.ErrNotFound) {
		writeError(w, http.StatusNotFound, "post_not_found", "Post not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "post_lookup_failed", err.Error(), nil)
		return
	}
	if !s.ensureAdminOrganizationSlugForRequest(w, r, existing.OrganizationSlug) {
		return
	}

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
	organizationSlug := r.URL.Query().Get("organization_slug")
	if !s.ensureAdminListScope(w, r, organizationSlug) {
		return
	}

	var (
		items []impact.Report
		err   error
	)
	if organizationSlug != "" {
		items, err = s.impact.ListAdminByOrganizationSlug(r.Context(), organizationSlug, limit)
	} else {
		items, err = s.impact.ListAdmin(r.Context(), limit)
	}
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
	if !s.ensureAdminOrganizationSlugForRequest(w, r, input.OrganizationSlug) {
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

	existing, err := s.impact.FindByID(r.Context(), r.PathValue("id"))
	if errors.Is(err, impact.ErrNotFound) {
		writeError(w, http.StatusNotFound, "impact_report_not_found", "Impact report not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "impact_report_lookup_failed", err.Error(), nil)
		return
	}
	if !s.ensureAdminOrganizationSlugForRequest(w, r, existing.OrganizationSlug) || !s.ensureAdminOrganizationSlugForRequest(w, r, input.OrganizationSlug) {
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
	existing, err := s.impact.FindByID(r.Context(), r.PathValue("id"))
	if errors.Is(err, impact.ErrNotFound) {
		writeError(w, http.StatusNotFound, "impact_report_not_found", "Impact report not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "impact_report_lookup_failed", err.Error(), nil)
		return
	}
	if !s.ensureAdminOrganizationSlugForRequest(w, r, existing.OrganizationSlug) {
		return
	}

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
	existing, err := s.impact.FindByID(r.Context(), r.PathValue("id"))
	if errors.Is(err, impact.ErrNotFound) {
		writeError(w, http.StatusNotFound, "impact_report_not_found", "Impact report not found", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "impact_report_lookup_failed", err.Error(), nil)
		return
	}
	if !s.ensureAdminOrganizationSlugForRequest(w, r, existing.OrganizationSlug) {
		return
	}

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

func (s *Server) ensureAdminListScope(w http.ResponseWriter, r *http.Request, organizationSlug string) bool {
	if organizationSlug != "" {
		return s.ensureAdminOrganizationSlugForRequest(w, r, organizationSlug)
	}
	if item, ok := principalFromContext(r); ok {
		if !item.AdminKey && item.User.Role == "superadmin" {
			return true
		}
		if !item.AdminKey {
			writeError(w, http.StatusForbidden, "admin_org_scope_required", "User sessions must include organization_slug", nil)
			return false
		}
	}
	if !s.adminScopeConfigured() {
		return true
	}

	writeError(w, http.StatusForbidden, "admin_org_scope_required", "Scoped admin keys must include organization_slug", nil)
	return false
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

func writeOrganizationValidationError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}

	switch {
	case errors.Is(err, organizations.ErrOrganizationNameRequired):
		writeError(w, http.StatusBadRequest, "organization_name_required", "Organization name is required", nil)
	case errors.Is(err, organizations.ErrOrganizationSlugRequired):
		writeError(w, http.StatusBadRequest, "organization_slug_required", "Organization slug is required", nil)
	case errors.Is(err, organizations.ErrOrganizationClaimStatusInvalid):
		writeError(w, http.StatusBadRequest, "organization_claim_status_invalid", "Organization claim_status must be unclaimed, pending, claimed, or rejected", nil)
	case errors.Is(err, organizations.ErrOrganizationOfficialEmailInvalid):
		writeError(w, http.StatusBadRequest, "organization_official_email_invalid", "Organization official_email must be a valid email address", nil)
	case errors.Is(err, organizations.ErrOrganizationJSONInvalid):
		writeError(w, http.StatusBadRequest, "organization_json_invalid", "Organization JSON fields must contain valid JSON values", nil)
	default:
		return false
	}
	return true
}

func decodeRelationshipPatchBody(w http.ResponseWriter, r *http.Request) (organizations.RelationshipInput, bool) {
	defer r.Body.Close()

	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&raw); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON", map[string]string{
			"error": err.Error(),
		})
		return organizations.RelationshipInput{}, false
	}

	for key := range raw {
		switch key {
		case "parent_organization_slug",
			"child_organization_slug",
			"relationship_type",
			"label",
			"status",
			"started_at",
			"ended_at",
			"metadata":
		default:
			writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON", map[string]string{
				"error": "json: unknown field " + key,
			})
			return organizations.RelationshipInput{}, false
		}
	}

	encoded, err := json.Marshal(raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON", map[string]string{
			"error": err.Error(),
		})
		return organizations.RelationshipInput{}, false
	}

	var input organizations.RelationshipInput
	if err := json.Unmarshal(encoded, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON", map[string]string{
			"error": err.Error(),
		})
		return organizations.RelationshipInput{}, false
	}
	if value, ok := raw["started_at"]; ok && strings.TrimSpace(string(value)) == "null" {
		input.ClearStartedAt = true
	}
	if value, ok := raw["ended_at"]; ok && strings.TrimSpace(string(value)) == "null" {
		input.ClearEndedAt = true
	}
	return input, true
}
