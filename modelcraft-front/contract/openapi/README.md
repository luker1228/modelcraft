# OpenAPI Specification Maintenance Guide

## Overview

ModelCraft uses **modular OpenAPI specifications** that are bundled into a single file and compiled into Go server code using `oapi-codegen`.

## Directory Structure

```
api/openapi/
├── README.md                  # This file
├── oapi-codegen.yaml          # Code generation config
├── openapi-root.yaml          # Main entry point (references all modules)
├── openapi.yaml               # Generated bundled spec (DO NOT EDIT MANUALLY)
├── common.yaml                # Common schemas (errors, base response, security)
├── auth.yaml                  # Authentication endpoints
├── org.yaml                   # Organization management endpoints
└── webhook.yaml               # Webhook endpoints
```

## Module Files

### openapi-root.yaml
- Main entry point that imports all other modules
- Defines paths by referencing module files
- Defines schemas by referencing module files
- **Always update this file** when adding new paths or schemas from modules

### Module Files (auth.yaml, org.yaml, webhook.yaml, etc.)
- Each module defines both `paths` and `schemas` sections
- Paths use relative references to schemas in the same module: `$ref: "org.yaml#/schemas/..."`
- Common schemas reference: `$ref: "common.yaml#/schemas/..."`

### common.yaml
- Shared schemas used across multiple modules
- Base response types
- Error response types
- Security schemes (BearerAuth)

## Adding New Endpoints

### Step 1: Edit Module File

Example: Adding `/api/orgs/initialize` in `org.yaml`:

```yaml
paths:
  /api/orgs/initialize:
    post:
      operationId: initializeOrganization
      summary: Initialize a new organization
      tags: [Organization]
      requestBody:
        required: false
        content:
          application/json:
            schema:
              $ref: "org.yaml#/schemas/InitializeOrganizationRequest"
      responses:
        "200":
          description: Organization initialized successfully
          content:
            application/json:
              schema:
                $ref: "org.yaml#/schemas/InitializeOrganizationResponse"
        "400":
          description: Invalid request
          content:
            application/json:
              schema:
                $ref: "org.yaml#/schemas/OrgInvalidInputError"
        "401":
          description: Unauthorized
          content:
            application/json:
              schema:
                $ref: "common.yaml#/schemas/AuthenticationFailedError"
        "500":
          description: Server error
          content:
            application/json:
              schema:
                $ref: "common.yaml#/schemas/SystemError"

schemas:
  InitializeOrganizationRequest:
    type: object
    properties:
      namePrefix:
        type: string
        description: Optional custom prefix for organization name

  InitializeOrganizationResponse:
    allOf:
      - $ref: "common.yaml#/schemas/BaseResponse"
      - type: object
        required:
          - success
          - organization
        properties:
          success:
            type: boolean
          organization:
            type: object
            properties:
              id:
                type: string
              name:
                type: string
```

### Step 2: Update openapi-root.yaml

Add path reference:
```yaml
paths:
  # Organization Endpoints
  /api/org/create:
    $ref: "org.yaml#/paths/~1api~1org~1create"
  /api/orgs/initialize:
    $ref: "org.yaml#/paths/~1api~1orgs~1initialize"
```

Add schema references:
```yaml
components:
  schemas:
    # -- Organization --
    OrgInvalidInputError:
      $ref: "org.yaml#/schemas/OrgInvalidInputError"
    CreateOrganizationRequest:
      $ref: "org.yaml#/schemas/CreateOrganizationRequest"
    CreateOrganizationResponse:
      $ref: "org.yaml#/schemas/CreateOrganizationResponse"
    InitializeOrganizationRequest:
      $ref: "org.yaml#/schemas/InitializeOrganizationRequest"
    InitializeOrganizationResponse:
      $ref: "org.yaml#/schemas/InitializeOrganizationResponse"
```

**Important**: Path references use URL encoding for slashes: `/api/org/create` → `~1api~1org~1create`

### Step 3: Generate Code

```bash
# Bundle and generate (recommended)
task generate-oapi

# Or step by step:
task bundle-oapi      # Bundle modular specs
oapi-codegen --config api/openapi/oapi-codegen.yaml api/openapi/openapi.yaml
```

This will:
1. Bundle all module files into `openapi.yaml`
2. Generate server interface in `internal/interfaces/http/generated/server.gen.go`

