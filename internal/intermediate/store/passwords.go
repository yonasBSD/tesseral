package store

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/google/uuid"
	"github.com/tesseral-labs/tesseral/internal/bcryptcost"
	"github.com/tesseral-labs/tesseral/internal/common/apierror"
	"github.com/tesseral-labs/tesseral/internal/intermediate/authn"
	intermediatev1 "github.com/tesseral-labs/tesseral/internal/intermediate/gen/tesseral/intermediate/v1"
	"github.com/tesseral-labs/tesseral/internal/intermediate/store/queries"
	"github.com/tesseral-labs/tesseral/internal/store/idformat"
	"golang.org/x/crypto/bcrypt"
)

const (
	// after this many failed attempts, lock out a user
	passwordLockoutAttempts = 5
	// how long to lock users out
	passwordLockoutDuration = time.Minute * 10
)

func (s *Store) RegisterPassword(ctx context.Context, req *intermediatev1.RegisterPasswordRequest) (*intermediatev1.RegisterPasswordResponse, error) {
	// Check if the password is compromised.
	pwned, err := s.hibp.Pwned(ctx, req.Password)
	if err != nil {
		return nil, fmt.Errorf("check password against HIBP: %w", err)
	}
	if pwned {
		return nil, apierror.NewPasswordCompromisedError("password is compromised", errors.New("password is compromised"))
	}

	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	qIntermediateSession, err := q.GetIntermediateSessionByID(ctx, authn.IntermediateSessionID(ctx))
	if err != nil {
		return nil, err
	}

	if qIntermediateSession.PasswordVerified {
		return nil, apierror.NewFailedPreconditionError("password already verified", fmt.Errorf("password already verified"))
	}

	qProject, err := q.GetProjectByID(ctx, authn.ProjectID(ctx))
	if err != nil {
		return nil, fmt.Errorf("get project by id: %w", err)
	}

	if err := enforceProjectLoginEnabled(qProject); err != nil {
		return nil, fmt.Errorf("enforce project login enabled: %w", err)
	}

	qOrg, err := q.GetProjectOrganizationByID(ctx, queries.GetProjectOrganizationByIDParams{
		ProjectID: authn.ProjectID(ctx),
		ID:        *qIntermediateSession.OrganizationID,
	})
	if err != nil {
		return nil, fmt.Errorf("get organization by id: %w", err)
	}

	if err := enforceOrganizationLoginEnabled(qOrg); err != nil {
		return nil, fmt.Errorf("enforce organization login enabled: %w", err)
	}

	// Ensure given organization is suitable for authentication over password:
	if !qOrg.LogInWithPassword {
		return nil, apierror.NewFailedPreconditionError("password authentication not enabled", fmt.Errorf("password authentication not enabled"))
	}

	emailVerified, err := s.getIntermediateSessionEmailVerified(ctx, q, authn.IntermediateSessionID(ctx))
	if err != nil {
		return nil, fmt.Errorf("get intermediate session email verified: %w", err)
	}
	if !emailVerified {
		return nil, apierror.NewFailedPreconditionError("email not verified", fmt.Errorf("email not verified"))
	}

	// only allow password registration if the matching user doesn't already
	// have one, or if the intermediate session has verified a password reset
	// code
	qUser, err := s.matchUser(ctx, q, qOrg, qIntermediateSession)
	if err != nil {
		return nil, fmt.Errorf("match user: %w", err)
	}

	if qUser != nil && qUser.PasswordBcrypt != nil && !qIntermediateSession.PasswordResetCodeVerified {
		return nil, apierror.NewFailedPreconditionError("user already has password configured", fmt.Errorf("user already has password configured"))
	}

	passwordBcryptBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptcost.Cost)
	if err != nil {
		return nil, fmt.Errorf("generate bcrypt hash: %w", err)
	}

	passwordBcrypt := string(passwordBcryptBytes)
	_, err = q.UpdateIntermediateSessionNewUserPasswordBcrypt(ctx, queries.UpdateIntermediateSessionNewUserPasswordBcryptParams{
		ID:                    authn.IntermediateSessionID(ctx),
		NewUserPasswordBcrypt: &passwordBcrypt,
	})
	if err != nil {
		return nil, fmt.Errorf("update intermediate session new user password bcrypt: %w", err)
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &intermediatev1.RegisterPasswordResponse{}, nil
}

