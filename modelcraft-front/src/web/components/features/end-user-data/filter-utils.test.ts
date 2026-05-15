import { describe, it, expect } from 'vitest'
import { isValidJson, formatJson, getFilterCount, filterRowsToWhereJson } from './filter-utils'
import type { FilterRow } from './filter-utils'

describe('isValidJson', () => {
  it('returns true for valid JSON object', () => {
    expect(isValidJson('{"name": {"contains": "张"}}')).toBe(true)
  })
  it('returns false for empty string', () => {
    expect(isValidJson('')).toBe(false)
  })
  it('returns false for invalid JSON', () => {
    expect(isValidJson('{bad json')).toBe(false)
  })
  it('returns false for JSON array (not an object)', () => {
    expect(isValidJson('[1,2,3]')).toBe(false)
  })
})

describe('formatJson', () => {
  it('pretty-prints valid JSON', () => {
    expect(formatJson('{"a":1}')).toBe('{\n  "a": 1\n}')
  })
  it('returns original string on invalid JSON', () => {
    expect(formatJson('{bad')).toBe('{bad')
  })
})

describe('getFilterCount', () => {
  it('returns null for null input', () => {
    expect(getFilterCount(null)).toBeNull()
  })
  it('returns AND array length', () => {
    expect(getFilterCount('{"AND":[{},{}]}')).toBe(2)
  })
  it('returns OR array length', () => {
    expect(getFilterCount('{"OR":[{}]}')).toBe(1)
  })
  it('returns bullet for single-field condition', () => {
    expect(getFilterCount('{"name":{"contains":"张"}}')).toBe('•')
  })
  it('returns null for invalid JSON', () => {
    expect(getFilterCount('{bad')).toBeNull()
  })
  it('returns null for empty object', () => {
    expect(getFilterCount('{}')).toBeNull()
  })
  it('returns null for empty AND array', () => {
    expect(getFilterCount('{"AND":[]}')).toBeNull()
  })
  it('returns null for empty OR array', () => {
    expect(getFilterCount('{"OR":[]}')).toBeNull()
  })
})

describe('filterRowsToWhereJson', () => {
  it('returns null for empty rows', () => {
    expect(filterRowsToWhereJson([])).toBeNull()
  })

  it('returns null when all rows have empty value', () => {
    const rows: FilterRow[] = [{ id: '1', field: 'name', operator: 'contains', value: '' }]
    expect(filterRowsToWhereJson(rows)).toBeNull()
  })

  it('returns null when field is not selected', () => {
    const rows: FilterRow[] = [{ id: '1', field: '', operator: 'contains', value: '张' }]
    expect(filterRowsToWhereJson(rows)).toBeNull()
  })

  it('wraps single string condition with AND array', () => {
    const rows: FilterRow[] = [{ id: '1', field: 'name', operator: 'contains', value: '张' }]
    expect(filterRowsToWhereJson(rows)).toEqual({
      AND: [{ name: { contains: '张' } }],
    })
  })

  it('wraps multiple conditions in AND array', () => {
    const rows: FilterRow[] = [
      { id: '1', field: 'name', operator: 'contains', value: '张' },
      { id: '2', field: 'age', operator: 'gt', value: '18' },
    ]
    expect(filterRowsToWhereJson(rows)).toEqual({
      AND: [
        { name: { contains: '张' } },
        { age: { gt: 18 } },
      ],
    })
  })

  it('coerces number string to number for numeric operators', () => {
    const rows: FilterRow[] = [{ id: '1', field: 'score', operator: 'gte', value: '90' }]
    expect(filterRowsToWhereJson(rows)).toEqual({
      AND: [{ score: { gte: 90 } }],
    })
  })

  it('keeps string value for equals on text field', () => {
    const rows: FilterRow[] = [{ id: '1', field: 'status', operator: 'equals', value: 'active', fieldType: 'STRING' }]
    expect(filterRowsToWhereJson(rows)).toEqual({
      AND: [{ status: { equals: 'active' } }],
    })
  })

  it('coerces to number for equals on number field', () => {
    const rows: FilterRow[] = [{ id: '1', field: 'age', operator: 'equals', value: '25', fieldType: 'NUMBER' }]
    expect(filterRowsToWhereJson(rows)).toEqual({
      AND: [{ age: { equals: 25 } }],
    })
  })

  it('skips rows with empty value and includes valid ones', () => {
    const rows: FilterRow[] = [
      { id: '1', field: 'name', operator: 'contains', value: '张' },
      { id: '2', field: 'age', operator: 'gt', value: '' },
    ]
    expect(filterRowsToWhereJson(rows)).toEqual({
      AND: [{ name: { contains: '张' } }],
    })
  })

  it('handles boolean equals: is true', () => {
    const rows: FilterRow[] = [{ id: '1', field: 'active', operator: 'equals', value: 'true', fieldType: 'BOOLEAN' }]
    expect(filterRowsToWhereJson(rows)).toEqual({
      AND: [{ active: { equals: true } }],
    })
  })

  it('handles boolean equals: is false', () => {
    const rows: FilterRow[] = [{ id: '1', field: 'active', operator: 'equals', value: 'false', fieldType: 'BOOLEAN' }]
    expect(filterRowsToWhereJson(rows)).toEqual({
      AND: [{ active: { equals: false } }],
    })
  })
})
