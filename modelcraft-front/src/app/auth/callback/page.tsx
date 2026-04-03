"use client";

import { Suspense, useEffect, useState, useRef } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import {
  getUserInfoFromToken,
  isTokenExpired,
} from "@bff/auth/casdoor";
import { useAuthStore } from "@shared/stores/auth-store";

interface TokenResponse {
  accessToken: string;
  expiresIn: number;
}

function AuthCallbackContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [error, setError] = useState<string | null>(null);
  const [status, setStatus] = useState<string>("Processing...");
  // Ref to prevent double execution in React StrictMode
  const isExchangingRef = useRef(false);

  useEffect(() => {
    async function handleCallback() {
      // Get authorization code from URL
      const code = searchParams.get("code");

      if (!code) {
        setError("No authorization code received");
        return;
      }

      // CRITICAL: Prevent double execution in React 18 StrictMode
      // OAuth authorization codes can ONLY be used ONCE
      // Use sessionStorage to persist flag across component mount cycles
      const exchangeKey = `auth_exchange_${code}`;
      const exchangeStatus = sessionStorage.getItem(exchangeKey);
      const isInProgress = exchangeStatus === "in_progress" || exchangeStatus === "true";

      const waitForToken = async () => {
        for (let attempt = 0; attempt < 10; attempt += 1) {
          const token = useAuthStore.getState().accessToken;
          if (token) {
            return token;
          }
          await new Promise((resolve) => setTimeout(resolve, 300));
        }
        return null;
      };

      if (exchangeStatus) {
        console.log(
          "Authorization code already exchanged, skipping duplicate call",
        );

        // Try to retrieve the stored token and continue the flow
        let storedToken = useAuthStore.getState().accessToken;

        if (!storedToken && isInProgress) {
          setStatus("Finishing authentication...");
          storedToken = await waitForToken();
        }

        if (!storedToken) {
          // Code was exchanged but no token found - redirect to login
          console.warn(
            "Code already used but no token found, redirecting to login",
          );
          setError("Session expired. Please login again.");
          setTimeout(() => {
            router.push("/login");
          }, 2000);
          return;
        }

        console.log("Found stored token, continuing authentication flow");
        setStatus("Redirecting...");

        const userInfo = getUserInfoFromToken(storedToken);

        if (!userInfo) {
          setError("Invalid token. Please login again.");
          setTimeout(() => {
            router.push("/login");
          }, 2000);
          return;
        }

        // Always redirect to org selector after login
        console.log("User authenticated, redirecting to org selector");
        router.push("/org-selector");
        return;
      }

      // Prevent race conditions within the same mount cycle
      if (isExchangingRef.current) {
        console.log("Token exchange already in progress in this mount cycle");
        return;
      }

      try {
        // Mark as exchanging BEFORE the async call
        isExchangingRef.current = true;
        sessionStorage.setItem(exchangeKey, "in_progress");

        setStatus("Exchanging authorization code...");

        // Exchange code for tokens via BFF endpoint
        const response = await fetch("/api/bff/auth/token", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            code,
            redirectUri: window.location.origin + "/auth/callback",
          }),
        });

        if (!response.ok) {
          const errData = await response.json().catch(() => ({})) as Record<string, string>;
          throw new Error(errData.error || "Failed to exchange token");
        }

        const data = await response.json() as TokenResponse;

        setStatus("Storing authentication token...");

        // Store access token in memory store (refresh token is in httpOnly Cookie)
        useAuthStore.getState().setAccessToken(data.accessToken, data.expiresIn);
        sessionStorage.setItem(exchangeKey, "done");

        // Extract user info from token
        const userInfo = getUserInfoFromToken(data.accessToken);

        if (!userInfo) {
          setError("Failed to extract user information from token");
          return;
        }

        setStatus("Authentication successful!");

        // Always redirect to org selector after login
        console.log("User authenticated, redirecting to org selector...");
        setStatus("Redirecting to organization selector...");
        setTimeout(() => {
          sessionStorage.removeItem(exchangeKey);
          router.push("/org-selector");
        }, 500);
      } catch (err) {
        console.error("Auth callback error:", err);

        // Check if it's a code reuse error
        const errorMessage =
          err instanceof Error ? err.message : "Authentication failed";

        if (
          errorMessage.includes("authorization code has been used") ||
          errorMessage.includes("invalid_grant")
        ) {
          // Authorization code was already used - check if we have a token stored
          const storedToken = useAuthStore.getState().accessToken;

          if (storedToken && !isTokenExpired(storedToken)) {
            console.log("Code already used but found valid token, redirecting to org selector");
            router.push("/org-selector");
            return;
          }

          // No valid token found - user needs to login again
          setError("This login link has expired. Please login again.");
          setTimeout(() => {
            sessionStorage.removeItem(exchangeKey);
            router.push("/login");
          }, 2000);
        } else {
          // Other authentication errors
          setError(errorMessage);
        }
      }
    }

    handleCallback();
  }, [searchParams, router]);

  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="w-full max-w-md space-y-8 p-8">
        <div className="text-center">
          {error ? (
            <>
              <h2 className="mb-4 text-3xl font-semibold text-foreground">
                Authentication Failed
              </h2>
              <div className="rounded-md bg-red-50 p-4">
                <div className="flex">
                  <div className="flex-shrink-0">
                    <svg
                      className="size-5 text-red-400"
                      xmlns="http://www.w3.org/2000/svg"
                      viewBox="0 0 20 20"
                      fill="currentColor"
                    >
                      <path
                        fillRule="evenodd"
                        d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                        clipRule="evenodd"
                      />
                    </svg>
                  </div>
                  <div className="ml-3">
                    <h3 className="text-sm font-medium text-red-800">
                      {error}
                    </h3>
                    <div className="mt-4">
                      <button
                        onClick={() => router.push("/login")}
                        className="text-sm font-medium text-red-600 hover:text-red-500"
                      >
                        Try again
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </>
          ) : (
            <>
              <h2 className="mb-4 text-3xl font-semibold text-foreground">
                Signing you in...
              </h2>
              <div className="rounded-md bg-primary/5 p-4">
                <div className="flex items-center justify-center">
                  <svg
                    className="size-8 animate-spin text-primary"
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                  >
                    <circle
                      className="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      strokeWidth="4"
                    />
                    <path
                      className="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    />
                  </svg>
                </div>
                <p className="mt-4 text-sm text-muted-foreground">{status}</p>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}

function LoadingFallback() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="w-full max-w-md space-y-8 p-8">
        <div className="text-center">
          <h2 className="mb-4 text-3xl font-semibold text-foreground">
            Signing you in...
          </h2>
          <div className="rounded-md bg-primary/5 p-4">
            <div className="flex items-center justify-center">
              <svg
                className="size-8 animate-spin text-primary"
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle
                  className="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  strokeWidth="4"
                />
                <path
                  className="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                />
              </svg>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default function AuthCallback() {
  return (
    <Suspense fallback={<LoadingFallback />}>
      <AuthCallbackContent />
    </Suspense>
  );
}
