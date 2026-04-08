export type HealthStatus = 'HEALTHY' | 'NEEDS_REPAIR' | 'BROKEN'

export type SchemaIssueType =
  | 'TABLE_MISSING'
  | 'FIELD_MISSING'
  | 'FIELD_TYPE_MISMATCH'
  | 'FIELD_CONSTRAINT_MISMATCH'
  | 'DATABASE_MISSING'
  | 'CLUSTER_NOT_FOUND'
  | 'FIELD_HAS_DEPENDENCIES'

export interface SchemaIssue {
  type: SchemaIssueType
  description: string
  tableName: string
  fieldName?: string
  details?: string
}
