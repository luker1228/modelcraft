import { describe, expect, it } from 'vitest'

import { getRecordPageCountText } from './recordPageCount'

describe('getRecordPageCountText', () => {
  it('returns current page count when no search keyword is present', () => {
    expect(
      getRecordPageCountText({
        pageCount: 20,
        filteredCount: 20,
        searchKeyword: '',
      })
    ).toBe('本页 20 条')
  })

  it('returns filtered count against current page count when search keyword is present', () => {
    expect(
      getRecordPageCountText({
        pageCount: 20,
        filteredCount: 8,
        searchKeyword: 'iphone',
      })
    ).toBe('页内搜索 8 / 本页 20')
  })
})
