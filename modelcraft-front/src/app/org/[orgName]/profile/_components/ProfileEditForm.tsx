'use client'

import { zodResolver } from '@hookform/resolvers/zod'
import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { Label } from '@web/components/ui/label'
import { Textarea } from '@web/components/ui/textarea'
import { Loader2 } from 'lucide-react'
import type { ProfileDomainError, UpdateMyProfileFormValues } from '@/types'

const profileEditSchema = z.object({
  nickname: z.string().trim().min(1, '昵称不能为空').max(64, '昵称最多 64 个字符'),
  avatarUrl: z.string().max(512, '头像地址最多 512 个字符').optional().or(z.literal('')),
  bio: z.string().max(280, '个人简介最多 280 个字符').optional().or(z.literal('')),
})

type ProfileEditFormValues = z.infer<typeof profileEditSchema>

export interface ProfileEditFormProps {
  initialValues: UpdateMyProfileFormValues
  saving: boolean
  submitError?: ProfileDomainError | null
  onSubmit: (values: UpdateMyProfileFormValues) => Promise<void>
  onCancel: () => void
}

export function ProfileEditForm({
  initialValues,
  saving,
  submitError,
  onSubmit,
  onCancel,
}: ProfileEditFormProps) {
  const {
    register,
    reset,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ProfileEditFormValues>({
    resolver: zodResolver(profileEditSchema),
    defaultValues: {
      nickname: initialValues.nickname ?? '',
      avatarUrl: initialValues.avatarUrl ?? '',
      bio: initialValues.bio ?? '',
    },
  })

  useEffect(() => {
    reset({
      nickname: initialValues.nickname ?? '',
      avatarUrl: initialValues.avatarUrl ?? '',
      bio: initialValues.bio ?? '',
    })
  }, [initialValues.avatarUrl, initialValues.bio, initialValues.nickname, reset])

  const submitting = saving || isSubmitting

  const handleFormSubmit = handleSubmit(async (values) => {
    await onSubmit({
      nickname: values.nickname.trim() || undefined,
      avatarUrl: values.avatarUrl?.trim() || undefined,
      bio: values.bio?.trim() || undefined,
    })
  })

  return (
    <form onSubmit={handleFormSubmit} className="space-y-5">
      <div className="space-y-2">
        <Label htmlFor="nickname" className="font-medium text-foreground">
          昵称
        </Label>
        <Input
          id="nickname"
          placeholder="请输入昵称"
          {...register('nickname')}
          aria-invalid={Boolean(errors.nickname)}
        />
        {errors.nickname && <p className="text-sm text-destructive">{errors.nickname.message}</p>}
      </div>

      <div className="space-y-2">
        <Label htmlFor="avatarUrl" className="font-medium text-foreground">
          头像地址
        </Label>
        <Input
          id="avatarUrl"
          placeholder="例如 /mocks/avatar-user.png"
          {...register('avatarUrl')}
          aria-invalid={Boolean(errors.avatarUrl)}
        />
        {errors.avatarUrl && <p className="text-sm text-destructive">{errors.avatarUrl.message}</p>}
      </div>

      <div className="space-y-2">
        <Label htmlFor="bio" className="font-medium text-foreground">
          个人简介
        </Label>
        <Textarea
          id="bio"
          rows={5}
          placeholder="介绍一下你自己..."
          {...register('bio')}
          aria-invalid={Boolean(errors.bio)}
        />
        {errors.bio && <p className="text-sm text-destructive">{errors.bio.message}</p>}
      </div>

      {submitError && (
        <div className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2">
          <p className="text-sm font-medium text-destructive">{submitError.message}</p>
          {submitError.suggestion && (
            <p className="mt-1 text-sm font-medium text-destructive">建议：{submitError.suggestion}</p>
          )}
        </div>
      )}

      <div className="flex items-center justify-end gap-3">
        <Button type="button" variant="outline" onClick={onCancel} disabled={submitting}>
          取消
        </Button>
        <Button type="submit" disabled={submitting}>
          {submitting && <Loader2 className="mr-2 size-4 animate-spin" />}
          保存资料
        </Button>
      </div>
    </form>
  )
}
