package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pendig/kelompok/internal/audit"
)

type setMaintenanceInput struct {
	Maintenance *bool `json:"maintenance"`
}

func (s *Server) isMaintenanceModeActive(ctx context.Context) bool {
	// 1. Env/Config takes absolute priority for deployment-safety.
	if s.config.MaintenanceMode {
		return true
	}

	// 2. If DB is not initialized (e.g. in tests), fallback to config.
	if s.db == nil {
		return false
	}

	// 3. Query the system_settings table.
	var value string
	err := s.db.QueryRow(ctx, "SELECT value FROM system_settings WHERE key = 'maintenance_mode'").Scan(&value)
	if err != nil {
		// If table doesn't exist or other error, fallback to false.
		return false
	}

	return value == "true"
}

func (s *Server) setMaintenanceMode(ctx context.Context, active bool) error {
	if s.db == nil {
		return nil
	}

	strVal := "false"
	if active {
		strVal = "true"
	}

	_, err := s.db.Exec(ctx, `
		INSERT INTO system_settings (key, value, updated_at)
		VALUES ('maintenance_mode', $1, now())
		ON CONFLICT (key) DO UPDATE
		SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at
	`, strVal)
	return err
}

func (s *Server) maintenanceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always allow health checks, readiness checks, and status endpoint.
		if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" || r.URL.Path == "/api/v1/maintenance" {
			next.ServeHTTP(w, r)
			return
		}

		if s.isMaintenanceModeActive(r.Context()) {
			// Check if the request is from a superadmin or uses the Admin API Key
			item, ok, err := s.adminPrincipal(r)
			if err == nil && ok && (item.AdminKey || (item.User.ID != "" && item.User.Role == "superadmin")) {
				next.ServeHTTP(w, r)
				return
			}

			// Block other requests with 503 Service Unavailable
			writeError(w, http.StatusServiceUnavailable, "maintenance_mode", "The system is currently undergoing maintenance. Please try again later.", nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleGetMaintenance(w http.ResponseWriter, r *http.Request) {
	active := s.isMaintenanceModeActive(r.Context())
	writeJSON(w, http.StatusOK, response{
		Data: map[string]bool{
			"maintenance": active,
		},
		Message: "ok",
	})
}

func (s *Server) handleUpdateMaintenance(w http.ResponseWriter, r *http.Request) {
	item, ok := principalFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "admin_auth_required", "Admin authorization is required", nil)
		return
	}

	// Restrict to superadmin or AdminKey only!
	if !item.AdminKey && item.User.Role != "superadmin" {
		writeError(w, http.StatusForbidden, "admin_forbidden", "Only superadmins are allowed to modify system maintenance mode", nil)
		return
	}

	var input setMaintenanceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON payload", nil)
		return
	}

	if input.Maintenance == nil {
		writeError(w, http.StatusBadRequest, "validation_failed", "Field 'maintenance' is required", nil)
		return
	}

	beforeActive := s.isMaintenanceModeActive(r.Context())
	active := *input.Maintenance

	// If the config env var is overriding, we reject dynamic changes to prevent confusion.
	if s.config.MaintenanceMode && !active {
		writeError(w, http.StatusConflict, "config_override", "Maintenance mode is locked by server configuration (env var)", nil)
		return
	}

	if err := s.setMaintenanceMode(r.Context(), active); err != nil {
		writeError(w, http.StatusInternalServerError, "maintenance_update_failed", "Failed to update maintenance mode", nil)
		return
	}

	// Audit log recording for permission/security compliance.
	actorType := "admin_key"
	var actorUserID *string
	if item.User.ID != "" {
		actorUserID = &item.User.ID
		actorType = "user_session"
	}

	_ = audit.Record(
		r.Context(),
		s.db,
		actorUserID,
		"system",
		nil,
		"update_maintenance_mode",
		map[string]any{"maintenance": beforeActive},
		map[string]any{"maintenance": active},
		map[string]any{
			"actor_type": actorType,
			"method":     "api",
		},
	)

	writeJSON(w, http.StatusOK, response{
		Data: map[string]bool{
			"maintenance": active,
		},
		Message: "Maintenance mode updated successfully",
	})
}
