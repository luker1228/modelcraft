'use client'

import { useState, useMemo, useCallback, useEffect, useRef } from 'react'
import { useMutation, useQuery } from '@apollo/client'
import { useProjectScopedClient } from '@bff/apollo/public'
import { createFieldEnumRelation, queryModelEnumContext } from '@bff/model-enum/public'
import { ADD_FIELDS, GET_LOGICAL_FOREIGN_KEYS } from '@web/graphql'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { Label } from '@web/components/ui/label'
import { Textarea } from '@web/components/ui/textarea'
import { Switch } from '@web/components/ui/switch'
import {
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
} from '@web/components/ui/drawer'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@web/components/ui/tooltip'
import { Columns, Loader2, Save, HelpCircle, Link2 } from 'lucide-react'
import type {
  EnumSourceOption,
  LogicalForeignKey,
  ModelEnumDomainError,
} from '@/types'

const FORMAT_TYPE_OPTIONS = [
  { value: 'STRING', label: '字符串 (String)' },
  { value: 'ENUM', label: '枚举 (Enum)' },
  { value: 'ENUM_LABEL', label: '枚举标签 (Enum Label)' },
  { value: 'UUID', label: 'UUID v7' },
  { value: 'DATE', label: '日期 (Date)' },
  { value: 'DATETIME', label: '日期时间 (DateTime)' },
  { value: 'TIME', label: '时间 (Time)' },
  { value: 'INTEGER', label: '整数 (Integer)' },
  { value: 'NUMBER', label: '浮点数 (Number)' },
  { value: 'DECIMAL', label: '精确小数 (Decimal)' },
  { value: 'BOOLEAN', label: '布尔值 (Boolean)' },
  { value: 'RELATION', label: '关联关系 (Relation)' },
]

const STRING_STORAGE_HINTS = [
  { value: '_default_', label: '默认 (VARCHAR 255)' },
  { value: 'VARCHAR(32)', label: 'VARCHAR(32) - 短文本' },
  { value: 'VARCHAR(64)', label: 'VARCHAR(64) - 标识符' },
  { value: 'VARCHAR(128)', label: 'VARCHAR(128) - 标题' },
  { value: 'VARCHAR(256)', label: 'VARCHAR(256) - 摘要' },
  { value: 'VARCHAR(512)', label: 'VARCHAR(512) - 长摘要' },
  { value: 'TEXT', label: 'TEXT - 长文本' },
]

const DEFAULT_FIELD_DATA = {
  name: '',
  title: '',
  format: 'STRING',
  storageHint: '',
  nonNull: false,
  required: false,
  unique: false,
  description: '',
  relateFkId: '',
  relateEnumName: '',
  sourceFieldName: '',
  relationDirection: 'NORMAL' as 'NORMAL' | 'REVERSE',
}

const FIELD_NAME_REGEX = /^[a-zA-Z][a-zA-Z0-9_]*$/

export interface InsertFieldSheetProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  modelId: string
  modelName?: string
  projectSlug: string
  orgName: string
  /** Existing field names in the model, used for duplicate name validation */
  existingFieldNames?: string[]
  /** Called after a field is successfully added */
  onSuccess?: () => void
}

