'use client'

import React from 'react'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { Loader2 } from 'lucide-react'
import { z } from 'zod'
import { Alert, AlertDescription } from '@web/components/ui/alert'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { Label } from '@web/components/ui/label'
import { Textarea } from '@web/components/ui/textarea'
import { hasSystemLabelSuffix, isSystemGeneratedLabelField } from '@/shared/model/system-field'
import type { ModelEnumDomainError, UpdateFieldMetaFormValues } from '@/types'

const schema = z.object({
  title: z.string().trim().min(1, '字段标题不能为空').max(64, '字段标题不能超过 64 个字符'),
  description: z.string().trim().max(500, '描述不能超过 500 个字符').optional(),
})

interface EditFieldImmutablePageProps {
  fieldName: string
  format: string
  title?: string
  description?: string
  relateEnumName?: string
  enumRelationId?: string
  loading: boolean
  error: ModelEnumDomainError | null
  onSubmit: (values: UpdateFieldMetaFormValues) => Promise<void>
  onCancel: () => void
}

export function EditFieldImmutablePage({
  fieldName,
  format,
  title,
  description,
  relateEnumName,
  enumRelationId,
  loading,
  error,
  onSubmit,
  onCancel,
}: EditFieldImmutablePageProps) {
  const isSystemField = isSystemGeneratedLabelField(
    { name: fieldName, format },
    [{ name: fieldName, format }],
  ) || hasSystemLabelSuffix(fieldName)

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<UpdateFieldMetaFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      title: title ?? '',
      description: description ?? '',
    },
  })

  React.useEffect(() => {
    reset({
      title: title ?? '',
      description: description ?? '',
    })
  }, [description, reset, title])

  const handleFormSubmit = handleSubmit(async (values) => {
    await onSubmit(values)
  })

  return (
    <form onSubmit={handleFormSubmit} className="flex flex-col gap-4 py-4">
      {error && (
        <Alert variant="destructive">
          <AlertDescription>
            <p>{error.message}</p>
            {error.code && <p className="mt-1 font-mono text-xs">错误码: {error.code}</p>}
          </AlertDescription>
        </Alert>
      )}

      <div className="space-y-1.5">
        <Label htmlFor="edit-field-name">字段名称</Label>
        <Input id="edit-field-name" value={fieldName} disabled className="bg-muted/40 font-mono" />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="edit-field-title">显示标题</Label>
        <Input
          id="edit-field-title"
          placeholder="输入字段标题"
          {...register('title')}
          disabled={isSystemField}
          aria-invalid={Boolean(errors.title)}
        />
        {errors.title && <p className="text-xs text-destructive">{errors.title.message}</p>}
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="edit-field-description">描述</Label>
        <Textarea
          id="edit-field-description"
          className="min-h-[96px]"
          placeholder="输入字段描述"
          {...register('description')}
          disabled={isSystemField}
          aria-invalid={Boolean(errors.description)}
        />
        {errors.description && <p className="text-xs text-destructive">{errors.description.message}</p>}
      </div>

      <div className="rounded-md border border-border bg-muted/30 p-3">
        <p className="mb-2 text-xs text-muted-foreground">格式信息（只读）</p>
        <div className="grid grid-cols-1 gap-3">
          <div className="space-y-1">
            <Label className="text-xs text-muted-foreground">字段类型</Label>
            <Input value={format || '-'} disabled className="bg-background font-mono text-sm" />
          </div>

          {format === 'ENUM' && (
            <div className="space-y-1">
              <Label className="text-xs text-muted-foreground">关联枚举</Label>
              <Input value={relateEnumName || '-'} disabled className="bg-background font-mono text-sm" />
            </div>
          )}

          {format === 'ENUM_LABEL' && (
            <div className="space-y-1">
              <Label className="text-xs text-muted-foreground">关联 relationId</Label>
              <Input value={enumRelationId || '-'} disabled className="bg-background font-mono text-sm" />
            </div>
          )}
        </div>
        <p className="mt-2 text-xs text-muted-foreground">字段 format/关联配置创建后不可修改。</p>
        {isSystemField && (
          <p className="mt-1 text-xs text-muted-foreground">系统生成字段只读，不可编辑。</p>
        )}
      </div>

      <div className="mt-2 flex items-center justify-end gap-2 border-t border-border pt-4">
        <Button type="button" variant="outline" size="sm" onClick={onCancel}>
          取消
        </Button>
        <Button type="submit" size="sm" disabled={loading || isSubmitting || isSystemField}>
          {(loading || isSubmitting) && <Loader2 className="mr-1.5 size-3.5 animate-spin" />}
          保存
        </Button>
      </div>
    </form>
  )
}
