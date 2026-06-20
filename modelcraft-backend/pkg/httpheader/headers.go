package httpheader

// Single source of truth for HTTP header names and common header values used
// across the backend. Keep header literals out of handlers and middleware.
const (
	Authorization = "Authorization"
	Accept        = "Accept"
	Origin        = "Origin"

	ContentType    = "Content-Type"
	ContentLength  = "Content-Length"
	AcceptEncoding = "Accept-Encoding"
	AcceptLanguage = "Accept-Language"
	CacheControl   = "Cache-Control"
	XCSRFToken     = "X-CSRF-Token" //nolint:gosec // G101: header name constant, not a credential
	XRequestedWith = "X-Requested-With"

	XAction          = "X-Action"
	XInternalToken   = "X-Internal-Token"
	XUserID          = "X-User-ID"
	XOrgName         = "X-Org-Name"
	XProjectSlug     = "X-Project-Slug"
	XRequestID       = "X-Request-Id"
	XClientRequestID = "X-Client-Request-Id"
	Traceparent      = "Traceparent"

	// XMCAuthUserID is injected by the gateway for RLS context propagation.
	// It carries the authenticated end-user ID into backend middleware.
	// XMCAuthUserIDInt is injected by the gateway for numeric user IDs (int64).
	// Mutually exclusive with XMCAuthUserIDStr.
	XMCAuthUserIDInt = "X-MC-Auth-Userid-Int"
	// XMCAuthUserIDStr is injected by the gateway for string user IDs.
	// Mutually exclusive with XMCAuthUserIDInt.
	XMCAuthUserIDStr = "X-MC-Auth-Userid-Str"
	// XMCAuthUserName is injected by the gateway for RLS context propagation.
	// It carries the authenticated end-user name into backend middleware.
	XMCAuthUserName = "X-MC-Auth-Username"
	// XMCAuthRoles is injected by the gateway for RLS context propagation.
	// It carries the caller's role list as a comma-separated string.
	XMCAuthRoles = "X-MC-Auth-Roles"
	// XMCAuthUseAdmin is set by the PAT caller to request admin-level access.
	// Valid value: "true".  Absent or any other value means no admin elevation.
	XMCAuthUseAdmin = "X-MC-Auth-Useadmin"

	AccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	AccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	AccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	AccessControlAllowMethods     = "Access-Control-Allow-Methods"
	AccessControlExposeHeaders    = "Access-Control-Expose-Headers"

	ContentTypeApplicationJSON = "application/json"
	ContentTypeTextHTMLUTF8    = "text/html; charset=utf-8"
)

var (
	CORSAllowRequestHeaders = []string{
		ContentType,
		ContentLength,
		AcceptEncoding,
		XCSRFToken,
		Authorization,
		Accept,
		Origin,
		CacheControl,
		XRequestedWith,
	}

	CORSExposeResponseHeaders = []string{
		ContentLength,
		ContentType,
	}
)
