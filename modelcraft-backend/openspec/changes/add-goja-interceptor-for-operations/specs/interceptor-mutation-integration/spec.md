# interceptor-mutation-integration Specification

## Purpose
Define how JavaScript interceptors integrate with GraphQL mutation operations (createOne, createMany, updateOne, updateMany, deleteOne, deleteMany) to modify data payloads, enforce validation rules, and implement fine-grained authorization.

## Why
Mutation interceptors enable automatic injection of audit fields (createdBy, updatedBy), tenant context, validation rules, and authorization checks before data is written to the database. This ensures data integrity and compliance without requiring developers to manually implement these patterns in every mutation.

## Related Specifications
- **interceptor-core**: Defines core interceptor execution engine
- **interceptor-query-integration**: Defines query operation integration patterns

## ADDED Requirements

### Requirement: Interceptor Integration for createOne Operation

The createOne operation SHALL support interceptor execution to modify data payloads, inject default values, and enforce validation rules before record creation.

#### Scenario: Interceptor injects default values into data payload

- **WHEN** a createOne mutation is executed with an interceptor configured
- **THEN** the interceptor SHALL receive `input.data` containing the fields to be inserted
- **AND** the interceptor can add new fields with default values
- **AND** the interceptor can modify existing field values
- **AND** the modified data payload SHALL be passed to the repository CreateOne method

**Example**:

**Original Mutation**:
```graphql
mutation {
  createOne(data: { name: "John Doe", email: "john@example.com" }) {
    id
    createdObj { id, name, email, createdBy }
  }
}
```

**Interceptor Script** (inject audit fields):
```javascript
function intercept() {
    // Add creator information
    input.data.createdBy = user.id;
    input.data.createdAt = new Date(environment.timestamp);

    // Add tenant context
    input.data.tenantId = user.metadata.tenantId;

    return input;
}

intercept();
```

**Effective Data Payload**:
```javascript
{
  name: "John Doe",
  email: "john@example.com",
  createdBy: "user_123",
  createdAt: "2025-01-06T10:30:00Z",
  tenantId: "tenant_456"
}
```

#### Scenario: Interceptor validates data before creation

- **WHEN** an interceptor performs validation on createOne data
- **THEN** the interceptor can throw an error if validation fails
- **AND** the CreateOne repository method SHALL NOT be called
- **AND** the client SHALL receive a validation error
- **AND** the error message SHALL describe the validation failure

**Example Script** (email domain validation):
```javascript
function intercept() {
    // Validate email domain
    const allowedDomains = ["company.com", "partner.com"];
    const emailDomain = input.data.email.split("@")[1];

    if (!contains(allowedDomains, emailDomain)) {
        throw new Error("Email domain not allowed: " + emailDomain);
    }

    // Validate required fields
    if (!input.data.department) {
        throw new Error("Department is required");
    }

    return input;
}

intercept();
```

#### Scenario: Interceptor sanitizes sensitive data

- **WHEN** an interceptor needs to prevent storage of sensitive information
- **THEN** the interceptor can remove or redact sensitive fields
- **AND** the modified data SHALL be stored in the database
- **AND** this prevents accidental storage of passwords, tokens, or PII

**Example Script**:
```javascript
function intercept() {
    // Remove sensitive fields that shouldn't be stored
    delete input.data.password;      // Passwords should be hashed separately
    delete input.data.ssn;           // SSN shouldn't be in this table
    delete input.data.creditCard;    // PCI compliance

    // Redact partial data
    if (input.data.phoneNumber) {
        // Keep only last 4 digits visible
        input.data.phoneNumber = "***-***-" + input.data.phoneNumber.slice(-4);
    }

    return input;
}

intercept();
```

#### Scenario: Interceptor denies creation based on permissions

- **WHEN** an interceptor determines the user lacks permission to create the record
- **THEN** the interceptor SHALL throw an error
- **AND** the CreateOne repository method SHALL NOT be called
- **AND** the error message SHALL indicate permission denial

**Example Script**:
```javascript
function intercept() {
    // Only managers can create users in specific departments
    if (input.data.department === "executive" && !hasPermission(user, "manager")) {
        throw new Error("Only managers can create executive department users");
    }

    // Users cannot create records with higher privilege levels
    if (input.data.privilegeLevel > user.metadata.privilegeLevel) {
        throw new Error("Cannot create user with higher privilege level");
    }

    return input;
}

intercept();
```

