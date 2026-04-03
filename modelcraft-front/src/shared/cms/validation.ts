/**
 * Form Validation Utilities
 * Provides validation functions for form data against JSON Schema
 */

import Ajv, { ErrorObject } from 'ajv'
import addFormats from 'ajv-formats'
import type { JSONSchemaProperty } from './schema-transformer'

// Create AJV instance with formats support
const ajv = new Ajv({ allErrors: true })
addFormats(ajv)

export interface ValidationError {
  field: string
  message: string
}

export interface ValidationResult {
  valid: boolean
  errors: ValidationError[]
}

interface ValidationRule {
  required?: boolean
  message: string
  type?: string
  min?: number
  max?: number
  pattern?: RegExp
  minimum?: number
  maximum?: number
  enum?: string[]
}

interface SchemaWithRequirements {
  required?: string[]
  [key: string]: unknown
}

/**
 * Validate form data against JSON Schema
 * @param schema - JSON Schema object
 * @param data - Form data to validate
 * @returns Validation result with errors
 */
export function validateAgainstSchema(
  schema: Record<string, unknown>,
  data: unknown
): ValidationResult {
  try {
    const validate = ajv.compile(schema)
    const valid = validate(data)

    if (valid) {
      return {
        valid: true,
        errors: [],
      }
    }

    const errors: ValidationError[] = (validate.errors || []).map(
      (error: ErrorObject) => ({
        field: error.instancePath.replace(/^\//, '') || 'root',
        message: formatErrorMessage(error),
      })
    )

    return {
      valid: false,
      errors,
    }
  } catch (error) {
    console.error('Validation error:', error)
    return {
      valid: false,
      errors: [
        {
          field: 'root',
          message: 'Schema validation failed',
        },
      ],
    }
  }
}

/**
 * Format AJV error message to be more user-friendly
 */
function formatErrorMessage(error: ErrorObject): string {
  const { keyword, message, params } = error

  switch (keyword) {
    case 'required':
      return `${String(params.missingProperty)} is required`
    case 'type':
      return `Must be of type ${String(params.type)}`
    case 'minimum':
      return `Must be at least ${String(params.limit)}`
    case 'maximum':
      return `Must be at most ${String(params.limit)}`
    case 'minLength':
      return `Must be at least ${String(params.limit)} characters`
    case 'maxLength':
      return `Must be at most ${String(params.limit)} characters`
    case 'pattern':
      return `Must match pattern: ${String(params.pattern)}`
    case 'enum': {
      if (Array.isArray(params.allowedValues)) {
        return `Must be one of: ${params.allowedValues.map(String).join(', ')}`
      }
      return message || 'Validation failed'
    }
    case 'format':
      return `Must be a valid ${String(params.format)}`
    default:
      return message || 'Validation failed'
  }
}

/**
 * Validate a single field against its schema definition
 */
export function validateField(
  fieldName: string,
  value: unknown,
  fieldSchema: JSONSchemaProperty
): ValidationError | null {
  const schema: Record<string, unknown> = {
    type: 'object',
    properties: {
      [fieldName]: fieldSchema,
    },
  }

  const data: Record<string, unknown> = { [fieldName]: value }
  const result = validateAgainstSchema(schema, data)

  if (result.valid) {
    return null
  }

  return result.errors[0] || null
}

/**
 * Check if a field is required in the schema
 */
export function isFieldRequired(
  fieldName: string,
  schema: SchemaWithRequirements
): boolean {
  return schema.required?.includes(fieldName) || false
}

/**
 * Get validation rules from JSON Schema for a specific field
 * Returns Formily-compatible validation rules
 */
export function getFieldValidationRules(fieldSchema: JSONSchemaProperty): ValidationRule[] {
  const rules: ValidationRule[] = []

  // Required rule
  if (fieldSchema.required) {
    rules.push({
      required: true,
      message: 'This field is required',
    })
  }

  // Type validation
  if (fieldSchema.type) {
    rules.push({
      type: fieldSchema.type,
      message: `Must be of type ${fieldSchema.type}`,
    })
  }

  // String length validation
  if (fieldSchema.type === 'string') {
    if (fieldSchema.minLength !== undefined) {
      rules.push({
        min: fieldSchema.minLength,
        message: `Minimum length is ${fieldSchema.minLength}`,
      })
    }
    if (fieldSchema.maxLength !== undefined) {
      rules.push({
        max: fieldSchema.maxLength,
        message: `Maximum length is ${fieldSchema.maxLength}`,
      })
    }
    if (fieldSchema.pattern) {
      rules.push({
        pattern: new RegExp(fieldSchema.pattern),
        message: `Must match pattern: ${fieldSchema.pattern}`,
      })
    }
  }

  // Number range validation
  if (fieldSchema.type === 'number' || fieldSchema.type === 'integer') {
    if (fieldSchema.minimum !== undefined) {
      rules.push({
        minimum: fieldSchema.minimum,
        message: `Minimum value is ${fieldSchema.minimum}`,
      })
    }
    if (fieldSchema.maximum !== undefined) {
      rules.push({
        maximum: fieldSchema.maximum,
        message: `Maximum value is ${fieldSchema.maximum}`,
      })
    }
  }

  // Enum validation
  if (fieldSchema.enum) {
    rules.push({
      enum: fieldSchema.enum,
      message: `Must be one of: ${fieldSchema.enum.join(', ')}`,
    })
  }

  return rules
}

/**
 * Transform validation errors to Formily format
 */
export function transformToFormilyErrors(
  errors: ValidationError[]
): Record<string, string> {
  const formilyErrors: Record<string, string> = {}

  for (const error of errors) {
    formilyErrors[error.field] = error.message
  }

  return formilyErrors
}

/**
 * Custom validator function for Formily
 * Can be used as x-validator in schema
 */
export function createSchemaValidator(jsonSchema: Record<string, unknown>) {
  return (value: unknown, _rule: unknown) => {
    const result = validateAgainstSchema(jsonSchema, value)

    if (result.valid) {
      return ''
    }

    return result.errors.map((e) => e.message).join('; ')
  }
}
