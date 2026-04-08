"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { getToken, getUserInfoFromToken, refreshAccessToken } from "@bff/auth/public";

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
        router.push("/login");
        return;
      }

      const userInfo = getUserInfoFromToken(token);
      console.log("[HomePage] User:", userInfo?.id, "→ redirecting to org-selector");
      setDebugInfo("Redirecting...");
      router.push("/org-selector");
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