---

### Requirement: Interceptor Integration for createMany Operation

The createMany operation SHALL support interceptor execution to modify batch data payloads and enforce consistent validation across all records.

#### Scenario: Interceptor processes each record in batch

- **WHEN** a createMany mutation is executed with an interceptor configured
- **THEN** the interceptor SHALL receive `input.data` as an array of records
- **AND** the interceptor can iterate over each record and modify it
- **AND** modifications SHALL be applied consistently to all records
- **AND** the modified array SHALL be passed to the repository CreateMany method

**Example**:

**Original Mutation**:
```graphql
mutation {
  createMany(data: [
    { name: "User 1", email: "user1@example.com" },
    { name: "User 2", email: "user2@example.com" }
  ]) {
    count
  }
}
```

**Interceptor Script**:
```javascript
function intercept() {
    // Add tenant ID and creator to all records
    input.data.forEach(function(record) {
        record.tenantId = user.metadata.tenantId;
        record.createdBy = user.id;
        record.createdAt = new Date(environment.timestamp);
    });

    return input;
}

intercept();
```

**Effective Data Payload**:
```javascript
[
  {
    name: "User 1",
    email: "user1@example.com",
    tenantId: "tenant_456",
    createdBy: "user_123",
    createdAt: "2025-01-06T10:30:00Z"
  },
  {
    name: "User 2",
    email: "user2@example.com",
    tenantId: "tenant_456",
    createdBy: "user_123",
    createdAt: "2025-01-06T10:30:00Z"
  }
]
```

#### Scenario: Interceptor validates entire batch before creation

- **WHEN** an interceptor performs validation on createMany data
- **THEN** the interceptor can validate each record individually
- **AND** if ANY record fails validation, the entire batch SHALL be rejected
- **AND** the error message SHALL indicate which record(s) failed validation
- **AND** this enforces atomicity (all-or-nothing)

**Example Script**:
```javascript
function intercept() {
    const errors = [];

    // Validate each record
    input.data.forEach(function(record, index) {
        if (!record.email || !record.email.includes("@")) {
            errors.push("Record " + index + ": Invalid email");
        }
        if (!record.department) {
            errors.push("Record " + index + ": Department is required");
        }
    });

    // Reject entire batch if any errors
    if (errors.length > 0) {
        throw new Error("Batch validation failed:\n" + errors.join("\n"));
    }

    return input;
}

intercept();
```

#### Scenario: Interceptor enforces batch size limits

- **WHEN** an interceptor needs to prevent excessively large batch operations
- **THEN** the interceptor can check the array length
- **AND** if the batch size exceeds the limit, an error SHALL be thrown
- **AND** this prevents performance issues and resource exhaustion

**Example Script**:
```javascript
function intercept() {
    const MAX_BATCH_SIZE = 1000;

    if (input.data.length > MAX_BATCH_SIZE) {
        throw new Error("Batch size " + input.data.length + " exceeds maximum of " + MAX_BATCH_SIZE);
    }

    return input;
}

intercept();
```

---

### Requirement: Interceptor Integration for updateOne Operation

The updateOne operation SHALL support interceptor execution to modify update payloads and enforce field-level permissions before record updates.

#### Scenario: Interceptor modifies update data payload

- **WHEN** an updateOne mutation is executed with an interceptor configured
- **THEN** the interceptor SHALL receive `input.data` containing fields to be updated
- **AND** the interceptor SHALL receive `input.where` containing the record identifier
- **AND** the interceptor can modify update values or add new fields
- **AND** the modified data and WHERE clause SHALL be passed to the repository UpdateOne method

**Example**:

**Original Mutation**:
```graphql
mutation {
  updateOne(
    where: { id: "123" }
    data: { name: "Updated Name", status: "active" }
  ) {
    success
    updatedObj { id, name, status, updatedBy, updatedAt }
  }
}
```

