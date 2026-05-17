package httpapi

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/pendig/kelompok/internal/impact"
	"github.com/pendig/kelompok/internal/organizations"
	"github.com/pendig/kelompok/internal/posts"
)

type publicOrganizationResponse struct {
	Slug          string                                   `json:"slug"`
	Name          string                                   `json:"name"`
	LegalName     string                                   `json:"legal_name,omitempty"`
	Description   string                                   `json:"description,omitempty"`
	History       string                                   `json:"history,omitempty"`
	Country       string                                   `json:"country,omitempty"`
	Region        string                                   `json:"region,omitempty"`
	City          string                                   `json:"city,omitempty"`
	WebsiteURL    string                                   `json:"website_url,omitempty"`
	ClaimStatus   string                                   `json:"claim_status"`
	ProfileData   json.RawMessage                          `json:"profile_data"`
	SDGSData      json.RawMessage                          `json:"sdgs_data"`
	ImpactData    json.RawMessage                          `json:"impact_data"`
	Relationships *publicOrganizationRelationshipsResponse `json:"relationships,omitempty"`
	CreatedAt     time.Time                                `json:"created_at"`
	UpdatedAt     time.Time                                `json:"updated_at"`
}

type publicOrganizationRefResponse struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type publicOrganizationRelationshipResponse struct {
	Organization     publicOrganizationRefResponse `json:"organization"`
	RelationshipType string                        `json:"relationship_type"`
	Label            string                        `json:"label,omitempty"`
	Status           string                        `json:"status"`
}

type publicOrganizationRelationshipsResponse struct {
	Parents  []publicOrganizationRelationshipResponse `json:"parents"`
	Children []publicOrganizationRelationshipResponse `json:"children"`
	Related  []publicOrganizationRelationshipResponse `json:"related"`
}

type publicPostResponse struct {
	Organization publicOrganizationRefResponse `json:"organization"`
	CategorySlug string                        `json:"category_slug,omitempty"`
	Slug         string                        `json:"slug"`
	Title        string                        `json:"title"`
	Summary      string                        `json:"summary,omitempty"`
	Content      string                        `json:"content,omitempty"`
	Status       string                        `json:"status"`
	PostData     json.RawMessage               `json:"post_data"`
	SEOData      json.RawMessage               `json:"seo_data"`
	PublishedAt  *time.Time                    `json:"published_at,omitempty"`
	CreatedAt    time.Time                     `json:"created_at"`
	UpdatedAt    time.Time                     `json:"updated_at"`
}

type publicImpactReportResponse struct {
	Organization      publicOrganizationRefResponse `json:"organization"`
	Title             string                        `json:"title"`
	Summary           string                        `json:"summary,omitempty"`
	ReportPeriodStart *time.Time                    `json:"report_period_start,omitempty"`
	ReportPeriodEnd   *time.Time                    `json:"report_period_end,omitempty"`
	SDGS              json.RawMessage               `json:"sdgs"`
	Metrics           json.RawMessage               `json:"metrics"`
	Status            string                        `json:"status"`
	PublishedAt       *time.Time                    `json:"published_at,omitempty"`
	CreatedAt         time.Time                     `json:"created_at"`
	UpdatedAt         time.Time                     `json:"updated_at"`
}

var organizationProfilePublicKeys = publicKeys(
	"focus",
	"founded_year",
	"languages",
	"members_count",
	"mission",
	"operating_areas",
	"programs",
	"public_contact",
	"social_links",
	"tags",
	"team_size",
	"vision",
)

var organizationSDGSPublicKeys = publicKeys(
	"confidence",
	"goals",
	"primary",
	"secondary",
	"signals",
)

var organizationImpactPublicKeys = publicKeys(
	"beneficiaries",
	"education_sessions",
	"highlights",
	"metrics",
	"neighborhoods",
	"projects",
	"trees_stewarded",
	"volunteers",
)

