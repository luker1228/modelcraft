import Sdk from "casdoor-js-sdk";
import { useAuthStore } from "@shared/stores/auth-store";

// Casdoor configuration
const casdoorConfig = {
  serverUrl: process.env.NEXT_PUBLIC_CASDOOR_URL || "",
  clientId: process.env.NEXT_PUBLIC_CASDOOR_CLIENT_ID || "",
  organizationName: process.env.NEXT_PUBLIC_CASDOOR_ORGANIZATION || "built-in",
  appName: process.env.NEXT_PUBLIC_CASDOOR_APP_NAME || "modelcraft",
  redirectPath: "/auth/callback",
};

// Lazy-initialize Casdoor SDK (only on client side, since SDK accesses `window`)
let _casdoorSdk: Sdk | null = null;

function getCasdoorSdk(): Sdk {
  if (!_casdoorSdk) {
    _casdoorSdk = new Sdk(casdoorConfig);
  }
  return _casdoorSdk;
}

/**
 * Redirect to Casdoor login page
 */
export function redirectToLogin() {
  const loginUrl = getCasdoorSdk().getSigninUrl();
  window.location.href = loginUrl;
}

/**
 * Redirect to Casdoor signup page (OAuth flow)
 * IMPORTANT: Use enablePassword=false to get OAuth signup URL with redirect_uri
 * Otherwise it returns a direct signup page URL without OAuth callback
 */
export function redirectToSignup() {
  const signupUrl = getCasdoorSdk().getSignupUrl(false);
  window.location.href = signupUrl;
}

/**
 * Get Casdoor login URL (for custom handling)
 */
export function getLoginUrl(state?: string): string {
  const sdk = getCasdoorSdk();
  if (state) {
    return sdk.getSigninUrl() + `&state=${encodeURIComponent(state)}`;
  }
  return sdk.getSigninUrl();
}

/**
 * Get Casdoor signup URL (for custom handling)
 * IMPORTANT: Use enablePassword=false to get OAuth signup URL with redirect_uri
 */
export function getSignupUrl(state?: string): string {
  const sdk = getCasdoorSdk();
  const signupUrl = sdk.getSignupUrl(false); // Use OAuth flow
  if (state) {
    return signupUrl + `&state=${encodeURIComponent(state)}`;
  }
  return signupUrl;
}

/**
 * @deprecated Use /api/bff/auth/token endpoint directly from the callback page.
 * This function is removed to prevent accidental use of the old /api/auth/token endpoint.
 */
export async function exchangeCodeForToken(_code: string): Promise<{ accessToken: string; refreshToken?: string }> {
  throw new Error('exchangeCodeForToken is deprecated. Use /api/bff/auth/token endpoint directly.')
}

interface JWTPayload {
  exp?: number
  user_id?: string
  sub?: string
  email?: string
  name?: string
}

/**
 * Decode JWT token (without verification - use for client-side display only)
 */
export function decodeJWT(token: string): JWTPayload | null {
  try {
    const base64Url = token.split(".")[1];
    const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split("")
        .map((c) => "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2))
        .join(""),
    );
    return JSON.parse(jsonPayload) as JWTPayload;
  } catch {
    return null;
  }
}

/**
 * Check if token is expired
 */
export function isTokenExpired(token: string): boolean {
  const decoded = decodeJWT(token);
  if (!decoded || !decoded.exp) {
    return true;
  }

  const currentTime = Math.floor(Date.now() / 1000);
  return decoded.exp < currentTime;
}

/**
 * Get user info from JWT token
 */
export interface UserInfo {
  id: string;        // user_id from JWT
  email: string;     // email from JWT
  name: string;      // name from JWT
}

export function getUserInfoFromToken(token: string): UserInfo | null {
  const decoded = decodeJWT(token);
  if (!decoded) {
    return null;
  }

  return {
    id: decoded.user_id || decoded.sub || "",  // user_id is the new field, sub for backward compatibility
    email: decoded.email || "",
    name: decoded.name || "",
  };
}