### Step 4: Implement Handler

The generated code creates a `ServerInterface` that you must implement:

```go
// internal/interfaces/http/server.go
func (s *Server) InitializeOrganization(w http.ResponseWriter, r *http.Request) {
    s.delegateToGin(w, r, s.orgInitializeHandler.Handle)
}
```

### Step 5: Configure Authentication

Update `chi_setup.go` if the endpoint needs special authentication:

```go
// User-only paths (JWT required, no organization context)
userOnlyPaths := map[string]bool{
    "/api/orgs/initialize": true,
}
```

## Common Patterns

### Error Responses

Always include standard error responses:
```yaml
responses:
  "400":
    description: Invalid request
    content:
      application/json:
        schema:
          $ref: "common.yaml#/schemas/SystemError"  # or domain-specific error
  "401":
    description: Unauthorized
    content:
      application/json:
        schema:
          $ref: "common.yaml#/schemas/AuthenticationFailedError"
  "500":
    description: Server error
    content:
      application/json:
        schema:
          $ref: "common.yaml#/schemas/SystemError"
```

### Base Response

All success responses should extend `BaseResponse`:
```yaml
schemas:
  MyResponse:
    allOf:
      - $ref: "common.yaml#/schemas/BaseResponse"
      - type: object
        properties:
          data:
            type: object
```

### Public vs Protected Endpoints

```yaml
# Public endpoint (no authentication)
security: []

# Protected endpoint (default, requires JWT)
# (no security field needed, inherits global security requirement)
```

## Code Generation Configuration

`oapi-codegen.yaml`:
```yaml
package: generated
output: internal/interfaces/http/generated/server.gen.go
generate:
  chi-server: true      # Generate Chi router handlers
  models: true          # Generate Go structs for schemas
  embedded-spec: true   # Embed OpenAPI spec in binary
output-options:
  skip-prune: false
```

## Verification

After generating code:

1. **Build the application**:
   ```bash
   task build
   ```

2. **Check generated interface**:
   ```bash
   grep -A 5 "InitializeOrganization" internal/interfaces/http/generated/server.gen.go
   ```

3. **Verify routes are registered**:
   - Start server
   - Check logs for route registration
   - Test endpoint with curl

4. **Test the endpoint**:
   ```bash
   curl -X POST http://localhost:8080/api/orgs/initialize \
     -H "Authorization: Bearer <token>" \
     -H "Content-Type: application/json" \
     -d '{"namePrefix": "myorg"}'
   ```

## Troubleshooting

### "404 Not Found" for new endpoint

**Cause**: Path not referenced in `openapi-root.yaml`

**Solution**: Add path reference to `openapi-root.yaml` and regenerate

### "Method not implemented" error

**Cause**: Handler not implemented in `server.go`

**Solution**: Implement the generated `ServerInterface` method

### Schema validation errors

**Cause**: Schema references incorrect or circular

**Solution**:
- Check module file schema references use correct format: `"module.yaml#/schemas/SchemaName"`
- Verify schemas are defined in the referenced module
- Use `$ref` for all schema references, don't inline

### Bundle fails with reference errors

**Cause**: Invalid $ref paths or missing schemas

**Solution**:
- Verify all `$ref` paths use correct format
- Check that referenced schemas exist
- Use URL encoding for path slashes: `~1api~1org~1create`

## Best Practices

1. **Keep modules focused**: One module per domain (auth, org, webhook, etc.)
2. **Use common.yaml**: Put shared schemas in `common.yaml`
3. **Consistent naming**:
   - Paths: `/api/domain/action` (e.g., `/api/orgs/initialize`)
   - Operations: `actionDomain` (e.g., `initializeOrganization`)
   - Request schemas: `ActionDomainRequest`
   - Response schemas: `ActionDomainResponse`
4. **Always extend BaseResponse**: For consistent response structure
5. **Document thoroughly**: Add descriptions to paths, parameters, and schemas
6. **Commit generated files**: Commit both module files and `openapi.yaml`

## Reference Documentation

- [OpenAPI 3.0 Specification](https://swagger.io/specification/)
- [oapi-codegen Documentation](https://github.com/oapi-codegen/oapi-codegen)
- [Redocly CLI Documentation](https://redocly.com/docs/cli/)
- [Chi Router Documentation](https://github.com/go-chi/chi)
