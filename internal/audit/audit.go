package audit

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Record(ctx context.Context, pool *pgxpool.Pool, actorUserID any, entityType string, entityID any, action string, beforeData any, afterData any, metadata any) error {
	if pool == nil {
		return nil
	}

	_, err := pool.Exec(ctx, `
		INSERT INTO audit_logs (
			actor_user_id,
			entity_type,
			entity_id,
			action,
			before_data,
			after_data,
			metadata
		)
		VALUES ($1, $2, $3::uuid, $4, $5::jsonb, $6::jsonb, $7::jsonb)
	`,
		normalizeUUID(actorUserID),
		entityType,
		normalizeUUID(entityID),
		action,
		normalizeJSON(beforeData),
		normalizeJSON(afterData),
		normalizeJSON(metadata),
	)
	return err
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
