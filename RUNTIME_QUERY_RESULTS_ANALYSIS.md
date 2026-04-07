# ModelCraft Frontend: Runtime GraphQL Query Results Display Analysis

## Overview
This analysis documents how the ModelCraft frontend displays runtime GraphQL query results in data tables/grids, with specific focus on how relation fields are currently handled.

---

## 1. Key File Paths

### Core Table Components
- **`modelcraft-front/src/web/components/model-editor/DynamicModelTable.tsx`** (695 lines)
  - Main wrapper component for runtime query results
  - Fetches model schema and field definitions
  - Manages create/edit/delete operations
  - Builds and executes `findMany` queries

- **`modelcraft-front/src/web/components/model-editor/ModelRecordTable.tsx`** (325 lines)
  - Actual table rendering component (React UI)
  - Receives pre-processed data and field info
  - Uses `renderCellValue()` to format cell content
  - Supports column resizing, tooltips, copy-to-clipboard

### Cell Value Rendering
- **`modelcraft-front/src/web/components/model-editor/fieldProtocol.ts`** (55 lines)
  - **KEY FILE**: Contains `renderCellValue()` function
  - Defines protocol for displaying different field types
  - **Has special handling for RELATION fields**

### Query Building
- **`modelcraft-front/src/bff/cms/runtime-query-builder.ts`** (299 lines)
  - Dynamically builds GraphQL queries using `gql-query-builder`
  - **KEY FUNCTION**: `buildFieldSelections()` - converts field array to GraphQL selection set
  - **Detects RELATION fields** and creates sub-selections

### Data Mapping
- **`modelcraft-front/src/web/components/model-editor/modelFieldMapping.ts`** (53 lines)
  - Maps backend model fields to runtime field definitions
  - `mapModelFieldsToRuntimeFields()` converts field metadata
  - Preserves field `format` property (used to detect RELATION)

---

## 2. How Query Results Are Displayed

### Query Flow
```
DynamicModelTable (parent)
  ├─ Fetches model metadata (GET_MODEL_QUERY)
  ├─ Fetches JSON Schema (MODEL_JSON_SCHEMA_QUERY)
  ├─ Builds findMany query via buildFindManyQuery()
  ├─ Extracts field definitions from model.fields
  ├─ Creates field→prop mapping for cell rendering
  └─ Passes data to ModelRecordTable (child)

ModelRecordTable (rendering)
  ├─ Receives:
  │  ├─ contentList: Array<Record<string, unknown>>
  │  ├─ displayFields: string[] (field names to show)
  │  ├─ propByName: Record<fieldName, schemaProperty>
  │  └─ getFieldInfo(): (fieldName) => ModelRecordTableFieldInfo
  │
  └─ For each row:
     ├─ For each field:
     │  ├─ Gets rawValue from row[fieldName]
     │  ├─ Gets schemaProperty from propByName[fieldName]
     │  ├─ Calls renderCellValue(rawValue, schemaProperty)
     │  └─ Displays rendered value in table cell
     └─ Displays edit/delete action buttons
```

### Cell Rendering Code (ModelRecordTable.tsx, lines 229-291)
```typescript
{visibleFields.map((field) => {
  const rawValue = item[field]
  
  if (rawValue === null || rawValue === undefined) {
    return <TableCell key={field}>
      <span className="font-mono text-xs text-muted-foreground/50">NULL</span>
    </TableCell>
  }
  
  const renderedValue = renderCellValue(rawValue, propByName[field] ?? {})
  
  return <TableCell key={field}>
    <Tooltip>
      <TooltipTrigger>
        <span className="block truncate text-sm font-normal text-foreground">
          {renderedValue}
        </span>
      </TooltipTrigger>
      {/* Tooltip with copy button */}
    </Tooltip>
  </TableCell>
})}
```

---

## 3. Relation Field Handling (Current Implementation)

### Detection Mechanism
**Relation fields are identified in the JSON Schema with these markers:**

```typescript
// fieldProtocol.ts, lines 27-32
if (prop.type === 'object' && (prop['x-relateFkId'] || prop['x-belongsToFkId'])) {
  // This is a relation field
}
```