var postDataPublicKeys = publicKeys(
	"canonical_url",
	"featured",
	"kind",
	"reading_time",
	"tags",
)

var postSEOPublicKeys = publicKeys(
	"description",
	"image_url",
	"title",
)

func publicOrganizations(items []organizations.Organization) []publicOrganizationResponse {
	responses := make([]publicOrganizationResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, publicOrganization(item))
	}
	return responses
}

func publicOrganization(item organizations.Organization) publicOrganizationResponse {
	return publicOrganizationResponse{
		Slug:        item.Slug,
		Name:        item.Name,
		LegalName:   item.LegalName,
		Description: item.Description,
		History:     item.History,
		Country:     item.Country,
		Region:      item.Region,
		City:        item.City,
		WebsiteURL:  item.WebsiteURL,
		ClaimStatus: item.ClaimStatus,
		ProfileData: publicJSONObject(item.ProfileData, "{}", organizationProfilePublicKeys),
		SDGSData:    publicJSONObject(item.SDGSData, "{}", organizationSDGSPublicKeys),
		ImpactData:  publicJSONObject(item.ImpactData, "{}", organizationImpactPublicKeys),
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func publicOrganizationWithRelationships(item organizations.Organization, relationships []organizations.Relationship) publicOrganizationResponse {
	response := publicOrganization(item)
	grouped := publicOrganizationRelationships(item.Slug, relationships)
	response.Relationships = &grouped
	return response
}

func publicOrganizationRelationships(slug string, relationships []organizations.Relationship) publicOrganizationRelationshipsResponse {
	response := publicOrganizationRelationshipsResponse{
		Parents:  []publicOrganizationRelationshipResponse{},
		Children: []publicOrganizationRelationshipResponse{},
		Related:  []publicOrganizationRelationshipResponse{},
	}

	for _, item := range relationships {
		if item.Status != "active" {
			continue
		}
		parentRef := publicOrganizationRefResponse{Slug: item.Parent.Slug, Name: item.Parent.Name}
		childRef := publicOrganizationRefResponse{Slug: item.Child.Slug, Name: item.Child.Name}
		if item.Child.Slug == slug && isHierarchicalRelationship(item.RelationshipType) {
			response.Parents = append(response.Parents, publicRelationshipRef(parentRef, item))
			continue
		}
		if item.Parent.Slug == slug && isHierarchicalRelationship(item.RelationshipType) {
			response.Children = append(response.Children, publicRelationshipRef(childRef, item))
			continue
		}
		if item.Child.Slug == slug {
			response.Related = append(response.Related, publicRelationshipRef(parentRef, item))
			continue
		}
		if item.Parent.Slug == slug {
			response.Related = append(response.Related, publicRelationshipRef(childRef, item))
		}
	}

	return response
}

func publicRelationshipRef(organization publicOrganizationRefResponse, item organizations.Relationship) publicOrganizationRelationshipResponse {
	return publicOrganizationRelationshipResponse{
		Organization:     organization,
		RelationshipType: item.RelationshipType,
		Label:            item.Label,
		Status:           item.Status,
	}
}

func isHierarchicalRelationship(relationshipType string) bool {
	return relationshipType == "structural_parent" || relationshipType == "autonomous_body"
}

func publicPosts(items []posts.Post) []publicPostResponse {
	responses := make([]publicPostResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, publicPost(item))
	}
	return responses
}

func publicPost(item posts.Post) publicPostResponse {
	return publicPostResponse{
		Organization: publicOrganizationRefResponse{
			Slug: item.OrganizationSlug,
			Name: item.OrganizationName,
		},
		CategorySlug: item.CategorySlug,
		Slug:         item.Slug,
		Title:        item.Title,
		Summary:      item.Summary,
		Content:      item.Content,
		Status:       item.Status,
		PostData:     publicJSONObject(item.PostData, "{}", postDataPublicKeys),
		SEOData:      publicJSONObject(item.SEOData, "{}", postSEOPublicKeys),
		PublishedAt:  item.PublishedAt,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
}

func publicImpactReports(items []impact.Report) []publicImpactReportResponse {
	responses := make([]publicImpactReportResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, publicImpactReport(item))
	}
	return responses
}

func publicImpactReport(item impact.Report) publicImpactReportResponse {
	return publicImpactReportResponse{
		Organization: publicOrganizationRefResponse{
			Slug: item.OrganizationSlug,
			Name: item.OrganizationName,
		},
		Title:             item.Title,
		Summary:           item.Summary,
		ReportPeriodStart: item.ReportPeriodStart,
		ReportPeriodEnd:   item.ReportPeriodEnd,
		SDGS:              publicJSON(item.SDGS, "[]"),
		Metrics:           publicJSON(item.Metrics, "{}"),
		Status:            item.Status,
		PublishedAt:       item.PublishedAt,
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
	}
}

func publicKeys(keys ...string) map[string]struct{} {
	allowed := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		allowed[key] = struct{}{}
	}
	return allowed
}

