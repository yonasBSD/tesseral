import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import {
  getPublishableKey,
  updatePublishableKey,
  deletePublishableKey,
} from '@/gen/tesseral/backend/v1/backend-BackendService_connectquery';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { DateTime } from 'luxon';
import { timestampDate } from '@bufbuild/protobuf/wkt';
import { toast } from 'sonner';
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { Button } from '@/components/ui/button';
import { z } from 'zod';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import {
  PageCodeSubtitle,
  PageContent,
  PageDescription,
  PageHeader,
  PageTitle,
} from '@/components/page';
import { ConsoleCardDetails } from '@/components/ui/console-card';

export const ViewPublishableKeyPage = () => {
  const { publishableKeyId } = useParams();
  const { data: getPublishableKeyResponse } = useQuery(getPublishableKey, {
    id: publishableKeyId,
  });

  return (
    <>
      <PageHeader>
        <PageTitle>
          {getPublishableKeyResponse?.publishableKey?.displayName}
        </PageTitle>
        <PageCodeSubtitle>{publishableKeyId}</PageCodeSubtitle>
        <PageDescription>
          Tesseral's client-side SDKs require a Publishable Key. Publishable
          Keys can safely be publicly accessible in your web or mobile app's
          client-side code.
        </PageDescription>
      </PageHeader>

      <PageContent>
        <Card className="my-8">
          <CardHeader className="flex-row justify-between items-center">
            <ConsoleCardDetails>
              <CardTitle>Configuration</CardTitle>
              <CardDescription>
                Details about this Publishable Key.
              </CardDescription>
            </ConsoleCardDetails>
            <EditPublishableKeyButton />
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-3 gap-x-2 text-sm">
              <div className="border-r border-gray-200 pr-8 flex flex-col gap-4">
                <div>
                  <div className="font-semibold">Display Name</div>
                  <div className="truncate">
                    {getPublishableKeyResponse?.publishableKey?.displayName}
                  </div>
                </div>
                <div>
                  <div className="font-semibold">Dev Mode</div>
                  <div className="truncate">
                    {getPublishableKeyResponse?.publishableKey?.devMode
                      ? 'Enabled'
                      : 'Disabled'}
                  </div>
                </div>
              </div>
              <div className="border-r border-gray-200 pr-8 pl-8 flex flex-col gap-4">
                <div>
                  <div className="font-semibold">Created</div>
                  <div className="truncate">
                    {getPublishableKeyResponse?.publishableKey?.createTime &&
                      DateTime.fromJSDate(
                        timestampDate(
                          getPublishableKeyResponse?.publishableKey?.createTime,
                        ),
                      ).toRelative()}
                  </div>
                </div>
              </div>
              <div className="border-gray-200 pl-8 flex flex-col gap-4">
                <div>
                  <div className="font-semibold">Updated</div>
                  <div className="truncate">
                    {getPublishableKeyResponse?.publishableKey?.updateTime &&
                      DateTime.fromJSDate(
                        timestampDate(
                          getPublishableKeyResponse?.publishableKey?.updateTime,
                        ),
                      ).toRelative()}
                  </div>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        <DangerZoneCard />
      </PageContent>
    </>
  );
};

const schema = z.object({
  displayName: z.string(),
});

const EditPublishableKeyButton = () => {
  const { publishableKeyId } = useParams();
  const { data: getPublishableKeyResponse, refetch } = useQuery(
    getPublishableKey,
    {
      id: publishableKeyId,
    },
  );
  const updatePublishableKeyMutation = useMutation(updatePublishableKey);

  // Currently there's an issue with the types of react-hook-form and zod
  // preventing the compiler from inferring the correct types.
  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: {
      displayName: '',
    },
  });
  useEffect(() => {
    if (getPublishableKeyResponse?.publishableKey) {
      form.reset({
        displayName: getPublishableKeyResponse.publishableKey.displayName,
      });
    }
  }, [getPublishableKeyResponse]);

  const [open, setOpen] = useState(false);

  const handleSubmit = async (values: z.infer<typeof schema>) => {
    await updatePublishableKeyMutation.mutateAsync({
      id: publishableKeyId,
      publishableKey: {
        displayName: values.displayName,
      },
    });
    await refetch();
    setOpen(false);
  };

  return (
    <AlertDialog open={open} onOpenChange={setOpen}>
      <AlertDialogTrigger>
        <Button variant="outline">Edit</Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Edit Publishable Key</AlertDialogTitle>
          <AlertDialogDescription>
            Edit Publishable Key settings.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <Form {...form}>
          {}
          {/**Currently there's an issue with the types of react-hook-form and zod
           preventing the compiler from inferring the correct types.*/}
          <form onSubmit={form.handleSubmit(handleSubmit)}>
            {}
            <FormField
              control={form.control}
              name="displayName"
              render={({ field }: { field: any }) => (
                <FormItem>
                  <FormLabel>Display Name</FormLabel>
                  <FormDescription>
                    An internal human-friendly name for the Publishable Key. Not
                    shown to your customers.
                  </FormDescription>
                  <FormControl>
                    <Input className="max-w-96" {...field} />
                  </FormControl>

                  <FormMessage />
                </FormItem>
              )}
            />
            <AlertDialogFooter className="mt-8">
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <Button type="submit">Save</Button>
            </AlertDialogFooter>
          </form>
        </Form>
      </AlertDialogContent>
    </AlertDialog>
  );
};

const DangerZoneCard = () => {
  const { publishableKeyId } = useParams();
  const { data: getPublishableKeyResponse } = useQuery(getPublishableKey, {
    id: publishableKeyId,
  });

  const [confirmDeleteOpen, setConfirmDeleteOpen] = useState(false);
  const handleDelete = () => {
    setConfirmDeleteOpen(true);
  };

  const deletePublishableKeyMutation = useMutation(deletePublishableKey);
  const navigate = useNavigate();
  const handleConfirmDelete = async () => {
    await deletePublishableKeyMutation.mutateAsync({
      id: publishableKeyId,
    });

    toast.success('Publishable Key deleted');
    navigate(`/project-settings/api-keys`);
  };

  return (
    <>
      <AlertDialog open={confirmDeleteOpen} onOpenChange={setConfirmDeleteOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              Delete {getPublishableKeyResponse?.publishableKey?.displayName}?
            </AlertDialogTitle>
            <AlertDialogDescription>
              Deleting a Publishable Key cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <Button variant="destructive" onClick={handleConfirmDelete}>
              Delete Publishable Key
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <Card className="border-destructive">
        <CardHeader>
          <CardTitle>Danger Zone</CardTitle>
        </CardHeader>

        <CardContent className="space-y-8">
          <div className="flex justify-between items-center">
            <div>
              <div className="text-sm font-semibold">
                Delete Publishable Key
              </div>
              <p className="text-sm">Delete this Publishable Key.</p>
            </div>

            <Button variant="destructive" onClick={handleDelete}>
              Delete Publishable Key
            </Button>
          </div>
        </CardContent>
      </Card>
    </>
  );
};
