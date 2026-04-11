'use client'

import { useState } from 'react'
import { useParams } from 'next/navigation'
import { useMutation, useQuery } from '@apollo/client'
import type { Reference, FieldPolicy } from '@apollo/client'
import { gql } from '@apollo/client'
import { Input } from '@web/components/ui/input'
import { Button } from '@web/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@web/components/ui/dialog'
import { Plus, Search, Edit, Trash2, RefreshCw, Eye } from 'lucide-react'
import { GET_ENUMS } from '@web/graphql/queries/enum'
import { DELETE_ENUM, CREATE_ENUM, UPDATE_ENUM } from '@web/graphql/mutations/enum'
import { useProjectScopedClient } from '@bff/apollo/public'

interface EnumDefinition {
  id: string
  projectSlug: string
  name: string
  displayName: string
  description: string
  isMultiSelect: boolean
  options: Array<{
    code: string
    label: string
    order: number
    description?: string
  }>
  createdAt: string
  updatedAt: string
}

interface OptionRow {
  code: string
  label: string
  description: string
}

const NAME_REGEX = /^[A-Za-z][A-Za-z0-9]*$/
const OPTION_CODE_REGEX = /^[A-Za-z][A-Za-z0-9_]*$/

interface EnumsQueryData {
  enums: EnumDefinition[]
}

interface EnumPayloadError {
  __typename: string
  message: string
  suggestion?: string
}

interface DeleteEnumResult {
  deleteEnum: {
    success?: boolean
    error?: EnumPayloadError
  }
}

interface CreateEnumResult {
  createEnum: {
    enum?: EnumDefinition
    error?: EnumPayloadError
  }
}

interface UpdateEnumResult {
  updateEnum: {
    enum?: EnumDefinition
    error?: EnumPayloadError
  }
}

