'use client'

import React, { useCallback, useEffect, useRef, useState } from 'react'
import { Button } from '@web/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@web/components/ui/table'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@web/components/ui/tooltip'
import { Check, Copy, Edit, Key, Loader2, Plus, Trash2 } from 'lucide-react'
import { renderCellValue } from './fieldProtocol'

export interface ModelRecordTableFieldInfo {
  name: string
  title?: string | null
  isPrimary?: boolean
  storageHint?: string | null
  schemaType?: string | null
}

export type ModelRecordTableRow = Record<string, unknown> & {
  id?: string | number | null
}

interface ModelRecordTableProps {
  contentLoading: boolean
  contentList: ModelRecordTableRow[]
  displayFields: string[]
  getFieldInfo: (fieldName: string) => ModelRecordTableFieldInfo | null
  getFieldTypeDisplay: (fieldInfo: ModelRecordTableFieldInfo | null) => string
  propByName: Record<string, Readonly<Record<string, unknown>>>
  onCreate: () => void
  onEdit: (id: string) => void
  onDelete: (id: string) => void
}

export function ModelRecordTable({
  contentLoading,
  contentList,
  displayFields,
  getFieldInfo,
  getFieldTypeDisplay,
  propByName,
  onCreate,
  onEdit,
  onDelete,
}: ModelRecordTableProps) {
  const [copiedCell, setCopiedCell] = useState<string | null>(null)

  const [columnWidths, setColumnWidths] = useState<Record<string, number>>({})
  const [resizingColumn, setResizingColumn] = useState<string | null>(null)
  const resizeStartX = useRef<number>(0)
  const resizeStartWidth = useRef<number>(0)
  const tableRef = useRef<HTMLTableElement>(null)

  const DEFAULT_COLUMN_WIDTH = 150
  const MIN_COLUMN_WIDTH = 60
  const INDEX_COLUMN_WIDTH = 50
  const ACTION_COLUMN_WIDTH = 100

  const getColumnWidth = useCallback(
    (field: string) => {
      return columnWidths[field] || DEFAULT_COLUMN_WIDTH
    },
    [columnWidths]
  )

  const handleResizeStart = useCallback(
    (e: React.MouseEvent, field: string) => {
      e.preventDefault()
      e.stopPropagation()
      setResizingColumn(field)
      resizeStartX.current = e.clientX
      resizeStartWidth.current = columnWidths[field] || DEFAULT_COLUMN_WIDTH
    },
    [columnWidths]
  )

  const handleResizeMove = useCallback(
    (e: MouseEvent) => {
      if (!resizingColumn) return

      const delta = e.clientX - resizeStartX.current
      const newWidth = Math.max(MIN_COLUMN_WIDTH, resizeStartWidth.current + delta)

      setColumnWidths((prev) => ({
        ...prev,
        [resizingColumn]: newWidth,
      }))
    },
    [resizingColumn]
  )

  const handleResizeEnd = useCallback(() => {
    setResizingColumn(null)
  }, [])

  useEffect(() => {
    if (resizingColumn) {
      document.addEventListener('mousemove', handleResizeMove)
      document.addEventListener('mouseup', handleResizeEnd)
      document.body.style.cursor = 'col-resize'
      document.body.style.userSelect = 'none'
    }

    return () => {
      document.removeEventListener('mousemove', handleResizeMove)
      document.removeEventListener('mouseup', handleResizeEnd)
      document.body.style.cursor = ''
      document.body.style.userSelect = ''
    }
  }, [resizingColumn, handleResizeMove, handleResizeEnd])

  const visibleFields = displayFields

  return (
    <TooltipProvider>
      <div className="flex flex-1 flex-col overflow-hidden bg-sidebar">
        <div className="flex-1 overflow-auto">
          <Table ref={tableRef} className="table-fixed">
            <TableHeader className="sticky top-0 z-10 bg-sidebar">
              <TableRow className="border-b border-border hover:bg-transparent">
                <TableHead
                  className="bg-sidebar py-2.5 text-xs font-semibold text-muted-foreground"
                  style={{ width: INDEX_COLUMN_WIDTH }}
                >
                  #
                </TableHead>
                {visibleFields.map((field) => {
                  const fieldInfo = getFieldInfo(field)
                  const typeDisplay = getFieldTypeDisplay(fieldInfo)
                  const isPrimary = fieldInfo?.isPrimary
                  const fieldTitle = (fieldInfo?.title ?? field).trim() || field
                  const headerLabel = `${fieldTitle} (${field})`

                  return (
                    <TableHead
                      key={field}
                      className="group relative bg-sidebar py-2.5 text-xs font-semibold text-foreground"
                      style={{ width: getColumnWidth(field) }}
                    >
                      <div className="flex min-w-0 items-center gap-1.5 pr-3">
                        {isPrimary && <Key className="size-3 flex-shrink-0 text-blue-500" />}
                        <span
                          className="truncate text-xs font-semibold text-foreground"
                          title={headerLabel}
                        >
                          {headerLabel}
                        </span>
                        {typeDisplay && (
                          <>
                            <span className="flex-shrink-0 text-[10px] text-muted-foreground/40">
                              ·
                            </span>
                            <span className="flex-shrink-0 font-mono text-[10px] font-normal text-muted-foreground">
                              {typeDisplay}
                            </span>
                          </>
                        )}
                      </div>
                      <div
                        className="absolute right-0 top-0 h-full w-1 cursor-col-resize transition-colors hover:bg-blue-500/40 group-hover:bg-blue-500/20"
                        onMouseDown={(e) => handleResizeStart(e, field)}
                        style={{
                          backgroundColor:
                            resizingColumn === field ? 'hsl(var(--primary))' : undefined,
                        }}
                      />
                    </TableHead>
                  )
                })}
                <TableHead
                  className="bg-sidebar py-2.5 text-right text-xs font-semibold text-muted-foreground"
                  style={{ width: ACTION_COLUMN_WIDTH }}
                >
                  操作
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {contentLoading ? (
                <TableRow>
                  <TableCell colSpan={visibleFields.length + 2} className="h-48 text-center">
                    <div className="flex items-center justify-center">
                      <Loader2 className="size-6 animate-spin text-muted-foreground" />
                    </div>
                  </TableCell>
                </TableRow>
              ) : contentList.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={visibleFields.length + 2} className="h-48 text-center">
                    <div className="flex flex-col items-center justify-center">
                      <p className="mb-3 text-sm text-muted-foreground">暂无数据</p>
                      <Button
                        size="sm"
                        className="border-0 bg-[#2563eb] text-white transition-colors duration-200 hover:bg-[#1d4ed8]"
                        onClick={onCreate}
                      >
                        <Plus className="mr-1.5 size-3.5" />
                        添加第一条数据
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ) : (
                contentList.map((item, index) => {
                  const rowId = String(item.id)

                  return (
                    <TableRow
                      key={rowId}
                      className="border-b border-border/50 bg-sidebar transition-colors hover:bg-selected"
                    >
                      <TableCell
                        className="py-2 text-xs tabular-nums text-muted-foreground"
                        style={{ width: INDEX_COLUMN_WIDTH }}
                      >
                        {index + 1}
                      </TableCell>
                      {visibleFields.map((field) => {
                        const rawValue = item[field]

                        if (rawValue === null || rawValue === undefined) {
                          return (
                            <TableCell
                              key={field}
                              className="py-2"
                              style={{ width: getColumnWidth(field) }}
                            >
                              <span className="font-mono text-xs text-muted-foreground/50">NULL</span>
                            </TableCell>
                          )
                        }

                        const renderedValue = renderCellValue(rawValue, propByName[field] ?? {})
                        const cellKey = `${rowId}-${field}`

                        return (
                          <TableCell
                            key={field}
                            className="py-2"
                            style={{ width: getColumnWidth(field) }}
                          >
                            <Tooltip delayDuration={300}>
                              <TooltipTrigger asChild>
                                <span
                                  className="block truncate text-sm font-normal text-foreground"
                                  style={{ maxWidth: getColumnWidth(field) - 16 }}
                                >
                                  {renderedValue}
                                </span>
                              </TooltipTrigger>
                              <TooltipContent
                                side="bottom"
                                align="start"
                                className="flex max-w-xs items-center gap-2 break-all font-mono text-xs"
                              >
                                <span>{renderedValue}</span>
                                <button
                                  className="ml-1 shrink-0 opacity-70 hover:opacity-100"
                                  onClick={async (e) => {
                                    e.stopPropagation()
                                    try {
                                      await navigator.clipboard.writeText(renderedValue)
                                    } catch {
                                      // ignore clipboard errors
                                    }
                                    setCopiedCell(cellKey)
                                    setTimeout(() => setCopiedCell(null), 1500)
                                  }}
                                >
                                  {copiedCell === cellKey ? (
                                    <Check className="size-3" />
                                  ) : (
                                    <Copy className="size-3" />
                                  )}
                                </button>
                              </TooltipContent>
                            </Tooltip>
                          </TableCell>
                        )
                      })}
                      <TableCell className="py-2 text-right" style={{ width: ACTION_COLUMN_WIDTH }}>
                        <div className="flex items-center justify-end gap-1">
                          <Button
                            variant="ghost"
                            size="sm"
                            className="size-7 p-0 text-muted-foreground transition-colors hover:bg-selected hover:text-foreground"
                            onClick={() => onEdit(rowId)}
                            title="编辑"
                          >
                            <Edit className="size-3.5" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="size-7 p-0 text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                            onClick={() => onDelete(rowId)}
                            title="删除"
                          >
                            <Trash2 className="size-3.5" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  )
                })
              )}
            </TableBody>
          </Table>
        </div>
      </div>
    </TooltipProvider>
  )
}
