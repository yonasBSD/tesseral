package store

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tesseral-labs/tesseral/internal/common/apierror"
	"github.com/tesseral-labs/tesseral/internal/googleoauth"
	"github.com/tesseral-labs/tesseral/internal/intermediate/authn"
	intermediatev1 "github.com/tesseral-labs/tesseral/internal/intermediate/gen/tesseral/intermediate/v1"
	"github.com/tesseral-labs/tesseral/internal/intermediate/store/queries"
)

func (s *Store) GetGoogleOAuthRedirectURL(ctx context.Context, req *intermediatev1.GetGoogleOAuthRedirectURLRequest) (*intermediatev1.GetGoogleOAuthRedirectURLResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	qProject, err := q.GetProjectByID(ctx, authn.ProjectID(ctx))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierror.NewNotFoundError("project not found", fmt.Errorf("get project by id: %v", err))
		}

		return nil, fmt.Errorf("get project by id: %v", err)
	}

	if err := enforceProjectLoginEnabled(qProject); err != nil {
		return nil, fmt.Errorf("enforce project login enabled: %w", err)
	}

	clientID := s.defaultGoogleOAuthClientID
	redirectURI := s.defaultGoogleOAuthRedirectURI

	if derefOrEmpty(qProject.GoogleOauthClientID) != "" {
		clientID = *qProject.GoogleOauthClientID
		redirectURI = req.RedirectUrl
	}

	state := uuid.NewString()
	stateSHA := sha256.Sum256([]byte(state))
	if _, err := q.UpdateIntermediateSessionGoogleOAuthStateSHA256(ctx, queries.UpdateIntermediateSessionGoogleOAuthStateSHA256Params{
		ID:                     authn.IntermediateSessionID(ctx),
		GoogleOauthStateSha256: stateSHA[:],
	}); err != nil {
		return nil, fmt.Errorf("update intermediate session google oauth state: %v", err)
	}

	url := googleoauth.GetAuthorizeURL(&googleoauth.GetAuthorizeURLRequest{
		GoogleOAuthClientID: clientID,
		RedirectURI:         redirectURI,
		State:               state,
	})

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %v", err)
	}

	return &intermediatev1.GetGoogleOAuthRedirectURLResponse{
		Url: url,
	}, nil
}

func (s *Store) RedeemGoogleOAuthCode(ctx context.Context, req *intermediatev1.RedeemGoogleOAuthCodeRequest) (*intermediatev1.RedeemGoogleOAuthCodeResponse, error) {
	qProject, qIntermediateSession, err := s.getProjectAndIntermediateSession(ctx)
	if err != nil {
		return nil, fmt.Errorf("get project and intermediate session: %v", err)
	}

	if err := enforceProjectLoginEnabled(*qProject); err != nil {
		return nil, fmt.Errorf("enforce project login enabled: %w", err)
	}

	clientID := s.defaultGoogleOAuthClientID
	clientSecret := s.defaultGoogleOAuthClientSecret
	redirectURI := s.defaultGoogleOAuthRedirectURI

	if derefOrEmpty(qProject.GoogleOauthClientID) != "" {
		clientID = *qProject.GoogleOauthClientID
		redirectURI = req.RedirectUrl

		decryptRes, err := s.kms.Decrypt(ctx, &kms.DecryptInput{
			CiphertextBlob:      qProject.GoogleOauthClientSecretCiphertext,
			EncryptionAlgorithm: types.EncryptionAlgorithmSpecRsaesOaepSha256,
			KeyId:               &s.googleOAuthClientSecretsKMSKeyID,
		})
		if err != nil {
			return nil, fmt.Errorf("decrypt google oauth client secret: %v", err)
		}

		clientSecret = string(decryptRes.Plaintext)
	}

	stateSHA := sha256.Sum256([]byte(req.State))
	if !bytes.Equal(qIntermediateSession.GoogleOauthStateSha256, stateSHA[:]) {
		return nil, apierror.NewInvalidArgumentError("invalid state", fmt.Errorf("invalid state"))
	}

	redeemRes, err := s.googleOAuthClient.RedeemCode(ctx, &googleoauth.RedeemCodeRequest{
		GoogleOAuthClientID:     clientID,
		GoogleOAuthClientSecret: clientSecret,
		RedirectURI:             redirectURI,
		Code:                    req.Code,
	})
	if err != nil {
		return nil, apierror.NewInvalidArgumentError("failed to redeem google oauth code", fmt.Errorf("redeem google oauth code: %v", err))
	}

	if qIntermediateSession.Email != nil && redeemRes.Email != *qIntermediateSession.Email {
		return nil, apierror.NewInvalidArgumentError("Email mismatch", fmt.Errorf("email mismatch"))
	}

	// start new tx now that all kms and oauth http i/o is done
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	if _, err := q.UpdateIntermediateSessionGoogleDetails(ctx, queries.UpdateIntermediateSessionGoogleDetailsParams{
		ID:                 authn.IntermediateSessionID(ctx),
		Email:              &redeemRes.Email,
		GoogleUserID:       &redeemRes.GoogleUserID,
		GoogleHostedDomain: refOrNil(redeemRes.GoogleHostedDomain),
		UserDisplayName:    refOrNil(redeemRes.DisplayName),
		ProfilePictureUrl:  refOrNil(redeemRes.ProfilePictureURL),
	}); err != nil {
		return nil, fmt.Errorf("update intermediate session google details: %v", err)
	}

	if redeemRes.EmailVerified {
		if _, err := q.CreateVerifiedEmail(ctx, queries.CreateVerifiedEmailParams{
			ID:           uuid.New(),
			ProjectID:    authn.ProjectID(ctx),
			Email:        redeemRes.Email,
			GoogleUserID: &redeemRes.GoogleUserID,
		}); err != nil {
			return nil, fmt.Errorf("create verified email: %v", err)
		}
	}

	primaryAuthFactor := queries.PrimaryAuthFactorGoogle
	if _, err := q.UpdateIntermediateSessionPrimaryAuthFactor(ctx, queries.UpdateIntermediateSessionPrimaryAuthFactorParams{
		ID:                authn.IntermediateSessionID(ctx),
		PrimaryAuthFactor: &primaryAuthFactor,
	}); err != nil {
		return nil, fmt.Errorf("update intermediate session primary auth factor: %v", err)
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &intermediatev1.RedeemGoogleOAuthCodeResponse{}, nil
}

func (s *Store) getProjectAndIntermediateSession(ctx context.Context) (*queries.Project, *queries.IntermediateSession, error) {
	// this function exists to avoid doing non-database i/o operations mid-tx
	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer rollback()

	qProject, err := q.GetProjectByID(ctx, authn.ProjectID(ctx))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, apierror.NewNotFoundError("project not found", fmt.Errorf("get project by id: %w", err))
		}

		return nil, nil, fmt.Errorf("get project by id: %v", err)
	}

	qIntermediateSession, err := q.GetIntermediateSessionByID(ctx, authn.IntermediateSessionID(ctx))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, apierror.NewNotFoundError("intermediate session not found", fmt.Errorf("get intermediate session by id: %w", err))
		}

		return nil, nil, fmt.Errorf("get intermediate session by id: %v", err)
	}

	return &qProject, &qIntermediateSession, nil
}