export default function EnumsPage() {
  const params = useParams()
  const projectSlug = params?.projectSlug as string
  const [searchQuery, setSearchQuery] = useState('')
  const [deletingName, setDeletingName] = useState<string | null>(null)
  const [deleteError, setDeleteError] = useState<string | null>(null)
  const [refreshing, setRefreshing] = useState(false)

  // View detail dialog state
  const [viewingEnum, setViewingEnum] = useState<EnumDefinition | null>(null)

  // Create/edit dialog state
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingEnum, setEditingEnum] = useState<EnumDefinition | null>(null)
  const [formName, setFormName] = useState('')
  const [formDisplayName, setFormDisplayName] = useState('')
  const [formDescription, setFormDescription] = useState('')
  const [options, setOptions] = useState<OptionRow[]>([{ code: '', label: '', description: '' }])
  const [submitting, setSubmitting] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)
  const [nameError, setNameError] = useState<string | null>(null)
  const [displayNameError, setDisplayNameError] = useState<string | null>(null)
  const [optionErrors, setOptionErrors] = useState<Record<number, { code?: string; label?: string }>>({})

  const projectClient = useProjectScopedClient(projectSlug)

  const { data, loading, error, refetch } = useQuery<EnumsQueryData>(GET_ENUMS, {
    client: projectClient,
    skip: !projectSlug,
  })

  const [deleteEnum] = useMutation<DeleteEnumResult>(DELETE_ENUM, {
    client: projectClient,
    update(cache, result, { variables }) {
      if (!result.data?.deleteEnum?.success) return
      cache.modify({
        fields: {
          enums(existing: readonly Reference[] = [], { readField }) {
            return existing.filter(
              (ref) => readField<string>('name', ref) !== variables?.name
            )
          },
        },
      })
    },
    onCompleted: (mutationData) => {
      if (mutationData.deleteEnum.success) {
        setDeletingName(null)
      } else if (mutationData.deleteEnum.error) {
        setDeleteError(mutationData.deleteEnum.error.message)
        setDeletingName(null)
      }
    },
    onError: (err) => {
      setDeleteError(err.message)
      setDeletingName(null)
    },
  })

  const [createEnum] = useMutation<CreateEnumResult>(CREATE_ENUM, {
    client: projectClient,
    update(cache, result) {
      const newEnum = result.data?.createEnum?.enum
      if (!newEnum) return
      cache.modify({
        fields: {
          enums(existing: readonly Reference[] = []) {
            const newRef = cache.writeFragment({
              data: newEnum,
              fragment: gql`
                fragment NewEnum on EnumDefinition {
                  id
                  projectSlug
                  name
                  displayName
                  description
                  isMultiSelect
                  options {
                    code
                    label
                    order
                    description
                  }
                  createdAt
                  updatedAt
                }
              `,
            })
            return [...existing, newRef as Reference]
          },
        },
      })
    },
  })

  const [updateEnum] = useMutation<UpdateEnumResult>(UPDATE_ENUM, {
    client: projectClient,
  })

  const handleRefresh = async () => {
    setRefreshing(true)
    await refetch()
    setRefreshing(false)
  }

  const handleDelete = async (name: string) => {
    if (!confirm(`确定删除枚举 "${name}" 吗？此操作不可撤销。`)) return
    setDeletingName(name)
    setDeleteError(null)
    await deleteEnum({ variables: { name } })
  }

  const handleEdit = (enumItem: EnumDefinition) => {
    setEditingEnum(enumItem)
    setFormName(enumItem.name)
    setFormDisplayName(enumItem.displayName)
    setFormDescription(enumItem.description || '')
    setOptions(enumItem.options.length > 0
      ? enumItem.options.map((opt) => ({ code: opt.code, label: opt.label, description: opt.description || '' }))
      : [{ code: '', label: '', description: '' }]
    )
    setFormError(null)
    setNameError(null)
    setDisplayNameError(null)
    setOptionErrors({})
    setDialogOpen(true)
  }

  const openDialog = () => {
    setEditingEnum(null)
    setFormName('')
    setFormDisplayName('')
    setFormDescription('')
    setOptions([{ code: '', label: '', description: '' }])
    setFormError(null)
    setNameError(null)
    setDisplayNameError(null)
    setOptionErrors({})
    setDialogOpen(true)
  }

  const addOption = () => {
    setOptions([...options, { code: '', label: '', description: '' }])
  }

  const removeOption = (idx: number) => {
    setOptions(options.filter((_, i) => i !== idx))
    setOptionErrors((prev) => {
      const next: Record<number, { code?: string; label?: string }> = {}
      Object.entries(prev).forEach(([key, val]) => {
        const k = Number(key)
        if (k < idx) next[k] = val
        else if (k > idx) next[k - 1] = val
      })
      return next
    })
  }

  const updateOption = (idx: number, field: keyof OptionRow, value: string) => {
    const next = [...options]
    next[idx] = { ...next[idx], [field]: value }
    setOptions(next)
    if (field === 'code' || field === 'label') {
      setOptionErrors((prev) => {
        const row = { ...prev[idx] }
        delete row[field]
        return { ...prev, [idx]: row }
      })
    }
  }

  const handleSubmit = async () => {
    let valid = true
    const newOptionErrors: Record<number, { code?: string; label?: string }> = {}

    // Validate name
    if (!formName) {
      setNameError('枚举名称不能为空')
      valid = false
    } else if (!NAME_REGEX.test(formName)) {
      setNameError('英文字母开头，仅含字母和数字')
      valid = false
    } else {
      setNameError(null)
    }

    // Validate displayName
    if (!formDisplayName) {
      setDisplayNameError('显示标题不能为空')
      valid = false
    } else {
      setDisplayNameError(null)
    }

    // Validate options
    options.forEach((opt, i) => {
      const rowErr: { code?: string; label?: string } = {}
      if (!opt.code.trim()) {
        rowErr.code = '选项代码不能为空'
        valid = false
      } else if (!OPTION_CODE_REGEX.test(opt.code.trim())) {
        rowErr.code = '字母开头，仅含字母、数字和下划线'
        valid = false
      }
      if (!opt.label.trim()) {
        rowErr.label = '显示名称不能为空'
        valid = false
      }
      if (Object.keys(rowErr).length > 0) {
        newOptionErrors[i] = rowErr
      }
    })

    // Check for duplicate codes
    const codes = options.map((o) => o.code.trim()).filter(Boolean)
    const seen = new Set<string>()
    codes.forEach((code, i) => {
      if (seen.has(code)) {
        if (!newOptionErrors[i]) {
          newOptionErrors[i] = {}
        }
        newOptionErrors[i].code = '选项代码不能重复'
        valid = false
      } else {
        seen.add(code)
      }
    })
    setOptionErrors(newOptionErrors)

    if (!valid) return

    setSubmitting(true)
    setFormError(null)

    try {
      if (editingEnum) {
        // Update existing enum
        const result = await updateEnum({
          variables: {
            name: editingEnum.name,
            input: {
              displayName: formDisplayName,
              description: formDescription || undefined,
              options: options.map((opt, i) => ({
                code: opt.code.trim(),
                label: opt.label.trim(),
                order: i,
                description: opt.description.trim() || undefined,
              })),
            },
          },
        })

        const payload = result.data?.updateEnum
        if (payload?.error) {
          const msg = payload.error.message || '更新失败'
          const suggestion = payload.error.suggestion
          setFormError(suggestion ? `${msg} — ${suggestion}` : msg)
          return
        }
      } else {
        // Create new enum
        const result = await createEnum({
          variables: {
            input: {
              name: formName,
              displayName: formDisplayName,
              description: formDescription || undefined,
              options: options.map((opt, i) => ({
                code: opt.code.trim(),
                label: opt.label.trim(),
                order: i,
                description: opt.description.trim() || undefined,
              })),
            },
          },
        })

        const payload = result.data?.createEnum
        if (payload?.error) {
          const msg = payload.error.message || '创建失败'
          const suggestion = payload.error.suggestion
          setFormError(suggestion ? `${msg} — ${suggestion}` : msg)
          return
        }
      }

      // Success
      setDialogOpen(false)
    } catch (err: unknown) {
      setFormError(err instanceof Error ? err.message : editingEnum ? '更新失败，请重试' : '创建失败，请重试')
    } finally {
      setSubmitting(false)
    }
  }

  const enums: EnumDefinition[] = data?.enums || []

  const filteredEnums = enums.filter(
    (e) =>
      e.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      e.displayName.toLowerCase().includes(searchQuery.toLowerCase()) ||
      e.description?.toLowerCase().includes(searchQuery.toLowerCase())
  )

  return (
    <div className="h-full overflow-auto bg-white">
      <div className="mx-auto max-w-7xl p-6">
        {/* Page Title */}
        <div className="mb-8">
          <h1 className="font-heading text-xl font-semibold tracking-tight text-foreground">
            枚举管理
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            管理系统中的枚举定义，用于构建类型安全的字段选项
          </p>
        </div>

        {/* Error Banner */}
        {deleteError && (
          <div className="mb-4 flex items-start gap-3 rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600">
            <span>{deleteError}</span>
            <button
              onClick={() => setDeleteError(null)}
              className="ml-auto shrink-0 text-red-400 hover:text-red-600"
            >
              ✕
            </button>
          </div>
        )}

        {/* Header Actions */}
        <div className="mb-6 flex items-center gap-3">
          <div className="relative max-w-xs flex-1">
            <Search
              className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
              strokeWidth={1.5}
            />
            <Input
              type="text"
              placeholder="搜索枚举..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="h-9 border border-slate-200 bg-white pl-9 text-sm text-foreground placeholder:text-muted-foreground focus:border-blue-600 focus:ring-1 focus:ring-blue-100"
            />
          </div>
          <button
            onClick={handleRefresh}
            disabled={refreshing || loading}
            className="inline-flex h-9 items-center justify-center rounded-md border border-slate-200 bg-white px-3 text-muted-foreground transition-colors hover:bg-slate-50 hover:text-foreground disabled:opacity-50"
            title="刷新列表"
          >
            <RefreshCw className={`size-4 ${refreshing ? 'animate-spin' : ''}`} strokeWidth={1.5} />
          </button>
          <Button
            onClick={openDialog}
            className="h-9 gap-2 rounded-md border-0 bg-blue-600 px-4 text-sm font-medium text-white transition-all duration-200 hover:bg-blue-700"
          >
            <Plus className="size-4" strokeWidth={1.5} />
            创建枚举
          </Button>
        </div>

        {/* Table */}
        <div className="overflow-hidden rounded-lg border border-slate-200 bg-white">
          {loading ? (
            <div className="flex items-center justify-center py-12 text-sm text-muted-foreground">
              加载中…
            </div>
          ) : error ? (
            <div className="flex items-center justify-center py-12 text-sm text-red-500">
              加载失败：{error.message}
            </div>
          ) : (
            <>
              <div className="overflow-x-auto">
                <table className="w-full border-collapse text-sm">
                  <thead className="border-b border-slate-200 bg-slate-50">
                    <tr>
                      <th className="h-10 px-4 py-3 text-left font-semibold text-foreground">枚举名称</th>
                      <th className="h-10 px-4 py-3 text-left font-semibold text-foreground">显示标题</th>
                      <th className="h-10 px-4 py-3 text-left font-semibold text-foreground">描述</th>
                      <th className="h-10 px-4 py-3 text-right font-semibold text-foreground">操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filteredEnums.map((enumItem) => (
                      <tr
                        key={enumItem.id}
                        className="border-b border-slate-200 transition-colors last:border-0 hover:bg-slate-50"
                      >
                        <td className="px-4 py-3 font-mono text-sm text-foreground">
                          {enumItem.name}
                        </td>
                        <td className="px-4 py-3 text-sm font-medium text-foreground">
                          {enumItem.displayName}
                        </td>
                        <td className="px-4 py-3 text-sm text-muted-foreground">
                          {enumItem.description}
                        </td>
                        <td className="px-4 py-3 text-right">
                          <div className="flex items-center justify-end gap-1">
                            <button
                              onClick={() => setViewingEnum(enumItem)}
                              className="inline-flex h-8 items-center justify-center gap-1.5 whitespace-nowrap rounded-md px-3 text-sm text-muted-foreground transition-colors hover:bg-slate-100 hover:text-foreground"
                              title="查看选项详情"
                            >
                              <Eye className="size-4" strokeWidth={1.5} />
                              <span>查看</span>
                            </button>
                            <button
                              onClick={() => handleEdit(enumItem)}
                              className="inline-flex h-8 items-center justify-center gap-1.5 whitespace-nowrap rounded-md px-3 text-sm text-muted-foreground transition-colors hover:bg-slate-100 hover:text-foreground"
                            >
                              <Edit className="size-4" strokeWidth={1.5} />
                              <span>编辑</span>
                            </button>
                            <button
                              onClick={() => handleDelete(enumItem.name)}
                              disabled={deletingName === enumItem.name}
                              className="inline-flex h-8 items-center justify-center gap-1.5 whitespace-nowrap rounded-md px-3 text-sm text-muted-foreground transition-colors hover:bg-red-50 hover:text-red-600 disabled:opacity-50"
                            >
                              <Trash2 className="size-4" strokeWidth={1.5} />
                              <span>{deletingName === enumItem.name ? '删除中…' : '删除'}</span>
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              {filteredEnums.length === 0 && (
                <div className="flex items-center justify-center py-12 text-sm text-muted-foreground">
                  {searchQuery ? '未找到匹配的枚举' : '暂无枚举定义'}
                </div>
              )}
            </>
          )}
        </div>
      </div>

      {/* View Options Detail Dialog */}
      <Dialog open={!!viewingEnum} onOpenChange={(open) => { if (!open) setViewingEnum(null) }}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>
              {viewingEnum?.displayName}
              <span className="ml-2 font-mono text-sm font-normal text-muted-foreground">
                ({viewingEnum?.name})
              </span>
            </DialogTitle>
          </DialogHeader>

          <div className="py-2">
            {viewingEnum?.description && (
              <p className="mb-4 text-sm text-muted-foreground">{viewingEnum.description}</p>
            )}

            <div className="mb-3 text-sm text-muted-foreground">
              <span>{viewingEnum?.options.length ?? 0} 个选项</span>
            </div>

            {viewingEnum && viewingEnum.options.length > 0 ? (
              <div className="overflow-hidden rounded-md border border-slate-200">
                <table className="w-full border-collapse text-sm">
                  <thead className="border-b border-slate-200 bg-slate-50">
                    <tr>
                      <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground">代码</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground">显示名称</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground">描述</th>
                    </tr>
                  </thead>
                  <tbody>
                    {viewingEnum.options
                      .slice()
                      .sort((a, b) => a.order - b.order)
                      .map((opt) => (
                        <tr key={opt.code} className="border-b border-slate-100 last:border-0">
                          <td className="px-3 py-2 font-mono text-xs text-foreground">{opt.code}</td>
                          <td className="px-3 py-2 text-sm text-foreground">{opt.label}</td>
                          <td className="px-3 py-2 text-sm text-muted-foreground">{opt.description || '—'}</td>
                        </tr>
                      ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="rounded-md border border-slate-200 py-6 text-center text-sm text-muted-foreground">
                暂无选项
              </div>
            )}
          </div>

          <DialogFooter>
            <button
              type="button"
              onClick={() => setViewingEnum(null)}
              className="h-9 rounded-md border border-slate-200 bg-white px-4 text-sm font-medium text-foreground hover:border-slate-300 hover:bg-slate-50"
            >
              关闭
            </button>
            <button
              type="button"
              onClick={() => {
                if (viewingEnum) {
                  setViewingEnum(null)
                  handleEdit(viewingEnum)
                }
              }}
              className="h-9 rounded-md bg-blue-600 px-4 text-sm font-medium text-white hover:bg-blue-700"
            >
              编辑
            </button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Create / Edit Enum Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{editingEnum ? '编辑枚举' : '创建枚举'}</DialogTitle>
          </DialogHeader>

          <div className="max-h-[60vh] overflow-y-auto pr-1">
            <div className="space-y-4 py-2">
              {/* Server error */}
              {formError && (
                <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-600">
                  {formError}
                </div>
              )}

              {/* Name */}
              <div>
                <label className="mb-1.5 block text-sm font-medium text-foreground">
                  枚举名称 <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  value={formName}
                  onChange={(e) => {
                    setFormName(e.target.value)
                    setNameError(null)
                  }}
                  disabled={!!editingEnum}
                  placeholder="例如：OrderStatus"
                  className="w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm text-foreground transition-all duration-200 placeholder:text-muted-foreground focus:border-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-100/50 disabled:cursor-not-allowed disabled:bg-slate-50 disabled:text-muted-foreground"
                />
                {nameError ? (
                  <p className="mt-1 text-xs text-red-500">{nameError}</p>
                ) : (
                  <p className="mt-1 text-xs text-muted-foreground">
                    {editingEnum ? '枚举名称创建后不可修改' : '英文字母开头，仅含字母和数字'}
                  </p>
                )}
              </div>

              {/* Title */}
              <div>
                <label className="mb-1.5 block text-sm font-medium text-foreground">
                  显示标题 <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  value={formDisplayName}
                  onChange={(e) => {
                    setFormDisplayName(e.target.value)
                    setDisplayNameError(null)
                  }}
                  placeholder="例如：订单状态"
                  className="w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm text-foreground transition-all duration-200 placeholder:text-muted-foreground focus:border-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-100/50"
                />
                {displayNameError && (
                  <p className="mt-1 text-xs text-red-500">{displayNameError}</p>
                )}
              </div>

              {/* Description */}
              <div>
                <label className="mb-1.5 block text-sm font-medium text-foreground">
                  描述
                </label>
                <textarea
                  value={formDescription}
                  onChange={(e) => setFormDescription(e.target.value)}
                  placeholder="可选，描述此枚举的用途"
                  rows={2}
                  className="w-full resize-none rounded-md border border-slate-200 bg-white px-3 py-2 text-sm text-foreground transition-all duration-200 placeholder:text-muted-foreground focus:border-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-100/50"
                />
              </div>

              {/* Options */}
              <div className="border-t border-slate-100 pt-4">
                <label className="mb-3 block text-sm font-medium text-foreground">
                  枚举选项 <span className="text-red-500">*</span>
                </label>

                <div className="mb-1 flex items-center gap-2 pr-9 text-xs text-muted-foreground">
                  <span className="flex-1">代码 Code</span>
                  <span className="flex-1">显示名称 Label</span>
                </div>

                <div className="space-y-2">
                  {options.map((opt, idx) => (
                    <div key={idx} className="flex items-start gap-2">
                      <div className="flex-1">
                        <input
                          type="text"
                          value={opt.code}
                          onChange={(e) => updateOption(idx, 'code', e.target.value)}
                          placeholder="如：admin"
                          className="w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm text-foreground transition-all duration-200 placeholder:text-muted-foreground focus:border-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-100/50"
                        />
                        {optionErrors[idx]?.code && (
                          <p className="mt-1 text-xs text-red-500">{optionErrors[idx].code}</p>
                        )}
                      </div>
                      <div className="flex-1">
                        <input
                          type="text"
                          value={opt.label}
                          onChange={(e) => updateOption(idx, 'label', e.target.value)}
                          placeholder="如：管理员"
                          className="w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm text-foreground transition-all duration-200 placeholder:text-muted-foreground focus:border-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-100/50"
                        />
                        {optionErrors[idx]?.label && (
                          <p className="mt-1 text-xs text-red-500">{optionErrors[idx].label}</p>
                        )}
                      </div>
                      <button
                        type="button"
                        onClick={() => removeOption(idx)}
                        disabled={options.length <= 1}
                        className={`mt-0.5 inline-flex size-7 shrink-0 items-center justify-center rounded text-muted-foreground transition-colors hover:bg-red-50 hover:text-red-500 ${options.length <= 1 ? 'invisible' : ''}`}
                        title="删除此选项"
                      >
                        <Trash2 className="size-3.5" strokeWidth={1.5} />
                      </button>
                    </div>
                  ))}
                </div>

                <button
                  type="button"
                  onClick={addOption}
                  className="mt-2 inline-flex h-8 items-center gap-1.5 rounded-md px-2 text-sm text-muted-foreground transition-colors hover:bg-slate-100 hover:text-foreground"
                >
                  <Plus className="size-3.5" strokeWidth={1.5} />
                  添加选项
                </button>
              </div>
            </div>
          </div>

          <DialogFooter>
            <button
              type="button"
              onClick={() => setDialogOpen(false)}
              disabled={submitting}
              className="h-9 rounded-md border border-slate-200 bg-white px-4 text-sm font-medium text-foreground hover:border-slate-300 hover:bg-slate-50 disabled:opacity-50"
            >
              取消
            </button>
            <button
              type="button"
              onClick={handleSubmit}
              disabled={submitting}
              className="h-9 rounded-md bg-blue-600 px-4 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
            >
              {submitting
                ? editingEnum ? '保存中...' : '创建中...'
                : editingEnum ? '保存' : '创建'
              }
            </button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
