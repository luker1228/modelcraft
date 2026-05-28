'use client'

import { useState } from 'react'
import { useParams } from 'next/navigation'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@web/components/ui/dialog'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { Label } from '@web/components/ui/label'
import { Textarea } from '@web/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import { RadioGroup, RadioGroupItem } from '@web/components/ui/radio-group'
import { Loader2 } from 'lucide-react'
import {
  useClusterRawDatabases,
  useRegisterModelDatabase,
  type DatabaseMode,
} from '@web/hooks/model-database/use-model-databases'

interface RegisterDatabaseDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function RegisterDatabaseDialog({ open, onOpenChange }: RegisterDatabaseDialogProps) {
  const params = useParams<{ orgName: string; projectSlug: string }>()
  const { rawDatabases, loading: rawLoading } = useClusterRawDatabases(params.projectSlug, !open)
  const { register, loading: registering } = useRegisterModelDatabase(params.projectSlug)

  const unregistered = rawDatabases.filter((db) => !db.isRegistered)

  const [selectedName, setSelectedName] = useState('')
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [mode, setMode] = useState<DatabaseMode>('SELF_HOSTED')
  const [error, setError] = useState('')

  const handleNameChange = (name: string) => {
    setSelectedName(name)
    if (!title || title === selectedName) {
      setTitle(name)
    }
  }

  const handleSubmit = async () => {
    if (!selectedName || !title) return
    setError('')
    try {
      await register({ name: selectedName, title, description: description || undefined, mode })
      onOpenChange(false)
      setSelectedName('')
      setTitle('')
      setDescription('')
      setMode('SELF_HOSTED')
    } catch (e) {
      setError(e instanceof Error ? e.message : '接管失败')
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>接管数据库</DialogTitle>
        </DialogHeader>
        <div className="flex flex-col gap-4 py-2">
          <div className="flex flex-col gap-2">
            <Label>选择数据库</Label>
            {rawLoading ? (
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Loader2 className="size-4 animate-spin" /> 加载中...
              </div>
            ) : (
              <Select value={selectedName} onValueChange={handleNameChange}>
                <SelectTrigger>
                  <SelectValue placeholder="选择要接管的数据库" />
                </SelectTrigger>
                <SelectContent>
                  {unregistered.length === 0 ? (
                    <div className="px-2 py-3 text-center text-sm text-muted-foreground">
                      所有数据库已接管
                    </div>
                  ) : (
                    unregistered.map((db) => (
                      <SelectItem key={db.name} value={db.name}>
                        {db.name}
                      </SelectItem>
                    ))
                  )}
                </SelectContent>
              </Select>
            )}
          </div>
          <div className="flex flex-col gap-2">
            <Label htmlFor="db-title">友好名称</Label>
            <Input
              id="db-title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="数据库显示名称"
            />
          </div>
          <div className="flex flex-col gap-2">
            <Label htmlFor="db-description">描述（可选）</Label>
            <Textarea
              id="db-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="简要描述此数据库的用途"
              rows={2}
            />
          </div>
          <div className="flex flex-col gap-2">
            <Label>访问模式</Label>
            <RadioGroup value={mode} onValueChange={(v) => setMode(v as DatabaseMode)}>
              <div className="flex items-start gap-3 rounded-md border border-border p-3">
                <RadioGroupItem value="SELF_HOSTED" id="mode-self" className="mt-0.5" />
                <div>
                  <label htmlFor="mode-self" className="cursor-pointer text-sm font-medium">
                    自建
                  </label>
                  <p className="text-xs text-muted-foreground">可读写，支持新建和导入模型</p>
                </div>
              </div>
              <div className="flex items-start gap-3 rounded-md border border-border p-3">
                <RadioGroupItem value="MANAGED" id="mode-managed" className="mt-0.5" />
                <div>
                  <label htmlFor="mode-managed" className="cursor-pointer text-sm font-medium">
                    托管
                  </label>
                  <p className="text-xs text-muted-foreground">只读，仅支持同步模型</p>
                </div>
              </div>
            </RadioGroup>
          </div>
          {error && <p className="text-sm text-destructive">{error}</p>}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={registering}>
            取消
          </Button>
          <Button onClick={handleSubmit} disabled={!selectedName || !title || registering}>
            {registering && <Loader2 className="mr-2 size-4 animate-spin" />}
            确认接管
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