func (s *Store) VerifyPassword(ctx context.Context, req *intermediatev1.VerifyPasswordRequest) (*intermediatev1.VerifyPasswordResponse, error) {
	if req.Email != "" {
		res, err := s.logInWithPassword(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("log in with password: %w", err)
		}
		return res, nil
	}

	intermediateSession := authn.IntermediateSession(ctx)

	if intermediateSession.PasswordVerified {
		return nil, apierror.NewFailedPreconditionError("user already verified for intermediate session", fmt.Errorf("user already verified for intermediate session"))
	}

	orgID, err := idformat.Organization.Parse(intermediateSession.OrganizationId)
	if err != nil {
		return nil, apierror.NewInvalidArgumentError("invalid organization id", fmt.Errorf("parse organization id: %w", err))
	}

	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	qProject, err := q.GetProjectByID(ctx, authn.ProjectID(ctx))
	if err != nil {
		return nil, fmt.Errorf("get project by id: %w", err)
	}

	if err := enforceProjectLoginEnabled(qProject); err != nil {
		return nil, fmt.Errorf("enforce project login enabled: %w", err)
	}

	qOrg, err := q.GetProjectOrganizationByID(ctx, queries.GetProjectOrganizationByIDParams{
		ProjectID: authn.ProjectID(ctx),
		ID:        orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("get organization by id: %w", err)
	}

	if err := enforceOrganizationLoginEnabled(qOrg); err != nil {
		return nil, fmt.Errorf("enforce organization login enabled: %w", err)
	}

	qIntermediateSession, err := q.GetIntermediateSessionByID(ctx, authn.IntermediateSessionID(ctx))
	if err != nil {
		return nil, fmt.Errorf("get intermediate session by id: %w", err)
	}

	// Ensure given organization is visible to intermediate session, and
	// suitable for authentication over password:
	//
	// 1. The organization must have passwords enabled,
	// 2. The intermediate session must have a verified email, and
	// 3. A user in that org must have the same email.
	if !qOrg.LogInWithPassword {
		return nil, apierror.NewFailedPreconditionError("password authentication not enabled", nil)
	}

	emailVerified, err := s.getIntermediateSessionEmailVerified(ctx, q, qIntermediateSession.ID)
	if err != nil {
		return nil, fmt.Errorf("get intermediate session verified: %w", err)
	}

	if !emailVerified {
		return nil, apierror.NewFailedPreconditionError("email not verified", nil)
	}

	qMatchingUser, err := s.matchEmailUser(ctx, q, qOrg, qIntermediateSession)
	if err != nil {
		return nil, fmt.Errorf("match email user: %w", err)
	}

	if qMatchingUser == nil {
		return nil, apierror.NewFailedPreconditionError("no corresponding user found", nil)
	}

	if err := s.attemptMatchPassword(ctx, q, *qMatchingUser, req.Password); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, apierror.NewIncorrectPasswordError("incorrect password", nil)
		}

		return nil, fmt.Errorf("attempt match password: %w", err)
	}

	// Re-write password back to database; this lets us progressively increase
	// bcrypt costs over time.
	//
	// We could avoid these writes by checking the PasswordBcrypt using
	// bcrypt.Cost, but for relatively small additional cost, not doing so
	// reduces the complexity and number of paths through this code.
	passwordBcryptBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptcost.Cost)
	if err != nil {
		return nil, fmt.Errorf("generate bcrypt hash: %w", err)
	}

	passwordBcrypt := string(passwordBcryptBytes)
	if _, err := q.UpdateUserPasswordBcrypt(ctx, queries.UpdateUserPasswordBcryptParams{
		ID:             qMatchingUser.ID,
		PasswordBcrypt: &passwordBcrypt,
	}); err != nil {
		return nil, fmt.Errorf("update user password bcrypt: %w", err)
	}

	if _, err := q.UpdateIntermediateSessionPasswordVerified(ctx, queries.UpdateIntermediateSessionPasswordVerifiedParams{
		OrganizationID: &qOrg.ID,
		ID:             qIntermediateSession.ID,
	}); err != nil {
		return nil, fmt.Errorf("update intermediate session password verified: %w", err)
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &intermediatev1.VerifyPasswordResponse{}, nil
}

