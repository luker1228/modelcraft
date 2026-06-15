'use client'

import {
  X,
  Loader2,
  Edit,
  Key,
  Plus,
  Table2,
  MoreVertical,
  Archive,
  Trash2,
  AlertTriangle,
} from 'lucide-react'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { Alert, AlertDescription } from '@web/components/ui/alert'
import {
  Drawer,
  DrawerContent,
} from '@web/components/ui/drawer'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@web/components/ui/dropdown-menu'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import { InsertFieldSheet } from '@web/components/features/model-editor/InsertFieldSheet'
import {
  getEnumDisplayFieldName,
  isSystemGeneratedLabelField,
} from '@/shared/model/system-field'
import { ForeignKeyPanel } from './ForeignKeyPanel'
import type { ModelEditorState } from '../_hooks'
import type { ModelCRUD } from '../_hooks'
import type { FieldOperations } from '../_hooks'
import type { ForeignKeyOps } from '../_hooks'

interface ModelDetailPanelProps {
  state: ModelEditorState
  crud: ModelCRUD
  fieldOps: FieldOperations
  fkOps: ForeignKeyOps
  orgName: string
  projectSlug: string
  onFieldAdded?: () => void
}

export function ModelDetailPanel({
  state,
  crud,
  fieldOps,
  fkOps,
  orgName,
  projectSlug,
  onFieldAdded,
}: ModelDetailPanelProps) {
  const displayFieldOptions = (state.editModelData?.fields || []).filter((field) => field.format !== 'RELATION')
  const orderedFields = [...(state.editModelData?.fields ?? [])].sort(
    (a, b) => Number(b.isPrimary === true) - Number(a.isPrimary === true)
  )
  const displayFieldSelectValue = state.metaDisplayField || '__display_field_none__'
  const isDisplayFieldUnset = state.metaDisplayField.trim() === ''
  const insertionOrderFieldSelectValue = state.metaInsertionOrderField || '__insertion_order_field_none__'
  const isManagedReadOnlyModel = state.editModelData?.isReadOnly === true

  return (
    <Drawer open={state.editModelOpen} onOpenChange={crud.handleCloseEditModel} direction="right">
      <DrawerContent direction="right" className="flex w-[680px] flex-col rounded-none">
        {/* 插入字段仅限关联关系（RELATION），物理字段由数据库同步 */}
        <InsertFieldSheet
          open={state.insertFieldOpen}
          onOpenChange={state.setInsertFieldOpen}
          modelId={state.editModelId || ''}
          modelName={state.editModelData?.name}
          projectSlug={projectSlug}
          orgName={orgName}
          existingFieldNames={(state.editModelData?.fields || []).map((f) => f.name)}
          onSuccess={() => {
            void crud.refreshModelDetail()
            onFieldAdded?.()
          }}
        />

        {/* Header */}
        <div className="flex shrink-0 items-start justify-between border-b border-border px-6 py-4">
          <div className="min-w-0">
            <h2 className="text-base font-semibold text-foreground">
              {state.editModelData?.title || state.editModelData?.name || '模型详情'}
            </h2>
            <p className="mt-0.5 font-mono text-xs text-muted-foreground">{state.editModelData?.name}</p>
          </div>
          <button
            onClick={crud.handleCloseEditModel}
            className="ml-4 shrink-0 rounded-md p-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
          >
            <X className="size-4" />
          </button>
        </div>

        {/* Scrollable body */}
        <div className="min-h-0 flex-1 overflow-y-auto">
          {state.editModelLoading ? (
            <div className="flex flex-col items-center justify-center py-24">
              <Loader2 className="mb-3 size-5 animate-spin text-muted-foreground" />
              <span className="text-sm text-muted-foreground">加载中...</span>
            </div>
          ) : state.editModelData ? (
            <div className="divide-y divide-border [&>div]:py-6">
              {isManagedReadOnlyModel && (
                <div className="px-6 pb-0 pt-6">
                  <Alert variant="warning" className="py-2">
                    <AlertTriangle className="size-4" />
                    <AlertDescription className="text-xs">
                      当前模型为托管模型，仅支持查看，不支持结构和字段修改。
                    </AlertDescription>
                  </Alert>
                </div>
              )}

              {/* Meta info section */}
              <div className="px-6">
                <div className="mb-3 flex items-center justify-between">
                  <span className="text-sm font-semibold text-foreground">元信息</span>
                  {!state.metaEditMode && (
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      className="h-7 text-xs"
                      onClick={() => state.setMetaEditMode(true)}
                      disabled={isManagedReadOnlyModel}
                    >
                      设置展示字段
                    </Button>
                  )}
                </div>
                {isDisplayFieldUnset && (
                  <Alert variant="warning" className="mb-3 py-2">
                    <AlertTriangle className="size-4" />
                    <AlertDescription className="flex items-center justify-between gap-2 text-xs">
                      <span>未设置展示字段，关系展示将显示空(id)。</span>
                      {!state.metaEditMode && (
                        <Button
                          type="button"
                          size="sm"
                          variant="outline"
                          className="h-6 px-2 text-xs"
                          onClick={() => state.setMetaEditMode(true)}
                          disabled={isManagedReadOnlyModel}
                        >
                          去设置
                        </Button>
                      )}
                    </AlertDescription>
                  </Alert>
                )}
                <div className="grid grid-cols-2 gap-x-6 gap-y-3">
                  <div className="space-y-1">
                    <label className="text-xs text-muted-foreground">标识名称</label>
                    <Input
                      value={state.editModelData.name}
                      disabled
                      className="h-8 bg-muted/30 font-mono text-xs"
                    />
                  </div>
                  <div className="space-y-1">
                    <label className="text-xs text-muted-foreground">显示标题</label>
                    <Input
                      value={state.metaTitle}
                      onChange={(e) => state.setMetaTitle(e.target.value)}
                      className="h-8 text-sm"
                      placeholder="输入显示标题"
                      disabled={!state.metaEditMode || isManagedReadOnlyModel}
                    />
                  </div>
                  <div className="space-y-1">
                    <label className="text-xs text-muted-foreground">数据库</label>
                    <Input
                      value={state.editModelData.databaseName}
                      disabled
                      className="h-8 bg-muted/30 font-mono text-xs"
                    />
                  </div>
                  <div className="space-y-1">
                    <label className="text-xs text-muted-foreground">描述</label>
                    <Input
                      value={state.metaDescription}
                      onChange={(e) => state.setMetaDescription(e.target.value)}
                      className="h-8 text-sm"
                      placeholder="输入模型描述"
                      disabled={!state.metaEditMode || isManagedReadOnlyModel}
                    />
                  </div>
                  <div className="space-y-1">
                    <label className="text-xs text-muted-foreground">展示字段</label>
                    {state.metaEditMode && !isManagedReadOnlyModel ? (
                      <Select
                        value={displayFieldSelectValue}
                        onValueChange={(value) => {
                          state.setMetaDisplayField(value === '__display_field_none__' ? '' : value)
                        }}
                      >
                        <SelectTrigger className="h-8 text-sm">
                          <SelectValue placeholder="选择展示字段" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="__display_field_none__" className="text-sm">
                            未设置
                          </SelectItem>
                          {displayFieldOptions.map((field) => (
                            <SelectItem key={field.name} value={field.name} className="font-mono text-xs">
                              {field.name}
                              {field.title && field.title !== field.name ? ` (${field.title})` : ''}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    ) : (
                      <Input
                        value={state.metaDisplayField || '未设置'}
                        disabled
                        className="h-8 bg-muted/30 text-sm"
                      />
                    )}
                  </div>
                  <div className="space-y-1">
                    <label className="text-xs text-muted-foreground">插入序字段</label>
                    {state.metaEditMode && !isManagedReadOnlyModel ? (
                      <>
                        <Select
                          value={insertionOrderFieldSelectValue}
                          onValueChange={(value) => {
                            state.setMetaInsertionOrderField(value === '__insertion_order_field_none__' ? '' : value)
                          }}
                        >
                          <SelectTrigger className="h-8 text-sm">
                            <SelectValue placeholder="未配置" />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="__insertion_order_field_none__" className="text-sm">
                              未配置
                            </SelectItem>
                            {displayFieldOptions.map((field) => (
                              <SelectItem key={field.name} value={field.name} className="font-mono text-xs">
                                {field.name}
                                {field.title && field.title !== field.name ? ` (${field.title})` : ''}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        {!state.metaInsertionOrderField && (
                          <p className="flex items-center gap-1 text-xs text-amber-600">
                            <span className="shrink-0">⚠</span>
                            未配置，<code className="rounded bg-muted px-1 font-mono">listByCursor</code> 分页稳定性无法保证。
                          </p>
                        )}
                        {state.metaInsertionOrderField && (
                          <p className="text-xs text-muted-foreground">
                            使用 <code className="font-mono">{state.metaInsertionOrderField}</code> 作为 cursor tiebreaker，分页稳定。
                          </p>
                        )}
                      </>
                    ) : (
                      <Input
                        value={state.metaInsertionOrderField || '未配置'}
                        disabled
                        className="h-8 bg-muted/30 text-sm"
                      />
                    )}
                  </div>
                </div>
                {state.metaEditMode && !isManagedReadOnlyModel && (
                  <div className="mt-4 flex items-center justify-end gap-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-7 text-xs"
                      onClick={() => {
                        state.setMetaTitle(state.editModelData!.title || '')
                        state.setMetaDescription(state.editModelData!.description || '')
                        state.setMetaDisplayField(state.editModelData!.displayField || '')
                        state.setMetaInsertionOrderField(state.editModelData!.insertionOrderField || '')
                        state.setMetaEditMode(false)
                      }}
                    >
                      取消
                    </Button>
                    <Button
                      size="sm"
                      className="h-7 px-4 text-xs"
                      disabled={state.metaSaving}
                      onClick={async () => {
                        await crud.handleSaveMeta()
                        state.setMetaEditMode(false)
                      }}
                    >
                      {state.metaSaving && <Loader2 className="mr-1.5 size-3 animate-spin" />}
                      保存更改
                    </Button>
                  </div>
                )}
              </div>

              {/* Fields section */}
              <div className="px-6">
                <div className="mb-3 flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-semibold text-foreground">字段定义</span>
                    <span className="rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground">
                      {state.editModelData.fields?.length || 0}
                    </span>
                  </div>
                  <button
                    type="button"
                    title={isManagedReadOnlyModel ? '托管模型仅支持查看' : '新增关联字段'}
                    className="rounded p-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground disabled:cursor-not-allowed disabled:opacity-50"
                    onClick={() => state.setInsertFieldOpen(true)}
                    disabled={isManagedReadOnlyModel}
                  >
                    <Plus className="size-3.5" />
                  </button>
                </div>

                <div className="overflow-hidden rounded-lg border border-border bg-card shadow-[0_1px_2px_rgba(0,0,0,0.04)]">
                  {state.editModelData.fields && state.editModelData.fields.length > 0 ? (
                    <div className="overflow-x-auto">
                      <table className="w-full text-sm">
                        <thead>
                          <tr className="border-b-2 border-border bg-card">
                            <th className="w-[180px] px-3 py-2 text-left text-[11px] font-medium uppercase tracking-wider text-foreground">
                              标识(名称)
                            </th>
                            <th className="w-[90px] px-3 py-2 text-left text-[11px] font-medium uppercase tracking-wider text-foreground">
                              格式
                            </th>
                            <th className="w-[90px] px-3 py-2 text-left text-[11px] font-medium uppercase tracking-wider text-foreground">
                              类型
                            </th>
                            <th className="w-[80px] px-3 py-2 text-left text-[11px] font-medium uppercase tracking-wider text-foreground">
                              默认值
                            </th>
                            <th className="w-[60px] px-3 py-2 text-center text-[11px] font-medium uppercase tracking-wider text-foreground">
                              主键
                            </th>
                            <th className="w-[50px] px-3 py-2 text-center text-[11px] font-medium uppercase tracking-wider text-foreground">
                              
                            </th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-border">
                          {orderedFields.map((field) => {
                            const enumDisplayFieldName = getEnumDisplayFieldName(field)
                            const isSystemField = isSystemGeneratedLabelField(field, state.editModelData?.fields ?? [])
                            const isFieldReadOnlyActionDisabled = isSystemField || isManagedReadOnlyModel

                            return (
                              <tr
                                key={field.name}
                                className="transition-colors hover:bg-muted/20"
                              >
                              <td className="px-3 py-2">
                                <div className="flex flex-col">
                                  <div className="flex items-center gap-2">
                                    <span className={`font-mono text-sm ${field.isDeprecated ? 'text-muted-foreground line-through' : 'text-foreground'}`}>
                                      {field.name}
                                    </span>
                                    {isSystemField && (
                                      <span className="inline-flex items-center rounded bg-muted px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground">
                                        系统字段
                                      </span>
                                    )}
                                    {field.isDeprecated && (
                                      <span className="bg-warning/10 text-warning inline-flex items-center rounded px-1.5 py-0.5 font-mono text-[11px]">
                                        已废弃
                                      </span>
                                    )}
                                  </div>
                                  {field.title && (
                                    <span className="truncate text-xs text-muted-foreground">
                                      {field.title}
                                    </span>
                                  )}
                                  {enumDisplayFieldName && (
                                    <span className="truncate text-xs text-muted-foreground">
                                      显示字段: <span className="font-mono">{enumDisplayFieldName}</span>
                                    </span>
                                  )}
                                  {isSystemField && (
                                    <span className="truncate text-xs text-muted-foreground">
                                      系统生成 / 只读 / 不可编辑 / 非物理列
                                    </span>
                                  )}
                                </div>
                              </td>
                              <td className="px-3 py-2">
                                <span className="inline-flex items-center rounded bg-primary/[0.08] px-1.5 py-0.5 font-mono text-[11px] text-primary">
                                  {field.format || '-'}
                                </span>
                              </td>
                              <td className="px-3 py-2">
                                <span className="inline-flex items-center rounded bg-muted px-1.5 py-0.5 font-mono text-[11px] text-muted-foreground">
                                  {field.dbColumn?.columnType || field.storageHint || field.schemaType || 'String'}
                                </span>
                              </td>
                              <td className="px-3 py-2">
                                <span className="font-mono text-xs text-muted-foreground">
                                  {field.dbColumn?.defaultValue !== undefined ? String(field.dbColumn.defaultValue) : '-'}
                                </span>
                              </td>
                              <td className="px-3 py-2 text-center">
                                {field.isPrimary ? (
                                    <span className="bg-warning/10 inline-flex size-5 items-center justify-center rounded">
                                        <Key className="text-warning size-3" />
                                      </span>
                                ) : (
                                  <span className="text-muted-foreground/30">-</span>
                                )}
                              </td>
                              <td className="p-2 text-center">
                                <DropdownMenu>
                                  <DropdownMenuTrigger asChild>
                                    <Button
                                      variant="ghost"
                                      size="sm"
                                      className="size-6 p-0 hover:bg-muted"
                                    >
                                      <MoreVertical className="size-3.5 text-muted-foreground" />
                                    </Button>
                                  </DropdownMenuTrigger>
                                  <DropdownMenuContent align="end" className="w-36">
                                    <DropdownMenuItem
                                      className={`text-xs ${isFieldReadOnlyActionDisabled ? 'cursor-not-allowed text-muted-foreground/50' : 'cursor-pointer'}`}
                                      onClick={() => fieldOps.handleOpenEditField(field)}
                                      disabled={isFieldReadOnlyActionDisabled}
                                    >
                                      <Edit className="mr-2 size-3.5" />
                                      编辑
                                    </DropdownMenuItem>
                                    <DropdownMenuItem
                                      className={`text-xs ${isFieldReadOnlyActionDisabled ? 'cursor-not-allowed text-muted-foreground/50' : 'cursor-pointer'}`}
                                      onClick={() => fieldOps.handleToggleDeprecate(field)}
                                      disabled={isFieldReadOnlyActionDisabled}
                                    >
                                      <Archive className="mr-2 size-3.5" />
                                      {field.isDeprecated ? '取消废弃' : '废弃'}
                                    </DropdownMenuItem>
                                    <DropdownMenuItem
                                      className={`cursor-pointer text-xs ${
                                        field.isDeprecated && !isFieldReadOnlyActionDisabled
                                          ? 'text-destructive focus:text-destructive'
                                          : 'cursor-not-allowed text-muted-foreground/50'
                                      }`}
                                      onClick={() => fieldOps.handleRemoveField(field)}
                                      disabled={!field.isDeprecated || isFieldReadOnlyActionDisabled}
                                    >
                                      <Trash2 className="mr-2 size-3.5" />
                                      删除
                                    </DropdownMenuItem>
                                  </DropdownMenuContent>
                                </DropdownMenu>
                              </td>
                              </tr>
                            )
                          })}
                        </tbody>
                      </table>
                    </div>
                  ) : (
                    <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
                      <Table2 className="mb-2 size-8 opacity-30" />
                      <p className="text-sm">暂无字段</p>
                      <p className="mt-1 text-xs">从数据库同步获取物理字段，或点击上方按钮添加关联字段</p>
                    </div>
                  )}
                </div>
              </div>

              {/* Foreign keys section */}
              <ForeignKeyPanel
                state={state}
                fkOps={fkOps}
                relationDatabaseNames={crud.relationDatabaseNames}
                getRelationModelsForDatabase={crud.getRelationModelsForDatabase}
                loadRelationModelsForDatabase={crud.loadRelationModelsForDatabase}
                isRelationModelsLoading={crud.isRelationModelsLoading}
              />
            </div>
          ) : null}
        </div>

        <div className="flex shrink-0 justify-end border-t border-border px-6 py-4">
          <Button
            variant="outline"
            size="sm"
            onClick={crud.handleCloseEditModel}
          >
            关闭
          </Button>
        </div>
      </DrawerContent>
    </Drawer>
  )
}
