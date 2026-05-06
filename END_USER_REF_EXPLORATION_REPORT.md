# END_USER_REF Field Format: Codebase Exploration Report

**Date:** 2026-05-06  
**Scope:** Backend (Go) + Frontend (TypeScript/React)  
**Focus:** Current END_USER_REF implementation, x-mc format system, runtime field rendering

---

## 1. WHAT IS `END_USER_REF`?

### Definition & Location
**Backend Definition:**
- **File:** `modelcraft-backend/internal/domain/modeldesign/field_definition.go`
- **Type Definition:**
  ```go
  type FormatType string
  
  const (
    // ... other formats
    FormatEndUserRef FormatType = "END_USER_REF" // 归属用户
  )
  ```
- **FieldType Mapping:**
  ```go
  FormatEndUserRef: {
    SchemaType: SchemaTypeString,
    Format: FormatEndUserRef,
    Title: "归属用户"
  }
  ```

### Purpose & Semantics
- **Purpose:** Indicates a field that stores EndUser ID for **Row-Level Security (RLS)** data ownership/attribution
- **Field Name Convention:** When added to a model, the field is named `owner` (fixed convention)
- **Data Type:** Always `STRING` (stores UUID of the EndUser)
- **Null Semantics:** Always `NonNull: true` (mandatory, every record must have an owner)
- **Schema Representation:** In JSON Schema, becomes `type: "string"` with no custom `format` value (kept in `x-mc` namespace instead)

### Current Usage Pattern
**Backend applies END_USER_REF automatically to new models:**
```go
// GetNewModelSystemFields returns system fields for newly-created models.
// Compared with GetSystemFields(), this adds owner (EndUserRef) to enable RLS by default.
func GetNewModelSystemFields() []*FieldDefinition {
  fields := GetSystemFields()
  fields = append(fields, &FieldDefinition{
    Name:        "owner",
    Title:       "Owner",
    Description: "System Field",
    Type:        GetFieldTypeByFormat(FormatEndUserRef),
    NonNull:     true,
  })
  return fields
}
```
- **Location:** `modelcraft-backend/internal/domain/modeldesign/field_service.go:159-169`
- **All new models** include an `owner` field by default for RLS enablement

---

## 2. JSON SCHEMA `x-mc` FORMAT SYSTEM

### Contract Documentation
**File:** `ai-metadata/backend/design/domain-model/8-runtime/jsonschema-contract.md`

### Core Design Principles
1. **Topmost JSON Schema fields only contain standard values** (type, format, title, required, etc.)
2. **All ModelCraft-specific metadata → `x-mc` namespace** (single, non-polluting extension point)
3. **`format` field reserved for JSON Schema standard values only:**
   - `uuid`, `date`, `date-time`, `time` (standard values)
   - **NOT used for domain enums** (ENUM, STRING, INTEGER, etc.)
4. **`x-mc.widget` directly encodes rendering intent** → frontend reads once, zero inference needed

### ALL Known Format Types in Backend
**Location:** `modelcraft-backend/internal/domain/modeldesign/field_definition.go`

```go
const (
  // Based on string
  FormatString   FormatType = "STRING"
  FormatUUID     FormatType = "UUID"
  FormatDate     FormatType = "DATE"
  FormatDateTime FormatType = "DATETIME"
  FormatTime     FormatType = "TIME"

  // Based on number
  FormatNumber  FormatType = "NUMBER"
  FormatInteger FormatType = "INTEGER"
  FormatDecimal FormatType = "DECIMAL"

  // Based on boolean
  FormatBoolean FormatType = "BOOLEAN"

  // Relations & enums
  FormatRelation FormatType = "RELATION"      // Virtual field, has one-to-many relation
  FormatEnum     FormatType = "ENUM"          // Single-select enum
  FormatEnumArray FormatType = "ENUM_ARRAY"   // Multi-select enum

  // RLS/Ownership
  FormatEndUserRef FormatType = "END_USER_REF" // Ownership field for RLS
)
```

### `x-mc` Extension Object Structure
**Location:** `ai-metadata/backend/design/domain-model/8-runtime/jsonschema-contract.md`

