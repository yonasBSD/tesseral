package store

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	auditlogv1 "github.com/tesseral-labs/tesseral/internal/auditlog/gen/tesseral/auditlog/v1"
	"github.com/tesseral-labs/tesseral/internal/common/apierror"
	"github.com/tesseral-labs/tesseral/internal/frontend/authn"
	frontendv1 "github.com/tesseral-labs/tesseral/internal/frontend/gen/tesseral/frontend/v1"
	"github.com/tesseral-labs/tesseral/internal/frontend/store/queries"
	"github.com/tesseral-labs/tesseral/internal/store/idformat"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Store) ListSAMLConnections(ctx context.Context, req *frontendv1.ListSAMLConnectionsRequest) (*frontendv1.ListSAMLConnectionsResponse, error) {
	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	qProject, err := q.GetProjectByID(ctx, authn.ProjectID(ctx))
	if err != nil {
		return nil, fmt.Errorf("get project by id: %w", err)
	}

	var startID uuid.UUID
	if err := s.pageEncoder.Unmarshal(req.PageToken, &startID); err != nil {
		return nil, fmt.Errorf("unmarshal page token: %w", err)
	}

	limit := 10
	qSAMLConnections, err := q.ListSAMLConnections(ctx, queries.ListSAMLConnectionsParams{
		OrganizationID: authn.OrganizationID(ctx),
		ID:             startID,
		Limit:          int32(limit + 1),
	})
	if err != nil {
		return nil, fmt.Errorf("list saml connections: %w", err)
	}

	var samlConnections []*frontendv1.SAMLConnection
	for _, qSAMLConn := range qSAMLConnections {
		samlConnections = append(samlConnections, parseSAMLConnection(qProject, qSAMLConn))
	}

	var nextPageToken string
	if len(samlConnections) == limit+1 {
		nextPageToken = s.pageEncoder.Marshal(qSAMLConnections[limit].ID)
		samlConnections = samlConnections[:limit]
	}

	return &frontendv1.ListSAMLConnectionsResponse{
		SamlConnections: samlConnections,
		NextPageToken:   nextPageToken,
	}, nil
}

func (s *Store) GetSAMLConnection(ctx context.Context, req *frontendv1.GetSAMLConnectionRequest) (*frontendv1.GetSAMLConnectionResponse, error) {
	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	samlConnectionID, err := idformat.SAMLConnection.Parse(req.Id)
	if err != nil {
		return nil, apierror.NewInvalidArgumentError("invalid saml connection id", fmt.Errorf("parse saml connection id: %w", err))
	}

	qProject, err := q.GetProjectByID(ctx, authn.ProjectID(ctx))
	if err != nil {
		return nil, fmt.Errorf("get project by id: %w", err)
	}

	qSAMLConnection, err := q.GetSAMLConnection(ctx, queries.GetSAMLConnectionParams{
		OrganizationID: authn.OrganizationID(ctx),
		ID:             samlConnectionID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierror.NewNotFoundError("saml connection not found", fmt.Errorf("get saml connection: %w", err))
		}

		return nil, fmt.Errorf("get saml connection: %w", err)
	}

	return &frontendv1.GetSAMLConnectionResponse{SamlConnection: parseSAMLConnection(qProject, qSAMLConnection)}, nil
}

