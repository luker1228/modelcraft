'use client'

import React from 'react'
import { Link2, Loader2, Plus } from 'lucide-react'
import { Button } from '@web/components/ui/button'
import { Label } from '@web/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import type { EnumRelationOption, EnumSourceOption } from '@/types'

interface EnumRelationSelectorProps {
  sourceFieldName: string
  sourceOptions: EnumSourceOption[]
  relationOptions: EnumRelationOption[]
  value: string
  disabled?: boolean
  onChange: (relationId: string) => void
  onCreateRelation: (sourceFieldName: string) => Promise<string | null>
}

export function EnumRelationSelector({
  sourceFieldName,
  sourceOptions,
  relationOptions,
  value,
  disabled,
  onChange,
  onCreateRelation,
}: EnumRelationSelectorProps) {
  const [creating, setCreating] = React.useState(false)

  const selectedSource = React.useMemo(
    () => sourceOptions.find((source) => source.fieldName === sourceFieldName),
    [sourceFieldName, sourceOptions],
  )

  const sourceRelations = React.useMemo(
    () => relationOptions.filter((relation) => relation.sourceFieldName === sourceFieldName),
    [relationOptions, sourceFieldName],
  )

  const handleCreateRelation = async () => {
    if (!sourceFieldName) {
      return
    }

    setCreating(true)
    try {
      const relationId = await onCreateRelation(sourceFieldName)
      if (relationId) {
        onChange(relationId)
      }
    } finally {
      setCreating(false)
    }
  }

  const relationDisabled = disabled || !sourceFieldName || selectedSource?.occupied

  return (
    <div className="flex flex-col gap-2">
      <Label className="text-xs text-muted-foreground">Relation</Label>
      <div className="flex items-center gap-2">
        <Select value={value} onValueChange={onChange} disabled={relationDisabled}>
          <SelectTrigger>
            <SelectValue placeholder={sourceFieldName ? '选择已有 relation' : '请先选择源字段'} />
          </SelectTrigger>
          <SelectContent>
            {sourceRelations.length > 0 ? (
              sourceRelations.map((relation) => (
                <SelectItem key={relation.id} value={relation.id}>
                  <span className="font-mono text-xs">{relation.id}</span>
                </SelectItem>
              ))
            ) : (
              <SelectItem value="__none__" disabled>
                暂无可用 relation
              </SelectItem>
            )}
          </SelectContent>
        </Select>

        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={handleCreateRelation}
          disabled={disabled || creating || !sourceFieldName || Boolean(selectedSource?.occupied)}
        >
          {creating ? (
            <Loader2 className="mr-1.5 size-3.5 animate-spin" />
          ) : (
            <Plus className="mr-1.5 size-3.5" />
          )}
          新建 relation
        </Button>
      </div>

      {selectedSource?.occupied && (
        <div className="rounded-md border border-destructive/20 bg-destructive/5 px-3 py-2 text-xs text-destructive">
          当前 source 字段已占用，不能重复创建 ENUM_LABEL。
        </div>
      )}

      {!selectedSource?.occupied && sourceFieldName && sourceRelations.length === 0 && (
        <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
          <Link2 className="size-3.5" />
          该 source 暂无 relation，可先新建后再保存。
        </div>
      )}
    </div>
  )
}
