"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { getToken, refreshAccessToken } from "@api-client/auth/public";
import { TENANT_LOGIN_PATH } from "@shared/constants/routes";

export default function Home() {
  const router = useRouter();
  const [debugInfo, setDebugInfo] = useState<string>("Restoring session...");

  useEffect(() => {
    async function init() {
      let token = getToken();
      console.log("[HomePage] In-memory token present:", !!token);

      if (!token) {
        console.log("[HomePage] Attempting silent refresh...");
        setDebugInfo("Restoring session...");
        token = await refreshAccessToken();
        console.log("[HomePage] Silent refresh:", token ? "success" : "failed");
      }

      if (!token) {
        // Middleware should have prevented reaching here, but be defensive
        console.warn("[HomePage] No token, redirecting to login");
        router.push(TENANT_LOGIN_PATH);
        return;
      }

      const defaultOrgName = localStorage.getItem("defaultOrgName");

      setDebugInfo("Redirecting...");
      if (defaultOrgName) {
        console.log("[HomePage] redirecting to default org:", defaultOrgName);
        router.push(`/org/${defaultOrgName}/workspace`);
        return;
      }

      console.log("[HomePage] default org not found, redirecting to login");
      router.push(TENANT_LOGIN_PATH);
    }

    init();
  }, [router]);

  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="space-y-4 text-center">
        <div className="mx-auto size-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
        <p className="mt-3 text-sm text-muted-foreground">{debugInfo}</p>
      </div>
    </div>
  );
}
