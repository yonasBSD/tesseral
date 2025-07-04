// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: queries-saml.sql

package queries

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createAuditLogEvent = `-- name: CreateAuditLogEvent :one
INSERT INTO audit_log_events (id, project_id, organization_id, actor_user_id, actor_session_id, resource_type, resource_id, event_name, event_time, event_details)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, coalesce($10, '{}'::jsonb))
RETURNING
    id, project_id, organization_id, actor_user_id, actor_session_id, actor_api_key_id, actor_console_user_id, actor_console_session_id, actor_backend_api_key_id, actor_intermediate_session_id, resource_type, resource_id, event_name, event_time, event_details
`

type CreateAuditLogEventParams struct {
	ID             uuid.UUID
	ProjectID      uuid.UUID
	OrganizationID *uuid.UUID
	ActorUserID    *uuid.UUID
	ActorSessionID *uuid.UUID
	ResourceType   *AuditLogEventResourceType
	ResourceID     *uuid.UUID
	EventName      string
	EventTime      *time.Time
	EventDetails   interface{}
}

func (q *Queries) CreateAuditLogEvent(ctx context.Context, arg CreateAuditLogEventParams) (AuditLogEvent, error) {
	row := q.db.QueryRow(ctx, createAuditLogEvent,
		arg.ID,
		arg.ProjectID,
		arg.OrganizationID,
		arg.ActorUserID,
		arg.ActorSessionID,
		arg.ResourceType,
		arg.ResourceID,
		arg.EventName,
		arg.EventTime,
		arg.EventDetails,
	)
	var i AuditLogEvent
	err := row.Scan(
		&i.ID,
		&i.ProjectID,
		&i.OrganizationID,
		&i.ActorUserID,
		&i.ActorSessionID,
		&i.ActorApiKeyID,
		&i.ActorConsoleUserID,
		&i.ActorConsoleSessionID,
		&i.ActorBackendApiKeyID,
		&i.ActorIntermediateSessionID,
		&i.ResourceType,
		&i.ResourceID,
		&i.EventName,
		&i.EventTime,
		&i.EventDetails,
	)
	return i, err
}

const createSession = `-- name: CreateSession :one
INSERT INTO sessions (id, user_id, expire_time, refresh_token_sha256, primary_auth_factor)
    VALUES ($1, $2, $3, $4, 'saml')
RETURNING
    id, user_id, create_time, expire_time, refresh_token_sha256, impersonator_user_id, last_active_time, primary_auth_factor
`

type CreateSessionParams struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	ExpireTime         *time.Time
	RefreshTokenSha256 []byte
}

func (q *Queries) CreateSession(ctx context.Context, arg CreateSessionParams) (Session, error) {
	row := q.db.QueryRow(ctx, createSession,
		arg.ID,
		arg.UserID,
		arg.ExpireTime,
		arg.RefreshTokenSha256,
	)
	var i Session
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.CreateTime,
		&i.ExpireTime,
		&i.RefreshTokenSha256,
		&i.ImpersonatorUserID,
		&i.LastActiveTime,
		&i.PrimaryAuthFactor,
	)
	return i, err
}

const createUser = `-- name: CreateUser :one
INSERT INTO users (id, organization_id, email, is_owner)
    VALUES ($1, $2, $3, $4)
RETURNING
    id, organization_id, password_bcrypt, google_user_id, microsoft_user_id, email, create_time, update_time, deactivate_time, is_owner, failed_password_attempts, password_lockout_expire_time, authenticator_app_secret_ciphertext, failed_authenticator_app_attempts, authenticator_app_lockout_expire_time, authenticator_app_recovery_code_sha256s, display_name, profile_picture_url, github_user_id
`

type CreateUserParams struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	Email          string
	IsOwner        bool
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser,
		arg.ID,
		arg.OrganizationID,
		arg.Email,
		arg.IsOwner,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.OrganizationID,
		&i.PasswordBcrypt,
		&i.GoogleUserID,
		&i.MicrosoftUserID,
		&i.Email,
		&i.CreateTime,
		&i.UpdateTime,
		&i.DeactivateTime,
		&i.IsOwner,
		&i.FailedPasswordAttempts,
		&i.PasswordLockoutExpireTime,
		&i.AuthenticatorAppSecretCiphertext,
		&i.FailedAuthenticatorAppAttempts,
		&i.AuthenticatorAppLockoutExpireTime,
		&i.AuthenticatorAppRecoveryCodeSha256s,
		&i.DisplayName,
		&i.ProfilePictureUrl,
		&i.GithubUserID,
	)
	return i, err
}

const getOrganizationDomains = `-- name: GetOrganizationDomains :many
SELECT
    DOMAIN
FROM
    organization_domains
WHERE
    organization_id = $1
`

