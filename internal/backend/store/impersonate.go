package store

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tesseral-labs/tesseral/internal/backend/authn"
	backendv1 "github.com/tesseral-labs/tesseral/internal/backend/gen/tesseral/backend/v1"
	"github.com/tesseral-labs/tesseral/internal/backend/store/queries"
	"github.com/tesseral-labs/tesseral/internal/common/apierror"
	"github.com/tesseral-labs/tesseral/internal/store/idformat"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const userImpersonationTokenDuration = time.Second * 30

func (s *Store) CreateUserImpersonationToken(ctx context.Context, req *backendv1.CreateUserImpersonationTokenRequest) (*backendv1.CreateUserImpersonationTokenResponse, error) {
	if err := validateIsConsoleSession(ctx); err != nil {
		return nil, fmt.Errorf("validate is console session: %w", err)
	}

	impersonatorID, err := idformat.User.Parse(authn.GetContextData(ctx).ConsoleSession.UserID)
	if err != nil {
		panic(fmt.Errorf("parse user id: %w", err))
	}

	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	qImpersonator, err := q.GetUser(ctx, queries.GetUserParams{
		ProjectID: *s.consoleProjectID,
		ID:        impersonatorID,
	})
	if err != nil {
		return nil, fmt.Errorf("get impersonator: %w", err)
	}

	if !qImpersonator.IsOwner {
		return nil, apierror.NewPermissionDeniedError("only owners may impersonate others", fmt.Errorf("impersonator is not an owner"))
	}

	impersonatedID, err := idformat.User.Parse(req.UserImpersonationToken.ImpersonatedId)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}

	// Ensure the target user belongs to the current project.
	if _, err := q.GetUser(ctx, queries.GetUserParams{
		ProjectID: authn.ProjectID(ctx),
		ID:        impersonatedID,
	}); err != nil {
		return nil, fmt.Errorf("get impersonated user: %w", err)
	}

	secretToken := uuid.New()
	secretTokenSHA256 := sha256.Sum256(secretToken[:])

	expireTime := time.Now().Add(userImpersonationTokenDuration)
	qUserImpersonationToken, err := q.CreateUserImpersonationToken(ctx, queries.CreateUserImpersonationTokenParams{
		ID:                uuid.New(),
		ImpersonatorID:    impersonatorID,
		ImpersonatedID:    impersonatedID,
		ExpireTime:        &expireTime,
		SecretTokenSha256: secretTokenSHA256[:],
	})
	if err != nil {
		return nil, fmt.Errorf("create user impersonation token: %w", err)
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	userImpersonationToken := parseUserImpersonationToken(qUserImpersonationToken)
	userImpersonationToken.SecretToken = idformat.UserImpersonationSecretToken.Format(secretToken)
	return &backendv1.CreateUserImpersonationTokenResponse{
		UserImpersonationToken: userImpersonationToken,
	}, nil
}

func parseUserImpersonationToken(qUserImpersonationToken queries.UserImpersonationToken) *backendv1.UserImpersonationToken {
	return &backendv1.UserImpersonationToken{
		Id:             idformat.UserImpersonationToken.Format(qUserImpersonationToken.ID),
		ImpersonatorId: idformat.User.Format(qUserImpersonationToken.ImpersonatorID),
		CreateTime:     timestamppb.New(*qUserImpersonationToken.CreateTime),
		ExpireTime:     timestamppb.New(*qUserImpersonationToken.ExpireTime),
		SecretToken:    "", // intentionally left blank
		ImpersonatedId: idformat.User.Format(qUserImpersonationToken.ImpersonatedID),
	}
}
