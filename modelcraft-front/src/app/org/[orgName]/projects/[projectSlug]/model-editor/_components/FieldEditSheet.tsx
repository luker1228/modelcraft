'use client'

import { Key, Settings } from 'lucide-react'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { Textarea } from '@web/components/ui/textarea'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@web/components/ui/sheet'
import type { ModelEditorState } from '../_hooks'
import type { FieldOperations } from '../_hooks'

interface FieldEditSheetProps {
  state: ModelEditorState
  fieldOps: FieldOperations
}

export function FieldEditSheet({ state, fieldOps }: FieldEditSheetProps) {
  return (
    <Sheet open={state.editFieldOpen} onOpenChange={state.setEditFieldOpen}>
      <SheetContent className="w-[480px] overflow-y-auto sm:max-w-[480px]">
        <SheetHeader>
          <SheetTitle className="flex items-center gap-2 font-heading text-base">
            <Settings className="size-4 text-[#2563eb]" />
            编辑字段
          </SheetTitle>
          <SheetDescription className="text-sm">
            编辑字段 <span className="font-mono text-[#2563eb]">{state.editingField?.name}</span> 的配置
          </SheetDescription>
        </SheetHeader>

        {state.editingField && (
          <div className="space-y-4 py-4">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-muted-foreground">字段名称</label>
              <Input
                value={state.editingField.name}
                disabled
                className="bg-muted/30 font-mono text-sm"
              />
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-medium text-foreground">显示名称 (DisplayName)</label>
              <Input
                value={state.editFieldTitle}
                onChange={(e) => state.setEditFieldTitle(e.target.value)}
                className="text-sm"
                placeholder="字段显示名称"
              />
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-medium text-muted-foreground">Format</label>
              <Input
                value={state.editingField.format || '-'}
                disabled
                className="bg-muted/30 font-mono text-sm"
              />
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-medium text-muted-foreground">Type</label>
              <Input
                value={state.editingField.storageHint || state.editingField.schemaType || 'String'}
                disabled
                className="bg-muted/30 font-mono text-sm"
              />
            </div>

            <div className="space-y-2">
              <label className="text-xs font-medium text-muted-foreground">属性</label>
              <div className="flex flex-wrap gap-2">
                {state.editingField.isPrimary && (
                  <span className="inline-flex items-center rounded px-2 py-1 text-xs" style={{ backgroundColor: '#fef3c7', color: '#d97706' }}>
                    <Key className="mr-1 size-3" />
                    主键
                  </span>
                )}
                {state.editingField.nonNull && (
                  <span className="inline-flex items-center rounded px-2 py-1 text-xs" style={{ backgroundColor: '#fee2e2', color: '#ef4444' }}>
                    必填
                  </span>
                )}
                {state.editingField.required && (
                  <span className="inline-flex items-center rounded px-2 py-1 text-xs" style={{ backgroundColor: '#fef3c7', color: '#d97706' }}>
                    必需
                  </span>
                )}
                {!state.editingField.isPrimary && !state.editingField.nonNull && !state.editingField.required && (
                  <span className="text-xs text-muted-foreground">无特殊属性</span>
                )}
              </div>
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-medium text-foreground">描述</label>
              <Textarea
                value={state.editFieldDescription}
                onChange={(e) => state.setEditFieldDescription(e.target.value)}
                className="min-h-[80px] resize-none text-sm"
                placeholder="字段描述"
              />
            </div>
          </div>
        )}

        <SheetFooter className="mt-4 flex gap-2 border-t border-border pt-4 sm:justify-end">
          <Button
            variant="outline"
            size="sm"
            onClick={() => state.setEditFieldOpen(false)}
          >
            取消
          </Button>
          <Button
            size="sm"
            className="border-0 bg-[#2563eb] text-white transition-colors duration-200 hover:bg-[#1d4ed8]"
            onClick={fieldOps.handleSaveField}
          >
            保存
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
