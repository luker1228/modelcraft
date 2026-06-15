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
	XCSRFToken     = "X-CSRF-Token"
	XRequestedWith = "X-Requested-With"

	XAction        = "X-Action"
	XInternalToken = "X-Internal-Token"
	XUserID        = "X-User-ID"
	// XUserType is injected by the gateway to distinguish end-user from tenant callers.
	// It is used to decide which APIs the end-user chain is allowed to call.
	XUserType        = "X-User-Type"
	XIsAdmin         = "X-Is-Admin"
	XOrgName         = "X-Org-Name"
	XProjectSlug     = "X-Project-Slug"
	XRequestID       = "X-Request-Id"
	XClientRequestID = "X-Client-Request-Id"
	Traceparent      = "Traceparent"

	// XMCAuthUserID is injected by the gateway for RLS context propagation.
	// It carries the authenticated end-user ID into backend middleware.
	XMCAuthUserID = "X-MC-Auth-Userid"
	// XMCAuthUserName is injected by the gateway for RLS context propagation.
	// It carries the authenticated end-user name into backend middleware.
	XMCAuthUserName = "X-MC-Auth-Username"
	// XMCAuthRoles is injected by the gateway for RLS context propagation.
	// It carries the caller's role list as a comma-separated string.
	XMCAuthRoles = "X-MC-Auth-Roles"

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