// logInWithPassword handles the flow where a user directly inputs an email and
// password.
func (s *Store) logInWithPassword(ctx context.Context, req *intermediatev1.VerifyPasswordRequest) (*intermediatev1.VerifyPasswordResponse, error) {
	// In this flow, we don't require verifying an email or any other previous
	// state on the intermediate session. We issue a user an intermediate
	// session for the unique user with that email and password.
	//
	// If there is no unique password-having user with the given email, then
	// refuse to proceed.

	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	qProject, err := q.GetProjectByID(ctx, authn.ProjectID(ctx))
	if err != nil {
		return nil, fmt.Errorf("get project by id: %w", err)
	}

	if err := enforceProjectLoginEnabled(qProject); err != nil {
		return nil, fmt.Errorf("enforce project login enabled: %w", err)
	}

	if !qProject.LogInWithPassword {
		return nil, apierror.NewFailedPreconditionError("log in with password not enabled", nil)
	}

	// this query checks for orgs with logins enabled and passwords enabled
	qUsers, err := q.GetUsersByForLogInWithPassword(ctx, queries.GetUsersByForLogInWithPasswordParams{
		ProjectID: authn.ProjectID(ctx),
		Email:     req.Email,
	})
	if err != nil {
		return nil, fmt.Errorf("get users by project id and email: %w", err)
	}

	if len(qUsers) != 1 {
		return nil, apierror.NewPasswordsUnavailableForEmailError("password-based login not available for this email", nil)
	}

	qMatchingUser := qUsers[0]

	if err := s.attemptMatchPassword(ctx, q, qMatchingUser, req.Password); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, apierror.NewIncorrectPasswordError("incorrect password", nil)
		}

		return nil, fmt.Errorf("attempt match password: %w", err)
	}

	passwordBcryptBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptcost.Cost)
	if err != nil {
		return nil, fmt.Errorf("generate bcrypt hash: %w", err)
	}

	passwordBcrypt := string(passwordBcryptBytes)
	if _, err := q.UpdateUserPasswordBcrypt(ctx, queries.UpdateUserPasswordBcryptParams{
		ID:             qMatchingUser.ID,
		PasswordBcrypt: &passwordBcrypt,
	}); err != nil {
		return nil, fmt.Errorf("update user password bcrypt: %w", err)
	}

	if _, err := q.UpdateIntermediateSessionEmail(ctx, queries.UpdateIntermediateSessionEmailParams{
		ID:    authn.IntermediateSessionID(ctx),
		Email: &req.Email,
	}); err != nil {
		return nil, fmt.Errorf("update intermediate session email: %w", err)
	}

	// if you have a working password, you count as having a verified email
	if _, err := q.UpdateIntermediateSessionEmailVerificationChallengeCompleted(ctx, authn.IntermediateSessionID(ctx)); err != nil {
		return nil, fmt.Errorf("update intermediate session email verification challenge completed: %w", err)
	}

	if _, err := q.UpdateIntermediateSessionPasswordVerified(ctx, queries.UpdateIntermediateSessionPasswordVerifiedParams{
		ID:             authn.IntermediateSessionID(ctx),
		OrganizationID: &qMatchingUser.OrganizationID,
	}); err != nil {
		return nil, fmt.Errorf("update intermediate session password verified: %w", err)
	}

	if _, err := q.UpdateIntermediateSessionPrimaryAuthFactor(ctx, queries.UpdateIntermediateSessionPrimaryAuthFactorParams{
		ID:                authn.IntermediateSessionID(ctx),
		PrimaryAuthFactor: refOrNil(queries.PrimaryAuthFactorPassword),
	}); err != nil {
		return nil, fmt.Errorf("update intermediate session primary auth factor: %w", err)
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &intermediatev1.VerifyPasswordResponse{}, nil
}

