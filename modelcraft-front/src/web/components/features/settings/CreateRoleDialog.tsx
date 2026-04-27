'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@web/components/ui/dialog'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { Textarea } from '@web/components/ui/textarea'
import { Checkbox } from '@web/components/ui/checkbox'
import { Label } from '@web/components/ui/label'
import { ScrollArea } from '@web/components/ui/scroll-area'

const AVAILABLE_PERMISSIONS = [
  { group: '项目', value: 'project:create', label: '创建项目' },
  { group: '项目', value: 'project:read', label: '查看项目' },
  { group: '项目', value: 'project:update', label: '编辑项目' },
  { group: '项目', value: 'project:delete', label: '删除项目' },

  { group: '模型', value: 'model:create', label: '创建模型' },
  { group: '模型', value: 'model:read', label: '查看模型' },
  { group: '模型', value: 'model:update', label: '编辑模型' },
  { group: '模型', value: 'model:delete', label: '删除模型' },

  { group: '集群', value: 'cluster:create', label: '创建集群' },
  { group: '集群', value: 'cluster:read', label: '查看集群' },
  { group: '集群', value: 'cluster:update', label: '编辑集群' },
  { group: '集群', value: 'cluster:delete', label: '删除集群' },

  { group: '枚举', value: 'enum:create', label: '创建枚举' },
  { group: '枚举', value: 'enum:read', label: '查看枚举' },
  { group: '枚举', value: 'enum:update', label: '编辑枚举' },
  { group: '枚举', value: 'enum:delete', label: '删除枚举' },

  { group: '用户管理', value: 'user:invite', label: '邀请成员' },
  { group: '用户管理', value: 'user:remove', label: '移除成员' },
  { group: '用户管理', value: 'user:list', label: '查看成员' },

  { group: '组织', value: 'organization:update', label: '编辑组织信息' },
]

const createRoleSchema = z.object({
  name: z.string().min(1, '角色名称不能为空').max(100),
  description: z.string().max(500).optional(),
})

type CreateRoleFormValues = z.infer<typeof createRoleSchema>

interface CreateRoleDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (input: { name: string; description?: string; permissions: string[] }) => void
  loading?: boolean
}

const groupedPermissions = AVAILABLE_PERMISSIONS.reduce(
  (acc, perm) => {
    if (!acc[perm.group]) acc[perm.group] = []
    acc[perm.group].push(perm)
    return acc
  },
  {} as Record<string, typeof AVAILABLE_PERMISSIONS>
)

export function CreateRoleDialog({
  open,
  onOpenChange,
  onSubmit,
  loading = false,
}: CreateRoleDialogProps) {
  const [selectedPermissions, setSelectedPermissions] = useState<Set<string>>(new Set())
  const [submitAttempted, setSubmitAttempted] = useState(false)

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateRoleFormValues>({
    resolver: zodResolver(createRoleSchema),
    defaultValues: { name: '', description: '' },
  })

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) {
      reset()
      setSelectedPermissions(new Set())
      setSubmitAttempted(false)
    }
    onOpenChange(newOpen)
  }

  const onFormSubmit = (data: CreateRoleFormValues) => {
    setSubmitAttempted(true)
    if (selectedPermissions.size === 0) return
    onSubmit({ ...data, permissions: Array.from(selectedPermissions) })
    reset()
    setSelectedPermissions(new Set())
    setSubmitAttempted(false)
  }

  const togglePermission = (permission: string) => {
    const next = new Set(selectedPermissions)
    if (next.has(permission)) {
      next.delete(permission)
    } else {
      next.add(permission)
    }
    setSelectedPermissions(next)
  }

  const permissionsError = submitAttempted && selectedPermissions.size === 0

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>新建角色</DialogTitle>
          <DialogDescription>
            创建自定义角色并指定对应权限。
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit(onFormSubmit)} className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="name">角色名称</Label>
            <Input
              id="name"
              placeholder="例如：项目负责人"
              {...register('name')}
              disabled={loading}
            />
            {errors.name && (
              <p className="text-xs text-destructive">{errors.name.message}</p>
            )}
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="description">
              描述{' '}
              <span className="font-normal text-muted-foreground">（可选）</span>
            </Label>
            <Textarea
              id="description"
              placeholder="说明该角色的用途"
              rows={2}
              {...register('description')}
              disabled={loading}
            />
          </div>

          <div className="space-y-1.5">
            <Label>权限</Label>
            <ScrollArea className="h-52 rounded-md border border-border p-3">
              <div className="space-y-4">
                {Object.entries(groupedPermissions).map(([group, perms]) => (
                  <div key={group}>
                    <p className="mb-1.5 text-xs font-medium uppercase tracking-wide text-muted-foreground">
                      {group}
                    </p>
                    <div className="space-y-1.5 pl-1">
                      {perms.map((perm) => (
                        <div key={perm.value} className="flex items-center gap-2">
                          <Checkbox
                            id={perm.value}
                            checked={selectedPermissions.has(perm.value)}
                            onCheckedChange={() => togglePermission(perm.value)}
                            disabled={loading}
                          />
                          <Label
                            htmlFor={perm.value}
                            className="cursor-pointer text-sm font-normal"
                          >
                            {perm.label}
                          </Label>
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            </ScrollArea>
            {permissionsError && (
              <p className="text-xs text-destructive">请至少选择一项权限。</p>
            )}
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => handleOpenChange(false)}
              disabled={loading}
            >
              取消
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? '创建中…' : '新建角色'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
