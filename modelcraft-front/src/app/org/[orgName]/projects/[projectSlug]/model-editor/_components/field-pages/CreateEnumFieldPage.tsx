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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import { Textarea } from '@web/components/ui/textarea'
import type { CreateEnumFieldFormValues, ModelEnumDomainError } from '@/types'

const schema = z.object({
  name: z
    .string()
    .trim()
    .min(1, '字段名称不能为空')
    .regex(/^[a-zA-Z_][a-zA-Z0-9_]*$/, '字段名称仅支持字母、数字和下划线，且不能数字开头'),
  title: z.string().trim().min(1, '字段标题不能为空').max(64, '字段标题不能超过 64 个字符'),
  description: z.string().trim().max(500, '描述不能超过 500 个字符').optional(),
  relateEnumName: z.string().trim().min(1, '请选择关联枚举'),
})

interface CreateEnumFieldPageProps {
  enumOptions: string[]
  loading: boolean
  error: ModelEnumDomainError | null
  onSubmit: (values: CreateEnumFieldFormValues) => Promise<void>
  onCancel: () => void
}

export function CreateEnumFieldPage({ enumOptions, loading, error, onSubmit, onCancel }: CreateEnumFieldPageProps) {
  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<CreateEnumFieldFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: '',
      title: '',
      description: '',
      relateEnumName: '',
    },
  })

  const relateEnumName = watch('relateEnumName')
  const submitDisabled = loading || isSubmitting || enumOptions.length === 0

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
        <Label htmlFor="create-enum-field-name">字段名称</Label>
        <Input id="create-enum-field-name" placeholder="例如 status_text" {...register('name')} aria-invalid={Boolean(errors.name)} />
        {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="create-enum-field-title">显示标题</Label>
        <Input id="create-enum-field-title" placeholder="例如 状态文案" {...register('title')} aria-invalid={Boolean(errors.title)} />
        {errors.title && <p className="text-xs text-destructive">{errors.title.message}</p>}
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="create-enum-field-description">描述</Label>
        <Textarea
          id="create-enum-field-description"
          placeholder="可选描述"
          className="min-h-[88px]"
          {...register('description')}
          aria-invalid={Boolean(errors.description)}
        />
        {errors.description && <p className="text-xs text-destructive">{errors.description.message}</p>}
      </div>

      <div className="space-y-1.5">
        <Label>关联枚举</Label>
        <Select value={relateEnumName} onValueChange={(value) => setValue('relateEnumName', value, { shouldValidate: true })}>
          <SelectTrigger aria-invalid={Boolean(errors.relateEnumName)}>
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
        {errors.relateEnumName && <p className="text-xs text-destructive">{errors.relateEnumName.message}</p>}
        {enumOptions.length === 0 && <p className="text-xs text-muted-foreground">请先创建枚举后再创建 ENUM 字段。</p>}
      </div>

      <div className="mt-2 flex items-center justify-end gap-2 border-t border-border pt-4">
        <Button type="button" variant="outline" size="sm" onClick={onCancel}>
          取消
        </Button>
        <Button type="submit" size="sm" disabled={submitDisabled}>
          {(loading || isSubmitting) && <Loader2 className="mr-1.5 size-3.5 animate-spin" />}
          保存
        </Button>
      </div>
    </form>
  )
}