/**
 * Get organization name from JWT token
 * @deprecated The simplified JWT no longer contains organization info.
 * Use /api/user/memberships to get user's organizations.
 * @returns null always
 */
export function getOrgNameFromToken(token: string): string | null {
  // No longer available in the simplified JWT
  // Use /api/user/memberships to get user's organizations
  return null;
}

/**
 * @deprecated Token is now stored in-memory via useAuthStore. This function is a no-op.
 * Store JWT token in localStorage
 */
export function storeToken(_token: string, _refreshToken?: string) {
  // no-op: tokens are now stored in-memory (access) + httpOnly Cookie (refresh)
}

/**
 * @deprecated Token is now stored in httpOnly Cookie. This function is a no-op.
 * Store refresh token in localStorage
 */
export function storeRefreshToken(_token: string) {
  // no-op: refresh token is now stored in httpOnly Cookie
}

/**
 * @deprecated Refresh token is now in httpOnly Cookie. Returns null always.
 * Get refresh token from localStorage
 */
export function getRefreshToken(): string | null {
  return null;
}

/**
 * @deprecated Refresh token is now in httpOnly Cookie. This function is a no-op.
 * Remove refresh token from localStorage
 */
export function removeRefreshToken() {
  // no-op
}

/**
 * @deprecated Use useAuthStore.getState().accessToken instead.
 * Get JWT token from in-memory store
 */
export function getToken(): string | null {
  return useAuthStore.getState().accessToken;
}

/**
 * @deprecated Use useAuthStore.getState().clearAccessToken() instead.
 * Remove JWT token (logout)
 */
export function removeToken() {
  useAuthStore.getState().clearAccessToken();
}

/**
 * Check if token expires within the given threshold (seconds).
 * Default threshold is 300 seconds (5 minutes).
 */
export function isTokenNearExpiry(token: string, thresholdSeconds = 300): boolean {
  const decoded = decodeJWT(token);
  if (!decoded || !decoded.exp) {
    return true;
  }
  const currentTime = Math.floor(Date.now() / 1000);
  return decoded.exp - currentTime < thresholdSeconds;
}

// Prevent concurrent refresh requests
let _isRefreshing = false;
let _refreshPromise: Promise<string | null> | null = null;

/**
 * Refresh the access token using the httpOnly Cookie refresh token.
 * Returns the new access token, or null if refresh failed.
 * Concurrent calls will share the same refresh request.
 */
export async function refreshAccessToken(): Promise<string | null> {
  if (_isRefreshing && _refreshPromise) {
    return _refreshPromise;
  }

  _isRefreshing = true;
  _refreshPromise = (async () => {
    try {
      // No body needed - refresh token is sent automatically via httpOnly Cookie
      const response = await fetch("/api/bff/auth/refresh", {
        method: "POST",
        credentials: "same-origin",
      });

      if (!response.ok) {
        // Refresh token invalid or expired - clear access token
        useAuthStore.getState().clearAccessToken();
        return null;
      }

      const data = (await response.json()) as { accessToken?: string; expiresIn?: number };
      const newAccessToken: string | undefined = data.accessToken;
      const expiresIn: number | undefined = data.expiresIn;

      if (newAccessToken && expiresIn) {
        useAuthStore.getState().setAccessToken(newAccessToken, expiresIn);
        return newAccessToken;
      }

      return null;
    } catch {
      return null;
    } finally {
      _isRefreshing = false;
      _refreshPromise = null;
    }
  })();

  return _refreshPromise;
}

/**
 * Check if user is authenticated
 */
export function isAuthenticated(): boolean {
  const { accessToken, isTokenExpired: checkExpired } = useAuthStore.getState();
  if (!accessToken) {
    return false;
  }
  return !checkExpired();
}
