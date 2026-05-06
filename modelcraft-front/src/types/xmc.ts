/**
 * x-mc namespace types — ModelCraft extensions to JSON Schema Draft 7.
 *
 * All non-standard fields injected by the backend live under the `x-mc` key.
 * Use `getXMC(prop)` to safely read extensions from a schema property.
 */

export interface XMCRelation {
  databaseName: string
  modelName: string
  belongsToFkId?: string
  relateFkId?: string
  relationType: 'ONE_TO_MANY' | 'MANY_TO_ONE'
  relationDirection: 'reverse' | 'normal'
}

export interface XMCEnum {
  labelFieldName: string
}

export type XMCWidget =
  | 'enum-select'
  | 'date'
  | 'datetime-local'
  | 'time'
  | 'textarea'
  | 'relation-selector'
  | 'relation-multi-readonly'
  | 'end-user-ref'

export interface XMC {
  widget?: XMCWidget
  format?: string
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

/**
 * Safely read the `x-mc` extension from a JSON Schema property object.
 */
export function getXMC(prop: Record<string, unknown>): XMC | undefined {
  return prop['x-mc'] as XMC | undefined
}
