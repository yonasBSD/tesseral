package store

import (
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	backendv1 "github.com/tesseral-labs/tesseral/internal/backend/gen/tesseral/backend/v1"
	"github.com/tesseral-labs/tesseral/internal/store/idformat"
)

func TestCreateSAMLConnection_SAMLEnabled(t *testing.T) {
	t.Parallel()

	ctx, u := newTestUtil(t)

	organizationID := u.Environment.NewOrganization(t, u.ProjectID, &backendv1.Organization{
		DisplayName:   "test",
		LogInWithSaml: refOrNil(true),
	})

	res, err := u.Store.CreateSAMLConnection(ctx, &backendv1.CreateSAMLConnectionRequest{
		SamlConnection: &backendv1.SAMLConnection{
			IdpRedirectUrl: "https://idp.example.com/saml/redirect",
			IdpEntityId:    "https://idp.example.com/saml/idp",
			OrganizationId: organizationID,
			Primary:        refOrNil(true),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, res.SamlConnection)
	require.NotEmpty(t, res.SamlConnection.SpAcsUrl)
	require.NotEmpty(t, res.SamlConnection.SpEntityId)
	require.Equal(t, "https://idp.example.com/saml/redirect", res.SamlConnection.IdpRedirectUrl)
	require.Equal(t, "https://idp.example.com/saml/idp", res.SamlConnection.IdpEntityId)
	require.Equal(t, organizationID, res.SamlConnection.OrganizationId)
	require.NotEmpty(t, res.SamlConnection.CreateTime)
	require.NotEmpty(t, res.SamlConnection.UpdateTime)
	require.True(t, res.SamlConnection.GetPrimary())
}

func TestCreateSAMLConnection_SAMLDisabled(t *testing.T) {
	t.Parallel()

	ctx, u := newTestUtil(t)

	organizationID := u.Environment.NewOrganization(t, u.ProjectID, &backendv1.Organization{
		DisplayName:   "test",
		LogInWithSaml: refOrNil(false), // SAML not enabled
	})

	_, err := u.Store.CreateSAMLConnection(ctx, &backendv1.CreateSAMLConnectionRequest{
		SamlConnection: &backendv1.SAMLConnection{
			IdpRedirectUrl: "https://idp.example.com/saml/redirect",
			IdpEntityId:    "https://idp.example.com/saml/idp",
			OrganizationId: organizationID,
		},
	})
	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	require.Equal(t, connect.CodeFailedPrecondition, connectErr.Code())
}

func TestGetSAMLConnection_Exists(t *testing.T) {
	t.Parallel()

	ctx, u := newTestUtil(t)

	organizationID := u.Environment.NewOrganization(t, u.ProjectID, &backendv1.Organization{
		DisplayName:   "test",
		LogInWithSaml: refOrNil(true),
	})
	samlConnection, err := u.Store.CreateSAMLConnection(ctx, &backendv1.CreateSAMLConnectionRequest{
		SamlConnection: &backendv1.SAMLConnection{
			IdpRedirectUrl: "https://idp.example.com/saml/redirect",
			IdpEntityId:    "https://idp.example.com/saml/idp",
			OrganizationId: organizationID,
		},
	})
	require.NoError(t, err, "failed to create SAML connection")

	res, err := u.Store.GetSAMLConnection(ctx, &backendv1.GetSAMLConnectionRequest{
		Id: samlConnection.SamlConnection.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, res.SamlConnection)
	require.Equal(t, samlConnection.SamlConnection.Id, res.SamlConnection.Id)
	require.NotEmpty(t, res.SamlConnection.SpAcsUrl)
	require.NotEmpty(t, res.SamlConnection.SpEntityId)
	require.Equal(t, "https://idp.example.com/saml/redirect", res.SamlConnection.IdpRedirectUrl)
	require.Equal(t, "https://idp.example.com/saml/idp", res.SamlConnection.IdpEntityId)
	require.Equal(t, organizationID, res.SamlConnection.OrganizationId)
	require.NotEmpty(t, res.SamlConnection.CreateTime)
	require.NotEmpty(t, res.SamlConnection.UpdateTime)
}

func TestGetSAMLConnection_DoesNotExist(t *testing.T) {
	t.Parallel()

	ctx, u := newTestUtil(t)

	_, err := u.Store.GetSAMLConnection(ctx, &backendv1.GetSAMLConnectionRequest{
		Id: idformat.SAMLConnection.Format(uuid.New()),
	})

	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	require.Equal(t, connect.CodeNotFound, connectErr.Code())
}

func TestUpdateSAMLConnection_UpdatesFields(t *testing.T) {
	t.Parallel()

	ctx, u := newTestUtil(t)

	organizationID := u.Environment.NewOrganization(t, u.ProjectID, &backendv1.Organization{
		DisplayName:   "test",
		LogInWithSaml: refOrNil(true),
	})
	createResp, err := u.Store.CreateSAMLConnection(ctx, &backendv1.CreateSAMLConnectionRequest{
		SamlConnection: &backendv1.SAMLConnection{
			IdpRedirectUrl: "https://idp.example.com/saml/redirect",
			IdpEntityId:    "https://idp.example.com/saml/idp",
			OrganizationId: organizationID,
		},
	})
	require.NoError(t, err)
	connID := createResp.SamlConnection.Id

	updateResp, err := u.Store.UpdateSAMLConnection(ctx, &backendv1.UpdateSAMLConnectionRequest{
		Id: connID,
		SamlConnection: &backendv1.SAMLConnection{
			IdpRedirectUrl: "https://idp.example.com/saml/redirect2",
			IdpEntityId:    "https://idp.example.com/saml/idp2",
			Primary:        refOrNil(true),
		},
	})
	require.NoError(t, err)
	updated := updateResp.SamlConnection
	require.Equal(t, "https://idp.example.com/saml/redirect2", updated.IdpRedirectUrl)
	require.Equal(t, "https://idp.example.com/saml/idp2", updated.IdpEntityId)
	require.True(t, updated.GetPrimary())
}

func TestUpdateSAMLConnection_SetPrimary(t *testing.T) {
	t.Parallel()

	ctx, u := newTestUtil(t)
	organizationID := u.Environment.NewOrganization(t, u.ProjectID, &backendv1.Organization{
		DisplayName:   "test",
		LogInWithSaml: refOrNil(true),
	})

	original, err := u.Store.CreateSAMLConnection(ctx, &backendv1.CreateSAMLConnectionRequest{
		SamlConnection: &backendv1.SAMLConnection{
			OrganizationId: organizationID,
			IdpRedirectUrl: "https://idp.example.com/saml/redirect2",
			IdpEntityId:    "https://idp.example.com/saml/idp2",
			Primary:        refOrNil(true),
		},
	})
	require.NoError(t, err)
	require.True(t, original.SamlConnection.GetPrimary())

	new, err := u.Store.CreateSAMLConnection(ctx, &backendv1.CreateSAMLConnectionRequest{
		SamlConnection: &backendv1.SAMLConnection{
			OrganizationId: organizationID,
			IdpRedirectUrl: "https://idp.example.com/saml/redirect2",
			IdpEntityId:    "https://idp.example.com/saml/idp2",
			Primary:        refOrNil(true),
		},
	})
	require.NoError(t, err)
	require.True(t, new.SamlConnection.GetPrimary())

	getResp, err := u.Store.GetSAMLConnection(ctx, &backendv1.GetSAMLConnectionRequest{Id: original.SamlConnection.Id})
	require.NoError(t, err)
	require.NotNil(t, getResp.SamlConnection)
	require.False(t, getResp.SamlConnection.GetPrimary(), "original connection should no longer be primary")
}

func TestDeleteSAMLConnection_RemovesConnection(t *testing.T) {
	t.Parallel()

	ctx, u := newTestUtil(t)

	organizationID := u.Environment.NewOrganization(t, u.ProjectID, &backendv1.Organization{
		DisplayName:   "test",
		LogInWithSaml: refOrNil(true),
	})
	createResp, err := u.Store.CreateSAMLConnection(ctx, &backendv1.CreateSAMLConnectionRequest{
		SamlConnection: &backendv1.SAMLConnection{
			IdpRedirectUrl: "https://idp.example.com/saml/redirect",
			IdpEntityId:    "https://idp.example.com/saml/idp",
			OrganizationId: organizationID,
		},
	})
	require.NoError(t, err)
	connID := createResp.SamlConnection.Id

	_, err = u.Store.DeleteSAMLConnection(ctx, &backendv1.DeleteSAMLConnectionRequest{Id: connID})
	require.NoError(t, err)

	res, err := u.Store.GetSAMLConnection(ctx, &backendv1.GetSAMLConnectionRequest{Id: connID})
	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	require.Equal(t, connect.CodeNotFound, connectErr.Code())
	require.Nil(t, res)
}

func TestCreateSAMLConnection_InvalidOrgID(t *testing.T) {
	t.Parallel()

	ctx, u := newTestUtil(t)

	_, err := u.Store.CreateSAMLConnection(ctx, &backendv1.CreateSAMLConnectionRequest{
		SamlConnection: &backendv1.SAMLConnection{
			OrganizationId: "invalid-id",
		},
	})
	require.Error(t, err)
}

func TestUpdateSAMLConnection_InvalidID(t *testing.T) {
	t.Parallel()

	ctx, u := newTestUtil(t)

	_, err := u.Store.UpdateSAMLConnection(ctx, &backendv1.UpdateSAMLConnectionRequest{
		Id:             "invalid-id",
		SamlConnection: &backendv1.SAMLConnection{},
	})
	require.Error(t, err)
}

func TestDeleteSAMLConnection_InvalidID(t *testing.T) {
	t.Parallel()

	ctx, u := newTestUtil(t)

	_, err := u.Store.DeleteSAMLConnection(ctx, &backendv1.DeleteSAMLConnectionRequest{Id: "invalid-id"})
	require.Error(t, err)
}

func TestCreateSAMLConnection_InvalidRedirectURL(t *testing.T) {
	t.Parallel()

	ctx, u := newTestUtil(t)

	organizationID := u.Environment.NewOrganization(t, u.ProjectID, &backendv1.Organization{
		DisplayName:   "test",
		LogInWithSaml: refOrNil(true),
	})
	_, err := u.Store.CreateSAMLConnection(ctx, &backendv1.CreateSAMLConnectionRequest{
		SamlConnection: &backendv1.SAMLConnection{
			IdpRedirectUrl: "not-a-url",
			OrganizationId: organizationID,
		},
	})

	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	require.Equal(t, connect.CodeInvalidArgument, connectErr.Code())
}

func TestListSAMLConnections_ReturnsAllForOrg(t *testing.T) {
	t.Parallel()
	ctx, u := newTestUtil(t)

	organizationID := u.Environment.NewOrganization(t, u.ProjectID, &backendv1.Organization{
		DisplayName:   "test",
		LogInWithSaml: refOrNil(true),
	})

	var ids []string
	for range 3 {
		resp, err := u.Store.CreateSAMLConnection(ctx, &backendv1.CreateSAMLConnectionRequest{
			SamlConnection: &backendv1.SAMLConnection{
				IdpRedirectUrl: "https://idp.example.com/saml/redirect",
				IdpEntityId:    "https://idp.example.com/saml/idp",
				OrganizationId: organizationID,
			},
		})
		require.NoError(t, err)
		ids = append(ids, resp.SamlConnection.Id)
	}

	listResp, err := u.Store.ListSAMLConnections(ctx, &backendv1.ListSAMLConnectionsRequest{
		OrganizationId: organizationID,
	})
	require.NoError(t, err)
	require.NotNil(t, listResp)
	require.Len(t, listResp.SamlConnections, 3)

	respIds := []string{}
	for _, conn := range listResp.SamlConnections {
		respIds = append(respIds, conn.Id)
	}

	require.ElementsMatch(t, ids, respIds)
}

func TestListSAMLConnections_Pagination(t *testing.T) {
	t.Parallel()
	ctx, u := newTestUtil(t)

	organizationID := u.Environment.NewOrganization(t, u.ProjectID, &backendv1.Organization{
		DisplayName:   "test",
		LogInWithSaml: refOrNil(true),
	})

	var createdIDs []string
	for range 15 {
		resp, err := u.Store.CreateSAMLConnection(ctx, &backendv1.CreateSAMLConnectionRequest{
			SamlConnection: &backendv1.SAMLConnection{
				IdpRedirectUrl: "https://idp.example.com/saml/redirect",
				IdpEntityId:    "https://idp.example.com/saml/idp",
				OrganizationId: organizationID,
			},
		})
		require.NoError(t, err)
		createdIDs = append(createdIDs, resp.SamlConnection.Id)
	}

	resp1, err := u.Store.ListSAMLConnections(ctx, &backendv1.ListSAMLConnectionsRequest{
		OrganizationId: organizationID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp1)
	require.Len(t, resp1.SamlConnections, 10)
	require.NotEmpty(t, resp1.NextPageToken)

	resp2, err := u.Store.ListSAMLConnections(ctx, &backendv1.ListSAMLConnectionsRequest{
		OrganizationId: organizationID,
		PageToken:      resp1.NextPageToken,
	})
	require.NoError(t, err)
	require.NotNil(t, resp2)
	require.Len(t, resp2.SamlConnections, 5)
	require.Empty(t, resp2.NextPageToken)

	var allIDs []string
	for _, c := range resp1.SamlConnections {
		allIDs = append(allIDs, c.Id)
	}
	for _, c := range resp2.SamlConnections {
		allIDs = append(allIDs, c.Id)
	}
	require.ElementsMatch(t, createdIDs, allIDs)
}