func (s *Store) CreateSAMLConnection(ctx context.Context, req *frontendv1.CreateSAMLConnectionRequest) (*frontendv1.CreateSAMLConnectionResponse, error) {
	if err := s.validateIsOwner(ctx); err != nil {
		return nil, fmt.Errorf("validate is owner: %w", err)
	}

	tx, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	qProject, err := q.GetProjectByID(ctx, authn.ProjectID(ctx))
	if err != nil {
		return nil, fmt.Errorf("get project by id: %w", err)
	}

	qOrg, err := q.GetOrganizationByID(ctx, authn.OrganizationID(ctx))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierror.NewNotFoundError("organization not found", err)
		}

		return nil, fmt.Errorf("get organization by id: %w", err)
	}

	if !qOrg.LogInWithSaml {
		return nil, apierror.NewFailedPreconditionError("saml is not enabled for the organization", fmt.Errorf("saml is not enabled for the organization"))
	}

	if req.SamlConnection.IdpRedirectUrl != "" {
		u, err := url.Parse(req.SamlConnection.IdpRedirectUrl)
		if err != nil {
			return nil, fmt.Errorf("invalid idp redirect url: %w", err)
		}

		if !u.IsAbs() {
			return nil, apierror.NewFailedPreconditionError("invalid idp redirect url", fmt.Errorf("invalid idp redirect url"))
		}
	}

	var idpCertificate []byte
	if req.SamlConnection.IdpX509Certificate != "" {
		block, _ := pem.Decode([]byte(req.SamlConnection.IdpX509Certificate))
		if block == nil || block.Type != "CERTIFICATE" {
			return nil, apierror.NewFailedPreconditionError("invalid idp x509 certificate", fmt.Errorf("invalid idp x509 certificate"))
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, apierror.NewFailedPreconditionError("invalid idp x509 certificate", fmt.Errorf("invalid idp x509 certificate: %w", err))
		}

		idpCertificate = cert.Raw
	}

	qSAMLConnection, err := q.CreateSAMLConnection(ctx, queries.CreateSAMLConnectionParams{
		ID:                 uuid.New(),
		OrganizationID:     authn.OrganizationID(ctx),
		IsPrimary:          derefOrEmpty(req.SamlConnection.Primary),
		IdpRedirectUrl:     &req.SamlConnection.IdpRedirectUrl,
		IdpX509Certificate: idpCertificate,
		IdpEntityID:        &req.SamlConnection.IdpEntityId,
	})
	if err != nil {
		return nil, fmt.Errorf("create saml connection: %w", err)
	}

	if req.SamlConnection.GetPrimary() {
		if err := q.UpdatePrimarySAMLConnection(ctx, queries.UpdatePrimarySAMLConnectionParams{
			OrganizationID: authn.OrganizationID(ctx),
			ID:             qSAMLConnection.ID,
		}); err != nil {
			return nil, fmt.Errorf("update primary saml connection: %w", err)
		}
	}

	samlConnection := parseSAMLConnection(qProject, qSAMLConnection)

	auditSAMLConnection, err := s.auditlogStore.GetSAMLConnection(ctx, tx, qSAMLConnection.ID)
	if err != nil {
		return nil, fmt.Errorf("get audit saml connection: %w", err)
	}

	if _, err := s.logAuditEvent(ctx, q, logAuditEventParams{
		EventName: "tesseral.saml_connections.create",
		EventDetails: &auditlogv1.CreateSAMLConnection{
			SamlConnection: auditSAMLConnection,
		},
		ResourceType: queries.AuditLogEventResourceTypeSamlConnection,
		ResourceID:   &qSAMLConnection.ID,
	}); err != nil {
		return nil, fmt.Errorf("create audit log event: %w", err)
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &frontendv1.CreateSAMLConnectionResponse{SamlConnection: samlConnection}, nil
}

