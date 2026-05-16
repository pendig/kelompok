package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pendig/kelompok/internal/jsonvalue"
)

type Log struct {
	ID             string          `json:"id"`
	ActorUserID    *string         `json:"actor_user_id,omitempty"`
	OrganizationID *string         `json:"organization_id,omitempty"`
	EntityType     string          `json:"entity_type"`
	EntityID       *string         `json:"entity_id,omitempty"`
	Action         string          `json:"action"`
	BeforeData     json.RawMessage `json:"before_data,omitempty"`
	AfterData      json.RawMessage `json:"after_data,omitempty"`
	Metadata       json.RawMessage `json:"metadata"`
	CreatedAt      time.Time       `json:"created_at"`
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func Record(ctx context.Context, pool *pgxpool.Pool, actorUserID any, entityType string, entityID any, action string, beforeData any, afterData any, metadata any) error {
	if pool == nil {
		return nil
	}

	_, err := pool.Exec(ctx, `
		INSERT INTO audit_logs (
			actor_user_id,
			organization_id,
			entity_type,
			entity_id,
			action,
			before_data,
			after_data,
			metadata
		)
		VALUES ($1, $2::uuid, $3, $4::uuid, $5, $6::jsonb, $7::jsonb, $8::jsonb)
	`,
		normalizeUUID(actorUserID),
		organizationID(entityType, entityID, metadata),
		entityType,
		normalizeUUID(entityID),
		action,
		normalizeJSON(beforeData),
		normalizeJSON(afterData),
		normalizeJSON(metadata),
	)
	return err
}

func (r *Repository) ListByOrganizationSlug(ctx context.Context, organizationSlug string, limit int) ([]Log, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			al.id::text,
			al.actor_user_id::text,
			al.organization_id::text,
			al.entity_type,
			al.entity_id::text,
			al.action,
			COALESCE(al.before_data::text, ''),
			COALESCE(al.after_data::text, ''),
			COALESCE(al.metadata::text, '{}'),
			al.created_at
		FROM audit_logs al
		JOIN organizations o ON o.id = al.organization_id
		WHERE o.slug = $1
		ORDER BY al.created_at DESC
		LIMIT $2
	`, organizationSlug, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Log, 0, limit)
	for rows.Next() {
		item, err := scanLog(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanLog(row interface{ Scan(dest ...any) error }) (Log, error) {
	var item Log
	var actorUserID sql.NullString
	var organizationID sql.NullString
	var entityID sql.NullString
	var beforeData string
	var afterData string
	var metadata string
	if err := row.Scan(
		&item.ID,
		&actorUserID,
		&organizationID,
		&item.EntityType,
		&entityID,
		&item.Action,
		&beforeData,
		&afterData,
		&metadata,
		&item.CreatedAt,
	); err != nil {
		return Log{}, err
	}
	if actorUserID.Valid {
		value := actorUserID.String
		item.ActorUserID = &value
	}
	if organizationID.Valid {
		value := organizationID.String
		item.OrganizationID = &value
	}
	if entityID.Valid {
		value := entityID.String
		item.EntityID = &value
	}
	item.BeforeData = jsonvalue.Raw(beforeData, "")
	item.AfterData = jsonvalue.Raw(afterData, "")
	item.Metadata = jsonvalue.Raw(metadata, "{}")
	return item, nil
}

func organizationID(entityType string, entityID any, metadata any) any {
	if entityType == "organization" {
		return normalizeUUID(entityID)
	}

	if value := metadataValue(metadata, "organization_id"); value != "" {
		return value
	}

	return nil
}

func metadataValue(metadata any, key string) string {
	switch typed := metadata.(type) {
	case nil:
		return ""
	case map[string]any:
		if value, ok := typed[key].(string); ok {
			return value
		}
	case map[string]string:
		return typed[key]
	}
	return ""
}

func normalizeUUID(value any) any {
	switch typed := value.(type) {
	case nil:
		return nil
	case string:
		if typed == "" {
			return nil
		}
		return typed
	default:
		return value
	}
}

func normalizeJSON(value any) any {
	if value == nil {
		return nil
	}

	switch typed := value.(type) {
	case json.RawMessage:
		if len(typed) == 0 {
			return nil
		}
		return string(typed)
	case []byte:
		if len(typed) == 0 {
			return nil
		}
		return string(typed)
	case string:
		if typed == "" {
			return nil
		}
		return typed
	default:
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil
		}
		return string(encoded)
	}
}