**Interceptor Script** (track modification):
```javascript
function intercept() {
    // Add modification tracking
    input.data.updatedBy = user.id;
    input.data.updatedAt = new Date(environment.timestamp);

    // Increment version for optimistic locking
    if (input.data.version !== undefined) {
        input.data.version = input.data.version + 1;
    }

    return input;
}

intercept();
```

**Effective Data Payload**:
```javascript
{
  name: "Updated Name",
  status: "active",
  updatedBy: "user_123",
  updatedAt: "2025-01-06T10:30:00Z",
  version: 2
}
```

#### Scenario: Interceptor prevents updates to protected fields

- **WHEN** an interceptor enforces field-level permissions
- **THEN** the interceptor can remove protected fields from the update payload
- **AND** if critical protected fields are modified, an error can be thrown
- **AND** this implements attribute-based access control (ABAC) at field level

**Example Script**:
```javascript
function intercept() {
    const protectedFields = ["createdAt", "createdBy", "tenantId", "id"];

    // Check if user is trying to modify protected fields
    protectedFields.forEach(function(field) {
        if (input.data[field] !== undefined) {
            if (hasPermission(user, "admin")) {
                // Admins can modify, but log it
                console.log("Admin modifying protected field: " + field);
            } else {
                // Non-admins cannot modify
                throw new Error("Permission denied: cannot modify field " + field);
            }
        }
    });

    return input;
}

intercept();
```

#### Scenario: Interceptor enforces row-level security on updates

- **WHEN** an interceptor needs to ensure users can only update their own records
- **THEN** the interceptor can add conditions to `input.where`
- **AND** the modified WHERE clause SHALL restrict which records can be updated
- **AND** if no records match the modified WHERE, the update SHALL fail

**Example Script**:
```javascript
function intercept() {
    // Non-admins can only update their own records
    if (!hasPermission(user, "admin")) {
        input.where.createdBy = { equals: user.id };
    }

    // Users in "restricted" status cannot update
    if (user.metadata.accountStatus === "restricted") {
        throw new Error("Account is restricted - updates not allowed");
    }

    return input;
}

intercept();
```

**SQL Generated** (non-admin user):
```sql
UPDATE users
SET name = 'Updated Name', status = 'active', updated_by = 'user_123'
WHERE id = '123' AND created_by = 'user_123';
```

---

### Requirement: Interceptor Integration for updateMany Operation

The updateMany operation SHALL support interceptor execution to modify batch update payloads and enforce WHERE clause restrictions.

#### Scenario: Interceptor modifies WHERE clause for batch updates

- **WHEN** an updateMany mutation is executed with an interceptor configured
- **THEN** the interceptor SHALL receive `input.where` containing the filter conditions
- **AND** the interceptor SHALL receive `input.data` containing fields to be updated
- **AND** the interceptor can add conditions to WHERE to restrict which records are updated
- **AND** the modified WHERE and data SHALL be passed to the repository UpdateMany method

**Example**:

**Original Mutation**:
```graphql
mutation {
  updateMany(
    where: { status: { equals: "pending" } }
    data: { status: "active" }
  ) {
    count
  }
}
```

**Interceptor Script** (tenant scoping):
```javascript
function intercept() {
    // Only update records in user's tenant
    input.where.tenantId = { equals: user.metadata.tenantId };

    // Add audit field
    input.data.updatedBy = user.id;

    return input;
}

intercept();
```

**SQL Generated**:
```sql
UPDATE users
SET status = 'active', updated_by = 'user_123'
WHERE status = 'pending' AND tenant_id = 'tenant_456';
```

#### Scenario: Interceptor prevents mass updates without filters

- **WHEN** an interceptor detects an updateMany without WHERE conditions
- **THEN** the interceptor can require specific filters to prevent accidental mass updates
- **AND** if WHERE is empty or too broad, an error can be thrown
- **AND** this protects against unintended data modification

**Example Script**:
```javascript
function intercept() {
    // Prevent updates without any WHERE conditions
    if (!input.where || Object.keys(input.where).length === 0) {
        throw new Error("updateMany requires WHERE clause - mass updates not allowed");
    }

    // Require at least one specific field filter
    const hasSpecificFilter = Object.keys(input.where).some(function(key) {
        return key !== 'AND' && key !== 'OR' && key !== 'NOT';
    });

    if (!hasSpecificFilter) {
        throw new Error("updateMany requires specific field filters");
    }

    return input;
}

intercept();
```

