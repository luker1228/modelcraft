import { describe, it, expect } from 'vitest'
import { isValidJson, formatJson, getFilterCount } from './filter-utils'

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
})