```json
{
  "x-mc": {
    "widget": "enum-select|date|datetime-local|time|textarea|relation-selector|relation-multi-readonly",
    "isPrimary": boolean,
    "isUnique": boolean,
    "displayOrder": "lexicographic sort key",
    "storageHint": "TEXT|NUMBER|BOOLEAN|...",
    "validateRule": "email|url|phone",
    "precision": number,           // DECIMAL fields
    "scale": number,               // DECIMAL fields
    "minDate": "YYYY-MM-DD",       // DATE validation
    "maxDate": "YYYY-MM-DD",
    "minTime": "HH:MM:SS",         // TIME validation
    "maxTime": "HH:MM:SS",
    "belongsToFkId": "fk-uuid",    // FK column field
    "relation": {
      "databaseName": "string",
      "modelName": "string"
    },
    "enum": {
      "name": "EnumName",
      "displayName": "Display Label",
      "isMultiSelect": false,
      "options": [
        { "code": "value", "label": "Label", "description": "" }
      ]
    }
  }
}
```

### Widget Mapping (Backend → Frontend)
**`x-mc.widget` determines RJSF widget:**

| `x-mc.widget` | Trigger Condition | Frontend Widget |
|---|---|---|
| `"enum-select"` | format=ENUM or ENUM_ARRAY | EnumSchemaSelect |
| `"date"` | format=DATE | Native `<input type="date">` |
| `"datetime-local"` | format=DATETIME | Native `<input type="datetime-local">` |
| `"time"` | format=TIME | Native `<input type="time">` |
| `"textarea"` | storageHint=TEXT | `<textarea>` |
| `"relation-selector"` | BelongsToFKID != nil | RelationSelector |
| `"relation-multi-readonly"` | Virtual RELATION field | RelationMultiReadonly |
| _(omitted)_ | All other fields | RJSF default by type |

### END_USER_REF in JSON Schema Output
**Key Observation:** `END_USER_REF` format is **NOT placed into JSON Schema `format` field**. Instead:
- Field appears as `type: "string"` in JSON Schema
- **No special `format` hint** in top-level JSON Schema
- Metadata **must be in `x-mc` namespace** (backend should set `x-mc.format: "END_USER_REF"`)
- **Currently NO dedicated widget for END_USER_REF** in `x-mc.widget` enum

---

## 3. RUNTIME FIELD RENDERING (FRONTEND)

### Widget Mapping System
**File:** `modelcraft-front/src/web/components/features/model-editor/model-record-form/build-ui-schema.ts`

```typescript
const WIDGET_MAP: Record<XMCWidget, string> = {
  'enum-select': 'EnumSelect',
  'date': 'date',
  'datetime-local': 'datetime-local',
  'time': 'time',
  'textarea': 'textarea',
  'relation-selector': 'RelationSelector',
  'relation-multi-readonly': 'RelationMultiReadonly',
}

export function buildUiSchema(jsonSchema: RJSFSchema): UiSchema {
  const uiSchema: UiSchema = {}
  if (!jsonSchema.properties) return uiSchema

  for (const [fieldName, prop] of Object.entries(jsonSchema.properties)) {
    const xmc = getXMC(prop as Record<string, unknown>)
    const widget = xmc?.widget

    if (widget && WIDGET_MAP[widget]) {
      uiSchema[fieldName] = { 'ui:widget': WIDGET_MAP[widget] }
    }
  }
  return uiSchema
}
```

**Behavior:** Direct 1:1 mapping from `x-mc.widget` → RJSF widget string. No widget defined yet for END_USER_REF.

### XMC Type Definitions (Frontend)
**File:** `modelcraft-front/src/types/xmc.ts`

```typescript
export type XMCWidget =
  | 'enum-select'
  | 'date'
  | 'datetime-local'
  | 'time'
  | 'textarea'
  | 'relation-selector'
  | 'relation-multi-readonly'

export interface XMC {
  widget?: XMCWidget
  format?: string        // Currently unused in runtime rendering
  isPrimary?: boolean
  isUnique?: boolean
  displayOrder?: string
  nullable?: boolean
  storageHint?: string
  validateRule?: string
  precision?: number
  scale?: number
  minDate?: string
  maxDate?: string
  minTime?: string
  maxTime?: string
  relation?: XMCRelation
  enum?: XMCEnum
}

export function getXMC(prop: Record<string, unknown>): XMC | undefined {
  return prop['x-mc'] as XMC | undefined
}
```

