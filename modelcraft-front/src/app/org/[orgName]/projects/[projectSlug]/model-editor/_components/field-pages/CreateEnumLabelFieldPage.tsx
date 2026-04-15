'use client'
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
import type {
  CreateEnumLabelFieldFormValues,
  EnumSourceOption,
  ModelEnumDomainError,
} from '@/types'

const schema = z.object({
  name: z
    .string()
    .trim()
    .min(1, '字段名称不能为空')
    .regex(/^[a-zA-Z_][a-zA-Z0-9_]*$/, '字段名称仅支持字母、数字和下划线，且不能数字开头'),
  title: z.string().trim().min(1, '字段标题不能为空').max(64, '字段标题不能超过 64 个字符'),
  description: z.string().trim().max(500, '描述不能超过 500 个字符').optional(),
  sourceFieldName: z.string().trim().min(1, '请选择 sourceField（本表 ENUM 字段）'),
})

interface CreateEnumLabelFieldPageProps {
  sourceOptions: EnumSourceOption[]
  loading: boolean
  error: ModelEnumDomainError | null
  onSubmit: (values: CreateEnumLabelFieldFormValues) => Promise<void>
  onCancel: () => void
}

export function CreateEnumLabelFieldPage({
  sourceOptions,
  loading,
  error,
  onSubmit,
  onCancel,
}: CreateEnumLabelFieldPageProps) {
  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<CreateEnumLabelFieldFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: '',
      title: '',
      description: '',
      sourceFieldName: '',
    },
  })

  const sourceFieldName = watch('sourceFieldName')
  const selectedSource = sourceOptions.find((source) => source.fieldName === sourceFieldName)

  const submitDisabled =
    loading || isSubmitting || sourceOptions.length === 0 || Boolean(selectedSource?.occupied)

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
        <Label htmlFor="create-enum-label-field-name">字段名称</Label>
        <Input id="create-enum-label-field-name" placeholder="例如 status_label" {...register('name')} aria-invalid={Boolean(errors.name)} />
        {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="create-enum-label-field-title">显示标题</Label>
        <Input id="create-enum-label-field-title" placeholder="例如 状态标签" {...register('title')} aria-invalid={Boolean(errors.title)} />
        {errors.title && <p className="text-xs text-destructive">{errors.title.message}</p>}
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="create-enum-label-field-description">描述</Label>
        <Textarea
          id="create-enum-label-field-description"
          className="min-h-[88px]"
          placeholder="可选描述"
          {...register('description')}
          aria-invalid={Boolean(errors.description)}
        />
        {errors.description && <p className="text-xs text-destructive">{errors.description.message}</p>}
      </div>

      <div className="space-y-1.5">
        <Label>源字段 (ENUM)</Label>
        <Select
          value={sourceFieldName}
          onValueChange={(value) => {
            setValue('sourceFieldName', value, { shouldValidate: true })
          }}
          disabled={loading || sourceOptions.length === 0}
        >
          <SelectTrigger aria-invalid={Boolean(errors.sourceFieldName)}>
            <SelectValue placeholder="请选择 sourceField（本表 ENUM 字段）" />
          </SelectTrigger>
          <SelectContent>
            {sourceOptions.length > 0 ? (
              sourceOptions.map((source) => (
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
        {errors.sourceFieldName && <p className="text-xs text-destructive">{errors.sourceFieldName.message}</p>}
      </div>

      <p className="-mt-1 text-xs text-muted-foreground">保存时会自动创建并绑定 relation。</p>

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
