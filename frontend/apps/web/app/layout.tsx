import '@/app/globals.css';
import SiteFooter from '@/components/SiteFooter';
import SiteHeader from '@/components/SiteHeader';
import AccountProvider from '@/components/providers/account-provider';
import { SessionProvider } from '@/components/providers/session-provider';
import { ThemeProvider } from '@/components/providers/theme-provider';
import { Toaster } from '@/components/ui/toaster';
import { fontSans } from '@/libs/fonts';
import { cn } from '@/libs/utils';
import { Metadata } from 'next';
import { ReactElement } from 'react';
import { auth, getDefaultProvider } from './api/auth/[...nextauth]/auth';
import { getSystemAppConfig } from './api/config/config';

export const metadata: Metadata = {
  title: 'Neosync',
  description: 'Open Source Test Data Management',
  icons: [{ rel: 'icon', url: 'favicon.ico' }],
};

export default async function RootLayout({
  children,
}: {
  children: React.ReactNode;
}): Promise<ReactElement> {
  const systemAppConfig = getSystemAppConfig();
  const session = systemAppConfig.isAuthEnabled ? await auth() : null;
  return (
    <html lang="en" suppressHydrationWarning>
      <head />
      <body
        className={cn(
          'min-h-screen bg-background font-sans antialiased overflow-scroll',
          fontSans.variable
        )}
      >
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <SessionProvider
            session={session}
            isAuthEnabled={systemAppConfig.isAuthEnabled}
            defaultProvider={getDefaultProvider()}
          >
            <AccountProvider>
              <div className="relative flex min-h-screen flex-col">
                <SiteHeader />
                <div className="flex-1 container" id="top-level-layout">
                  {children}
                </div>
                <SiteFooter />
                <Toaster />
              </div>
            </AccountProvider>
          </SessionProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
