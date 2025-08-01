import { timestampDate } from "@bufbuild/protobuf/wkt";
import { useMutation, useQuery } from "@connectrpc/connect-query";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  ArrowLeft,
  Ban,
  LoaderCircle,
  Trash,
  TriangleAlert,
} from "lucide-react";
import { DateTime } from "luxon";
import React, { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { Link, useNavigate, useParams } from "react-router";
import { toast } from "sonner";
import { z } from "zod";

import { Title } from "@/components/core/Title";
import { ValueCopier } from "@/components/core/ValueCopier";
import { PageContent } from "@/components/page";
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import {
  deleteSCIMAPIKey,
  getSCIMAPIKey,
  revokeSCIMAPIKey,
  updateSCIMAPIKey,
} from "@/gen/tesseral/frontend/v1/frontend-FrontendService_connectquery";

const schema = z.object({
  displayName: z.string().min(1, "Display name is required"),
});

export function ScimApiKeyPage() {
  const { scimApiKeyId } = useParams();

  const { data: getScimApiKeyResponse, refetch } = useQuery(
    getSCIMAPIKey,
    {
      id: scimApiKeyId,
    },
    {
      retry: 3,
    },
  );
  const updateScimApiKeyMutation = useMutation(updateSCIMAPIKey);

  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: {
      displayName: getScimApiKeyResponse?.scimApiKey?.displayName || "",
    },
  });

  async function handleSubmit(data: z.infer<typeof schema>) {
    await updateScimApiKeyMutation.mutateAsync({
      id: scimApiKeyId,
      scimApiKey: {
        displayName: data.displayName,
      },
    });
    await refetch();
    form.reset(data);
    toast.success("SCIM API Key updated successfully");
  }

  useEffect(() => {
    if (getScimApiKeyResponse?.scimApiKey) {
      form.reset({
        displayName: getScimApiKeyResponse.scimApiKey.displayName,
      });
    }
  }, [getScimApiKeyResponse, form]);

  return (
    <PageContent>
      <Title title={`SCIM API Key ${scimApiKeyId}`} />

      <div>
        <Link to={`/organization/authentication`}>
          <Button variant="ghost" size="sm">
            <ArrowLeft />
            Back to Authentication
          </Button>
        </Link>
      </div>

      <div>
        <div>
          <h1 className="text-2xl font-semibold">
            {getScimApiKeyResponse?.scimApiKey?.displayName || "SCIM API Key"}
          </h1>
          <ValueCopier
            value={getScimApiKeyResponse?.scimApiKey?.id || ""}
            label="SCIM API Key ID"
          />
          <div className="flex flex-wrap mt-2 gap-2 text-muted-foreground/50">
            {getScimApiKeyResponse?.scimApiKey?.revoked ? (
              <Badge variant="secondary">Revoked</Badge>
            ) : (
              <Badge>Active</Badge>
            )}
            <Badge className="border-0" variant="outline">
              Created{" "}
              {getScimApiKeyResponse?.scimApiKey?.createTime &&
                DateTime.fromJSDate(
                  timestampDate(getScimApiKeyResponse.scimApiKey.createTime),
                ).toRelative()}
            </Badge>
            <div>•</div>
            <Badge className="border-0" variant="outline">
              Updated{" "}
              {getScimApiKeyResponse?.scimApiKey?.updateTime &&
                DateTime.fromJSDate(
                  timestampDate(getScimApiKeyResponse.scimApiKey.updateTime),
                ).toRelative()}
            </Badge>
          </div>
        </div>
      </div>

      <Form {...form}>
        <form onSubmit={form.handleSubmit(handleSubmit)}>
          <Card>
            <CardHeader>
              <CardTitle>SCIM API Key Details</CardTitle>
              <CardDescription>
                Update basic information about this SCIM API Key.
              </CardDescription>
              <CardAction>
                <Button
                  type="submit"
                  disabled={
                    !form.formState.isDirty ||
                    updateScimApiKeyMutation.isPending
                  }
                >
                  {updateScimApiKeyMutation.isPending && (
                    <LoaderCircle className="animate-spin" />
                  )}
                  {updateScimApiKeyMutation.isPending
                    ? "Saving changes"
                    : "Save changes"}
                </Button>
              </CardAction>
            </CardHeader>
            <CardContent className="space-y-6">
              <FormField
                control={form.control}
                name="displayName"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Display Name</FormLabel>
                    <FormDescription>
                      The human-friendly name for this SCIM API Key.
                    </FormDescription>
                    <FormMessage />
                    <FormControl>
                      <Input
                        className="max-w-2xl"
                        placeholder="Display name"
                        {...field}
                      />
                    </FormControl>
                  </FormItem>
                )}
              />
            </CardContent>
          </Card>
        </form>
      </Form>

      <Card>
        <CardHeader>
          <CardTitle>Service Provider Details</CardTitle>
          <CardDescription>
            The configuration here needs to be inputted into your Identity
            Provider.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <div className="font-medium text-sm">SCIM Base URL</div>
            <ValueCopier
              value={`https://${location.host}/api/scim/v1`}
              label="SCIM Base URL"
            />
          </div>
        </CardContent>
      </Card>

      <DangerZoneCard />
    </PageContent>
  );
}

