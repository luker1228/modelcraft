/**
 * Check if a string is a valid JSON object (not array, not primitive).
 */
export function isValidJson(value: string): boolean {
  if (!value.trim()) return false
  try {
    const parsed = JSON.parse(value) as unknown
    return typeof parsed === 'object' && parsed !== null && !Array.isArray(parsed)
  } catch {
    return false
  }
}

/**
 * Pretty-print JSON. Returns original string on parse failure.
 */
export function formatJson(value: string): string {
  try {
    return JSON.stringify(JSON.parse(value), null, 2)
  } catch {
    return value
  }
}

/**
 * Count top-level filter conditions for the filter button badge.
 *
 * Returns:
 * - number  — length of AND/OR array
 * - '•'     — non-AND/OR structure (single-field condition)
 * - null    — no active filter (empty, invalid JSON, or empty object)
 *
 * AI note: this reads `whereJsonCommitted`. AI agents should write a valid
 * where JSON string and this function will derive the badge automatically.
 */
export function getFilterCount(whereJson: string | null): number | '•' | null {
  if (!whereJson) return null
  try {
    const parsed = JSON.parse(whereJson) as unknown
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return null
    const obj = parsed as Record<string, unknown>
    if (Array.isArray(obj.AND)) return obj.AND.length || null
    if (Array.isArray(obj.OR)) return obj.OR.length || null
    const keys = Object.keys(obj).filter((k) => k !== 'NOT')
    return keys.length > 0 ? '•' : null
  } catch {
    return null
  }
}

/** 数值操作符集合——这些操作符的值需要 coerce 为 number */
const NUMERIC_OPERATORS = new Set(['gt', 'gte', 'lt', 'lte'])

/**
 * 单条筛选行的数据结构。
 * fieldType 仅在 equals/not 时影响值的类型推断；
 * gt/gte/lt/lte 始终 coerce 为 number。
 */
export interface FilterRow {
  id: string
  field: string
  operator: string
  value: string
  /** 字段的存储类型，用于 equals/not 时决定值类型。
   *  来源：FieldDefinition.storageHint?.toUpperCase() 或 schemaType?.toUpperCase()
   *  可选，缺失时按字符串处理。
   */
  fieldType?: string
}

/**
 * 将结构化条件行数组转为后端 where JSON 对象。
 *
 * 规则：
 * - 跳过 field 为空或 value 为空的行
 * - 所有有效行用 AND 数组包裹
 * - 只有 1 行有效时仍然包在 AND 数组中（始终统一格式）
 * - 无有效行返回 null
 * - gt/gte/lt/lte 操作符：value 强制转为 number
 * - equals/not：fieldType 为 NUMBER/INT/FLOAT/DECIMAL → 转 number；
 *               fieldType 为 BOOLEAN → "true"/"false" 转 boolean；
 *               其他 → 保持 string
 */
export function filterRowsToWhereJson(
  rows: FilterRow[]
): Record<string, unknown> | null {
  const conditions: Record<string, unknown>[] = []

  for (const row of rows) {
    if (!row.field.trim() || !row.value.trim()) continue

    let coercedValue: unknown = row.value

    if (NUMERIC_OPERATORS.has(row.operator)) {
      const n = Number(row.value)
      coercedValue = Number.isNaN(n) ? row.value : n
    } else if (row.operator === 'equals' || row.operator === 'not') {
      const ft = row.fieldType?.toUpperCase() ?? ''
      if (['NUMBER', 'INT', 'INTEGER', 'FLOAT', 'DECIMAL'].includes(ft)) {
        const n = Number(row.value)
        coercedValue = Number.isNaN(n) ? row.value : n
      } else if (ft === 'BOOLEAN') {
        coercedValue = row.value === 'true'
      }
    }

    conditions.push({ [row.field]: { [row.operator]: coercedValue } })
  }

  if (conditions.length === 0) return null
  return { AND: conditions }
}
