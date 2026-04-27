'use client'

import { useState } from 'react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@web/components/ui/table'
import { Badge } from '@web/components/ui/badge'
import { Button } from '@web/components/ui/button'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@web/components/ui/alert-dialog'
import type { Role } from '@/types'
import { Shield, Trash2 } from 'lucide-react'

interface RoleTableProps {
  roles: Role[]
  onDelete?: (roleId: string) => void
}

export function RoleTable({ roles, onDelete }: RoleTableProps) {
  const [deleteTarget, setDeleteTarget] = useState<Role | null>(null)

  if (roles.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <Shield className="mb-3 size-8 text-muted-foreground/40" strokeWidth={1.5} />
        <p className="text-sm text-muted-foreground">No roles defined yet.</p>
        {onDelete !== undefined && (
          <p className="mt-1 text-xs text-muted-foreground/70">
            Create a role to assign permissions to team members.
          </p>
        )}
      </div>
    )
  }

  const handleConfirmDelete = () => {
    if (deleteTarget && onDelete) {
      onDelete(deleteTarget.id)
    }
    setDeleteTarget(null)
  }

  return (
    <>
      <div className="rounded-md border border-border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Permissions</TableHead>
              <TableHead>Type</TableHead>
              {onDelete && <TableHead className="w-16" />}
            </TableRow>
          </TableHeader>
          <TableBody>
            {roles.map((role) => (
              <TableRow key={role.id}>
                <TableCell className="font-medium text-foreground">
                  {role.name}
                </TableCell>
                <TableCell className="max-w-[200px] truncate text-sm text-muted-foreground">
                  {role.description || <span className="text-muted-foreground/50">—</span>}
                </TableCell>
                <TableCell>
                  <div className="flex flex-wrap gap-1">
                    {role.permissions.slice(0, 4).map((perm) => (
                      <Badge key={perm} variant="secondary" className="font-mono text-xs">
                        {perm}
                      </Badge>
                    ))}
                    {role.permissions.length > 4 && (
                      <Badge variant="outline" className="text-xs text-muted-foreground">
                        +{role.permissions.length - 4}
                      </Badge>
                    )}
                  </div>
                </TableCell>
                <TableCell>
                  {role.isSystem ? (
                    <Badge variant="secondary" className="text-xs">System</Badge>
                  ) : (
                    <Badge variant="outline" className="text-xs text-muted-foreground">Custom</Badge>
                  )}
                </TableCell>
                {onDelete && (
                  <TableCell className="text-right">
                    {!role.isSystem && (
                      <Button
                        variant="ghost"
                        size="sm"
                        className="size-7 p-0 text-muted-foreground/50 hover:text-destructive"
                        onClick={() => setDeleteTarget(role)}
                        aria-label={`Delete role ${role.name}`}
                      >
                        <Trash2 className="size-3.5" />
                      </Button>
                    )}
                  </TableCell>
                )}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      {deleteTarget && (
        <DeleteConfirmDialog
          roleName={deleteTarget.name}
          open={!!deleteTarget}
          onOpenChange={(open) => !open && setDeleteTarget(null)}
          onConfirm={handleConfirmDelete}
        />
      )}
    </>
  )
}

function DeleteConfirmDialog({
  roleName,
  open,
  onOpenChange,
  onConfirm,
}: {
  roleName: string
  open: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: () => void
}) {
  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete role</AlertDialogTitle>
          <AlertDialogDescription>
            Delete &quot;{roleName}&quot;? Users with this role will lose their assigned
            permissions. This cannot be undone.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={onConfirm}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          >
            Delete
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