func publicJSONObject(raw json.RawMessage, fallback string, allowed map[string]struct{}) json.RawMessage {
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return json.RawMessage(fallback)
	}

	filtered := make(map[string]any, len(data))
	for key, value := range data {
		if _, ok := allowed[key]; !ok {
			continue
		}
		if isSensitivePublicKey(key, "") {
			continue
		}
		cleaned, ok := cleanPublicJSONValue(value, key)
		if ok {
			filtered[key] = cleaned
		}
	}

	return marshalPublicJSON(filtered, fallback)
}

func publicJSON(raw json.RawMessage, fallback string) json.RawMessage {
	var data any
	if err := json.Unmarshal(raw, &data); err != nil {
		return json.RawMessage(fallback)
	}

	cleaned, ok := cleanPublicJSONValue(data, "")
	if !ok {
		return json.RawMessage(fallback)
	}

	return marshalPublicJSON(cleaned, fallback)
}

func cleanPublicJSONValue(value any, parentKey string) (any, bool) {
	switch typed := value.(type) {
	case nil:
		return nil, true
	case string, float64, bool:
		return typed, true
	case []any:
		cleaned := make([]any, 0, len(typed))
		for _, item := range typed {
			cleanedItem, ok := cleanPublicJSONValue(item, parentKey)
			if ok {
				cleaned = append(cleaned, cleanedItem)
			}
		}
		return cleaned, true
	case map[string]any:
		cleaned := make(map[string]any, len(typed))
		for key, value := range typed {
			if isSensitivePublicKey(key, parentKey) {
				continue
			}
			cleanedValue, ok := cleanPublicJSONValue(value, key)
			if ok {
				cleaned[key] = cleanedValue
			}
		}
		return cleaned, true
	default:
		return nil, false
	}
}

func marshalPublicJSON(value any, fallback string) json.RawMessage {
	encoded, err := json.Marshal(value)
	if err != nil {
		return json.RawMessage(fallback)
	}
	return json.RawMessage(encoded)
}

func isSensitivePublicKey(key, parentKey string) bool {
	normalized := strings.NewReplacer("_", "", "-", "", " ", "").Replace(strings.ToLower(key))
	parent := strings.NewReplacer("_", "", "-", "", " ", "").Replace(strings.ToLower(parentKey))
	if parent == "publiccontact" && (normalized == "email" || normalized == "phone") {
		return false
	}

	sensitiveFragments := []string{
		"apikey",
		"auth",
		"claim",
		"cookie",
		"credential",
		"email",
		"evidence",
		"internal",
		"password",
		"phone",
		"private",
		"raw",
		"secret",
		"source",
		"token",
	}

	for _, fragment := range sensitiveFragments {
		if strings.Contains(normalized, fragment) {
			return true
		}
	}

	return false
}