**Observation:** `x-mc.format` field exists but is **not actively used in runtime rendering logic**. Widget selection depends solely on `x-mc.widget`.

### Field Filtering for Forms
**Files:**
- `modelcraft-front/src/web/components/features/model-editor/model-record-form/runtime/model-field-mapping.ts`
- `modelcraft-front/src/api-client/runtime-query/runtime-query-builder.ts`

**Writable Field Extraction:**
```typescript
export function extractWritableFieldNamesFromSchema(
  schema: { properties?: Record<string, unknown> } | null | undefined
): string[] {
  if (!schema?.properties) {
    return []
  }

  return Object.entries(schema.properties)
    .filter(([, prop]) => (prop as Record<string, unknown>).readOnly !== true)
    .map(([name]) => name)
}
```

**Current Logic:** A field is writable if:
1. `readOnly !== true` in JSON Schema
2. **Primary key fields** are marked `readOnly: true` (auto-excluded)
3. **RELATION virtual fields** are marked `readOnly: true` (auto-excluded)

**Question:** Are `END_USER_REF` fields currently marked `readOnly: true`? Need to verify in backend JSON schema generation.

### Mutation Data Sanitization
**File:** `modelcraft-front/src/api-client/runtime-query/runtime-query-builder.ts`

```typescript
export function sanitizeMutationInputData(
  data: Record<string, unknown> | null | undefined,
  allowedFieldNames: readonly string[]
): Record<string, unknown> {
  if (!data || typeof data !== 'object' || allowedFieldNames.length === 0) {
    return {}
  }

  const allowed = new Set(allowedFieldNames)

  return Object.fromEntries(
    Object.entries(data).filter(
      ([key, value]) => allowed.has(key) && value !== undefined
    )
  )
}
```

**Behavior:** Whitelist-based filtering—only fields in `allowedFieldNames` (i.e., not `readOnly`) are sent to backend.

---

## 4. INSERT/CREATE DATA FLOW (FRONTEND)

### Form Submission Flow
**File:** `modelcraft-front/src/web/components/features/model-editor/model-record-form/RuntimeRecordWorkspace.tsx`

**Entry Point: User clicks "Add Data" button**
```typescript
const handleCreate = () => {
  if (isManagedReadOnlyModel) {
    showManagedReadonlyToast()
    return
  }
  setCreateDataOpen(true)
}
```

**Form Display & Submission:**
```typescript
<Sheet open={createDataOpen} onOpenChange={setCreateDataOpen}>
  <SheetContent side="right" className="w-[450px] overflow-y-auto sm:max-w-[500px]">
    <SheetHeader>
      <SheetTitle>添加数据</SheetTitle>
      <SheetDescription>向 <span className="font-mono">{model.name}</span> 添加一条新记录</SheetDescription>
    </SheetHeader>

    {jsonSchema && (
      <ModelRecordForm
        jsonSchema={jsonSchema}
        onSubmit={async (data) => {
          if (isManagedReadOnlyModel) {
            showManagedReadonlyToast()
            throw new Error('托管模型仅支持查看')
          }
          setCreateSaving(true)
          try {
            // *** KEY STEP: Sanitize form data before mutation ***
            const sanitizedData = sanitizeMutationInputData(data, writableFieldNames)
            await createContent({ variables: { data: sanitizedData } })
            setCreateDataOpen(false)
          } catch (error) {
            throw error
          } finally {
            setCreateSaving(false)
          }
        }}
        onCancel={() => setCreateDataOpen(false)}
        isSubmitting={createSaving}
        orgName={orgName}
        projectSlug={projectSlug}
        databaseName={model.databaseName ?? ''}
        modelId={modelId}
      />
    )}
  </SheetContent>
</Sheet>
```

### Data Pipeline Summary
1. **User enters form data** → ModelRecordForm (RJSF-based)
2. **Form calls `onSubmit(data)`** with all form fields (including hidden/readonly fields)
3. **Frontend sanitizes:** `sanitizeMutationInputData(data, writableFieldNames)`
   - Extracts only non-readOnly fields
   - Removes undefined values
   - **Result:** Only user-editable fields sent to backend
4. **GraphQL mutation:** `createContent({ variables: { data: sanitizedData } })`
5. **Mutation response:** `{ id }` (only ID returned)
6. **Success callback:** `refetch()` to reload table

