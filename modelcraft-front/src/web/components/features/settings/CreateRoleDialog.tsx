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
  { group: 'Project', value: 'project:create', label: 'Create Projects' },
  { group: 'Project', value: 'project:read', label: 'View Projects' },
  { group: 'Project', value: 'project:update', label: 'Edit Projects' },
  { group: 'Project', value: 'project:delete', label: 'Delete Projects' },

  { group: 'Model', value: 'model:create', label: 'Create Models' },
  { group: 'Model', value: 'model:read', label: 'View Models' },
  { group: 'Model', value: 'model:update', label: 'Edit Models' },
  { group: 'Model', value: 'model:delete', label: 'Delete Models' },

  { group: 'Cluster', value: 'cluster:create', label: 'Create Clusters' },
  { group: 'Cluster', value: 'cluster:read', label: 'View Clusters' },
  { group: 'Cluster', value: 'cluster:update', label: 'Edit Clusters' },
  { group: 'Cluster', value: 'cluster:delete', label: 'Delete Clusters' },

  { group: 'Enum', value: 'enum:create', label: 'Create Enums' },
  { group: 'Enum', value: 'enum:read', label: 'View Enums' },
  { group: 'Enum', value: 'enum:update', label: 'Edit Enums' },
  { group: 'Enum', value: 'enum:delete', label: 'Delete Enums' },

  { group: 'User Management', value: 'user:invite', label: 'Invite Users' },
  { group: 'User Management', value: 'user:remove', label: 'Remove Users' },
  { group: 'User Management', value: 'user:list', label: 'View Users' },

  { group: 'Organization', value: 'organization:update', label: 'Update Organization' },
]

const createRoleSchema = z.object({
  name: z.string().min(1, 'Role name is required').max(100),
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
          <DialogTitle>Create role</DialogTitle>
          <DialogDescription>
            Define a custom role with specific permissions.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit(onFormSubmit)} className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="name">Role name</Label>
            <Input
              id="name"
              placeholder="e.g., Project Lead"
              {...register('name')}
              disabled={loading}
            />
            {errors.name && (
              <p className="text-xs text-destructive">{errors.name.message}</p>
            )}
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="description">
              Description{' '}
              <span className="text-muted-foreground font-normal">(optional)</span>
            </Label>
            <Textarea
              id="description"
              placeholder="Describe the purpose of this role"
              rows={2}
              {...register('description')}
              disabled={loading}
            />
          </div>

          <div className="space-y-1.5">
            <Label>Permissions</Label>
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
              <p className="text-xs text-destructive">Select at least one permission.</p>
            )}
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => handleOpenChange(false)}
              disabled={loading}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? 'Creating…' : 'Create role'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
