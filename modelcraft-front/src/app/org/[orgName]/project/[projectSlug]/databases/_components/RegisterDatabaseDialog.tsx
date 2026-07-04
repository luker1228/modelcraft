'use client'

import { useState } from 'react'
import { useParams } from 'next/navigation'
import { generateUUID } from '@/shared/utils/uuid'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@web/components/ui/dialog'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import { Loader2, Plus, Trash2 } from 'lucide-react'
import {
  useClusterRawDatabases,
  useBatchRegisterModelDatabase,
  type ModelDatabase,
  type RegisterModelDatabaseInput,
} from '@web/hooks/model-database/use-model-databases'

interface DbRow {
  id: string
  name: string
  title: string
  description: string
  error?: string
}

interface RegisterDatabaseDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onRegistered?: (databases: ModelDatabase[]) => void
}

interface DbRowFormProps {
  row: DbRow
  options: { name: string }[]
  onUpdate: (patch: Partial<DbRow>) => void
  onRemove: () => void
}

function DbRowForm({ row, options, onUpdate, onRemove }: DbRowFormProps) {
  return (
    <div className="flex flex-col gap-2 rounded-md border border-border p-3">
      <div className="flex items-start gap-2">
        <Select
          value={row.name}
          onValueChange={(name) => {
            onUpdate({ name, title: row.title || name })
          }}
        >
          <SelectTrigger className="w-44 shrink-0">
            <SelectValue placeholder="选择数据库" />
          </SelectTrigger>
          <SelectContent>
            {options.map((db) => (
              <SelectItem key={db.name} value={db.name}>
                {db.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Input
          value={row.title}
          onChange={(e) => onUpdate({ title: e.target.value })}
          placeholder="友好名称"
          className="flex-1"
        />
        <Input
          value={row.description}
          onChange={(e) => onUpdate({ description: e.target.value })}
          placeholder="描述（可选）"
          className="flex-1"
        />
        <Button
          variant="ghost"
          size="icon"
          onClick={onRemove}
          className="shrink-0 text-muted-foreground hover:text-destructive"
        >
          <Trash2 className="size-4" />
        </Button>
      </div>
      {row.error && <p className="text-sm text-destructive">{row.error}</p>}
    </div>
  )
}

export function RegisterDatabaseDialog({
  open,
  onOpenChange,
  onRegistered,
}: RegisterDatabaseDialogProps) {
  const params = useParams<{ orgName: string; projectSlug: string }>()
  const { rawDatabases, loading: rawLoading } = useClusterRawDatabases(params.projectSlug, !open)
  const { batchRegister, loading: submitting } = useBatchRegisterModelDatabase(params.projectSlug)

  const unregistered = rawDatabases.filter((db) => !db.isRegistered)

  const [rows, setRows] = useState<DbRow[]>([])

  const takenNames = new Set(rows.map((r) => r.name).filter(Boolean))
  const availableFor = (row: DbRow) =>
    unregistered.filter((db) => db.name === row.name || !takenNames.has(db.name))

  const addRow = () => {
    setRows((prev) => [
      ...prev,
      { id: generateUUID(), name: '', title: '', description: '' },
    ])
  }

  const removeRow = (id: string) => setRows((prev) => prev.filter((r) => r.id !== id))

  const updateRow = (id: string, patch: Partial<DbRow>) =>
    setRows((prev) => prev.map((r) => (r.id === id ? { ...r, ...patch } : r)))

  const handleOpenChange = (v: boolean) => {
    if (!v) setRows([])
    onOpenChange(v)
  }

  const canSubmit = rows.length > 0 && rows.every((r) => r.name && r.title) && !submitting
  const allTaken =
    unregistered.length > 0 && takenNames.size >= unregistered.length

  const handleSubmit = async () => {
    setRows((prev) => prev.map((r) => ({ ...r, error: undefined })))
    const inputs: RegisterModelDatabaseInput[] = rows.map((r) => ({
      name: r.name,
      title: r.title,
      description: r.description || undefined,
      mode: 'SELF_HOSTED',
    }))
    const result = await batchRegister(inputs)
    const successNames = new Set(result.succeeded.map((d) => d.name))
    const failMap = new Map(result.failed.map((f) => [f.name, f.message]))
    if (result.succeeded.length > 0) {
      onRegistered?.(result.succeeded)
    }
    setRows((prev) => {
      const remaining = prev
        .filter((r) => !successNames.has(r.name))
        .map((r) => ({ ...r, error: failMap.get(r.name) }))
      return remaining
    })
    if (result.failed.length === 0) {
      handleOpenChange(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-3xl">
        <DialogHeader>
          <DialogTitle>批量注册数据库</DialogTitle>
        </DialogHeader>
        <div className="flex flex-col gap-3 py-2">
          <p className="text-sm text-muted-foreground">
            数据库统一按可写模式注册，支持导入和后续模型维护。
          </p>
          {rawLoading ? (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <Loader2 className="size-4 animate-spin" /> 加载数据库列表...
            </div>
          ) : (
            <>
              {rows.length === 0 && (
                <p className="py-4 text-center text-sm text-muted-foreground">
                  点击下方按钮添加要注册的数据库
                </p>
              )}
              <div className="flex max-h-96 flex-col gap-2 overflow-y-auto">
                {rows.map((row) => (
                  <DbRowForm
                    key={row.id}
                    row={row}
                    options={availableFor(row)}
                    onUpdate={(patch) => updateRow(row.id, patch)}
                    onRemove={() => removeRow(row.id)}
                  />
                ))}
              </div>
              {unregistered.length === 0 ? (
                <p className="text-center text-sm text-muted-foreground">所有数据库已注册</p>
              ) : (
                <Button
                  variant="outline"
                  onClick={addRow}
                  disabled={allTaken}
                  className="gap-1.5"
                >
                  <Plus className="size-4" />
                  添加数据库
                </Button>
              )}
            </>
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => handleOpenChange(false)} disabled={submitting}>
            取消
          </Button>
          <Button onClick={handleSubmit} disabled={!canSubmit}>
            {submitting && <Loader2 className="mr-2 size-4 animate-spin" />}
            确认注册
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
