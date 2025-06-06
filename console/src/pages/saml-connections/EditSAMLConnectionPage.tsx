import { useNavigate, useParams } from 'react-router';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import {
  getOrganization,
  getSAMLConnection,
  updateSAMLConnection,
} from '@/gen/tesseral/backend/v1/backend-BackendService_connectquery';
import React, { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { Button } from '@/components/ui/button';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Switch } from '@/components/ui/switch';
import { Link } from 'react-router-dom';
import {
  ConsoleCard,
  ConsoleCardContent,
  ConsoleCardDescription,
  ConsoleCardDetails,
  ConsoleCardHeader,
  ConsoleCardTitle,
} from '@/components/ui/console-card';
import { Input } from '@/components/ui/input';
import { toast } from 'sonner';
import { PageContent, PageHeader, PageTitle } from '@/components/page';

const schema = z.object({
  primary: z.boolean(),
  idpEntityId: z.string().min(1, {
    message: 'IDP Entity ID must be non-empty.',
  }),
  idpRedirectUrl: z.string().url({
    message: 'IDP Redirect URL must be a valid URL.',
  }),
  idpX509Certificate: z.string().startsWith('-----BEGIN CERTIFICATE-----', {
    message: 'IDP Certificate must be a PEM-encoded X.509 certificate.',
  }),
});

