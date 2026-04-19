'use client'

import { useEffect, useMemo, useState } from 'react'
import { Plus, Loader2, Trash2, Link2 } from 'lucide-react'
import { Button } from '@web/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import type { ModelEditorState, EditorModel, EditorModelField } from '../_hooks'
import type { ForeignKeyOps } from '../_hooks'

interface ForeignKeyPanelProps {
  state: ModelEditorState
  fkOps: ForeignKeyOps
  relationDatabaseNames: string[]
  getRelationModelsForDatabase: (databaseName: string) => EditorModel[]
  loadRelationModelsForDatabase: (databaseName: string) => Promise<void>
  isRelationModelsLoading: (databaseName: string) => boolean
}

export function ForeignKeyPanel({
  state,
  fkOps,
  relationDatabaseNames,
  getRelationModelsForDatabase,
  loadRelationModelsForDatabase,
  isRelationModelsLoading,
}: ForeignKeyPanelProps) {
  const currentDatabase = state.editModelData?.databaseName ?? ''
  const [relationDatabaseName, setRelationDatabaseName] = useState('')

  useEffect(() => {
    if (state.fkFormOpen && !relationDatabaseName && currentDatabase) {
      setRelationDatabaseName(currentDatabase)
    }
  }, [state.fkFormOpen, relationDatabaseName, currentDatabase])

  useEffect(() => {
    if (!state.fkFormOpen || !relationDatabaseName) return
    void loadRelationModelsForDatabase(relationDatabaseName)
  }, [state.fkFormOpen, relationDatabaseName, loadRelationModelsForDatabase])

  const relationModels = useMemo(
    () => getRelationModelsForDatabase(relationDatabaseName),
    [getRelationModelsForDatabase, relationDatabaseName]
  )
  const relationModelsLoading = relationDatabaseName
    ? isRelationModelsLoading(relationDatabaseName)
    : false

  const selectableModels = relationModels
    .filter((m) => m.id !== state.editModelId)
    .sort((a, b) => {
      const dbOrder = a.databaseName.localeCompare(b.databaseName)
      if (dbOrder !== 0) return dbOrder
      return a.name.localeCompare(b.name)
    })

  return (
    <div className="px-6">
      <div className="mb-3 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-sm font-semibold text-foreground">关联关系</span>
          <span className="inline-flex items-center rounded-full bg-muted px-1.5 py-0.5 text-xs font-medium text-muted-foreground">
            {state.fkList.length}
          </span>
        </div>
        {!state.fkFormOpen && (
          <button
            type="button"
            title="添加关系"
            className="rounded p-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
            onClick={() => {
              state.setFkFormOpen(true)
              setRelationDatabaseName(currentDatabase)
              state.setFkRefModelId('')
              state.setFkMappings([{ sourceField: '', targetField: '' }])
            }}
          >
            <Plus className="size-3.5" />
          </button>
        )}
      </div>

      {/* FK list table */}
      {state.fkLoading ? (
        <div className="flex items-center gap-2 py-4 text-sm text-muted-foreground">
          <Loader2 className="size-4 animate-spin" />
          加载中...
        </div>
      ) : state.fkList.length > 0 ? (
        <div className="overflow-hidden rounded-lg border border-border bg-card">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border bg-muted/30">
                  <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground">
                    关系
                  </th>
                  <th className="w-[80px] px-3 py-2 text-center text-xs font-medium text-muted-foreground">
                    方向
                  </th>
                  <th className="w-[50px] px-3 py-2 text-center text-xs font-medium text-muted-foreground" />
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {state.fkList.map((fk) => {
                  const srcName = fk.modelName
                  const refName = fk.refModelName
                  const pairs = fk.sourceFields.map((sf, i) => {
                    const tf = fk.targetFields[i] ?? '?'
                    return `${srcName}.${sf} → ${refName}.${tf}`
                  })
                  const isNormal = fk.direction === 'NORMAL'
                  return (
                    <tr key={fk.id} className="transition-colors hover:bg-muted/20">
                      <td className="px-3 py-2">
                        <div className="flex flex-col gap-0.5">
                          {pairs.map((p, i) => (
                            <span key={i} className="font-mono text-xs text-foreground">
                              {p}
                            </span>
                          ))}
                        </div>
                      </td>
                      <td className="px-3 py-2 text-center">
                        <span className={`inline-flex items-center rounded px-1.5 py-0.5 text-xs font-medium ${isNormal ? 'bg-blue-500/10 text-blue-600 dark:text-blue-400' : 'bg-orange-500/10 text-orange-600 dark:text-orange-400'}`}>
                          {isNormal ? '多→一' : '一→多'}
                        </span>
                      </td>
                      <td className="p-2 text-center">
                        {state.fkDeleteConfirm === fk.pairId ? (
                          <div className="flex items-center gap-1">
                            <button
                              className="rounded px-1.5 py-0.5 text-xs text-destructive hover:bg-destructive/10"
                              onClick={() => fkOps.handleDeleteFK(fk.pairId)}
                            >
                              确认
                            </button>
                            <button
                              className="rounded px-1.5 py-0.5 text-xs text-muted-foreground hover:bg-muted"
                              onClick={() => state.setFkDeleteConfirm(null)}
                            >
                              取消
                            </button>
                          </div>
                        ) : (
                          <Button
                            variant="ghost"
                            size="sm"
                            className="size-6 p-0 hover:bg-muted hover:text-destructive"
                            onClick={() => state.setFkDeleteConfirm(fk.pairId)}
                          >
                            <Trash2 className="size-3.5 text-muted-foreground" />
                          </Button>
                        )}
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </div>
      ) : !state.fkFormOpen ? (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed border-border py-6 text-muted-foreground">
          <Link2 className="mb-2 size-6 opacity-30" />
          <p className="text-sm">暂无逻辑外键</p>
          <p className="mt-1 text-xs">点击&ldquo;添加关系&rdquo;创建逻辑外键</p>
        </div>
      ) : null}

      {/* Inline create form */}
      {state.fkFormOpen && (
        <div className="space-y-3 rounded-lg border border-border bg-muted/10 p-4">
          <h4 className="text-xs font-semibold text-foreground">新建逻辑外键</h4>

          {/* Ref model select */}
          <div className="flex items-center gap-2">
            <label className="w-20 shrink-0 text-xs text-muted-foreground">引用库</label>
            <Select
              value={relationDatabaseName}
              onValueChange={(v) => {
                setRelationDatabaseName(v)
                state.setFkRefModelId('')
                state.setFkMappings([{ sourceField: '', targetField: '' }])
              }}
            >
              <SelectTrigger className="h-7 text-xs">
                <SelectValue placeholder="选择数据库" />
              </SelectTrigger>
              <SelectContent>
                {relationDatabaseNames.map((dbName) => (
                  <SelectItem key={dbName} value={dbName} className="font-mono text-xs">
                    {dbName}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="flex items-center gap-2">
            <label className="w-20 shrink-0 text-xs text-muted-foreground">引用模型</label>
            <Select
              value={state.fkRefModelId}
              disabled={!relationDatabaseName || relationModelsLoading}
              onValueChange={(v) => {
                state.setFkRefModelId(v)
                state.setFkMappings([{ sourceField: '', targetField: '' }])
              }}
            >
              <SelectTrigger className="h-7 text-xs">
                <SelectValue
                  placeholder={
                    !relationDatabaseName
                      ? '先选择数据库'
                      : relationModelsLoading
                        ? '加载模型中...'
                        : '选择引用模型'
                  }
                />
              </SelectTrigger>
              <SelectContent>
                {relationModelsLoading && (
                  <div className="px-2 py-1.5 text-xs text-muted-foreground">加载中...</div>
                )}
                {selectableModels.map(m => (
                    <SelectItem key={m.id} value={m.id} className="font-mono text-xs">
                      {m.databaseName !== currentDatabase ? `${m.databaseName}.` : ''}
                      {m.name}
                      {m.title && m.title !== m.name ? ` (${m.title})` : ''}
                    </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Field mappings */}
          <div className="space-y-2">
            <label className="text-xs text-muted-foreground">字段映射（源字段 → 目标字段）</label>
            {state.fkMappings.map((mapping, idx) => (
              <div key={idx} className="flex items-center gap-2">
                <Select
                  value={mapping.sourceField}
                  onValueChange={(v) => {
                    const next = [...state.fkMappings]
                    next[idx] = { ...next[idx], sourceField: v }
                    state.setFkMappings(next)
                  }}
                >
                  <SelectTrigger className="h-7 text-xs">
                    <SelectValue placeholder="源字段" />
                  </SelectTrigger>
                  <SelectContent>
                    {state.editModelData?.fields?.map((f: EditorModelField) => (
                      <SelectItem key={f.name} value={f.name} className="font-mono text-xs">
                        {f.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>

                <span className="shrink-0 text-xs text-muted-foreground">→</span>

                <Select
                  value={mapping.targetField}
                  disabled={!state.fkRefModelId || state.fkRefModelLoading}
                  onValueChange={(v) => {
                    const next = [...state.fkMappings]
                    next[idx] = { ...next[idx], targetField: v }
                    state.setFkMappings(next)
                  }}
                >
                  <SelectTrigger className="h-7 text-xs">
                    <SelectValue placeholder={state.fkRefModelLoading ? '加载中...' : '目标字段'} />
                  </SelectTrigger>
                  <SelectContent>
                    {state.fkRefModelDetail?.fields?.map((f: EditorModelField) => (
                      <SelectItem key={f.name} value={f.name} className="font-mono text-xs">
                        {f.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>

                {state.fkMappings.length > 1 && (
                  <Button
                    variant="ghost"
                    size="sm"
                    className="size-6 shrink-0 p-0 hover:text-destructive"
                    onClick={() => state.setFkMappings(state.fkMappings.filter((_, i) => i !== idx))}
                  >
                    <Trash2 className="size-3" />
                  </Button>
                )}
              </div>
            ))}

            <Button
              variant="ghost"
              size="sm"
              className="h-6 text-xs text-muted-foreground hover:text-foreground"
              onClick={() => state.setFkMappings([...state.fkMappings, { sourceField: '', targetField: '' }])}
            >
              <Plus className="mr-1 size-3" />
              添加映射
            </Button>
          </div>

          {/* Action buttons */}
          <div className="flex justify-end gap-2 pt-1">
            <Button
              variant="outline"
              size="sm"
              className="h-7 text-xs"
              onClick={() => {
                state.setFkFormOpen(false)
                setRelationDatabaseName('')
                state.setFkRefModelId('')
                state.setFkMappings([{ sourceField: '', targetField: '' }])
              }}
            >
              取消
            </Button>
            <Button
              size="sm"
              className="h-7 text-xs"
              disabled={
                state.fkSubmitting ||
                !state.fkRefModelId ||
                state.fkMappings.every(m => !m.sourceField || !m.targetField)
              }
              onClick={fkOps.handleCreateFK}
            >
              {state.fkSubmitting && <Loader2 className="mr-1 size-3 animate-spin" />}
              创建外键
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