export function InsertFieldSheet({
  open,
  onOpenChange,
  modelId,
  modelName,
  projectSlug,
  orgName,
  existingFieldNames = [],
  onSuccess,
}: InsertFieldSheetProps) {
  const projectClient = useProjectScopedClient(projectSlug)

  const projectScopedContext = useMemo(() => {
    if (!orgName || !projectSlug) return undefined
    return { uri: `/graphql/org/${orgName}/project/${projectSlug}/` }
  }, [orgName, projectSlug])

  const [fieldData, setFieldData] = useState(DEFAULT_FIELD_DATA)
  const [saving, setSaving] = useState(false)
  const [submitError, setSubmitError] = useState<string | null>(null)
  const [continueMode, setContinueMode] = useState(false)
  const continueModeRef = useRef(false)

  const trimmedFieldName = fieldData.name.trim()
  const isDuplicateName = trimmedFieldName !== '' && existingFieldNames.includes(trimmedFieldName)
  const isInvalidFieldNameFormat = trimmedFieldName !== '' && !FIELD_NAME_REGEX.test(trimmedFieldName)
  const [enumContextLoading, setEnumContextLoading] = useState(false)
  const [enumContextError, setEnumContextError] = useState<ModelEnumDomainError | null>(null)
  const [enumSources, setEnumSources] = useState<EnumSourceOption[]>([])

  // 查询逻辑外键列表（仅当抽屉打开时）
  const { data: fkData, loading: fkLoading } = useQuery<{ logicalForeignKeys: LogicalForeignKey[] }>(GET_LOGICAL_FOREIGN_KEYS, {
    skip: !open || !modelId,
    variables: { modelId },
    context: projectScopedContext,
    client: projectClient,
    fetchPolicy: 'network-only',
  })

  const logicalForeignKeys: LogicalForeignKey[] = fkData?.logicalForeignKeys ?? []

  // 按关联方向过滤外键列表
  // NORMAL（多对一）：当前模型持有 FK 列，direction = NORMAL
  // REVERSE（一对多）：当前模型被引用，direction = REVERSE
  const filteredForeignKeys = logicalForeignKeys.filter(
    fk => fk.direction === fieldData.relationDirection
  )

  const enumOptions = useMemo(
    () => Array.from(new Set(enumSources.map((source) => source.enumName))),
    [enumSources],
  )

  const selectedSource = useMemo(
    () => enumSources.find((source) => source.fieldName === fieldData.sourceFieldName),
    [enumSources, fieldData.sourceFieldName],
  )

  const refreshEnumContext = useCallback(async () => {
    if (!open || !modelId) {
      setEnumSources([])
      return null
    }

    setEnumContextLoading(true)
    setEnumContextError(null)

    try {
      const result = await queryModelEnumContext(
        { orgName, projectSlug, modelId },
        projectClient,
      )

      if (result.error) {
        setEnumContextError(result.error)
        setEnumSources([])
        return null
      }

      setEnumSources(result.enumSources)
      return result
    } catch {
      const error: ModelEnumDomainError = {
        type: 'Unknown',
        code: 'UNKNOWN',
        message: '加载 ENUM 上下文失败，请稍后重试。',
      }
      setEnumContextError(error)
      setEnumSources([])
      return null
    } finally {
      setEnumContextLoading(false)
    }
  }, [modelId, open, orgName, projectSlug, projectClient])

  useEffect(() => {
    if (!open) {
      setEnumContextError(null)
      setEnumSources([])
      return
    }

    if (fieldData.format === 'ENUM' || fieldData.format === 'ENUM_LABEL') {
      void refreshEnumContext()
    }
  }, [fieldData.format, open, refreshEnumContext])

  const [addField] = useMutation(ADD_FIELDS, {
    client: projectClient,
    context: projectScopedContext,
    onCompleted: (data: { addFields?: { error?: { message?: string } | null } | null }) => {
      const bizError = data?.addFields?.error
      if (bizError) {
        setSubmitError(bizError.message ?? '添加字段失败')
        setSaving(false)
        return
      }
      setFieldData(DEFAULT_FIELD_DATA)
      setSubmitError(null)
      setSaving(false)
      onSuccess?.()
      if (!continueModeRef.current) {
        onOpenChange(false)
      }
    },
    onError: (error) => {
      setSubmitError('添加字段失败: ' + error.message)
      setSaving(false)
    },
    refetchQueries: ['GetModel', 'GetModelJsonSchema'],
  })

  const handleSave = async (andContinue = false) => {
    continueModeRef.current = andContinue
    setContinueMode(andContinue)
    if (!fieldData.name || !fieldData.title || !fieldData.format) {
      setSubmitError('请填写字段名、标题和类型')
      return
    }

    if (isInvalidFieldNameFormat) {
      setSubmitError("字段标识格式无效：只允许字母、数字、下划线，且必须以字母开头（不允许 '_' 前缀）")
      return
    }

    if (isDuplicateName) {
      setSubmitError(`字段标识 '${fieldData.name}' 已存在`)
      return
    }

    if (fieldData.format === 'RELATION' && !fieldData.relateFkId) {
      setSubmitError('请选择关联的逻辑外键')
      return
    }

    if (fieldData.format === 'ENUM' && !fieldData.relateEnumName) {
      setSubmitError('请选择关联枚举')
      return
    }

    if (fieldData.format === 'ENUM_LABEL') {
      if (!fieldData.sourceFieldName) {
        setSubmitError('请选择 sourceField（本表 ENUM 字段）')
        return
      }

      if (selectedSource?.occupied) {
        setSubmitError('当前 source 字段已占用，不能重复创建 ENUM_LABEL')
        return
      }
    }

    const input: Record<string, unknown> = {
      name: trimmedFieldName,
      title: fieldData.title,
      format: fieldData.format,
      storageHint: fieldData.storageHint !== '' ? fieldData.storageHint : undefined,
      nonNull: fieldData.nonNull,
      required: fieldData.required,
      isUnique: fieldData.unique,
      description: fieldData.description || undefined,
      relateFkId: fieldData.relateFkId || undefined,
    }

    if (fieldData.format === 'ENUM') {
      input.relateEnumName = fieldData.relateEnumName
    }

    setSubmitError(null)
    setSaving(true)
    try {
      if (fieldData.format === 'ENUM_LABEL') {
        const source = enumSources.find((option) => option.fieldName === fieldData.sourceFieldName)
        if (!source) {
          setSubmitError('未找到 source 字段，请重新选择')
          setSaving(false)
          return
        }

        const relationResult = await createFieldEnumRelation({
          orgName,
          projectSlug,
          modelId,
          sourceFieldName: source.fieldName,
          enumName: source.enumName,
          labelFieldName: trimmedFieldName,
        })

        if (!relationResult.success) {
          setSubmitError(relationResult.error?.message ?? '创建 relation 失败，请稍后重试。')
          setSaving(false)
          return
        }

        const latestContext = await refreshEnumContext()
        const matchedRelation = latestContext?.relations.find(
          (relation) => relation.sourceFieldName === source.fieldName && relation.labelFieldName === trimmedFieldName,
        )

        if (!matchedRelation) {
          setSubmitError('relation 创建成功但未查询到关系，请重试。')
          setSaving(false)
          return
        }

        input.enumRelationId = matchedRelation.id
      }

      await addField({
        variables: {
          projectSlug: projectSlug,
          modelID: modelId,
          input: [input],
        },
      })
    } catch {
      // error handled in onError
    }
  }

  const handleOpenChange = (nextOpen: boolean) => {
    if (!nextOpen) {
      setFieldData(DEFAULT_FIELD_DATA)
      setSubmitError(null)
      setEnumContextError(null)
      setEnumSources([])
    }
    onOpenChange(nextOpen)
  }

  /** 格式化逻辑外键的显示标签 */
  const formatFkLabel = (fk: LogicalForeignKey) => {
    const fieldPairs = fk.sourceFields.map((sf, i) => {
      const tf = fk.targetFields[i] ?? '?'
      return `${fk.modelName}.${sf} → ${fk.refModelName}.${tf}`
    }).join(', ')
    return fieldPairs
  }

  const isSubmitDisabled =
    saving
    || !fieldData.name
    || !fieldData.title
    || isInvalidFieldNameFormat
    || isDuplicateName
    || (fieldData.format === 'RELATION' && !fieldData.relateFkId)
    || (fieldData.format === 'ENUM' && !fieldData.relateEnumName)
    || (
      fieldData.format === 'ENUM_LABEL'
      && (!fieldData.sourceFieldName || Boolean(selectedSource?.occupied))
    )

  return (
    <Drawer open={open} onOpenChange={handleOpenChange} direction="right">
      <DrawerContent className="left-auto right-0 h-screen max-h-screen w-[600px] rounded-none border-l border-border">
        <div className="flex h-full flex-col overflow-hidden">
          {/* Header */}
          <DrawerHeader className="border-b border-border px-6 py-4">
            <DrawerTitle className="flex items-center gap-2 text-lg font-semibold">
              <Columns className="size-5 text-primary" strokeWidth={1.5} />
              插入字段
            </DrawerTitle>
            <DrawerDescription className="mt-1 text-sm">
              为模型 <span className="font-mono text-primary">{modelName}</span> 添加新字段
            </DrawerDescription>
          </DrawerHeader>

          {/* Content - Scrollable */}
          <div className="no-scrollbar flex-1 overflow-y-auto px-6">
            <div className="divide-y divide-border py-4">
              {/* 基础属性 */}
              <div className="grid grid-cols-12 gap-4 py-5">
                <div className="col-span-4">
                  <p className="text-sm font-semibold text-foreground">基础属性</p>
                </div>
                <div className="col-span-8 flex flex-col gap-4">

                  {/* 字段标识 */}
                  <div className="space-y-2">
                    <Label className="flex items-center gap-1.5 text-sm text-foreground">
                      字段标识 <span className="text-xs text-destructive">*</span>
                    </Label>
                    <Input
                      value={fieldData.name}
                      onChange={(e) => {
                        setFieldData(prev => ({ ...prev, name: e.target.value }))
                        setSubmitError(null)
                      }}
                      placeholder="例如：title, status"
                      className={`font-mono ${isDuplicateName ? 'border-destructive focus-visible:ring-destructive' : ''}`}
                    />
                    {isInvalidFieldNameFormat ? (
                      <p className="text-xs text-destructive">
                        字段标识格式无效：只允许字母、数字、下划线，且必须以字母开头（不允许 &apos;_&apos; 前缀）
                      </p>
                    ) : isDuplicateName ? (
                      <p className="text-xs text-destructive">字段标识 &apos;{fieldData.name}&apos; 已存在</p>
                    ) : (
                      <p className="text-xs text-muted-foreground">字母开头，可包含字母、数字和下划线</p>
                    )}
                  </div>

                  {/* 显示名字 */}
                  <div className="space-y-2">
                    <Label className="flex items-center gap-1.5 text-sm text-foreground">
                      显示名字 <span className="text-xs text-destructive">*</span>
                    </Label>
                    <Input
                      value={fieldData.title}
                      onChange={(e) => setFieldData(prev => ({ ...prev, title: e.target.value }))}
                      placeholder="例如：标题、描述、状态"
                    />
                  </div>

                  {/* 字段描述 */}
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <Label className="text-sm text-foreground">字段描述</Label>
                      <span className="text-xs text-muted-foreground">可选</span>
                    </div>
                    <Textarea
                      value={fieldData.description}
                      onChange={(e) => setFieldData(prev => ({ ...prev, description: e.target.value }))}
                      placeholder="输入字段描述"
                      rows={2}
                    />
                  </div>
                </div>
              </div>

              {/* 类型设置 */}
              <div className="grid grid-cols-12 gap-4 py-5">
                <div className="col-span-4">
                  <p className="text-sm font-semibold text-foreground">类型设置</p>
                </div>
                <div className="col-span-8 flex flex-col gap-4">

                  {/* 字段类型 */}
                  <div className="space-y-2">
                    <Label className="flex items-center gap-1.5 text-sm text-foreground">
                      字段类型 <span className="text-xs text-destructive">*</span>
                      {fieldData.format === 'UUID' && (
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <HelpCircle className="size-3.5 cursor-help text-muted-foreground" strokeWidth={1.5} />
                            </TooltipTrigger>
                            <TooltipContent side="left" className="max-w-[220px] text-xs leading-relaxed">
                              <p className="mb-1 font-semibold">UUID v7</p>
                              <p>基于时间戳的有序 UUID，相比 v4 随机 UUID：</p>
                              <ul className="mt-1 list-inside list-disc space-y-0.5">
                                <li>按时间自然排序，索引性能更好</li>
                                <li>可从 ID 推断创建时间</li>
                                <li>适合作为主键使用</li>
                              </ul>
                            </TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                      )}
                      {fieldData.format === 'RELATION' && (
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <HelpCircle className="size-3.5 cursor-help text-muted-foreground" strokeWidth={1.5} />
                            </TooltipTrigger>
                            <TooltipContent side="left" className="max-w-[240px] text-xs leading-relaxed">
                              <p className="mb-1 font-semibold">关联关系字段</p>
                              <p>基于逻辑外键生成的虚拟关联字段，用于 GraphQL API 中展示关联数据。</p>
                              <ul className="mt-1 list-inside list-disc space-y-0.5">
                                <li>不在数据库中产生实际列</li>
                                <li>需选择已定义的逻辑外键</li>
                                <li>在 API 中自动展开关联对象</li>
                              </ul>
                            </TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                      )}
                    </Label>
                    <Select
                      value={fieldData.format}
                      onValueChange={(value) => setFieldData(prev => ({
                        ...prev,
                        format: value,
                        storageHint: value === 'STRING' ? prev.storageHint : '',
                        relateFkId: value === 'RELATION' ? prev.relateFkId : '',
                        relateEnumName: value === 'ENUM' ? prev.relateEnumName : '',
                        sourceFieldName: value === 'ENUM_LABEL' ? prev.sourceFieldName : '',
                        relationDirection: value === 'RELATION' ? prev.relationDirection : 'NORMAL',
                      }))}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="选择字段类型" />
                      </SelectTrigger>
                      <SelectContent>
                        {FORMAT_TYPE_OPTIONS.map((option) => (
                          <SelectItem key={option.value} value={option.value}>
                            {option.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  {/* 枚举绑定 - 仅 ENUM 类型 */}
                  {fieldData.format === 'ENUM' && (
                    <div className="space-y-2">
                      <Label className="text-sm text-foreground">
                        关联枚举 <span className="text-xs text-destructive">*</span>
                      </Label>
                      {enumContextLoading ? (
                        <div className="flex h-9 items-center gap-2 rounded-md border border-border px-3 text-sm text-muted-foreground">
                          <Loader2 className="size-3.5 animate-spin" strokeWidth={1.5} />
                          加载枚举列表...
                        </div>
                      ) : (
                        <Select
                          value={fieldData.relateEnumName}
                          onValueChange={(value) => setFieldData(prev => ({ ...prev, relateEnumName: value }))}
                          disabled={enumOptions.length === 0}
                        >
                          <SelectTrigger>
                            <SelectValue placeholder="请选择枚举" />
                          </SelectTrigger>
                          <SelectContent>
                            {enumOptions.length > 0 ? (
                              enumOptions.map((enumName) => (
                                <SelectItem key={enumName} value={enumName}>
                                  <span className="font-mono text-xs">{enumName}</span>
                                </SelectItem>
                              ))
                            ) : (
                              <SelectItem value="__empty__" disabled>
                                暂无可用枚举
                              </SelectItem>
                            )}
                          </SelectContent>
                        </Select>
                      )}
                      {enumContextError && <p className="text-xs text-destructive">{enumContextError.message}</p>}
                      {enumOptions.length === 0 && !enumContextLoading && (
                        <p className="text-xs text-muted-foreground">请先创建枚举后再创建 ENUM 字段。</p>
                      )}
                    </div>
                  )}

                  {/* 枚举标签绑定 - 仅 ENUM_LABEL 类型 */}
                  {fieldData.format === 'ENUM_LABEL' && (
                    <div className="space-y-4">
                      <div className="space-y-2">
                        <Label className="text-sm text-foreground">
                          源字段 (ENUM) <span className="text-xs text-destructive">*</span>
                        </Label>
                        {enumContextLoading ? (
                          <div className="flex h-9 items-center gap-2 rounded-md border border-border px-3 text-sm text-muted-foreground">
                            <Loader2 className="size-3.5 animate-spin" strokeWidth={1.5} />
                            加载 sourceField...
                          </div>
                        ) : (
                          <Select
                            value={fieldData.sourceFieldName}
                            onValueChange={(value) => setFieldData(prev => ({
                              ...prev,
                              sourceFieldName: value,
                            }))}
                            disabled={enumSources.length === 0}
                          >
                            <SelectTrigger>
                              <SelectValue placeholder="请选择 sourceField（本表 ENUM 字段）" />
                            </SelectTrigger>
                            <SelectContent>
                              {enumSources.length > 0 ? (
                                enumSources.map((source) => (
                                  <SelectItem key={source.fieldName} value={source.fieldName} disabled={source.occupied}>
                                    <div className="flex items-center gap-2">
                                      <span className="font-mono text-xs">{source.fieldName}</span>
                                      <span className="text-xs text-muted-foreground">→ {source.enumName}</span>
                                      {source.occupied && <span className="text-xs text-destructive">(已占用)</span>}
                                    </div>
                                  </SelectItem>
                                ))
                              ) : (
                                <SelectItem value="__empty__" disabled>
                                  暂无 ENUM 字段，请先创建
                                </SelectItem>
                              )}
                            </SelectContent>
                          </Select>
                        )}
                        {enumSources.length === 0 && !enumContextLoading && (
                          <p className="text-xs text-muted-foreground">当前模型暂无 ENUM 字段，请先创建。</p>
                        )}
                      </div>

                      <p className="text-xs text-muted-foreground">保存时自动创建并绑定 relation。</p>
                      {selectedSource?.occupied && (
                        <p className="text-xs text-destructive">当前 source 字段已占用，不能重复创建 ENUM_LABEL。</p>
                      )}

                      {enumContextError && <p className="text-xs text-destructive">{enumContextError.message}</p>}
                    </div>
                  )}

                  {/* 存储长度 - 仅字符串类型 */}
                  {fieldData.format === 'STRING' && (
                    <div className="space-y-2">
                      <Label className="text-sm text-foreground">存储长度</Label>
                      <Select
                        value={fieldData.storageHint || '_default_'}
                        onValueChange={(value) => setFieldData(prev => ({ ...prev, storageHint: value === '_default_' ? '' : value }))}
                      >
                        <SelectTrigger>
                          <SelectValue placeholder="选择存储长度" />
                        </SelectTrigger>
                        <SelectContent>
                          {STRING_STORAGE_HINTS.map((option) => (
                            <SelectItem key={option.value} value={option.value}>
                              {option.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <p className="text-xs text-muted-foreground">根据字段用途选择合适的长度，默认 VARCHAR(255)</p>
                    </div>
                  )}

                  {/* 逻辑外键选择 - 仅关联关系类型 */}
                  {fieldData.format === 'RELATION' && (
                    <div className="space-y-3">
                      {/* 关联类型选择 */}
                      <div className="space-y-2">
                        <Label className="flex items-center gap-1.5 text-sm text-foreground">
                          <Link2 className="size-3.5 text-primary" strokeWidth={1.5} />
                          关联类型 <span className="text-xs text-destructive">*</span>
                        </Label>
                        <Select
                          value={fieldData.relationDirection}
                          onValueChange={(value: 'NORMAL' | 'REVERSE') =>
                            setFieldData(prev => ({ ...prev, relationDirection: value, relateFkId: '' }))
                          }
                        >
                          <SelectTrigger>
                            <SelectValue placeholder="选择关联类型" />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="NORMAL">
                              <div className="flex flex-col gap-0.5">
                                <span>多对一（BelongsTo）</span>
                                <span className="text-xs text-muted-foreground">当前模型持有外键列，关联单条目标记录</span>
                              </div>
                            </SelectItem>
                            <SelectItem value="REVERSE">
                              <div className="flex flex-col gap-0.5">
                                <span>一对多（HasMany）</span>
                                <span className="text-xs text-muted-foreground">目标模型持有外键列，关联多条目标记录</span>
                              </div>
                            </SelectItem>
                          </SelectContent>
                        </Select>
                        <p className="text-xs text-muted-foreground">
                          {fieldData.relationDirection === 'NORMAL'
                            ? '多对一：此模型的多条记录可关联目标模型的同一条记录'
                            : '一对多：此模型的一条记录可对应目标模型的多条记录'}
                        </p>
                      </div>

                      {/* 逻辑外键选择 */}
                      <div className="space-y-2">
                        <Label className="text-sm text-foreground">
                          关联的逻辑外键 <span className="text-xs text-destructive">*</span>
                        </Label>
                        {fkLoading ? (
                          <div className="flex h-9 items-center gap-2 rounded-md border border-border px-3 text-sm text-muted-foreground">
                            <Loader2 className="size-3.5 animate-spin" strokeWidth={1.5} />
                            加载外键列表...
                          </div>
                        ) : filteredForeignKeys.length === 0 ? (
                          <div className="rounded-md border border-dashed border-border bg-muted/50 p-3">
                            <p className="text-xs text-muted-foreground">
                              {logicalForeignKeys.length === 0
                                ? '当前模型暂无逻辑外键。请先在模型编辑器的「逻辑外键」区域定义关系，再添加关联字段。'
                                : fieldData.relationDirection === 'NORMAL'
                                  ? '当前模型无多对一方向的逻辑外键（NORMAL）。请先定义一条以此模型为主表的逻辑外键。'
                                  : '当前模型无一对多方向的逻辑外键（REVERSE）。请先定义一条以此模型为被引用表的逻辑外键。'}
                            </p>
                          </div>
                        ) : (
                          <Select
                            value={fieldData.relateFkId}
                            onValueChange={(value) => setFieldData(prev => ({ ...prev, relateFkId: value }))}
                          >
                            <SelectTrigger>
                              <SelectValue placeholder="选择逻辑外键" />
                            </SelectTrigger>
                            <SelectContent>
                              {filteredForeignKeys.map((fk) => (
                                <SelectItem key={fk.id} value={fk.id}>
                                  <span className="font-mono text-xs">{formatFkLabel(fk)}</span>
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        )}
                        <p className="text-xs text-muted-foreground">
                          关联字段将引用此逻辑外键，在 GraphQL API 中展开关联的模型数据
                        </p>
                      </div>
                    </div>
                  )}
                </div>
              </div>

              {/* 约束设置 - 关联关系类型不显示约束 */}
              {fieldData.format !== 'RELATION' && (
                <div className="grid grid-cols-12 gap-4 py-5">
                  <div className="col-span-4">
                    <p className="text-sm font-semibold text-foreground">约束设置</p>
                  </div>
                  <div className="col-span-8 flex flex-col gap-5">

                    <div className="flex flex-row gap-4">
                      <Switch
                        checked={fieldData.nonNull}
                        onCheckedChange={(checked) => setFieldData(prev => ({ ...prev, nonNull: checked }))}
                        className="data-[state=checked]:bg-primary"
                      />
                      <div className="flex flex-col gap-1">
                        <Label className="text-sm text-foreground">非空约束</Label>
                        <p className="text-xs leading-normal text-muted-foreground">数据库层面的 NOT NULL 约束，插入时必须提供值</p>
                      </div>
                    </div>

                    <div className="flex flex-row gap-4">
                      <Switch
                        checked={fieldData.required}
                        onCheckedChange={(checked) => setFieldData(prev => ({ ...prev, required: checked }))}
                        className="data-[state=checked]:bg-primary"
                      />
                      <div className="flex flex-col gap-1">
                        <Label className="text-sm text-foreground">必填字段</Label>
                        <p className="text-xs leading-normal text-muted-foreground">表单提交时此字段必须填写</p>
                      </div>
                    </div>

                    <div className="flex flex-row gap-4">
                      <Switch
                        checked={fieldData.unique}
                        onCheckedChange={(checked) => setFieldData(prev => ({ ...prev, unique: checked }))}
                        className="data-[state=checked]:bg-primary"
                      />
                      <div className="flex flex-col gap-1">
                        <Label className="text-sm text-foreground">唯一约束</Label>
                        <p className="text-xs leading-normal text-muted-foreground">强制该列的值在所有行中唯一</p>
                      </div>
                    </div>

                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Footer */}
          <DrawerFooter className="border-t border-border px-6 py-4">
            {submitError && (
              <p className="mb-2 rounded-md bg-destructive/10 px-3 py-2 text-xs text-destructive">
                {submitError}
              </p>
            )}
            <div className="flex gap-3">
              <Button
                type="button"
                variant="outline"
                onClick={() => handleOpenChange(false)}
                className="flex-1"
              >
                取消
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={() => handleSave(true)}
                disabled={isSubmitDisabled}
                className="flex-1"
              >
                {saving && continueMode ? (
                  <>
                    <Loader2 className="mr-2 size-4 animate-spin" strokeWidth={1.5} />
                    保存中...
                  </>
                ) : (
                  '保存并继续'
                )}
              </Button>
              <Button
                type="button"
                onClick={() => handleSave(false)}
                disabled={isSubmitDisabled}
                className="flex-1"
              >
                {saving && !continueMode ? (
                  <>
                    <Loader2 className="mr-2 size-4 animate-spin" strokeWidth={1.5} />
                    保存中...
                  </>
                ) : (
                  <>
                    <Save className="mr-2 size-4" strokeWidth={1.5} />
                    保存
                  </>
                )}
              </Button>
            </div>
          </DrawerFooter>
        </div>
      </DrawerContent>
    </Drawer>
  )
}
