import type { Metadata } from "next";
import type { ReactNode } from "react";

import { tenantRequest } from "@/components/branding/tenant-frame";
import { TenantStyle } from "@/components/branding/tenant-style";
import { ThemeProvider } from "@/components/layout/theme-provider";
import { runtimeConfig } from "@/lib/config";

import "./globals.css";

// Identity is read per request, never captured at build time.
export const dynamic = "force-dynamic";

export function generateMetadata(): Metadata {
  const { productName } = runtimeConfig();
  return { title: productName };
}

// G069 (Plan r1 S5, S7, §2.4): the head carries the request's tenant stylesheet (empty on a loopback or unknown host),
// so the first HTML byte has the tenant's values in both appearances; the theme provider opens the body and writes the
// resolved light or dark class on <html> before the first paint.
export default async function RootLayout({ children }: { children: ReactNode }) {
  const { tenant } = await tenantRequest();
  return (
    // The hydration-warning exemption covers only <html>'s own attributes: the theme provider sets its class before
    // hydration. Nothing below <html> is exempt from hydration checks.
    <html lang="en" suppressHydrationWarning>
      <head>
        <TenantStyle tokens={tenant?.tokens} />
      </head>
      <body className="min-h-screen bg-background text-foreground antialiased">
        <ThemeProvider>{children}</ThemeProvider>
      </body>
    </html>
  );
}
