"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { useRequireAuth } from "@web/hooks/auth/use-auth";
import { useOrganizationStore } from "@shared/stores/organization";

export default function OrgLayout({ children }: { children: React.ReactNode }) {
  const params = useParams();
  const router = useRouter();
  const orgName = params.orgName as string;

  // Restore access token if page was refreshed (middleware already checked cookie)
  const { isLoading: authLoading, user } = useRequireAuth()

  const [isVerifying, setIsVerifying] = useState(true);
  const { setCurrentOrg, loadMemberships } = useOrganizationStore();

  useEffect(() => {
    if (authLoading) return

    async function verifyOrgAccess() {
      console.log("[OrgLayout] Verifying org access:", orgName, "user:", user?.id);

      try {
        const { getToken } = await import('@api-client/auth/public')
        const token = getToken()
        if (!token) {
          // Should not happen — middleware guards this, but be safe
          console.warn("[OrgLayout] No token after auth restore")
          router.push("/login")
          return
        }

        console.log("[OrgLayout] Loading memberships...");
        const memberships = await loadMemberships(token, false);
        console.log("[OrgLayout] Memberships:", memberships.map((m) => m.orgName));

        const hasAccess = memberships.some((m) => m.orgName === orgName);
        console.log("[OrgLayout] Has access to org:", orgName, "→", hasAccess);

        if (!hasAccess) {
          const fallbackOrgName = memberships[0]?.orgName;

          if (fallbackOrgName) {
            console.warn(
              `[OrgLayout] Access denied to "${orgName}", redirecting to fallback org "${fallbackOrgName}"`
            );
            localStorage.setItem("defaultOrgName", fallbackOrgName);
            router.push(`/org/${fallbackOrgName}/workspace`);
            return;
          }

          console.warn(
            `[OrgLayout] Access denied to "${orgName}" and no memberships found, redirecting to org creation`
          );
          localStorage.removeItem("defaultOrgName");
          router.push("/org/create");
          return;
        }

        setCurrentOrg(orgName);
        localStorage.setItem("defaultOrgName", orgName);
        console.log("[OrgLayout] Org access verified ✓");
        setIsVerifying(false);
      } catch (error) {
        console.error("[OrgLayout] Error verifying org access:", error);
        localStorage.removeItem("defaultOrgName");
        router.push("/login");
      }
    }

    verifyOrgAccess();
  }, [authLoading, orgName, router, setCurrentOrg, loadMemberships, user?.id]);

  if (authLoading || isVerifying) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    );
  }

  return <>{children}</>;
}
