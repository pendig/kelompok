package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pendig/kelompok/internal/audit"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrInvalidSession      = errors.New("invalid session")
	ErrNotFound            = errors.New("user not found")
	ErrUserExists          = errors.New("user already has a password")
	ErrProfileNameRequired = errors.New("profile name is required")
	ErrProfileNameTooLong  = errors.New("profile name must be at most 120 characters")
)

const SessionTTL = 30 * 24 * time.Hour

const (
	UserRoleSuperadmin = "superadmin"

	OrganizationRoleOwner  = "owner"
	OrganizationRoleAdmin  = "admin"
	OrganizationRoleMember = "member"
	OrganizationRoleViewer = "viewer"
)

type Repository struct {
	db *pgxpool.Pool
}

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateProfileInput struct {
	Name string `json:"name"`
}

type Session struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      User      `json:"user"`
}

type OrganizationRole struct {
	OrganizationID   string    `json:"organization_id"`
	OrganizationSlug string    `json:"organization_slug"`
	OrganizationName string    `json:"organization_name"`
	Role             string    `json:"role"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// OrganizationClaim is a per-user, organization-scoped view of a claim
// request that surfaces only the fields the account page needs to render
// the claim journey (status, timestamps, organization context). It is
// intentionally narrower than organizations.ClaimRequest so the public
// /auth/me payload never leaks reviewer-only metadata or evidence.
type OrganizationClaim struct {
	ID                      string     `json:"id"`
	OrganizationID          string     `json:"organization_id"`
	OrganizationSlug        string     `json:"organization_slug"`
	OrganizationName        string     `json:"organization_name"`
	OrganizationClaimStatus string     `json:"organization_claim_status"`
	Method                  string     `json:"method"`
	Target                  string     `json:"target"`
	Status                  string     `json:"status"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
	ReviewedAt              *time.Time `json:"reviewed_at,omitempty"`
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Register(ctx context.Context, input RegisterInput) (User, error) {
	name := strings.TrimSpace(input.Name)
	email := normalizeEmail(input.Email)
	password := strings.TrimSpace(input.Password)
	if email == "" {
		return User{}, errors.New("email is required")
	}
	if !validEmail(email) {
		return User{}, errors.New("email must be valid")
	}
	if password == "" {
		return User{}, errors.New("password is required")
	}
	if len(password) < 8 {
		return User{}, errors.New("password must be at least 8 characters")
	}
	if name == "" {
		name = strings.Split(email, "@")[0]
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	row := r.db.QueryRow(ctx, `
		INSERT INTO users (name, email, password_hash, role)
		VALUES ($1, $2, $3, 'viewer')
		ON CONFLICT (email) DO NOTHING
		RETURNING id::text, name, email, role, created_at, updated_at
	`, name, email, string(hash))

	user, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		user, err = r.setInitialPassword(ctx, name, email, string(hash))
	}
	if err == nil {
		_ = audit.Record(ctx, r.db, user.ID, "user", user.ID, "register", nil, user, nil)
	}
	return user, err
}

func (r *Repository) setInitialPassword(ctx context.Context, name string, email string, passwordHash string) (User, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE users
		SET
			name = COALESCE(NULLIF($2, ''), name),
			password_hash = $3,
			updated_at = now()
		WHERE email = $1
			AND password_hash IS NULL
		RETURNING id::text, name, email, role, created_at, updated_at
	`, email, name, passwordHash)

	user, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrUserExists
	}
	return user, err
}

func (r *Repository) Login(ctx context.Context, input LoginInput) (Session, error) {
	email := normalizeEmail(input.Email)
	password := strings.TrimSpace(input.Password)
	if email == "" || password == "" {
		return Session{}, ErrInvalidCredentials
	}

	var user User
	var passwordHash string
	err := r.db.QueryRow(ctx, `
		SELECT id::text, name, email, role, created_at, updated_at, COALESCE(password_hash, '')
		FROM users
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt, &passwordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrInvalidCredentials
	}
	if err != nil {
		return Session{}, err
	}
	if passwordHash == "" || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) != nil {
		return Session{}, ErrInvalidCredentials
	}

	token, tokenHash, err := newSessionToken()
	if err != nil {
		return Session{}, err
	}
	expiresAt := time.Now().UTC().Add(SessionTTL)

	if _, err := r.db.Exec(ctx, `
		INSERT INTO user_sessions (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, user.ID, tokenHash, expiresAt); err != nil {
		return Session{}, err
	}

	_ = audit.Record(ctx, r.db, user.ID, "user_session", nil, "login", nil, nil, nil)
	return Session{Token: token, ExpiresAt: expiresAt, User: user}, nil
}

func (r *Repository) UserBySessionToken(ctx context.Context, token string) (User, error) {
	tokenHash := hashToken(token)
	if tokenHash == "" {
		return User{}, ErrInvalidSession
	}

	row := r.db.QueryRow(ctx, `
		UPDATE user_sessions
		SET last_used_at = now()
		WHERE token_hash = $1
			AND revoked_at IS NULL
			AND expires_at > now()
		RETURNING user_id::text
	`, tokenHash)

	var userID string
	if err := row.Scan(&userID); errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrInvalidSession
	} else if err != nil {
		return User{}, err
	}

	user, err := r.FindUserByID(ctx, userID)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (r *Repository) Logout(ctx context.Context, token string) error {
	tokenHash := hashToken(token)
	if tokenHash == "" {
		return ErrInvalidSession
	}

	tag, err := r.db.Exec(ctx, `
		UPDATE user_sessions
		SET revoked_at = now()
		WHERE token_hash = $1
			AND revoked_at IS NULL
	`, tokenHash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrInvalidSession
	}
	return nil
}

func (r *Repository) FindUserByID(ctx context.Context, id string) (User, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id::text, name, email, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id)
	user, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return user, err
}

func (r *Repository) UpdateProfile(ctx context.Context, userID string, input UpdateProfileInput) (User, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return User{}, ErrProfileNameRequired
	}
	if utf8.RuneCountInString(name) > 120 {
		return User{}, ErrProfileNameTooLong
	}

	before, err := r.FindUserByID(ctx, userID)
	if err != nil {
		return User{}, err
	}

	row := r.db.QueryRow(ctx, `
		UPDATE users
		SET
			name = $2,
			updated_at = now()
		WHERE id = $1
		RETURNING id::text, name, email, role, created_at, updated_at
	`, userID, name)

	user, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}

	_ = audit.Record(ctx, r.db, user.ID, "user", user.ID, "update_profile", before, user, nil)
	return user, nil
}

func (r *Repository) RolesByUserID(ctx context.Context, userID string) ([]OrganizationRole, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			our.organization_id::text,
			o.slug,
			o.name,
			our.role,
			our.created_at,
			our.updated_at
		FROM organization_user_roles our
		JOIN organizations o ON o.id = our.organization_id
		WHERE our.user_id = $1
		ORDER BY o.name ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]OrganizationRole, 0)
	for rows.Next() {
		var item OrganizationRole
		if err := rows.Scan(&item.OrganizationID, &item.OrganizationSlug, &item.OrganizationName, &item.Role, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// ClaimsByUserID returns every organization claim request submitted by the
// given user, joined with the organization context the account page needs
// to render the claim journey (organization name/slug + the organization's
// current claim_status). The list is ordered by recency so the account UI
// can always show the most recent submission first without re-sorting.
func (r *Repository) ClaimsByUserID(ctx context.Context, userID string) ([]OrganizationClaim, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			cr.id::text,
			cr.organization_id::text,
			o.slug,
			o.name,
			o.claim_status,
			cr.method,
			cr.target,
			cr.status,
			cr.reviewed_at,
			cr.created_at,
			cr.updated_at
		FROM claim_requests cr
		JOIN organizations o ON o.id = cr.organization_id
		WHERE cr.user_id = $1
		ORDER BY cr.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]OrganizationClaim, 0)
	for rows.Next() {
		var item OrganizationClaim
		if err := rows.Scan(
			&item.ID,
			&item.OrganizationID,
			&item.OrganizationSlug,
			&item.OrganizationName,
			&item.OrganizationClaimStatus,
			&item.Method,
			&item.Target,
			&item.Status,
			&item.ReviewedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CanManageOrganization(ctx context.Context, user User, organizationSlug string) (bool, error) {
	if user.Role == UserRoleSuperadmin {
		return true, nil
	}

	var allowed bool
	if err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM organization_user_roles our
			JOIN organizations o ON o.id = our.organization_id
			WHERE our.user_id = $1
				AND o.slug = $2
				AND our.role = ANY($3)
		)
	`, user.ID, organizationSlug, ManageOrganizationRoles()).Scan(&allowed); err != nil {
		return false, err
	}
	return allowed, nil
}