---

### Requirement: Interceptor Integration for deleteOne Operation

The deleteOne operation SHALL support interceptor execution to enforce deletion permissions and implement soft-delete patterns before record deletion.

#### Scenario: Interceptor enforces row-level security on deletes

- **WHEN** a deleteOne mutation is executed with an interceptor configured
- **THEN** the interceptor SHALL receive `input.where` containing the record identifier
- **AND** the interceptor can add conditions to WHERE to restrict which records can be deleted
- **AND** the modified WHERE clause SHALL be passed to the repository DeleteOne method
- **AND** if no records match the modified WHERE, the deletion SHALL fail

**Example**:

**Original Mutation**:
```graphql
mutation {
  deleteOne(where: { id: "123" }) {
    success
  }
}
```

**Interceptor Script** (ownership check):
```javascript
function intercept() {
    // Non-admins can only delete their own records
    if (!hasPermission(user, "admin")) {
        input.where.createdBy = { equals: user.id };
    }

    // Additional tenant check
    input.where.tenantId = { equals: user.metadata.tenantId };

    return input;
}

intercept();
```

**SQL Generated** (non-admin user):
```sql
DELETE FROM users
WHERE id = '123' AND created_by = 'user_123' AND tenant_id = 'tenant_456';
```

#### Scenario: Interceptor implements soft-delete pattern

- **WHEN** an interceptor implements soft-delete instead of hard-delete
- **THEN** the interceptor can transform the delete operation into an update
- **AND** instead of calling DeleteOne, the system SHALL call UpdateOne
- **AND** the record SHALL be marked as deleted (e.g., `deleted: true`, `deletedAt: timestamp`)
- **AND** this enables data recovery and audit trails

**Example Script**:
```javascript
function intercept() {
    // Transform delete into soft-delete (update operation)
    // Note: This requires special handling in the resolver to detect soft-delete

    input._operation = "update"; // Signal to use update instead of delete
    input.data = {
        deleted: true,
        deletedAt: new Date(environment.timestamp),
        deletedBy: user.id
    };

    return input;
}

intercept();
```

**Note**: This scenario requires resolver support to detect `input._operation = "update"` and call UpdateOne instead of DeleteOne.

#### Scenario: Interceptor prevents deletion of protected records

- **WHEN** an interceptor needs to prevent deletion of critical records
- **THEN** the interceptor can check record attributes and throw an error
- **AND** the DeleteOne repository method SHALL NOT be called
- **AND** this protects system data and critical configurations

**Example Script**:
```javascript
function intercept() {
    // Prevent deletion of system accounts
    if (input.where.id === "system_admin" || input.where.email === "admin@system.com") {
        throw new Error("Cannot delete system accounts");
    }

    // Prevent deletion during restricted hours
    const hour = new Date(environment.timestamp).getHours();
    if (hour < 9 || hour > 17) {
        throw new Error("Deletion operations only allowed during business hours (9 AM - 5 PM)");
    }

    return input;
}

intercept();
```

---

### Requirement: Interceptor Integration for deleteMany Operation

The deleteMany operation SHALL support interceptor execution to enforce WHERE clause restrictions and prevent mass deletions.

#### Scenario: Interceptor restricts deleteMany scope

- **WHEN** a deleteMany mutation is executed with an interceptor configured
- **THEN** the interceptor SHALL receive `input.where` containing the filter conditions
- **AND** the interceptor can add conditions to WHERE to restrict which records can be deleted
- **AND** the modified WHERE clause SHALL be passed to the repository DeleteMany method
- **AND** only records matching the modified WHERE SHALL be deleted

**Example**:

**Original Mutation**:
```graphql
mutation {
  deleteMany(where: { status: { equals: "archived" } }) {
    count
  }
}
```

**Interceptor Script**:
```javascript
function intercept() {
    // Only delete records in user's tenant
    input.where.tenantId = { equals: user.metadata.tenantId };

    // Only allow deleting archived records older than 90 days
    const ninetyDaysAgo = new Date();
    ninetyDaysAgo.setDate(ninetyDaysAgo.getDate() - 90);

    input.where.archivedAt = { lt: ninetyDaysAgo.toISOString() };

    return input;
}

intercept();
```

