'use client'

import { useState, useEffect } from 'react'
import { useParams } from 'next/navigation'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetFooter,
} from '@web/components/ui/sheet'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { Label } from '@web/components/ui/label'
import { Textarea } from '@web/components/ui/textarea'
import { Loader2 } from 'lucide-react'
import {
  useUpdateModelDatabase,
  type ModelDatabase,
  type DatabaseMode,
} from '@web/hooks/model-database/use-model-databases'

interface EditDatabaseSheetProps {
  database: ModelDatabase | null
  onClose: () => void
}

export function EditDatabaseSheet({ database, onClose }: EditDatabaseSheetProps) {
  const params = useParams<{ orgName: string; projectSlug: string }>()
  const { update, loading } = useUpdateModelDatabase(params.projectSlug)

  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [mode, setMode] = useState<DatabaseMode>('MANAGED')

  useEffect(() => {
    if (database) {
      setTitle(database.title)
      setDescription(database.description)
      setMode(database.mode)
    }
  }, [database])

  const handleSave = async () => {
    if (!database) return
    await update(database.id, { title, description, mode })
    onClose()
  }

  return (
    <Sheet open={!!database} onOpenChange={(open) => !open && onClose()}>
      <SheetContent>
        <SheetHeader>
          <SheetTitle>编辑数据库</SheetTitle>
        </SheetHeader>
        <div className="flex flex-col gap-4 py-4">
          <div className="flex flex-col gap-2">
            <Label className="text-muted-foreground">数据库名</Label>
            <p className="text-sm font-medium">{database?.name}</p>
          </div>
          <div className="flex flex-col gap-2">
            <Label htmlFor="edit-title">友好名称</Label>
            <Input
              id="edit-title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
            />
          </div>
          <div className="flex flex-col gap-2">
            <Label htmlFor="edit-description">描述</Label>
            <Textarea
              id="edit-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
            />
          </div>
          <div className="flex flex-col gap-2">
            <Label>访问模式</Label>
            <div className="flex items-start gap-3 rounded-md border border-border bg-muted/30 p-3">
              <div>
                <p className="text-sm font-medium">托管</p>
                <p className="text-xs text-muted-foreground">由 ModelCraft 托管，自动处理连接与凭据</p>
              </div>
            </div>
          </div>
        </div>
        <SheetFooter>
          <Button variant="outline" onClick={onClose} disabled={loading}>
            取消
          </Button>
          <Button onClick={handleSave} disabled={!title || loading}>
            {loading && <Loader2 className="mr-2 size-4 animate-spin" />}
            保存
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