func (s *Store) UpdateSAMLConnection(ctx context.Context, req *frontendv1.UpdateSAMLConnectionRequest) (*frontendv1.UpdateSAMLConnectionResponse, error) {
	if err := s.validateIsOwner(ctx); err != nil {
		return nil, fmt.Errorf("validate is owner: %w", err)
	}

	tx, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	qProject, err := q.GetProjectByID(ctx, authn.ProjectID(ctx))
	if err != nil {
		return nil, fmt.Errorf("get project by id: %w", err)
	}

	samlConnectionID, err := idformat.SAMLConnection.Parse(req.Id)
	if err != nil {
		return nil, apierror.NewInvalidArgumentError("invalid saml connection id", fmt.Errorf("parse saml connection id: %w", err))
	}

	// authz
	qSAMLConnection, err := q.GetSAMLConnection(ctx, queries.GetSAMLConnectionParams{
		OrganizationID: authn.OrganizationID(ctx),
		ID:             samlConnectionID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierror.NewNotFoundError("saml connection not found", fmt.Errorf("get saml connection: %w", err))
		}

		return nil, fmt.Errorf("get saml connection: %w", err)
	}

	auditPreviousSAMLConnection, err := s.auditlogStore.GetSAMLConnection(ctx, tx, qSAMLConnection.ID)
	if err != nil {
		return nil, fmt.Errorf("get audit previous saml connection: %w", err)
	}

	updates := queries.UpdateSAMLConnectionParams{
		ID:                 samlConnectionID,
		IsPrimary:          qSAMLConnection.IsPrimary,
		IdpRedirectUrl:     qSAMLConnection.IdpRedirectUrl,
		IdpX509Certificate: qSAMLConnection.IdpX509Certificate,
		IdpEntityID:        qSAMLConnection.IdpEntityID,
	}

	if req.SamlConnection.IdpRedirectUrl != "" {
		u, err := url.Parse(req.SamlConnection.IdpRedirectUrl)
		if err != nil {
			return nil, apierror.NewFailedPreconditionError("invalid idp redirect url", fmt.Errorf("invalid idp redirect url: %w", err))
		}

		if !u.IsAbs() {
			return nil, apierror.NewFailedPreconditionError("invalid ipd redirect url", fmt.Errorf("invalid idp redirect url"))
		}

		updates.IdpRedirectUrl = &req.SamlConnection.IdpRedirectUrl
	}

	if req.SamlConnection.IdpX509Certificate != "" {
		block, _ := pem.Decode([]byte(req.SamlConnection.IdpX509Certificate))
		if block == nil || block.Type != "CERTIFICATE" {
			return nil, apierror.NewFailedPreconditionError("invalid idp x509 certificate", fmt.Errorf("invalid idp x509 certificate"))
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, apierror.NewFailedPreconditionError("invalid idp x509 certificate", fmt.Errorf("invalid idp x509 certificate: %w", err))
		}

		updates.IdpX509Certificate = cert.Raw
	}

	if req.SamlConnection.IdpEntityId != "" {
		updates.IdpEntityID = &req.SamlConnection.IdpEntityId
	}

	if req.SamlConnection.Primary != nil {
		updates.IsPrimary = *req.SamlConnection.Primary
	}

	qUpdatedSAMLConnection, err := q.UpdateSAMLConnection(ctx, updates)
	if err != nil {
		return nil, fmt.Errorf("update saml connection: %w", err)
	}

	if req.SamlConnection.GetPrimary() {
		if err := q.UpdatePrimarySAMLConnection(ctx, queries.UpdatePrimarySAMLConnectionParams{
			OrganizationID: qSAMLConnection.OrganizationID,
			ID:             samlConnectionID,
		}); err != nil {
			return nil, fmt.Errorf("update primary saml connection: %w", err)
		}
	}

	auditSAMLConnection, err := s.auditlogStore.GetSAMLConnection(ctx, tx, qUpdatedSAMLConnection.ID)
	if err != nil {
		return nil, fmt.Errorf("get audit saml connection: %w", err)
	}

	if _, err := s.logAuditEvent(ctx, q, logAuditEventParams{
		EventName: "tesseral.saml_connections.update",
		EventDetails: &auditlogv1.UpdateSAMLConnection{
			SamlConnection:         auditSAMLConnection,
			PreviousSamlConnection: auditPreviousSAMLConnection,
		},
		ResourceType: queries.AuditLogEventResourceTypeSamlConnection,
		ResourceID:   &qSAMLConnection.ID,
	}); err != nil {
		return nil, fmt.Errorf("create audit log event: %w", err)
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &frontendv1.UpdateSAMLConnectionResponse{SamlConnection: parseSAMLConnection(qProject, qUpdatedSAMLConnection)}, nil
}

func (s *Store) DeleteSAMLConnection(ctx context.Context, req *frontendv1.DeleteSAMLConnectionRequest) (*frontendv1.DeleteSAMLConnectionResponse, error) {
	if err := s.validateIsOwner(ctx); err != nil {
		return nil, fmt.Errorf("validate is owner: %w", err)
	}

	tx, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	samlConnectionID, err := idformat.SAMLConnection.Parse(req.Id)
	if err != nil {
		return nil, apierror.NewInvalidArgumentError("invalid saml connection id", fmt.Errorf("parse saml connection id: %w", err))
	}

	// authz
	qSAMLConnection, err := q.GetSAMLConnection(ctx, queries.GetSAMLConnectionParams{
		OrganizationID: authn.OrganizationID(ctx),
		ID:             samlConnectionID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierror.NewNotFoundError("saml connection not found", fmt.Errorf("get saml connection: %w", err))
		}

		return nil, fmt.Errorf("get saml connection: %w", err)
	}

	auditSAMLConnection, err := s.auditlogStore.GetSAMLConnection(ctx, tx, qSAMLConnection.ID)
	if err != nil {
		return nil, fmt.Errorf("get audit saml connection: %w", err)
	}

	if err := q.DeleteSAMLConnection(ctx, samlConnectionID); err != nil {
		return nil, fmt.Errorf("delete saml connection: %w", err)
	}

	if _, err := s.logAuditEvent(ctx, q, logAuditEventParams{
		EventName: "tesseral.saml_connections.delete",
		EventDetails: &auditlogv1.DeleteSAMLConnection{
			SamlConnection: auditSAMLConnection,
		},
		ResourceType: queries.AuditLogEventResourceTypeSamlConnection,
		ResourceID:   &qSAMLConnection.ID,
	}); err != nil {
		return nil, fmt.Errorf("create audit log event: %w", err)
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &frontendv1.DeleteSAMLConnectionResponse{}, nil
}

func parseSAMLConnection(qProject queries.Project, qSAMLConnection queries.SamlConnection) *frontendv1.SAMLConnection {
	var certPEM string
	if len(qSAMLConnection.IdpX509Certificate) != 0 {
		cert, err := x509.ParseCertificate(qSAMLConnection.IdpX509Certificate)
		if err != nil {
			panic(err)
		}

		certPEM = string(pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		}))
	}

	spACSURL := fmt.Sprintf("https://%s/api/saml/v1/%s/acs", qProject.VaultDomain, idformat.SAMLConnection.Format(qSAMLConnection.ID))
	spEntityID := fmt.Sprintf("https://%s/api/saml/v1/%s", qProject.VaultDomain, idformat.SAMLConnection.Format(qSAMLConnection.ID))

	return &frontendv1.SAMLConnection{
		Id:                 idformat.SAMLConnection.Format(qSAMLConnection.ID),
		CreateTime:         timestamppb.New(*qSAMLConnection.CreateTime),
		UpdateTime:         timestamppb.New(*qSAMLConnection.UpdateTime),
		Primary:            &qSAMLConnection.IsPrimary,
		SpAcsUrl:           spACSURL,
		SpEntityId:         spEntityID,
		IdpRedirectUrl:     derefOrEmpty(qSAMLConnection.IdpRedirectUrl),
		IdpX509Certificate: certPEM,
		IdpEntityId:        derefOrEmpty(qSAMLConnection.IdpEntityID),
	}
}
