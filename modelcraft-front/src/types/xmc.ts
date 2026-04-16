/**
 * x-mc namespace types — ModelCraft extensions to JSON Schema Draft 7.
 *
 * All non-standard fields injected by the backend live under the `x-mc` key.
 * Use `getXMC(prop)` to safely read extensions from a schema property.
 */

export interface XMCRelation {
  databaseName: string
  modelName: string
}

export interface XMCEnumOption {
  code: string
  label: string
  description?: string
}

export interface XMCEnum {
  name: string
  displayName: string
  description?: string
  isMultiSelect: boolean
  options: XMCEnumOption[]
}

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
  belongsToFkId?: string
  relation?: XMCRelation
  relateFkId?: string
  enum?: XMCEnum
}

/**
 * Safely read the `x-mc` extension from a JSON Schema property object.
 */
export function getXMC(prop: Record<string, unknown>): XMC | undefined {
  return prop['x-mc'] as XMC | undefined
}