### Key Files in Flow
| File | Role |
|------|------|
| `RuntimeRecordWorkspace.tsx` | Orchestration, mutation setup, form container |
| `ModelRecordForm/index.tsx` | RJSF form wrapper, calls onSubmit |
| `build-ui-schema.ts` | Maps `x-mc.widget` → RJSF uiSchema |
| `runtime-query-builder.ts` | Builds GraphQL mutation, sanitizes data |
| `model-field-mapping.ts` | Maps model fields to table/form info |

---

## 5. ENDUSER CONCEPT

### What is EndUser in ModelCraft?
**Architectural Context:** Dual-UI system with two workspaces:
- **Tenant Workspace:** Admin/developer authoring models, managing data, setting permissions
- **EndUser Workspace:** Regular app users consuming data through a frontend interface

### EndUser in Architecture
**Files:**
- `ai-metadata/prd/dual-ui-architecture/00-overview.md`
- `ai-metadata/prd/dual-ui-architecture/02-enduser-login.md`
- `ai-metadata/prd/dual-ui-architecture/03-enduser-workspace.md`

**Key Concepts:**
1. **EndUser Identity:** Logged-in user in the app (identified by JWT token in httpOnly cookie)
2. **Data Scoping:** Row-Level Security (RLS) restricts which records EndUser can see/edit
3. **Owner Field:** The `owner` field (END_USER_REF) stores the EndUser ID who created/owns the record
4. **Access Control:** Backend checks `record.owner == current_user.id` before allowing CRUD

### EndUser Workspace Flow
```
/end-user/[orgName]/workspace
  ↓
  EndUser can view Projects they have access to
  ↓
  /end-user/[orgName]/workspace/[projectSlug]/data
    ↓
    View records filtered by RLS policy (e.g., owner == current_user.id)
    ↓
    Create new record (owner auto-populated to current_user.id by backend)
    ↓
    Edit/delete own records
```

### RuntimeRecordWorkspace (EndUser Runtime)
**File:** `modelcraft-front/src/web/components/features/model-editor/model-record-form/RuntimeRecordWorkspace.tsx`

**Purpose:** EndUser data workspace (not design-time)
- Uses **end-user scoped GraphQL client** (authenticated with end-user token)
- **Cannot insert fields** (field lifecycle = design-time only)
- **Cannot delete fields**
- **Can only CRUD records** (create, read, update, delete)
- Queries automatically scoped to current EndUser by backend RLS policy

```typescript
const endUserToken = getEndUserToken() // from localStorage
const managementClient = useMemo(() => {
  if (!endUserToken) return null
  return createEndUserScopedClient(orgName, projectSlug, endUserToken)
}, [orgName, projectSlug, endUserToken])
```

---

## 6. CRITICAL OBSERVATIONS & GAPS

### Gap 1: END_USER_REF Not in WIDGET_MAP
**Current State:** No frontend widget defined for END_USER_REF format
- `WIDGET_MAP` has 7 widget types, but no `'end-user-ref'` entry
- If backend sets `x-mc.widget: 'end-user-ref'`, frontend will silently ignore it → falls back to RJSF default (text input)
- **Design Decision Needed:** Should END_USER_REF be read-only (hidden) or editable by EndUser?

### Gap 2: Writable Field Detection
**Current Logic:** Depends on `readOnly: true` in JSON Schema
- Primary keys → `readOnly: true` ✓
- RELATION fields → `readOnly: true` ✓
- **END_USER_REF field → status unclear**
  - If `readOnly: false`, field is submittable in mutations
  - If `readOnly: true`, field is auto-filtered out by `sanitizeMutationInputData`
  - **Question:** Should backend mark `owner` as `readOnly: true`?

### Gap 3: Backend JSON Schema Generation for END_USER_REF
**Unknown:** How does backend currently emit END_USER_REF in JSON Schema?
- Does it set `x-mc.widget`? (If yes, to what value?)
- Does it set `readOnly`? (If yes, should be `true` to prevent user override)
- Does it set `x-mc.format: "END_USER_REF"`? (For frontend awareness)