**Field metadata flow:**
```
Backend API generates field.format = 'RELATION'
  ↓
mapModelFieldsToRuntimeFields() preserves format property
  ↓
buildFieldSelections() detects format === 'RELATION'
  ↓
GraphQL query includes: { relationField: ['id'] }
  ↓
Runtime returns: relationField: { id: "...", name: "...", __typename: "..." }
  ↓
renderCellValue() handles as object type
```

### Rendering Logic for Relations (fieldProtocol.ts, lines 24-36)

```typescript
export function renderCellValue(value: unknown, prop: SchemaProperty): string {
  if (value === null || value === undefined) return ''
  
  // RELATION field handling:
  if (prop.type === 'object' && (prop['x-relateFkId'] || prop['x-belongsToFkId'])) {
    if (typeof value === 'object' && value !== null) {
      const rel = value as Record<string, unknown>
      // Display: rel.name first, fallback to rel.id
      return String(rel.name ?? rel.id ?? '')
    }
    return ''
  }
  
  // For all other types: convert to string, limit 100 chars
  return String(value).slice(0, 100)
}
```

**Key Points:**
- ✅ Relations ARE detected and handled specially
- ✅ Displays `name` attribute if available, falls back to `id`
- ✅ Only includes `['id']` in query selection set to minimize payload
- ✅ Schema includes `x-relateFkId` or `x-belongsToFkId` to mark relations
- ✅ Backend schema builder marks relation fields with `type: 'object'`

---

## 4. Query Building Details

### buildFieldSelections() Function (runtime-query-builder.ts, lines 43-59)

```typescript
function buildFieldSelections(
  fields: string[] | FieldDefinition[]
): (string | Record<string, string[]>)[] {
  if (fields.length === 0) {
    return ['id']
  }

  return fields.map((field) => {
    if (typeof field === 'string') {
      return field  // scalar field
    }
    
    // RELATION field detected
    if (isRelationField(field)) {
      return { [field.name]: ['id'] }  // nested selection set
    }
    
    return field.name
  })
}

// Check function (line 34-36)
function isRelationField(field: FieldDefinition): boolean {
  return field.format === 'RELATION'
}
```

### Query Output Example
For a model with fields: `id`, `name`, `category` (relation), `author` (relation)

```graphql
query FindMany($take: Int, $skip: Int) {
  findMany(take: $take, skip: $skip) {
    timeCost
    reqId
    items {
      id
      name
      category {
        id
      }
      author {
        id
      }
    }
  }
}
```

---

## 5. How Relation Fields Might Be Skipped (Potential Issues)

### Current Behavior Summary
- ✅ Relation fields ARE included in the table display
- ✅ Relation fields ARE queried with `{ relationField: ['id'] }` selection
- ✅ Relation objects ARE rendered with `name ?? id` fallback logic

### However, Potential Gaps:

1. **No `name` field in relation sub-selection**
   - Current: `{ relationField: ['id'] }`
   - Problem: Backend returns only `id`, but frontend tries to display `name`
   - Result: Falls back to `id` (works, but shows ID not name)
   - **ISSUE**: The relation sub-selection doesn't request the `name` field!

2. **Silent fallback if relation object lacks both `name` and `id`**
   - Returns empty string if object exists but both fields missing
   - Frontend just displays nothing (appears as blank cell)

3. **No handling for array relations**
   - If a relation field returns an array of objects
   - Current code treats it as single object
   - Might error or display `[object Object]`

4. **Display order limited by displayFields**
   - Only first 6 fields shown (line 123 of ModelRecordTable.tsx)
   - Relation fields might be cut off in large models

---

## 6. Component Hierarchy

```
DynamicModelTable
├─ Query data from:
│  ├─ projectClient.useQuery(GET_MODEL_QUERY) → model metadata + fields
│  ├─ projectClient.useQuery(MODEL_JSON_SCHEMA_QUERY) → JSON schema
│  └─ runtimeClient.useQuery(findManyQuery) → actual records
│
├─ Process:
│  ├─ Extract fields from model.fields
│  ├─ Build field selections via buildFieldSelections()
│  ├─ Create propByName mapping from JSON schema
│  ├─ Build fieldInfo array (name, title, type, storageHint)
│  └─ Extract first 6 fields as displayFields
│
└─ Render ModelRecordTable
   ├─ Props:
   │  ├─ contentList: findMany result items
   │  ├─ displayFields: first 6 field names
   │  ├─ propByName: name → schema property mapping
   │  ├─ getFieldInfo: callback to get field metadata
   │  └─ getFieldTypeDisplay: format type for header
   │
   └─ For each cell:
      ├─ Get rawValue from contentList item
      ├─ Get prop from propByName[fieldName]
      ├─ Call renderCellValue(rawValue, prop)
      └─ Render in TableCell
```

