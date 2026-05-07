# Dropdown Component Analysis Report

## Executive Summary

The "— 不指定 —" dropdown is the **EndUserSelectorWidget** component used in the ModelCraft tenant (admin) workspace for selecting which EndUser should be the owner of a data record. It represents the **END_USER_REF field type** in models.

---

## 1. The Dropdown Component

### Location
- **File**: `modelcraft-front/src/web/components/features/model-editor/model-record-form/widgets/EndUserSelectorWidget.tsx`

### Purpose
- Allows **tenant admin users** (in design/admin workspace) to manually assign an owner (EndUser) to a record when creating/editing data
- Fetches available EndUsers from the organization and presents them in a dropdown
- Includes a special "— 不指定 —" (unspecified) option that maps to `undefined` / `null`

### Implementation Details
```tsx
export function EndUserSelectorWidget(props: WidgetProps) {
  const value = props.value as string | undefined
  const onChange = props.onChange
  
  // Fetches users from org-scoped findUsers query
  const handleChange = (val: string) => {
    onChange(val === '__none__' ? undefined : val)  // '__none__' → undefined
  }

  return (
    <Select value={value ?? '__none__'} onValueChange={handleChange}>
      <SelectTrigger>
        <SelectValue placeholder={loading ? '加载中...' : '选择用户'} />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="__none__">— 不指定 —</SelectItem>  {/* The dropdown option */}
        {users.map((user) => (
          <SelectItem key={user.id} value={user.id}>
            {user.username}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
```

---

## 2. What Field/Model It Represents

### Field Type: END_USER_REF
- **Backend Format Definition**: `FormatEndUserRef = "END_USER_REF"`
- **Location**: `modelcraft-backend/internal/domain/modeldesign/field_definition.go`

### Model Integration
- **Field Name**: Always named `owner` (convention)
- **Data Type**: STRING (stores UUID of the EndUser)
- **NonNull**: Always `true` (mandatory—every record must have an owner)
- **Purpose**: Stores the ID of the EndUser who owns/created the record for Row-Level Security (RLS)

### Automatic System Field
When a new model is created, the backend automatically adds an `owner` field:
```go
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

---

## 3. JSON Schema & x-mc Widget System

### JSON Schema Representation
The END_USER_REF field appears in JSON Schema as:
```json
{
  "owner": {
    "type": "string",
    "title": "Owner",
    "description": "System Field",
    "x-mc": {
      "widget": "end-user-ref",
      "format": "END_USER_REF",
      "isPrimary": false,
      "isUnique": false,
      "nullable": false,
      "displayOrder": "..."
    }
  }
}
```

### Widget Pipeline
1. **Backend** (`jsonschema_generator.go` line 304-305):
   - Detects field type `FormatEndUserRef`
   - Sets `x-mc.widget = "end-user-ref"`

2. **Frontend** (`build-ui-schema.ts`):
   ```typescript
   if (widget === 'end-user-ref') {
     if (workspaceMode === 'end_user') {
       uiSchema[fieldName] = { 'ui:widget': 'hidden' }  // Hidden in end-user workspace
     } else {
       uiSchema[fieldName] = { 'ui:widget': 'EndUserSelectorWidget' }  // Shown in design workspace
     }
   }
   ```

3. **RJSF** renders using the `EndUserSelectorWidget` component

### Workspace Mode Handling
- **`design` mode** (tenant admin): Shows `EndUserSelectorWidget` dropdown
- **`end_user` mode** (app user): Hides the field (backend auto-injects current user ID)

---

## 4. "— 不指定 —" (Unspecified) Option

### Semantics
- **Display Text**: "— 不指定 —" (Chinese: "unspecified")
- **Internal Value**: `"__none__"`
- **Transformed Value**: `undefined` / `null`

### Usage
When this option is selected, the form submission sends `owner: undefined`, which means:
- No specific owner is assigned at the moment
- Depends on backend/mutation logic to handle this case

### Context in Form Submission
```typescript
const handleChange = (val: string) => {
  onChange(val === '__none__' ? undefined : val)
}
```
The component converts the placeholder value `"__none__"` to JavaScript `undefined`.

---

## 5. RBAC & Data Access Control Context

### Row-Level Security (RLS)
The `owner` field (END_USER_REF) is fundamental to ModelCraft's Row-Level Security:

1. **Data Ownership**: Each record has an `owner` ID that identifies which EndUser created/owns it
2. **Access Filtering**: EndUser queries are automatically filtered by `WHERE owner = <current_user_id>`
3. **Permission Scopes**: RBAC roles can specify row access policies:
   - `READ_WRITE_ALL`: Full access (no owner field needed)
   - `READ_ALL`: Read-only, all records
   - `READ_WRITE_OWNER`: Read/write only own records (requires `owner` field)
   - `READ_ALL_WRITE_OWNER`: Read all, write only own (requires `owner` field)

### "Platform Admin Marker" - NOT FOUND
After thorough exploration, **there is NO concept called "platform admin marker"** in the codebase. 

Possible interpretations:
- **If you meant "system marker"**: The `owner` field itself serves as a marker of data ownership in the system
- **If you meant "admin/tenant workspace distinction"**: The `workspaceMode` parameter marks whether we're in `design` (admin) vs `end_user` (app user) mode
- **If you meant something else**: Please clarify—the term doesn't appear in PRDs, specs, or design docs

---

## 6. Data Flow: Creating a Record with Owner

### Tenant (Admin) Workspace Flow
```
1. Admin opens Model detail page → "Add Data" button
2. Form sheet opens with `EndUserSelectorWidget` for owner field
3. Admin selects an EndUser from dropdown (or leaves "— 不指定 —")
4. Form submission:
   - Data: { owner: "<selected-user-id>" | undefined, ...otherFields }
   - Sanitization: Only non-readOnly fields sent to backend
   - Mutation: createContent({ data: sanitizedData })
