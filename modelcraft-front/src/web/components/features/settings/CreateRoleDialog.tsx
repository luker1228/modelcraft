'use client'

import { useState } from 'react'
import { useForm, Controller } from 'react-hook-form'
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
  // Project Permissions
  { group: 'Project', value: 'project:create', label: 'Create Projects' },
  { group: 'Project', value: 'project:read', label: 'View Projects' },
  { group: 'Project', value: 'project:update', label: 'Edit Projects' },
  { group: 'Project', value: 'project:delete', label: 'Delete Projects' },

  // Model Permissions
  { group: 'Model', value: 'model:create', label: 'Create Models' },
  { group: 'Model', value: 'model:read', label: 'View Models' },
  { group: 'Model', value: 'model:update', label: 'Edit Models' },
  { group: 'Model', value: 'model:delete', label: 'Delete Models' },

  // Cluster Permissions
  { group: 'Cluster', value: 'cluster:create', label: 'Create Clusters' },
  { group: 'Cluster', value: 'cluster:read', label: 'View Clusters' },
  { group: 'Cluster', value: 'cluster:update', label: 'Edit Clusters' },
  { group: 'Cluster', value: 'cluster:delete', label: 'Delete Clusters' },

  // Enum Permissions
  { group: 'Enum', value: 'enum:create', label: 'Create Enums' },
  { group: 'Enum', value: 'enum:read', label: 'View Enums' },
  { group: 'Enum', value: 'enum:update', label: 'Edit Enums' },
  { group: 'Enum', value: 'enum:delete', label: 'Delete Enums' },

  // User/Team Permissions
  { group: 'User Management', value: 'user:invite', label: 'Invite Users' },
  { group: 'User Management', value: 'user:remove', label: 'Remove Users' },
  { group: 'User Management', value: 'user:list', label: 'View Users' },

  // Organization Permissions
  { group: 'Organization', value: 'organization:update', label: 'Update Organization' },
]

const createRoleSchema = z.object({
  name: z.string().min(1, 'Role name is required').max(100),
  description: z.string().max(500).optional(),
  permissions: z.array(z.string()).min(1, 'At least one permission must be selected'),
})

type CreateRoleInput = z.infer<typeof createRoleSchema>

interface CreateRoleDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (input: CreateRoleInput) => void
  loading?: boolean
}

export function CreateRoleDialog({
  open,
  onOpenChange,
  onSubmit,
  loading = false,
}: CreateRoleDialogProps) {
  const [selectedPermissions, setSelectedPermissions] = useState<Set<string>>(
    new Set()
  )

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateRoleInput>({
    resolver: zodResolver(createRoleSchema),
    defaultValues: {
      name: '',
      description: '',
      permissions: [],
    },
  })

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) {
      reset()
      setSelectedPermissions(new Set())
    }
    onOpenChange(newOpen)
  }

  const onFormSubmit = (data: CreateRoleInput) => {
    onSubmit({
      ...data,
      permissions: Array.from(selectedPermissions),
    })
    reset()
    setSelectedPermissions(new Set())
  }

  const togglePermission = (permission: string) => {
    const newSet = new Set(selectedPermissions)
    if (newSet.has(permission)) {
      newSet.delete(permission)
    } else {
      newSet.add(permission)
    }
    setSelectedPermissions(newSet)
  }

  const groupedPermissions = AVAILABLE_PERMISSIONS.reduce(
    (acc, perm) => {
      if (!acc[perm.group]) {
        acc[perm.group] = []
      }
      acc[perm.group].push(perm)
      return acc
    },
    {} as Record<string, typeof AVAILABLE_PERMISSIONS>
  )

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Create Role</DialogTitle>
          <DialogDescription>
            Create a new custom role with specific permissions.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit(onFormSubmit)} className="space-y-4">
          {/* Role Name */}
          <div className="space-y-2">
            <Label htmlFor="name">Role Name</Label>
            <Input
              id="name"
              placeholder="e.g., Project Lead"
              {...register('name')}
              disabled={loading}
            />
            {errors.name && (
              <p className="text-sm text-destructive">{errors.name.message}</p>
            )}
          </div>

          {/* Description */}
          <div className="space-y-2">
            <Label htmlFor="description">Description (Optional)</Label>
            <Textarea
              id="description"
              placeholder="Describe the purpose of this role"
              rows={2}
              {...register('description')}
              disabled={loading}
            />
            {errors.description && (
              <p className="text-sm text-destructive">
                {errors.description.message}
              </p>
            )}
          </div>

          {/* Permissions */}
          <div className="space-y-2">
            <Label>Permissions</Label>
            <ScrollArea className="h-48 rounded-md border p-3">
              <div className="space-y-3">
                {Object.entries(groupedPermissions).map(([group, perms]) => (
                  <div key={group}>
                    <p className="mb-2 text-xs font-semibold uppercase text-muted-foreground">
                      {group}
                    </p>
                    <div className="space-y-2 pl-2">
                      {perms.map((perm) => (
                        <div
                          key={perm.value}
                          className="flex items-center space-x-2"
                        >
                          <Checkbox
                            id={perm.value}
                            checked={selectedPermissions.has(perm.value)}
                            onCheckedChange={() =>
                              togglePermission(perm.value)
                            }
                            disabled={loading}
                          />
                          <Label
                            htmlFor={perm.value}
                            className="cursor-pointer font-normal"
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
            {selectedPermissions.size === 0 && (
              <p className="text-sm text-destructive">
                At least one permission must be selected
              </p>
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
              {loading ? 'Creating...' : 'Create Role'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
