package service

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/tesseral-labs/tesseral/internal/intermediate/authn"
	intermediatev1 "github.com/tesseral-labs/tesseral/internal/intermediate/gen/tesseral/intermediate/v1"
)

func (s *Service) ExchangeIntermediateSessionForSession(ctx context.Context, req *connect.Request[intermediatev1.ExchangeIntermediateSessionForSessionRequest]) (*connect.Response[intermediatev1.ExchangeIntermediateSessionForSessionResponse], error) {
	res, err := s.Store.ExchangeIntermediateSessionForSession(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	accessToken, err := s.AccessTokenIssuer.NewAccessToken(ctx, authn.ProjectID(ctx), res.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("issue access token: %w", err)
	}

	res.AccessToken = accessToken

	expiredIntermediateAccessTokenCookie, err := s.Cookier.ExpiredIntermediateAccessToken(ctx, authn.ProjectID(ctx))
	if err != nil {
		return nil, fmt.Errorf("create expired intermediate access token cookie: %w", err)
	}

	refreshTokenCookie, err := s.Cookier.NewRefreshToken(ctx, authn.ProjectID(ctx), res.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("issue refresh token cookie: %w", err)
	}

	accessTokenCookie, err := s.Cookier.NewAccessToken(ctx, authn.ProjectID(ctx), accessToken)
	if err != nil {
		return nil, fmt.Errorf("issue access token cookie: %w", err)
	}

	connectRes := connect.NewResponse(res)
	connectRes.Header().Add("Set-Cookie", expiredIntermediateAccessTokenCookie)
	connectRes.Header().Add("Set-Cookie", refreshTokenCookie)
	connectRes.Header().Add("Set-Cookie", accessTokenCookie)
	return connectRes, nil
}