// CanReviewRelatedClaim allows a manager of a parent organization to review
// pending claims for active child organizations created under that parent.
func (r *Repository) CanReviewRelatedClaim(ctx context.Context, user User, claimOrganizationSlug string) (bool, error) {
	if user.Role == UserRoleSuperadmin {
		return true, nil
	}

	var allowed bool
	if err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM organization_user_roles our
			JOIN organizations managed ON managed.id = our.organization_id
			JOIN organization_relationships rel ON rel.parent_organization_id = managed.id
			JOIN organizations claimed ON claimed.id = rel.child_organization_id
			WHERE our.user_id = $1
				AND claimed.slug = $2
				AND rel.status = 'active'
				AND our.role = ANY($3)
		)
	`, user.ID, claimOrganizationSlug, ManageOrganizationRoles()).Scan(&allowed); err != nil {
		return false, err
	}
	if allowed {
		return true, nil
	}

	return r.CanManageOrganization(ctx, user, claimOrganizationSlug)
}

// CanManageOrganizationRole is the shared permission rule for org-scoped
// console/admin actions that mutate or moderate organization data.
func CanManageOrganizationRole(role string) bool {
	role = strings.TrimSpace(strings.ToLower(role))
	return role == OrganizationRoleOwner || role == OrganizationRoleAdmin
}

// ManageOrganizationRoles returns the canonical role allow-list used by SQL
// guards that need to match CanManageOrganizationRole.
func ManageOrganizationRoles() []string {
	return []string{OrganizationRoleOwner, OrganizationRoleAdmin}
}

func (r *Repository) AssignOrganizationRole(ctx context.Context, organizationID string, userID string, role string, actorUserID any) error {
	role = strings.TrimSpace(strings.ToLower(role))
	if role != OrganizationRoleOwner && role != OrganizationRoleAdmin && role != OrganizationRoleMember && role != OrganizationRoleViewer {
		return errors.New("unsupported organization role")
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO organization_user_roles (organization_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (organization_id, user_id) DO UPDATE SET
			role = EXCLUDED.role,
			updated_at = now()
	`, organizationID, userID, role)
	if err == nil {
		_ = audit.Record(ctx, r.db, actorUserID, "organization_user_role", organizationID, "assign", nil, nil, roleAssignAuditMetadata(organizationID, userID, role))
	}
	return err
}

