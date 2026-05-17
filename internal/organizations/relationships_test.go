package organizations

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestNormalizeRelationshipInputRejectsSelfLink(t *testing.T) {
	_, err := normalizeRelationshipInput(RelationshipInput{
		ParentOrganizationSlug: "ipm",
		ChildOrganizationSlug:  "IPM",
		RelationshipType:       "structural_parent",
	})
	if !errors.Is(err, ErrRelationshipSelfLink) {
		t.Fatalf("expected self-link error, got %v", err)
	}
}

func TestNormalizeRelationshipInputDefaultsAndValidates(t *testing.T) {
	item, err := normalizeRelationshipInput(RelationshipInput{
		ParentOrganizationSlug: " Muhammadiyah ",
		ChildOrganizationSlug:  "Ikatan Pelajar Muhammadiyah",
	})
	if err != nil {
		t.Fatalf("normalize relationship input: %v", err)
	}

	if item.ParentOrganizationSlug != "muhammadiyah" {
		t.Fatalf("parent slug = %q", item.ParentOrganizationSlug)
	}
	if item.ChildOrganizationSlug != "ikatan-pelajar-muhammadiyah" {
		t.Fatalf("child slug = %q", item.ChildOrganizationSlug)
	}
	if item.RelationshipType != "related" {
		t.Fatalf("relationship type = %q", item.RelationshipType)
	}
	if item.Status != "active" {
		t.Fatalf("status = %q", item.Status)
	}
}

func TestRelationshipWriteErrorMapsDuplicateConstraint(t *testing.T) {
	err := relationshipWriteError(&pgconn.PgError{
		Code:           "23505",
		ConstraintName: "organization_relationships_unique_parent_child_type",
	})

	if !errors.Is(err, ErrRelationshipDuplicate) {
		t.Fatalf("expected duplicate error, got %v", err)
	}
}