function DangerZoneCard() {
  const { scimApiKeyId } = useParams();
  const navigate = useNavigate();

  const { data: getScimApiKeyResponse, refetch } = useQuery(getSCIMAPIKey, {
    id: scimApiKeyId,
  });
  const revokeScimApiKeyMutation = useMutation(revokeSCIMAPIKey);
  const deleteScimApiKeyMutation = useMutation(deleteSCIMAPIKey);

  const [deleteOpen, setDeleteOpen] = useState(false);
  const [revokeOpen, setRevokeOpen] = useState(false);

  async function handleRevoke() {
    await revokeScimApiKeyMutation.mutateAsync({
      id: scimApiKeyId,
    });
    toast.success("SCIM API Key revoked successfully");
    setRevokeOpen(false);
    await refetch();
  }

  async function handleDelete() {
    await deleteScimApiKeyMutation.mutateAsync({
      id: scimApiKeyId,
    });
    toast.success("SCIM API Key deleted successfully");
    navigate(`/organization/authentication`);
  }

  return (
    <>
      <Card className="bg-red-50/50 border-red-200 dark:bg-red-900/40 dark:border-red-700">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-destructive">
            <TriangleAlert className="h-4 w-4" />
            <span>Danger Zone</span>
          </CardTitle>
          <CardDescription>
            This section contains actions that can have significant
            consequences. Proceed with caution.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between gap-8 w-full lg:w-auto flex-wrap lg:flex-nowrap">
            <div className="space-y-1">
              <div className="text-sm font-semibold flex items-center gap-2">
                <Ban className="w-4 h-4" />
                <span>Revoke SCIM API Key</span>
              </div>
              <div className="text-sm text-muted-foreground">
                Revoke the SCIM API Key. This cannot be undone.
              </div>
            </div>
            <Button
              variant="destructive"
              size="sm"
              onClick={() => setRevokeOpen(true)}
              disabled={getScimApiKeyResponse?.scimApiKey?.revoked}
            >
              Revoke SCIM API Key
            </Button>
          </div>
          <div className="flex items-center justify-between gap-8 w-full lg:w-auto flex-wrap lg:flex-nowrap">
            <div className="space-y-1">
              <div className="text-sm font-semibold flex items-center gap-2">
                <Trash className="w-4 h-4" />
                <span>Delete SCIM API Key</span>
              </div>
              <div className="text-sm text-muted-foreground">
                Completely delete the SCIM API Key. This cannot be undone.
              </div>
            </div>
            <Button
              variant="destructive"
              size="sm"
              onClick={() => setDeleteOpen(true)}
              disabled={!getScimApiKeyResponse?.scimApiKey?.revoked}
            >
              Delete SCIM API Key
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Revoke Confirmation Dialog */}
      <AlertDialog open={revokeOpen} onOpenChange={setRevokeOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle className="flex items-center gap-2">
              <TriangleAlert />
              Are you sure?
            </AlertDialogTitle>
            <AlertDialogDescription>
              This will permanently revoke the{" "}
              <span className="font-semibold">
                {getScimApiKeyResponse?.scimApiKey?.displayName ||
                  getScimApiKeyResponse?.scimApiKey?.id}
              </span>{" "}
              SCIM API Key. This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <Button variant="outline" onClick={() => setRevokeOpen(false)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleRevoke}>
              Revoke SCIM API Key
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle className="flex items-center gap-2">
              <TriangleAlert />
              Are you sure?
            </AlertDialogTitle>
            <AlertDialogDescription>
              This will permanently delete the{" "}
              <span className="font-semibold">
                {getScimApiKeyResponse?.scimApiKey?.displayName ||
                  getScimApiKeyResponse?.scimApiKey?.id}
              </span>{" "}
              SCIM API Key. This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <Button variant="outline" onClick={() => setDeleteOpen(false)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDelete}>
              Delete SCIM API Key
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