**SQL Generated**:
```sql
DELETE FROM documents
WHERE status = 'archived'
  AND tenant_id = 'tenant_456'
  AND archived_at < '2024-10-08T00:00:00Z';
```

#### Scenario: Interceptor prevents deleteMany without filters

- **WHEN** an interceptor detects a deleteMany without WHERE conditions
- **THEN** the interceptor SHALL throw an error to prevent accidental mass deletion
- **AND** this protects against catastrophic data loss

**Example Script**:
```javascript
function intercept() {
    // Absolutely prevent deleteMany without WHERE
    if (!input.where || Object.keys(input.where).length === 0) {
        throw new Error("deleteMany without WHERE is not allowed - specify filters");
    }

    return input;
}

intercept();
```

---

### Requirement: Mutation Operation Error Handling

The system SHALL handle interceptor errors consistently across all mutation operations.

#### Scenario: Interceptor validation error aborts mutation

- **WHEN** an interceptor throws a validation error during a mutation operation
- **THEN** the repository mutation method SHALL NOT be called
- **AND** the database SHALL NOT be modified
- **AND** the client SHALL receive a GraphQL error with the validation message
- **AND** the error response SHALL follow standard GraphQL error format

**Error Response Example**:
```json
{
  "errors": [
    {
      "message": "Validation failed: Email domain not allowed",
      "extensions": {
        "code": "INTERCEPTOR_VALIDATION_ERROR",
        "modelName": "User",
        "operation": "createOne",
        "interceptorId": 123
      }
    }
  ],
  "data": {
    "createOne": null
  }
}
```

#### Scenario: Permission denied error in mutation

- **WHEN** an interceptor denies a mutation operation due to insufficient permissions
- **THEN** the error code SHALL be "INTERCEPTOR_PERMISSION_DENIED"
- **AND** the error message SHALL describe the permission requirement
- **AND** this distinguishes authorization failures from validation failures

---

### Requirement: Audit Logging for Mutation Operations

The system SHALL log interceptor actions on mutation operations for compliance and security auditing.

#### Scenario: Data modifications are logged

- **WHEN** an interceptor modifies data in a mutation operation
- **THEN** an audit log entry SHALL be created (optional, configurable)
- **AND** the log SHALL include original and modified data
- **AND** this enables compliance auditing (GDPR, SOX, HIPAA)

**Audit Log Example**:
```json
{
  "timestamp": "2025-01-06T10:30:00Z",
  "level": "info",
  "type": "interceptor_mutation",
  "requestId": "01930c8a-1234-7890-abcd-ef1234567890",
  "userId": "user_123",
  "operation": {
    "modelName": "User",
    "operation": "createOne"
  },
  "interceptor": {
    "id": 456,
    "name": "audit-injector"
  },
  "changes": {
    "addedFields": ["createdBy", "tenantId"],
    "originalData": { "name": "John Doe", "email": "john@example.com" },
    "modifiedData": {
      "name": "John Doe",
      "email": "john@example.com",
      "createdBy": "user_123",
      "tenantId": "tenant_456"
    }
  },
  "executionTimeMs": 2
}
```

#### Scenario: Denied mutations are logged

- **WHEN** an interceptor denies a mutation operation
- **THEN** an audit log entry SHALL be created
- **AND** the log SHALL include the denial reason
- **AND** this enables security monitoring and threat detection

---

### Requirement: Performance Impact on Mutation Operations

The system SHALL minimize performance overhead of interceptors on mutation operations.

#### Scenario: Interceptor overhead is measured for mutations

- **WHEN** a mutation operation executes with an interceptor
- **THEN** the interceptor execution time SHALL be measured separately
- **AND** typical overhead SHALL be < 5ms for simple interceptors
- **AND** batch operations (createMany, updateMany, deleteMany) SHALL have overhead proportional to batch size
- **AND** performance metrics SHALL be exposed for monitoring

**Performance Targets**:
- createOne/updateOne/deleteOne: < 5ms interceptor overhead
- createMany (100 records): < 20ms interceptor overhead
- updateMany/deleteMany (any size): < 5ms interceptor overhead (WHERE modification only)
