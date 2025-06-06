import React from 'react';
import { cn } from '@/lib/utils';
import { Outlet } from 'react-router';
import { AccessTokenProvider, useAccessToken } from '@/lib/AccessTokenProvider';
import ConsoleNavigation from './ConsoleNavigation';

export const PageShell = () => {
  return (
    <AccessTokenProvider>
      <PageShellInner />
    </AccessTokenProvider>
  );
};

function PageShellInner() {
  const accessToken = useAccessToken();
  if (!accessToken) {
    return null;
  }

  return (
    <>
      {/* <SidebarProvider>
        <ConsoleSidebar />
        <SidebarInset> */}
      <main className="bg-gray-100 w-full min-h-screen">
        <ConsoleNavigation />

        <div className="pt-16">
          <Outlet />
        </div>
      </main>
      {/* </SidebarInset>
      </SidebarProvider> */}
    </>
  );
}

export const PageTitle = ({
  className,
  ...props
}: React.HTMLAttributes<HTMLHeadingElement>) => (
  <h1 className={cn('mt-4 font-semibold text-3xl ', className)} {...props} />
);
PageTitle.displayName = 'PageTitle';

export const PageCodeSubtitle = ({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) => (
  <div
    className={cn(
      'mt-2 inline-block rounded py-1 px-2 font-mono text-xs bg-white/15 text-white/80',
      className,
    )}
    {...props}
  />
);
PageCodeSubtitle.displayName = 'PageCodeSubtitle';

export const PageDescription = ({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) => (
  <div className={cn('mt-4', className)} {...props} />
);
PageDescription.displayName = 'PageDescription';

export const PageHeader = ({
  className,
  children,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) => (
  <div className="bg-zinc-950 pt-16 pb-32 w-full relative overflow-hidden -mb-32 z-0">
    <div className="absolute flex justify-center items-center blur-3xl w-full z-0">
      <div className="relative rounded-full w-[750px] h-[750px] bg-indigo-600/30 blur-3xl m-auto" />
    </div>

    <div
      className={cn('container p-4 m-auto text-white relative z-5', className)}
      {...props}
    >
      {children}
    </div>
  </div>
);

export const PageContent = ({
  className,
  children,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) => (
  <div
    className={cn('relative container p-4 pb-16 m-auto z-5', className)}
    {...props}
  >
    {children}
  </div>
);
PageContent.displayName = 'PageContent';
