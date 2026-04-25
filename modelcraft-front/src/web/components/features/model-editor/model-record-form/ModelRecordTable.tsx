'use client'

import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
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
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@web/components/ui/dropdown-menu'
import { Archive, Check, Copy, Edit, Key, Link2, Loader2, Plus, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { renderCellValue } from './runtime/field-protocol'
import { getXMC } from '@/types/xmc'

export interface ModelRecordTableFieldInfo {
  name: string
  title?: string | null
  isPrimary?: boolean
  isDeprecated?: boolean
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
  onManageRelations?: (id: string) => void
  onToggleFieldDeprecated?: (fieldInfo: ModelRecordTableFieldInfo) => void
  onDeleteField?: (fieldInfo: ModelRecordTableFieldInfo) => void
  canManageFieldLifecycle?: boolean
}

type PairRole = 'label' | 'code'

interface PairMeta {
  pairBase: string
  role: PairRole
  pairedField: string
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
  onManageRelations,
  onToggleFieldDeprecated,
  onDeleteField,
  canManageFieldLifecycle = true,
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
  const ACTION_COLUMN_WIDTH = 132

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

  const { visibleFields, pairMetaByField } = useMemo(() => {
    const ordered: string[] = []
    const seen = new Set<string>()
    const fieldSet = new Set(displayFields)
    const pairMeta: Record<string, PairMeta> = {}
    const configuredLabelFieldByCodeField = new Map<string, string>()

    Object.entries(propByName).forEach(([codeField, prop]) => {
      const xmc = getXMC(prop as Record<string, unknown>)
      const enumLabelFieldName = xmc?.enum?.labelFieldName?.trim()
      if (!enumLabelFieldName || enumLabelFieldName === codeField) return
      if (!fieldSet.has(codeField) || !fieldSet.has(enumLabelFieldName)) return
      configuredLabelFieldByCodeField.set(codeField, enumLabelFieldName)
    })

    displayFields.forEach((field) => {
      if (seen.has(field)) return

      const configuredLabelField = configuredLabelFieldByCodeField.get(field)
      if (configuredLabelField) {
        ordered.push(configuredLabelField, field)
        seen.add(configuredLabelField)
        seen.add(field)
        pairMeta[configuredLabelField] = { pairBase: field, role: 'label', pairedField: field }
        pairMeta[field] = { pairBase: field, role: 'code', pairedField: configuredLabelField }
        return
      }

      const configuredCodeField = Array.from(configuredLabelFieldByCodeField.entries()).find(
        ([, labelField]) => labelField === field
      )?.[0]
      if (configuredCodeField && fieldSet.has(configuredCodeField)) {
        ordered.push(field, configuredCodeField)
        seen.add(field)
        seen.add(configuredCodeField)
        pairMeta[field] = { pairBase: configuredCodeField, role: 'label', pairedField: configuredCodeField }
        pairMeta[configuredCodeField] = {
          pairBase: configuredCodeField,
          role: 'code',
          pairedField: field,
        }
        return
      }

      ordered.push(field)
      seen.add(field)
    })

    return { visibleFields: ordered, pairMetaByField: pairMeta }
  }, [displayFields, propByName])

  const copyText = useCallback(async (text: string) => {
    if (navigator.clipboard && window.isSecureContext) {
      try {
        await navigator.clipboard.writeText(text)
        return true
      } catch {
        // fallback to execCommand
      }
    }

    const textarea = document.createElement('textarea')
    textarea.value = text
    textarea.style.cssText = 'position:fixed;top:0;left:0;opacity:0'
    document.body.appendChild(textarea)
    textarea.focus()
    textarea.select()
    const copied = document.execCommand('copy')
    document.body.removeChild(textarea)
    return copied
  }, [])

  return (
    <TooltipProvider>
      <div className="flex flex-1 flex-col overflow-hidden bg-card">
        <div className="flex-1 overflow-auto">
          <Table ref={tableRef} className="table-fixed">
            <TableHeader className="sticky top-0 z-10 bg-card">
              <TableRow className="border-b border-border hover:bg-transparent">
                <TableHead
                  className="bg-card py-2.5 text-xs font-medium uppercase tracking-wider text-muted-foreground"
                  style={{ width: INDEX_COLUMN_WIDTH }}
                >
                  #
                </TableHead>
                {visibleFields.map((field) => {
                  const fieldInfo = getFieldInfo(field)
                  const pairMeta = pairMetaByField[field]
                  const pairBaseInfo = pairMeta ? getFieldInfo(pairMeta.pairBase) : null
                  const typeDisplay = getFieldTypeDisplay(fieldInfo)
                  const isPrimary = fieldInfo?.isPrimary
                  const isDeprecated = fieldInfo?.isDeprecated === true
                  const fieldTitle =
                    pairMeta?.role === 'label'
                      ? `${((pairBaseInfo?.title ?? pairMeta.pairBase) || pairMeta.pairBase).trim()} Label`
                      : (fieldInfo?.title ?? field).trim() || field
                  const headerLabel = `${fieldTitle} (${field})`
                  const schemaProp = propByName[field] as { enum?: unknown[] } | undefined
                  const enumOptions = Array.isArray(schemaProp?.enum) ? schemaProp.enum : []
                  const isEnumField = enumOptions.length > 0
                  const showFieldName = fieldTitle !== field
                  const isPairedLabel = pairMeta?.role === 'label'
                  const isPairedCode = pairMeta?.role === 'code'
                  const isDerivedLabelField = isPairedLabel
                  const headerFieldInfo: ModelRecordTableFieldInfo = {
                    name: field,
                    title: fieldInfo?.title ?? null,
                    isPrimary: fieldInfo?.isPrimary,
                    isDeprecated: fieldInfo?.isDeprecated,
                    storageHint: fieldInfo?.storageHint ?? null,
                    schemaType: fieldInfo?.schemaType ?? null,
                  }

                  const triggerContent = (
                    <>
                      <span className="flex min-w-0 items-center gap-1.5">
                        {isPrimary && <Key className="size-3 flex-shrink-0 text-[#D97706]" />}
                        <span
                          className={`truncate text-xs font-semibold ${
                            isDeprecated ? 'text-muted-foreground line-through' : 'text-foreground'
                          }`}
                        >
                          {fieldTitle}
                        </span>
                        {isEnumField && (
                          <span className="inline-flex flex-shrink-0 items-center rounded border border-primary/20 bg-primary/[0.08] px-1.5 py-0 text-[9px] font-medium uppercase leading-4 text-primary">
                            Enum
                          </span>
                        )}
                        {isPairedLabel && (
                          <span className="inline-flex flex-shrink-0 items-center rounded border border-[rgba(5,150,105,0.2)] bg-[rgba(5,150,105,0.08)] px-1.5 py-0 text-[9px] font-medium leading-4 text-[#059669]">
                            Label
                          </span>
                        )}
                        {isPairedCode && (
                          <span className="inline-flex flex-shrink-0 items-center rounded border border-border/80 bg-muted/60 px-1.5 py-0 text-[9px] font-semibold leading-4 text-muted-foreground">
                            Code
                          </span>
                        )}
                        {isDeprecated && (
                          <span className="inline-flex flex-shrink-0 items-center rounded border border-[rgba(217,119,6,0.2)] bg-[rgba(217,119,6,0.08)] px-1.5 py-0 text-[9px] font-medium leading-4 text-[#D97706]">
                            已废弃
                          </span>
                        )}
                      </span>
                      <span className="flex min-w-0 items-center gap-1.5 text-[10px] leading-4 text-muted-foreground">
                        {showFieldName && (
                          <span className="truncate font-mono font-normal text-muted-foreground/90">
                            {field}
                          </span>
                        )}
                        {isEnumField && (
                          <span className="flex-shrink-0 rounded border border-border/70 bg-muted/50 px-1.5 py-0 text-[9px] font-medium text-muted-foreground">
                            {enumOptions.length} 项
                          </span>
                        )}
                        {typeDisplay && (
                          <span className="flex-shrink-0 font-mono font-normal text-muted-foreground/90">
                            {typeDisplay}
                          </span>
                        )}
                      </span>
                    </>
                  )

                  return (
                    <TableHead
                      key={field}
                      className={`group relative py-2.5 text-xs font-semibold text-foreground ${
                        isPairedLabel || isPairedCode ? 'bg-primary/[0.04]' : 'bg-card'
                      }`}
                      style={{ width: getColumnWidth(field) }}
                    >
                      {canManageFieldLifecycle ? (
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <button
                              type="button"
                              className="flex min-w-0 flex-col items-start gap-1 pr-3 text-left"
                              title={`${headerLabel}（点击管理字段）`}
                            >
                              {triggerContent}
                            </button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="start" className="w-36">
                            <DropdownMenuItem
                              className={`text-xs ${
                                isDerivedLabelField ? 'cursor-not-allowed text-muted-foreground/50' : 'cursor-pointer'
                              }`}
                              onClick={() => onToggleFieldDeprecated?.(headerFieldInfo)}
                              disabled={isDerivedLabelField}
                            >
                              <Archive className="mr-2 size-3.5" />
                              {isDeprecated ? '取消废弃' : '废弃字段'}
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              className={`cursor-pointer text-xs ${
                                !isDerivedLabelField && isDeprecated
                                  ? 'text-destructive focus:text-destructive'
                                  : 'cursor-not-allowed text-muted-foreground/50'
                              }`}
                              onClick={() => onDeleteField?.(headerFieldInfo)}
                              disabled={!isDeprecated || isDerivedLabelField}
                            >
                              <Trash2 className="mr-2 size-3.5" />
                              删除字段
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      ) : (
                        <div className="flex min-w-0 flex-col items-start gap-1 pr-3 text-left" title={headerLabel}>
                          {triggerContent}
                        </div>
                      )}
                      <div
                        className="absolute right-0 top-0 h-full w-1 cursor-col-resize transition-colors hover:bg-primary/40 group-hover:bg-primary/20"
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
                  className="bg-card py-2.5 text-right text-xs font-medium uppercase tracking-wider text-muted-foreground"
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
                        className="border-0 bg-primary text-white transition-colors duration-150 hover:bg-primary/90"
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
                      className="border-b border-border/60 bg-card transition-colors hover:bg-foreground/[0.015]"
                    >
                      <TableCell
                        className="py-2 text-xs tabular-nums text-muted-foreground"
                        style={{ width: INDEX_COLUMN_WIDTH }}
                      >
                        {index + 1}
                      </TableCell>
                      {visibleFields.map((field) => {
                        const pairMeta = pairMetaByField[field]
                        const isPairedLabel = pairMeta?.role === 'label'
                        const isPairedCode = pairMeta?.role === 'code'
                        const rawValue = item[field]
                        const pairedRawValue = pairMeta ? item[pairMeta.pairedField] : undefined

                        if (rawValue === null || rawValue === undefined) {
                          return (
                            <TableCell
                              key={field}
                              className={`py-2 ${isPairedLabel || isPairedCode ? 'bg-primary/[0.03]' : ''}`}
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
                            className={`py-2 ${isPairedLabel || isPairedCode ? 'bg-primary/[0.03]' : ''}`}
                            style={{ width: getColumnWidth(field) }}
                          >
                            <Tooltip delayDuration={300}>
                              <TooltipTrigger asChild>
                                <div className="space-y-0.5" style={{ maxWidth: getColumnWidth(field) - 16 }}>
                                  <span
                                    className={`block truncate text-sm ${
                                      isPairedCode
                                        ? 'font-mono font-normal text-muted-foreground'
                                        : 'font-normal text-foreground'
                                    }`}
                                  >
                                    {renderedValue}
                                  </span>
                                  {isPairedCode && pairedRawValue !== undefined && pairedRawValue !== null && (
                                    <span className="block truncate text-[10px] text-muted-foreground/80">
                                      {renderCellValue(pairedRawValue, propByName[pairMeta.pairedField] ?? {})}
                                    </span>
                                  )}
                                  {isPairedLabel && pairedRawValue !== undefined && pairedRawValue !== null && (
                                    <span className="block truncate font-mono text-[10px] text-muted-foreground/80">
                                      {renderCellValue(pairedRawValue, propByName[pairMeta.pairedField] ?? {})}
                                    </span>
                                  )}
                                </div>
                              </TooltipTrigger>
                              <TooltipContent
                                side="bottom"
                                align="start"
                                className="flex max-w-xs items-center gap-2 break-all font-mono text-xs"
                              >
                                <span>{renderedValue}</span>
                                <button
                                  type="button"
                                  className="ml-1 shrink-0 opacity-70 hover:opacity-100"
                                  onClick={async (e) => {
                                    e.stopPropagation()
                                    const copied = await copyText(renderedValue)
                                    if (!copied) {
                                      toast.error('复制失败，请手动复制')
                                      return
                                    }

                                    setCopiedCell(cellKey)
                                    setTimeout(() => {
                                      setCopiedCell((current) => (current === cellKey ? null : current))
                                    }, 1500)
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
                            className="size-7 p-0 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                            onClick={() => onManageRelations?.(rowId)}
                            title="关联管理"
                          >
                            <Link2 className="size-3.5" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="size-7 p-0 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
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