func (s *Store) attemptMatchPassword(ctx context.Context, q *queries.Queries, qUser queries.User, password string) error {
	if qUser.PasswordBcrypt == nil {
		return apierror.NewFailedPreconditionError("user does not have password configured", nil)
	}

	if qUser.PasswordLockoutExpireTime != nil && qUser.PasswordLockoutExpireTime.After(time.Now()) {
		return apierror.NewFailedPreconditionError("too many password attempts; user is temporarily locked out", nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*qUser.PasswordBcrypt), []byte(password)); err != nil {
		attempts := qUser.FailedPasswordAttempts + 1
		if attempts >= passwordLockoutAttempts {
			// lock the user out
			passwordLockoutExpireTime := time.Now().Add(passwordLockoutDuration)
			if _, err := q.UpdateUserPasswordLockoutExpireTime(ctx, queries.UpdateUserPasswordLockoutExpireTimeParams{
				ID:                        qUser.ID,
				PasswordLockoutExpireTime: &passwordLockoutExpireTime,
			}); err != nil {
				return fmt.Errorf("update user password lockout expire time: %w", err)
			}

			// reset fail count
			if _, err := q.UpdateUserFailedPasswordAttempts(ctx, queries.UpdateUserFailedPasswordAttemptsParams{
				ID:                     qUser.ID,
				FailedPasswordAttempts: 0,
			}); err != nil {
				return fmt.Errorf("update user failed password attempts: %w", err)
			}

			return apierror.NewFailedPreconditionError("too many password attempts; user is temporarily locked out", nil)
		}

		// update fail count, but do not lock out
		if _, err := q.UpdateUserFailedPasswordAttempts(ctx, queries.UpdateUserFailedPasswordAttemptsParams{
			ID:                     qUser.ID,
			FailedPasswordAttempts: attempts,
		}); err != nil {
			return fmt.Errorf("update user failed password attempts: %w", err)
		}

		return fmt.Errorf("bcrypt: %w", err)
	}

	return nil
}

