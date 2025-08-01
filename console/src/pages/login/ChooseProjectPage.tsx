import { useQuery } from "@connectrpc/connect-query";
import React, { useEffect } from "react";
import { useNavigate } from "react-router";
import { Link } from "react-router-dom";

import { LoginFlowCard } from "@/components/login/LoginFlowCard";
import { Title } from "@/components/page/Title";
import { Button } from "@/components/ui/button";
import { CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { listOrganizations } from "@/gen/tesseral/intermediate/v1/intermediate-IntermediateService_connectquery";

export function ChooseProjectPage() {
  const { data: listOrganizationsResponse } = useQuery(listOrganizations);

  const navigate = useNavigate();
  useEffect(() => {
    if (listOrganizationsResponse?.organizations) {
      if (listOrganizationsResponse.organizations.length === 0) {
        navigate("/create-sandbox-project");
      }
    }
  }, [listOrganizationsResponse, navigate]);

  return (
    <LoginFlowCard>
      <Title title="Choose a project" />
      <CardHeader>
        <CardTitle>Choose a project</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {listOrganizationsResponse?.organizations?.map((org) => (
            <Button key={org.id} className="w-full" variant="outline" asChild>
              <Link to={`/organizations/${org.id}/login`}>
                {org.displayName}
              </Link>
            </Button>
          ))}
        </div>

        <div className="block relative w-full cursor-default my-6">
          <div className="absolute inset-0 flex items-center border-muted-foreground">
            <span className="w-full border-t"></span>
          </div>
          <div className="relative flex justify-center text-xs uppercase">
            <span className="bg-card px-2 text-muted-foreground">or</span>
          </div>
        </div>

        <Button className="w-full" variant="outline" asChild>
          <Link to="/create-organization">Create a new project</Link>
        </Button>
      </CardContent>
    </LoginFlowCard>
  );
}