export const EditSAMLConnectionPage = () => {
  const navigate = useNavigate();
  const { organizationId, samlConnectionId } = useParams();
  const { data: getOrganizationResponse } = useQuery(getOrganization, {
    id: organizationId,
  });
  const { data: getSAMLConnectionResponse } = useQuery(getSAMLConnection, {
    id: samlConnectionId,
  });
  /* eslint-disable @typescript-eslint/no-unsafe-call */
  // Currently there's an issue with the types of react-hook-form and zod
  // preventing the compiler from inferring the correct types.
  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: {},
  });
  const updateSAMLConnectionMutation = useMutation(updateSAMLConnection);

  useEffect(() => {
    if (getSAMLConnectionResponse?.samlConnection) {
      form.reset({
        primary: getSAMLConnectionResponse.samlConnection.primary,
        idpEntityId: getSAMLConnectionResponse.samlConnection.idpEntityId,
        idpRedirectUrl: getSAMLConnectionResponse.samlConnection.idpRedirectUrl,
        idpX509Certificate:
          getSAMLConnectionResponse.samlConnection.idpX509Certificate,
      });
    }
  }, [getSAMLConnectionResponse]);
  /* eslint-enable @typescript-eslint/no-unsafe-call */

  const onSubmit = async (values: z.infer<typeof schema>) => {
    await updateSAMLConnectionMutation.mutateAsync({
      id: samlConnectionId,
      samlConnection: {
        primary: values.primary,
        idpEntityId: values.idpEntityId,
        idpRedirectUrl: values.idpRedirectUrl,
        idpX509Certificate: values.idpX509Certificate,
      },
    });

    toast.success('SAML Connection updated');
    navigate(
      `/organizations/${organizationId}/saml-connections/${samlConnectionId}`,
    );
  };

  return (
    // TODO remove padding when app shell in place
    <>
      <PageHeader>
        <PageTitle>Edit SAML Connection</PageTitle>
      </PageHeader>

      <PageContent>
        <Form {...form}>
          {/* eslint-disable @typescript-eslint/no-unsafe-call */}
          {/** Currently there's an issue with the types of react-hook-form and zod 
        preventing the compiler from inferring the correct types.*/}
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
            {/* eslint-enable @typescript-eslint/no-unsafe-call */}
            <ConsoleCard>
              <ConsoleCardHeader>
                <ConsoleCardDetails>
                  <ConsoleCardTitle>SAML connection settings</ConsoleCardTitle>
                  <ConsoleCardDescription>
                    Configure basic settings on this SAML connection.
                  </ConsoleCardDescription>
                </ConsoleCardDetails>
              </ConsoleCardHeader>
              <ConsoleCardContent className="space-y-8">
                <FormField
                  control={form.control}
                  name="primary"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Primary</FormLabel>
                      <FormDescription>
                        A primary SAML connection gets used by default within
                        its organization.
                      </FormDescription>
                      <FormControl>
                        <Switch
                          className="block"
                          checked={field.value}
                          onCheckedChange={field.onChange}
                        />
                      </FormControl>

                      <FormMessage />
                    </FormItem>
                  )}
                />
              </ConsoleCardContent>
            </ConsoleCard>

            <ConsoleCard>
              <ConsoleCardHeader>
                <ConsoleCardDetails>
                  <ConsoleCardTitle>Service Provider settings</ConsoleCardTitle>
                  <ConsoleCardDescription>
                    The configuration here is assigned automatically by
                    Tesseral, and needs to be inputted into your customer's
                    Identity Provider by their IT admin.
                  </ConsoleCardDescription>
                </ConsoleCardDetails>
              </ConsoleCardHeader>
              <ConsoleCardContent className="space-y-8">
                <div>
                  <div className="text-sm font-medium leading-none">
                    Assertion Consumer Service (ACS) URL
                  </div>
                  <div className="mt-1">
                    {getSAMLConnectionResponse?.samlConnection?.spAcsUrl}
                  </div>
                </div>
                <div>
                  <div className="text-sm font-medium leading-none">
                    SP Entity ID
                  </div>
                  <div className="mt-1">
                    {getSAMLConnectionResponse?.samlConnection?.spEntityId}
                  </div>
                </div>
              </ConsoleCardContent>
            </ConsoleCard>
            <ConsoleCard>
              <ConsoleCardHeader>
                <ConsoleCardDetails>
                  <ConsoleCardTitle>
                    Identity Provider settings
                  </ConsoleCardTitle>
                  <ConsoleCardDescription>
                    The configuration here needs to be copied over from the
                    customer's Identity Provider ("IDP").
                  </ConsoleCardDescription>
                </ConsoleCardDetails>
              </ConsoleCardHeader>
              <ConsoleCardContent className="space-y-8">
                <FormField
                  control={form.control}
                  name="idpEntityId"
                  render={({ field }: { field: any }) => (
                    <FormItem>
                      <FormLabel>IDP Entity ID</FormLabel>
                      <FormDescription>
                        The IDP Entity ID, as configured in the customer's
                        Identity Provider.
                      </FormDescription>
                      <FormControl>
                        <Input className="max-w-96" {...field} />
                      </FormControl>

                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="idpRedirectUrl"
                  render={({ field }: { field: any }) => (
                    <FormItem>
                      <FormLabel>IDP Redirect URL</FormLabel>
                      <FormDescription>
                        The IDP Redirect URL, as configured in the customer's
                        Identity Provider.
                      </FormDescription>
                      <FormControl>
                        <Input className="max-w-96" {...field} />
                      </FormControl>

                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="idpX509Certificate"
                  render={({
                    field: { onChange },
                  }: {
                    field: { onChange: (value: string) => void };
                  }) => (
                    <FormItem>
                      <FormLabel>IDP Certificate</FormLabel>
                      <FormDescription>
                        IDP Certificate, as a PEM-encoded X.509 certificate.
                        These start with '-----BEGIN CERTIFICATE-----' and end
                        with '-----END CERTIFICATE-----'.
                      </FormDescription>
                      <FormControl>
                        <Input
                          className="max-w-96"
                          type="file"
                          onChange={async (e) => {
                            // File inputs are special; they are necessarily "uncontrolled", and their value is a FileList.
                            // We just copy over the file's contents to the react-form-hook state manually on input change.
                            if (e.target.files) {
                              onChange(await e.target.files[0].text());
                            }
                          }}
                        />
                      </FormControl>

                      <FormMessage />
                    </FormItem>
                  )}
                />
              </ConsoleCardContent>
            </ConsoleCard>

            <div className="flex justify-end gap-x-4 pb-8">
              <Button variant="outline" asChild>
                <Link to={`/organizations/${organizationId}`}>Cancel</Link>
              </Button>
              <Button type="submit">Save Changes</Button>
            </div>
          </form>
        </Form>
      </PageContent>
    </>
  );
};