---

## 7. Data Flow for Relation Fields

### Example: User model with `department` relation field

```
1. Model Definition
   ├─ field.name: "department"
   ├─ field.format: "RELATION"
   └─ field.schemaType: "string" (or object)

2. Runtime Field Processing
   ├─ FieldDefinition created with format: "RELATION"
   └─ buildFieldSelections() returns: { department: ['id'] }

3. GraphQL Query
   findMany(take: 50, skip: 0) {
     items {
       id
       name
       department {
         id
       }
     }
   }

4. API Response (from runtime endpoint)
   {
     "findMany": {
       "items": [
         {
           "id": "user123",
           "name": "John",
           "department": {
             "id": "dept456",
             "__typename": "Department"
           }
         }
       ]
     }
   }

5. Table Cell Rendering
   ├─ rawValue: { id: "dept456", __typename: "Department" }
   ├─ prop: { type: "object", "x-relateFkId": "fk123" }
   ├─ renderCellValue(rawValue, prop):
   │  └─ Finds: prop.type === 'object' && prop['x-relateFkId']
   │  └─ Reads: rawValue.name ?? rawValue.id
   │  └─ Result: "dept456" (only id available)
   └─ Display: "dept456"
```

---

## 8. Schema Property Structure

### How JSON Schema Marks Relation Fields

From backend, a relation field looks like:
```typescript
{
  properties: {
    department: {
      type: "object",
      "x-relateFkId": "fk_user_department",
      "x-belongsToFkId": null,  // OR this for reverse relations
      title: "Department",
      // ... other properties
    }
  }
}
```

### getFieldProtocols() Function (fieldProtocol.ts, lines 42-54)
```typescript
export function getFieldProtocols(
  schema: RJSFSchema
): Array<{ name: string; prop: SchemaProperty }> {
  if (!schema.properties) return []
  
  return Object.entries(schema.properties)
    .map(([name, prop]) => ({ name, prop: prop as SchemaProperty }))
    .sort((a, b) => {
      // Sort by x-displayOrder
      const oa = String(a.prop['x-displayOrder'] ?? '')
      const ob = String(b.prop['x-displayOrder'] ?? '')
      return oa.localeCompare(ob)
    })
}
```

This builds the `propByName` lookup that's passed to `renderCellValue()`.

---

## 9. Key Findings Summary

| Aspect | Status | Details |
|--------|--------|---------|
| **Relation Field Detection** | ✅ Implemented | Via `prop.type === 'object' && (x-relateFkId or x-belongsToFkId)` |
| **Query Building** | ✅ Implemented | `buildFieldSelections()` creates `{ relationField: ['id'] }` |
| **Cell Rendering** | ✅ Implemented | `renderCellValue()` displays `name ?? id ?? ''` |
| **Fallback Handling** | ✅ Implemented | Falls back to `id` if `name` missing |
| **Display in Table** | ✅ Implemented | Limited to first 6 fields (line 123) |
| **Nested Object Support** | ⚠️ Limited | Works for single objects, not arrays |
| **Full Relation Data** | ❌ Not Included | Only queries `['id']` sub-selection, doesn't request `name` |
| **Array Relations** | ⚠️ Unclear | No explicit handling for array-type relations |

---

## 10. Code References

| File | Lines | Purpose |
|------|-------|---------|
| `DynamicModelTable.tsx` | 265-282 | Query building and execution |
| `DynamicModelTable.tsx` | 242-247 | propByName creation from schema |
| `ModelRecordTable.tsx` | 229-291 | Cell rendering loop |
| `fieldProtocol.ts` | 24-36 | renderCellValue() implementation |
| `fieldProtocol.ts` | 42-54 | getFieldProtocols() for schema parsing |
| `runtime-query-builder.ts` | 43-59 | buildFieldSelections() with RELATION detection |
| `runtime-query-builder.ts` | 34-36 | isRelationField() check |
| `modelFieldMapping.ts` | 18-26 | Field format preservation |
| `RelationPicker.tsx` | 70-186 | Form widget for selecting relations |