func (q *Queries) GetOrganizationDomains(ctx context.Context, organizationID uuid.UUID) ([]string, error) {
	rows, err := q.db.Query(ctx, getOrganizationDomains, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var domain string
		if err := rows.Scan(&domain); err != nil {
			return nil, err
		}
		items = append(items, domain)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getProject = `-- name: GetProject :one
SELECT
    id, organization_id, log_in_with_password, log_in_with_google, log_in_with_microsoft, google_oauth_client_id, microsoft_oauth_client_id, google_oauth_client_secret_ciphertext, microsoft_oauth_client_secret_ciphertext, display_name, create_time, update_time, logins_disabled, log_in_with_authenticator_app, log_in_with_passkey, log_in_with_email, log_in_with_saml, redirect_uri, after_login_redirect_uri, after_signup_redirect_uri, vault_domain, email_send_from_domain, cookie_domain, email_quota_daily, stripe_customer_id, entitled_custom_vault_domains, entitled_backend_api_keys, log_in_with_github, github_oauth_client_id, github_oauth_client_secret_ciphertext, api_keys_enabled, api_key_secret_token_prefix, audit_logs_enabled, log_in_with_oidc
FROM
    projects
WHERE
    id = $1
`

func (q *Queries) GetProject(ctx context.Context, id uuid.UUID) (Project, error) {
	row := q.db.QueryRow(ctx, getProject, id)
	var i Project
	err := row.Scan(
		&i.ID,
		&i.OrganizationID,
		&i.LogInWithPassword,
		&i.LogInWithGoogle,
		&i.LogInWithMicrosoft,
		&i.GoogleOauthClientID,
		&i.MicrosoftOauthClientID,
		&i.GoogleOauthClientSecretCiphertext,
		&i.MicrosoftOauthClientSecretCiphertext,
		&i.DisplayName,
		&i.CreateTime,
		&i.UpdateTime,
		&i.LoginsDisabled,
		&i.LogInWithAuthenticatorApp,
		&i.LogInWithPasskey,
		&i.LogInWithEmail,
		&i.LogInWithSaml,
		&i.RedirectUri,
		&i.AfterLoginRedirectUri,
		&i.AfterSignupRedirectUri,
		&i.VaultDomain,
		&i.EmailSendFromDomain,
		&i.CookieDomain,
		&i.EmailQuotaDaily,
		&i.StripeCustomerID,
		&i.EntitledCustomVaultDomains,
		&i.EntitledBackendApiKeys,
		&i.LogInWithGithub,
		&i.GithubOauthClientID,
		&i.GithubOauthClientSecretCiphertext,
		&i.ApiKeysEnabled,
		&i.ApiKeySecretTokenPrefix,
		&i.AuditLogsEnabled,
		&i.LogInWithOidc,
	)
	return i, err
}

const getSAMLConnection = `-- name: GetSAMLConnection :one
SELECT
    saml_connections.id, saml_connections.organization_id, saml_connections.create_time, saml_connections.is_primary, saml_connections.idp_redirect_url, saml_connections.idp_x509_certificate, saml_connections.idp_entity_id, saml_connections.update_time
FROM
    saml_connections
    JOIN organizations ON saml_connections.organization_id = organizations.id
WHERE
    organizations.project_id = $1
    AND organizations.log_in_with_saml
    AND saml_connections.id = $2
`

type GetSAMLConnectionParams struct {
	ProjectID uuid.UUID
	ID        uuid.UUID
}

func (q *Queries) GetSAMLConnection(ctx context.Context, arg GetSAMLConnectionParams) (SamlConnection, error) {
	row := q.db.QueryRow(ctx, getSAMLConnection, arg.ProjectID, arg.ID)
	var i SamlConnection
	err := row.Scan(
		&i.ID,
		&i.OrganizationID,
		&i.CreateTime,
		&i.IsPrimary,
		&i.IdpRedirectUrl,
		&i.IdpX509Certificate,
		&i.IdpEntityID,
		&i.UpdateTime,
	)
	return i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT
    id, organization_id, password_bcrypt, google_user_id, microsoft_user_id, email, create_time, update_time, deactivate_time, is_owner, failed_password_attempts, password_lockout_expire_time, authenticator_app_secret_ciphertext, failed_authenticator_app_attempts, authenticator_app_lockout_expire_time, authenticator_app_recovery_code_sha256s, display_name, profile_picture_url, github_user_id
FROM
    users
WHERE
    organization_id = $1
    AND email = $2
`

type GetUserByEmailParams struct {
	OrganizationID uuid.UUID
	Email          string
}

func (q *Queries) GetUserByEmail(ctx context.Context, arg GetUserByEmailParams) (User, error) {
	row := q.db.QueryRow(ctx, getUserByEmail, arg.OrganizationID, arg.Email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.OrganizationID,
		&i.PasswordBcrypt,
		&i.GoogleUserID,
		&i.MicrosoftUserID,
		&i.Email,
		&i.CreateTime,
		&i.UpdateTime,
		&i.DeactivateTime,
		&i.IsOwner,
		&i.FailedPasswordAttempts,
		&i.PasswordLockoutExpireTime,
		&i.AuthenticatorAppSecretCiphertext,
		&i.FailedAuthenticatorAppAttempts,
		&i.AuthenticatorAppLockoutExpireTime,
		&i.AuthenticatorAppRecoveryCodeSha256s,
		&i.DisplayName,
		&i.ProfilePictureUrl,
		&i.GithubUserID,
	)
	return i, err
}
