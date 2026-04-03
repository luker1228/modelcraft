/**
 * Schema Transformer
 * Converts JSON Schema from Model API to Formily Schema
 */

export interface JSONSchemaProperty {
  type: string
  title?: string
  description?: string
  format?: string
  enum?: string[]
  minimum?: number
  maximum?: number
  minLength?: number
  maxLength?: number
  pattern?: string
  default?: unknown
  required?: boolean
}

export interface JSONSchema {
  type: 'object'
  properties: Record<string, JSONSchemaProperty>
  required?: string[]
  title?: string
  description?: string
}

export interface FormilyValidatorRule {
  pattern?: RegExp
  message: string
  min?: number
  max?: number
  minimum?: number
  maximum?: number
}

export interface FormilySchemaProperty {
  type: string
  title?: string
  description?: string
  'x-decorator': string
  'x-component': string
  'x-component-props'?: Record<string, unknown>
  'x-validator'?: FormilyValidatorRule[]
  'x-reactions'?: unknown
  default?: unknown
  required?: boolean
  enum?: string[]
}

export interface FormilySchema {
  type: 'object'
  properties: Record<string, FormilySchemaProperty>
}

/**
 * Get Formily component type based on JSON Schema type and format
 */
function getComponentType(
  type: string,
  format?: string,
  enumValues?: string[]
): string {
  // Enum fields use Select
  if (enumValues && enumValues.length > 0) {
    return 'Select'
  }

  // Handle different types
  switch (type) {
    case 'string':
      if (format === 'date') return 'DatePicker'
      if (format === 'datetime') return 'DatePicker'
      if (format === 'time') return 'TimePicker'
      return 'Input'

    case 'number':
    case 'integer':
      return 'NumberPicker'

    case 'boolean':
      return 'Switch'

    case 'array':
      return 'Select' // Multi-select for arrays

    default:
      return 'Input'
  }
}

/**
 * Get Formily component props based on JSON Schema constraints
 */
function getComponentProps(
  property: JSONSchemaProperty
): Record<string, unknown> {
  const props: Record<string, unknown> = {}

  // Enum options
  if (property.enum) {
    props.options = property.enum.map((value) => ({
      label: value,
      value: value,
    }))
  }

  // Number constraints
  if (property.type === 'number' || property.type === 'integer') {
    if (property.minimum !== undefined) {
      props.min = property.minimum
    }
    if (property.maximum !== undefined) {
      props.max = property.maximum
    }
  }

  // String length constraints
  if (property.type === 'string') {
    if (property.maxLength !== undefined) {
      props.maxLength = property.maxLength
    }
  }

  // Date format props
  if (property.format === 'datetime') {
    props.showTime = true
    props.format = 'YYYY-MM-DD HH:mm:ss'
  } else if (property.format === 'date') {
    props.format = 'YYYY-MM-DD'
  } else if (property.format === 'time') {
    props.format = 'HH:mm:ss'
  }

  // Placeholder from description
  if (property.description) {
    props.placeholder = property.description
  }

  return props
}

/**
 * Get Formily validator rules based on JSON Schema constraints
 */
function getValidator(property: JSONSchemaProperty): FormilyValidatorRule[] | undefined {
  const validators: FormilyValidatorRule[] = []

  // Pattern validation
  if (property.pattern) {
    validators.push({
      pattern: new RegExp(property.pattern),
      message: `Value must match pattern: ${property.pattern}`,
    })
  }

  // Length validation for strings
  if (property.type === 'string') {
    if (property.minLength !== undefined) {
      validators.push({
        min: property.minLength,
        message: `Minimum length is ${property.minLength}`,
      })
    }
    if (property.maxLength !== undefined) {
      validators.push({
        max: property.maxLength,
        message: `Maximum length is ${property.maxLength}`,
      })
    }
  }

  // Range validation for numbers
  if (property.type === 'number' || property.type === 'integer') {
    if (property.minimum !== undefined) {
      validators.push({
        minimum: property.minimum,
        message: `Minimum value is ${property.minimum}`,
      })
    }
    if (property.maximum !== undefined) {
      validators.push({
        maximum: property.maximum,
        message: `Maximum value is ${property.maximum}`,
      })
    }
  }

  return validators.length > 0 ? validators : undefined
}

/**
 * Transform JSON Schema to Formily Schema
 */
export function transformToFormilySchema(jsonSchema: JSONSchema): FormilySchema {
  const properties: Record<string, FormilySchemaProperty> = {}

  for (const [key, value] of Object.entries(jsonSchema.properties)) {
    const component = getComponentType(value.type, value.format, value.enum)
    const componentProps = getComponentProps(value)
    const validator = getValidator(value)
    const isRequired = jsonSchema.required?.includes(key) || false

    properties[key] = {
      type: value.type,
      title: value.title || key,
      description: value.description,
      'x-decorator': 'FormItem',
      'x-component': component,
      'x-component-props': componentProps,
      required: isRequired,
      default: value.default,
    }

    if (validator) {
      properties[key]['x-validator'] = validator
    }

    // Add enum for Select components
    if (value.enum) {
      properties[key].enum = value.enum
    }
  }

  return {
    type: 'object',
    properties,
  }
}

/**
 * Parse JSON Schema string and transform to Formily Schema
 */
export function parseAndTransformSchema(schemaJSON: string): FormilySchema {
  try {
    const jsonSchema = JSON.parse(schemaJSON) as JSONSchema
    return transformToFormilySchema(jsonSchema)
  } catch (error) {
    console.error('Failed to parse JSON Schema:', error)
    throw new Error('Invalid JSON Schema format')
  }
}

/**
 * Add custom field reactions for conditional display
 * This is a helper to add x-reactions to schema properties
 */
export function addFieldReactions(
  schema: FormilySchema,
  fieldName: string,
  reactions: unknown
): FormilySchema {
  if (!schema.properties[fieldName]) {
    console.warn(`Field ${fieldName} not found in schema`)
    return schema
  }

  schema.properties[fieldName]['x-reactions'] = reactions
  return schema
}