### Gap 4: Mutation Input for END_USER_REF
**Unknown:** What happens when EndUser tries to submit a form with `owner` field?
- **Scenario A:** Backend enforces `owner` server-side → frontend can send any value (ignored)
- **Scenario B:** Backend validates `owner == current_user.id` → frontend must not send `owner` at all
- **Scenario C:** Backend allows frontend to override `owner` → security risk

---

## 7. RECOMMENDED DESIGN QUESTIONS

### For Runtime Field Handling, Answer These:

1. **Visibility:** Should `owner` (END_USER_REF) field:
   - [ ] Be completely hidden from the form? (readOnly + hidden widget)
   - [ ] Be displayed as read-only? (readOnly + display widget)
   - [ ] Be editable by EndUser? (normal input widget)

2. **Mutation Handling:** When EndUser creates a record:
   - [ ] Frontend must NOT include `owner` in mutation data (backend auto-assigns)
   - [ ] Frontend sends `owner` but backend always overrides with `current_user.id`
   - [ ] Frontend sends `owner` and backend honors it (trust EndUser? security model?)

3. **Widget Rendering:**
   - [ ] Use dedicated widget (new entry in WIDGET_MAP)
   - [ ] Treat as readonly text display (no widget needed)
   - [ ] Treat as hidden field (x-mc.widget = null, readOnly = true)

4. **Schema Marking:**
   - [ ] Extend `x-mc.format` to include "END_USER_REF" (informational for frontend)?
   - [ ] Rely only on `readOnly` flag for filtering?
   - [ ] Add new `x-mc.field-type` enum?

---

## 8. SUMMARY TABLE

| Aspect | Current Status | Location |
|--------|---|---|
| **Format Enum Definition** | ✓ Defined as `FormatEndUserRef` | `modelcraft-backend/.../field_definition.go` |
| **FieldType Mapping** | ✓ Maps to SchemaTypeString | `modelcraft-backend/.../field_definition.go` |
| **Auto-Add to New Models** | ✓ Added as `owner` field | `modelcraft-backend/.../field_service.go:159-169` |
| **x-mc Contract** | ✓ Documented in RFC | `ai-metadata/backend/design/domain-model/8-runtime/jsonschema-contract.md` |
| **Frontend XMC Types** | ✓ Defined (but `format` unused) | `modelcraft-front/src/types/xmc.ts` |
| **Widget Mapping** | ✗ No END_USER_REF widget | `modelcraft-front/.../build-ui-schema.ts` |
| **Writable Field Filter** | ✓ Uses `readOnly` check | `modelcraft-front/.../runtime-query-builder.ts` |
| **Form Sanitization** | ✓ Whitelist-based | `modelcraft-front/.../runtime-query-builder.ts` |
| **EndUser Scoping** | ✓ Implemented in RuntimeRecordWorkspace | `modelcraft-front/.../RuntimeRecordWorkspace.tsx` |
| **JSON Schema Output for END_USER_REF** | ? Unknown (needs verification) | `modelcraft-backend/.../jsonschema_generator.go` |

---

## 9. FILES TO INSPECT/MODIFY

### Backend (Go)
- `modelcraft-backend/internal/domain/modeldesign/field_definition.go` — Format enum
- `modelcraft-backend/internal/domain/modeldesign/field_service.go` — System field setup
- `modelcraft-backend/internal/domain/modeldesign/jsonschema_generator.go` — **KEY:** How END_USER_REF is emitted in JSON Schema
- `modelcraft-backend/api/graph/project/schema/runtime-json-schema-contract.md` — API contract documentation

### Frontend (TypeScript/React)
- `modelcraft-front/src/types/xmc.ts` — XMC type definitions (may need `XMCWidget` extension)
- `modelcraft-front/src/web/components/features/model-editor/model-record-form/build-ui-schema.ts` — Widget mapping
- `modelcraft-front/src/web/components/features/model-editor/model-record-form/runtime/model-field-mapping.ts` — Field filtering logic
- `modelcraft-front/src/api-client/runtime-query/runtime-query-builder.ts` — Writable field extraction, sanitization

### Documentation
- `ai-metadata/backend/design/domain-model/8-runtime/jsonschema-contract.md` — Master contract
- `modelcraft-backend/api/graph/project/schema/runtime-json-schema-contract.md` — Synced copy
- `modelcraft-front/contract/graph/project/schema/runtime-json-schema-contract.md` — Frontend contract

