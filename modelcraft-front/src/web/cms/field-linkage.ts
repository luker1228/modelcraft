/**
 * Field Linkage Support
 * Utilities and examples for adding field reactions in Formily forms
 *
 * Formily supports field linkage through the x-reactions property.
 * This allows conditional display, value calculations, and dynamic options.
 */

import { type FormilySchema } from '@shared/cms/schema-transformer'

/** Reaction object used in Formily's x-reactions field linkage */
export type FormilyReaction = {
  dependencies?: string[]
  fulfill?: {
    state?: Record<string, unknown>
    run?: string
  }
  target?: string
  when?: string
  otherwise?: Record<string, unknown>
}

/**
 * Example: Conditional Display
 * Show/hide field based on another field's value
 */
export function addConditionalDisplay(
  schema: FormilySchema,
  targetField: string,
  sourceField: string,
  condition: (value: unknown) => boolean
): FormilySchema {
  if (!schema.properties[targetField]) {
    console.warn(`Field ${targetField} not found in schema`)
    return schema
  }

  schema.properties[targetField]['x-reactions'] = {
    dependencies: [sourceField],
    fulfill: {
      state: {
        visible: `{{$deps[0] && ${condition.toString()}($deps[0])}}`,
      },
    },
  }

  return schema
}

/**
 * Example: Dynamic Required
 * Make field required based on another field's value
 */
export function addDynamicRequired(
  schema: FormilySchema,
  targetField: string,
  sourceField: string,
  condition: string
): FormilySchema {
  if (!schema.properties[targetField]) {
    console.warn(`Field ${targetField} not found in schema`)
    return schema
  }

  schema.properties[targetField]['x-reactions'] = {
    dependencies: [sourceField],
    fulfill: {
      state: {
        required: `{{${condition}}}`,
      },
    },
  }

  return schema
}

/**
 * Example: Value Calculation
 * Calculate field value based on other fields
 */
export function addValueCalculation(
  schema: FormilySchema,
  targetField: string,
  sourceFields: string[],
  calculation: string
): FormilySchema {
  if (!schema.properties[targetField]) {
    console.warn(`Field ${targetField} not found in schema`)
    return schema
  }

  schema.properties[targetField]['x-reactions'] = {
    dependencies: sourceFields,
    fulfill: {
      state: {
        value: `{{${calculation}}}`,
      },
    },
  }

  return schema
}

/**
 * Common Field Linkage Patterns
 */

// Pattern 1: Show field when enum equals specific value
export const showWhenEquals = (
  sourceField: string,
  value: string
) => ({
  dependencies: [sourceField],
  fulfill: {
    state: {
      visible: `{{$deps[0] === '${value}'}}`,
    },
  },
})

// Pattern 2: Show field when enum is one of multiple values
export const showWhenIn = (
  sourceField: string,
  values: string[]
) => ({
  dependencies: [sourceField],
  fulfill: {
    state: {
      visible: `{{${JSON.stringify(values)}.includes($deps[0])}}`,
    },
  },
})

// Pattern 3: Required when another field has value
export const requiredWhenHasValue = (sourceField: string) => ({
  dependencies: [sourceField],
  fulfill: {
    state: {
      required: `{{!!$deps[0]}}`,
    },
  },
})

// Pattern 4: Calculate sum of two fields
export const calculateSum = (field1: string, field2: string) => ({
  dependencies: [field1, field2],
  fulfill: {
    state: {
      value: `{{($deps[0] || 0) + ($deps[1] || 0)}}`,
    },
  },
})

// Pattern 5: Calculate percentage
export const calculatePercentage = (
  numeratorField: string,
  denominatorField: string
) => ({
  dependencies: [numeratorField, denominatorField],
  fulfill: {
    state: {
      value: `{{$deps[1] ? (($deps[0] || 0) / $deps[1] * 100).toFixed(2) : 0}}`,
    },
  },
})

/**
 * Example Usage:
 *
 * ```typescript
 * import { transformToFormilySchema } from '@shared/cms/schema-transformer'
 * import { showWhenEquals, calculateSum } from '@web/cms/field-linkage'
 *
 * // Transform JSON Schema
 * const formilySchema = transformToFormilySchema(jsonSchema)
 *
 * // Add field linkage
 * formilySchema.properties.companyName['x-reactions'] = showWhenEquals('userType', 'company')
 * formilySchema.properties.total['x-reactions'] = calculateSum('price', 'quantity')
 *
 * // Use in FormRenderer
 * <FormRenderer schema={formilySchema} />
 * ```
 */

/**
 * Complex Example: Multi-step Form with Conditional Steps
 */
export const conditionalSteps = {
  // Step 1: User Type Selection
  userType: {
    type: 'string',
    title: 'User Type',
    enum: ['individual', 'company'],
    'x-decorator': 'FormItem',
    'x-component': 'Select',
  },

  // Step 2: Individual Fields (only shown for individual users)
  firstName: {
    type: 'string',
    title: 'First Name',
    'x-decorator': 'FormItem',
    'x-component': 'Input',
    'x-reactions': showWhenEquals('userType', 'individual'),
  },

  lastName: {
    type: 'string',
    title: 'Last Name',
    'x-decorator': 'FormItem',
    'x-component': 'Input',
    'x-reactions': showWhenEquals('userType', 'individual'),
  },

  // Step 3: Company Fields (only shown for company users)
  companyName: {
    type: 'string',
    title: 'Company Name',
    'x-decorator': 'FormItem',
    'x-component': 'Input',
    'x-reactions': showWhenEquals('userType', 'company'),
  },

  taxId: {
    type: 'string',
    title: 'Tax ID',
    'x-decorator': 'FormItem',
    'x-component': 'Input',
    'x-reactions': showWhenEquals('userType', 'company'),
  },
}

/**
 * Helper: Apply multiple reactions to a field
 */
export function applyReactions(
  schema: FormilySchema,
  fieldName: string,
  reactions: FormilyReaction[]
): FormilySchema {
  if (!schema.properties[fieldName]) {
    console.warn(`Field ${fieldName} not found in schema`)
    return schema
  }

  // If single reaction, apply directly
  if (reactions.length === 1) {
    schema.properties[fieldName]['x-reactions'] = reactions[0]
    return schema
  }

  // If multiple reactions, use array
  schema.properties[fieldName]['x-reactions'] = reactions

  return schema
}