5. Backend:
   - Receives owner: "<user-id>"
   - Stores record with that owner
   - Returns { id } on success
6. Frontend: Refetches table data, shows new record
```

### EndUser (App) Workspace Flow
```
1. EndUser opens Model data table
2. Form sheet opens for new record
3. owner field is HIDDEN (readOnly + hidden widget)
4. Form submission:
   - Data: { ...otherFields } (owner NOT included)
   - Mutation: createContent({ data: sanitizedData })
5. Backend:
   - Receives owner undefined/not in payload
   - Auto-injects owner = current_user_id from JWT
   - Stores record with current EndUser as owner
6. Frontend: Refetches, shows new record filtered by owner
```

---

## 7. Key Files & References

| Component | File | Purpose |
|-----------|------|---------|
| Widget | `modelcraft-front/src/web/components/features/model-editor/model-record-form/widgets/EndUserSelectorWidget.tsx` | Dropdown selector for END_USER_REF fields |
| UI Schema Builder | `modelcraft-front/src/web/components/features/model-editor/model-record-form/build-ui-schema.ts` | Routes `end-user-ref` widget to EndUserSelectorWidget |
| JSON Schema Generator | `modelcraft-backend/internal/domain/modeldesign/jsonschema_generator.go` | Sets `x-mc.widget = "end-user-ref"` |
| Field Types | `modelcraft-backend/internal/domain/modeldesign/field_definition.go` | Defines `FormatEndUserRef` constant |
| x-mc Types (Frontend) | `modelcraft-front/src/types/xmc.ts` | TypeScript types for `x-mc` metadata |
| Domain Model Spec | `ai-metadata/backend/design/domain-model/4-rbac.md` | RBAC architecture & Membership model |
| Bundle Versioning | `ai-metadata/prd/rbac/05-bundle-versioning.md` | Permission bundle versioning (related to RBAC) |

---

## 8. Related RBAC Concepts

### EndUserPermissionBundle
- Represents a collection of permissions granted to EndUsers on a Project
- Stores both column-level and row-level policies
- Can reference END_USER_REF field for "owner-based" access patterns

### Row Scope Options (via RowScopeSelector)
```typescript
const SCOPE_OPTIONS = [
  { value: 'ALL', label: '全部行' },  // All rows, no filtering
  { value: 'SELF', label: '仅自己的行', requires: 'owner field (END_USER_REF)' },
  { value: 'DEPT', label: '所在部门的行', requires: 'dept_id field' },
  { value: 'DEPT_AND_CHILDREN', label: '部门及下级的行', requires: 'dept_id field' },
]
```

The `SELF` option is only available if the model has an `owner` (END_USER_REF) field.

---

## Summary

| Aspect | Answer |
|--------|--------|
| **What is the dropdown?** | EndUserSelectorWidget—a form input for selecting an EndUser (record owner) |
| **Field it represents** | `owner` field, type END_USER_REF |
| **Data type** | STRING (stores EndUser UUID) |
| **"— 不指定 —" meaning** | Placeholder option that sends `undefined` to backend |
| **Model context** | System field automatically added to all new models for RLS |
| **Workspace scope** | Visible in tenant/design workspace; hidden in end-user workspace |
| **Platform admin marker** | **Not found in codebase**—clarification needed |
| **RBAC relevance** | Owner field enables role-based row-level access control |

