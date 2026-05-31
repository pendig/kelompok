package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pendig/kelompok/internal/config"
)

func newMaintenanceTestServer(maintenanceMode bool, adminKey string) *Server {
	return New(config.Config{
		APIAddr:         ":0",
		MaintenanceMode: maintenanceMode,
		AdminAPIKey:     adminKey,
	}, nil)
}

func TestMaintenanceModeMiddlewareInterception(t *testing.T) {
	server := newMaintenanceTestServer(true, "secret-key")

	t.Run("public endpoint / is blocked with 503 when maintenance active", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		recorder := httptest.NewRecorder()

		server.Handler().ServeHTTP(recorder, request)

		if recorder.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected status 503 Service Unavailable, got %d", recorder.Code)
		}

		body := recorder.Body.String()
		if !strings.Contains(body, `"code":"maintenance_mode"`) {
			t.Fatalf("expected maintenance_mode error code in response, got %s", body)
		}
	})

	t.Run("healthz is allowed under maintenance mode", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		recorder := httptest.NewRecorder()

		server.Handler().ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected healthz to return 200 OK, got %d", recorder.Code)
		}
	})

	t.Run("readyz is allowed under maintenance mode", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		recorder := httptest.NewRecorder()

		server.Handler().ServeHTTP(recorder, request)

		// It returns 503 database_not_ready instead of maintenance_mode because db is nil,
		// which means it bypassed the maintenance middleware and reached the handler!
		if recorder.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected readyz to bypass maintenance and hit DB handler (503 database_not_ready), got %d", recorder.Code)
		}
		if !strings.Contains(recorder.Body.String(), "database_not_ready") {
			t.Fatalf("expected database_not_ready error, got %s", recorder.Body.String())
		}
	})

	t.Run("get maintenance status is allowed under maintenance mode", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/maintenance", nil)
		recorder := httptest.NewRecorder()

		server.Handler().ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected maintenance status to return 200 OK, got %d", recorder.Code)
		}

		body := recorder.Body.String()
		if !strings.Contains(body, `"maintenance":true`) {
			t.Fatalf("expected maintenance status to be true, got %s", body)
		}
	})

	t.Run("admin key request is allowed to bypass maintenance mode", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.Header.Set("X-Kelompok-Admin-Key", "secret-key")
		recorder := httptest.NewRecorder()

		server.Handler().ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected admin request to return 200 OK, got %d", recorder.Code)
		}

		body := recorder.Body.String()
		if !strings.Contains(body, "kelompok-api") {
			t.Fatalf("expected service info in response, got %s", body)
		}
	})
}

func TestMaintenanceModeOff(t *testing.T) {
	server := newMaintenanceTestServer(false, "secret-key")

	t.Run("public endpoint / returns 200 OK when maintenance inactive", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		recorder := httptest.NewRecorder()

		server.Handler().ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", recorder.Code)
		}
	})
}