// roleAssignAuditMetadata builds the audit metadata bag emitted whenever a
// user is granted (or re-granted) an organization role. The "organization_id"
// key is required so audit.Record's organizationID() resolver can pin the
// resulting audit row to the affected organization (audit_logs.organization_id
// is otherwise NULL for non-"organization" entity types and the row would not
// surface in the org-scoped audit listing endpoint).
func roleAssignAuditMetadata(organizationID string, userID string, role string) map[string]any {
	return map[string]any{
		"organization_id": organizationID,
		"user_id":         userID,
		"role":            role,
	}
}

type userRow interface {
	Scan(dest ...any) error
}

func scanUser(row userRow) (User, error) {
	var user User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	return user, err
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validEmail(email string) bool {
	address, err := mail.ParseAddress(email)
	return err == nil && address.Address == email
}

func newSessionToken() (string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(bytes)
	return token, hashToken(token), nil
}

func hashToken(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (r *Repository) FindOrCreateUserByEmail(ctx context.Context, email, name string) (User, error) {
	email = normalizeEmail(email)
	if email == "" {
		return User{}, errors.New("email is required")
	}

	// 1. Try to find the user first
	var user User
	err := r.db.QueryRow(ctx, `
		SELECT id::text, name, email, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err == nil {
		return user, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return User{}, err
	}

	// 2. If user doesn't exist, create them (with password_hash = NULL)
	if name == "" {
		name = strings.Split(email, "@")[0]
	}

	row := r.db.QueryRow(ctx, `
		INSERT INTO users (name, email, password_hash, role)
		VALUES ($1, $2, NULL, 'viewer')
		RETURNING id::text, name, email, role, created_at, updated_at
	`, name, email)

	user, err = scanUser(row)
	if err == nil {
		_ = audit.Record(ctx, r.db, user.ID, "user", user.ID, "register", nil, user, nil)
	}
	return user, err
}

func (r *Repository) CreateSessionForUser(ctx context.Context, user User) (Session, error) {
	token, tokenHash, err := newSessionToken()
	if err != nil {
		return Session{}, err
	}
	expiresAt := time.Now().UTC().Add(SessionTTL)

	if _, err := r.db.Exec(ctx, `
		INSERT INTO user_sessions (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, user.ID, tokenHash, expiresAt); err != nil {
		return Session{}, err
	}

	_ = audit.Record(ctx, r.db, user.ID, "user_session", nil, "login", nil, nil, nil)
	return Session{Token: token, ExpiresAt: expiresAt, User: user}, nil
}
