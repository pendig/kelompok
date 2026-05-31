package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

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

	claims, err := s.auth.ClaimsByUserID(r.Context(), item.User.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "claims_lookup_failed", "Failed to load organization claims", nil)
		return
	}

	writeJSON(w, http.StatusOK, response{
		Data: map[string]any{
			"user":                item.User,
			"organization_roles":  roles,
			"organization_claims": claims,
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
	if errors.Is(err, auth.ErrProfileNameRequired) {
		writeError(w, http.StatusBadRequest, "profile_name_required", "Name is required", nil)
		return
	}
	if errors.Is(err, auth.ErrProfileNameTooLong) {
		writeError(w, http.StatusBadRequest, "profile_name_too_long", "Name must be at most 120 characters", nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "profile_update_failed", "Failed to update profile", nil)
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

type GoogleLoginInput struct {
	Code        string `json:"code"`
	RedirectURI string `json:"redirect_uri"`
}

func (s *Server) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	var input GoogleLoginInput
	if !decodeJSONBody(w, r, &input) {
		return
	}

	code := strings.TrimSpace(input.Code)
	redirectURI := strings.TrimSpace(input.RedirectURI)
	if code == "" {
		writeError(w, http.StatusBadRequest, "google_login_failed", "Authorization code is required", nil)
		return
	}
	if redirectURI == "" {
		writeError(w, http.StatusBadRequest, "google_login_failed", "Redirect URI is required", nil)
		return
	}

	// Exchange code for Google Access Token
	accessToken, err := s.exchangeGoogleCode(r.Context(), code, redirectURI)
	if err != nil {
		writeError(w, http.StatusBadRequest, "google_login_failed", fmt.Sprintf("Failed to exchange code: %v", err), nil)
		return
	}

	// Fetch user details from Google
	googleUser, err := s.fetchGoogleUserInfo(r.Context(), accessToken)
	if err != nil {
		writeError(w, http.StatusBadRequest, "google_login_failed", fmt.Sprintf("Failed to get Google user info: %v", err), nil)
		return
	}

	if googleUser.Email == "" {
		writeError(w, http.StatusBadRequest, "google_login_failed", "Email not provided by Google", nil)
		return
	}

	// Create or resolve local user
	user, err := s.auth.FindOrCreateUserByEmail(r.Context(), googleUser.Email, googleUser.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "google_login_failed", fmt.Sprintf("Failed to resolve user: %v", err), nil)
		return
	}

	// Create a session for the user
	session, err := s.auth.CreateSessionForUser(r.Context(), user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "google_login_failed", fmt.Sprintf("Failed to create session: %v", err), nil)
		return
	}

	writeJSON(w, http.StatusOK, response{Data: session, Message: "ok"})
}

type GoogleUserInfo struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Picture       string `json:"picture"`
}

func (s *Server) exchangeGoogleCode(ctx context.Context, code, redirectURI string) (string, error) {
	if s.config.GoogleOAuthClientID == "" || s.config.GoogleOAuthClientSecret == "" {
		return "", errors.New("Google OAuth is not configured on the server")
	}

	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", s.config.GoogleOAuthClientID)
	data.Set("client_secret", s.config.GoogleOAuthClientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errData map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errData)
		return "", fmt.Errorf("Google token exchange failed (status %d): %v", resp.StatusCode, errData)
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.AccessToken, nil
}

func (s *Server) fetchGoogleUserInfo(ctx context.Context, accessToken string) (GoogleUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return GoogleUserInfo{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return GoogleUserInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return GoogleUserInfo{}, fmt.Errorf("Google userinfo request failed (status %d)", resp.StatusCode)
	}

	var info GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return GoogleUserInfo{}, err
	}

	return info, nil
}

