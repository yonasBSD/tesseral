import { useQuery } from "@connectrpc/connect-query";
import React from "react";
import { useParams } from "react-router";

import {
  getOrganization,
  getProject,
} from "@/gen/tesseral/backend/v1/backend-BackendService_connectquery";

import { ListOrganizationOidcConnectionsCard } from "./authentication/ListOrganizationOidcConnectionsCard";
import { ListOrganizationSamlConnectionsCard } from "./authentication/ListOrganizationSamlConnectionsCard";
import { ListOrganizationScimApiKeysCard } from "./authentication/ListOrganizationScimApiKeysCard";
import { OrganizationBasicAuthCard } from "./authentication/OrganizationBasicAuthCard";
import { OrganizationMFACard } from "./authentication/OrganizationMfaCard";
import { OrganizationOAuthCard } from "./authentication/OrganizationOauthCard";
import { OrganizationScimCard } from "./authentication/OrganizationScimCard";
import { OrganizationSsoCard } from "./authentication/OrganizationSsoCard";

export function OrganizationAuthentication() {
  const { organizationId } = useParams();

  const { data: getOrganizationResponse } = useQuery(getOrganization, {
    id: organizationId,
  });
  const { data: getProjectResponse } = useQuery(getProject);

  return (
    <div className="space-y-8">
      <div className="grid grid-cols-1 gap-8 lg:grid-cols-3">
        {(getProjectResponse?.project?.logInWithEmail ||
          getProjectResponse?.project?.logInWithPassword) && (
          <OrganizationBasicAuthCard />
        )}

        {(getProjectResponse?.project?.logInWithGoogle ||
          getProjectResponse?.project?.logInWithGithub ||
          getProjectResponse?.project?.logInWithMicrosoft) && (
          <OrganizationOAuthCard />
        )}

        {(getProjectResponse?.project?.logInWithAuthenticatorApp ||
          getProjectResponse?.project?.logInWithPasskey) && (
          <OrganizationMFACard />
        )}

        {(getProjectResponse?.project?.logInWithSaml ||
          getProjectResponse?.project?.logInWithOidc) && (
          <OrganizationSsoCard />
        )}

        <OrganizationScimCard />
      </div>

      {getProjectResponse?.project?.logInWithSaml &&
        getOrganizationResponse?.organization?.logInWithSaml && (
          <ListOrganizationSamlConnectionsCard />
        )}
      {getProjectResponse?.project?.logInWithOidc &&
        getOrganizationResponse?.organization?.logInWithOidc && (
          <ListOrganizationOidcConnectionsCard />
        )}
      {getOrganizationResponse?.organization?.scimEnabled && (
        <ListOrganizationScimApiKeysCard />
      )}
    </div>
  );
}