func (s *Store) IssuePasswordResetCode(ctx context.Context, req *intermediatev1.IssuePasswordResetCodeRequest) (*intermediatev1.IssuePasswordResetCodeResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	qProject, err := q.GetProjectByID(ctx, authn.ProjectID(ctx))
	if err != nil {
		return nil, fmt.Errorf("get project by id: %w", err)
	}

	qIntermediateSession, err := q.GetIntermediateSessionByID(ctx, authn.IntermediateSessionID(ctx))
	if err != nil {
		return nil, fmt.Errorf("get intermediate session by id: %w", err)
	}

	// we don't expect to issue these codes until after the initial email
	// verification process, so enforce that expectation here

	emailVerified, err := s.getIntermediateSessionEmailVerified(ctx, q, qIntermediateSession.ID)
	if err != nil {
		return nil, fmt.Errorf("get intermediate session email verified: %w", err)
	}

	if !emailVerified {
		return nil, apierror.NewFailedPreconditionError("email not verified", fmt.Errorf("email not verified"))
	}

	passwordResetCodeUUID := uuid.New()
	passwordResetCodeSHA256 := sha256.Sum256(passwordResetCodeUUID[:])

	if _, err := q.UpdateIntermediateSessionPasswordResetCodeSHA256(ctx, queries.UpdateIntermediateSessionPasswordResetCodeSHA256Params{
		ID:                      qIntermediateSession.ID,
		PasswordResetCodeSha256: passwordResetCodeSHA256[:],
	}); err != nil {
		return nil, fmt.Errorf("update intermediate session password reset code sha256: %w", err)
	}

	qEmailDailyQuotaUsage, err := q.IncrementProjectEmailDailyQuotaUsage(ctx, authn.ProjectID(ctx))
	if err != nil {
		return nil, fmt.Errorf("increment project email daily quota usage: %w", err)
	}

	emailQuotaDaily := defaultEmailQuotaDaily
	if qProject.EmailQuotaDaily != nil {
		emailQuotaDaily = *qProject.EmailQuotaDaily
	}

	slog.InfoContext(ctx, "email_daily_quota_usage", "usage", qEmailDailyQuotaUsage.QuotaUsage, "quota", emailQuotaDaily)

	if qEmailDailyQuotaUsage.QuotaUsage > emailQuotaDaily {
		slog.InfoContext(ctx, "email_daily_quota_exceeded")
		return nil, apierror.NewFailedPreconditionError("email daily quota exceeded", fmt.Errorf("email daily quota exceeded"))
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	if err := s.sendPasswordResetCode(ctx, *qIntermediateSession.Email, idformat.PasswordResetCode.Format(passwordResetCodeUUID)); err != nil {
		return nil, fmt.Errorf("send password reset code: %w", err)
	}

	return &intermediatev1.IssuePasswordResetCodeResponse{}, nil
}

var passwordResetEmailBodyTmpl = template.Must(template.New("emailVerificationEmailBody").Parse(`Hello,

Someone has requested a password reset for your {{ .ProjectDisplayName }} account. If you did not request this, please ignore this email.

To continue logging in to {{ .ProjectDisplayName }}, please go back to the "Forgot password" page and enter this verification code:

{{ .PasswordResetCode }}

If you did not request this verification, please ignore this email.
`))

func (s *Store) sendPasswordResetCode(ctx context.Context, toAddress string, passwordResetCode string) error {
	qProject, err := s.q.GetProjectByID(ctx, authn.ProjectID(ctx))
	if err != nil {
		return fmt.Errorf("get project by id: %w", err)
	}

	subject := fmt.Sprintf("%s - Reset password", qProject.DisplayName)

	var body bytes.Buffer
	if err := passwordResetEmailBodyTmpl.Execute(&body, struct {
		ProjectDisplayName string
		PasswordResetCode  string
	}{
		ProjectDisplayName: qProject.DisplayName,
		PasswordResetCode:  passwordResetCode,
	}); err != nil {
		return fmt.Errorf("execute password reset email body template: %w", err)
	}

	if _, err := s.ses.SendEmail(ctx, &sesv2.SendEmailInput{
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data: &subject,
				},
				Body: &types.Body{
					Text: &types.Content{
						Data: aws.String(body.String()),
					},
				},
			},
		},
		Destination: &types.Destination{
			ToAddresses: []string{toAddress},
		},
		FromEmailAddress: aws.String(fmt.Sprintf("noreply@%s", qProject.EmailSendFromDomain)),
	}); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}

func (s *Store) VerifyPasswordResetCode(ctx context.Context, req *intermediatev1.VerifyPasswordResetCodeRequest) (*intermediatev1.VerifyPasswordResetCodeResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	qIntermediateSession, err := q.GetIntermediateSessionByID(ctx, authn.IntermediateSessionID(ctx))
	if err != nil {
		return nil, fmt.Errorf("get intermediate session by id: %w", err)
	}

	passwordResetCodeUUID, err := idformat.PasswordResetCode.Parse(req.PasswordResetCode)
	if err != nil {
		return nil, apierror.NewInvalidArgumentError("invalid password reset code", fmt.Errorf("parse password reset code: %w", err))
	}

	passwordResetCodeSHA256 := sha256.Sum256(passwordResetCodeUUID[:])
	if !bytes.Equal(passwordResetCodeSHA256[:], qIntermediateSession.PasswordResetCodeSha256) {
		return nil, apierror.NewFailedPreconditionError("invalid password reset code", fmt.Errorf("invalid password reset code"))
	}

	if _, err := q.UpdateIntermediateSessionPasswordResetCodeVerified(ctx, authn.IntermediateSessionID(ctx)); err != nil {
		return nil, fmt.Errorf("update intermediate session password reset code verified: %w", err)
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &intermediatev1.VerifyPasswordResetCodeResponse{}, nil
}
