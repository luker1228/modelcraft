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
        <p className="text-sm text-muted-foreground">暂无角色。</p>
        {onDelete !== undefined && (
          <p className="mt-1 text-xs text-muted-foreground/70">
            新建角色以为成员分配权限。
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
              <TableHead>名称</TableHead>
              <TableHead>描述</TableHead>
              <TableHead>权限</TableHead>
              <TableHead>类型</TableHead>
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
                    <Badge variant="secondary" className="text-xs">系统</Badge>
                  ) : (
                    <Badge variant="outline" className="text-xs text-muted-foreground">自定义</Badge>
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
                        aria-label={`删除角色 ${role.name}`}
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
          <AlertDialogTitle>删除角色</AlertDialogTitle>
          <AlertDialogDescription>
            确认删除角色「{roleName}」？该角色下的成员将失去对应权限，此操作不可撤销。
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>取消</AlertDialogCancel>
          <AlertDialogAction
            onClick={onConfirm}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          >
            删除
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
