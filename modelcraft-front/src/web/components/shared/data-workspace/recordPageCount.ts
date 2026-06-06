interface GetRecordPageCountTextOptions {
  pageCount: number
  filteredCount: number
  searchKeyword: string
}

export function getRecordPageCountText({
  pageCount,
  filteredCount,
  searchKeyword,
}: GetRecordPageCountTextOptions): string {
  if (!searchKeyword.trim()) {
    return `本页 ${pageCount} 条`
  }

  return `页内搜索 ${filteredCount} / 本页 ${pageCount}`
}
