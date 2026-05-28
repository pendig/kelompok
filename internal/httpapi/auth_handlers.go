package httpapi

import (
	"errors"
	"net/http"

	"github.com/pendig/kelompok/internal/auth"
)

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var input auth.RegisterInput
	if !decodeJSONBody(w, r, &input) {
		return
	}

	user, err := s.auth.Register(r.Context(), input)
	if errors.Is(err, auth.ErrUserExists) {
		writeError(w, http.StatusConflict, "user_exists", "User already exists", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "register_failed", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusCreated, response{Data: user, Message: "ok"})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var input auth.LoginInput
	if !decodeJSONBody(w, r, &input) {
		return
	}

	session, err := s.auth.Login(r.Context(), input)
	if errors.Is(err, auth.ErrInvalidCredentials) {
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "login_failed", err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, response{Data: session, Message: "ok"})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	item, _ := principalFromContext(r)
	if err := s.auth.Logout(r.Context(), item.Token); err != nil {
		writeError(w, http.StatusUnauthorized, "logout_failed", "Session is invalid or expired", nil)
		return
	}

	writeJSON(w, http.StatusOK, response{Data: map[string]bool{"revoked": true}, Message: "ok"})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	item, _ := principalFromContext(r)
	roles, err := s.auth.RolesByUserID(r.Context(), item.User.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "roles_lookup_failed", "Failed to load organization roles", nil)
		return
	}

	writeJSON(w, http.StatusOK, response{
		Data: map[string]any{
			"user":               item.User,
			"organization_roles": roles,
		},
		Message: "ok",
	})
}

func (s *Server) handleUpdateMe(w http.ResponseWriter, r *http.Request) {
	item, _ := principalFromContext(r)

	var input auth.UpdateProfileInput
	if !decodeJSONBody(w, r, &input) {
		return
	}

	user, err := s.auth.UpdateProfile(r.Context(), item.User.ID, input)
	if errors.Is(err, auth.ErrNotFound) {
		writeError(w, http.StatusUnauthorized, "session_invalid", "Session is invalid or expired", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "profile_update_failed", err.Error(), nil)
		return
	}

	roles, err := s.auth.RolesByUserID(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "roles_lookup_failed", "Failed to load organization roles", nil)
		return
	}

	writeJSON(w, http.StatusOK, response{
		Data: map[string]any{
			"user":               user,
			"organization_roles": roles,
		},
		Message: "ok",
	})
}
