import { describe, expect, it } from 'vitest'
import {
  buildListByPageOrderBy,
  getDefaultListByPageOrderBy,
} from './useRuntimeListByPage'

describe('getDefaultListByPageOrderBy', () => {
  it('uses id when the model fields include id', () => {
    expect(getDefaultListByPageOrderBy(['created_at', 'id', 'name'])).toEqual([
      { id: 'desc' },
    ])
  })

  it('uses the first field when id is unavailable', () => {
    expect(getDefaultListByPageOrderBy(['email', 'username'])).toEqual([
      { email: 'asc' },
    ])
  })

  it('returns undefined when no fields are available', () => {
    expect(getDefaultListByPageOrderBy([])).toBeUndefined()
  })
})

describe('buildListByPageOrderBy', () => {
  it('appends id desc when stable sorting is enabled', () => {
    expect(
      buildListByPageOrderBy({ field: 'created_at', direction: 'desc' }, true, [
        'created_at',
        'id',
      ])
    ).toEqual([{ created_at: 'desc' }, { id: 'desc' }])
  })

  it('does not duplicate id when id is the primary sort field', () => {
    expect(
      buildListByPageOrderBy({ field: 'id', direction: 'desc' }, true, ['created_at', 'id'])
    ).toEqual([{ id: 'desc' }])
  })

  it('does not append id when stable sorting is disabled', () => {
    expect(
      buildListByPageOrderBy({ field: 'created_at', direction: 'desc' }, false, [
        'created_at',
        'id',
      ])
    ).toEqual([{ created_at: 'desc' }])
  })
})
