'use client'

import { Loader2 } from 'lucide-react'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@web/components/ui/sheet'
import type { ModelEditorState } from '../_hooks'
import type { ModelCRUD } from '../_hooks'

interface CreateModelDialogProps {
  state: ModelEditorState
  crud: ModelCRUD
}

export function CreateModelDialog({ state, crud }: CreateModelDialogProps) {
  return (
    <Sheet open={state.createModelOpen} onOpenChange={state.setCreateModelOpen}>
      <SheetContent className="w-[400px] sm:max-w-[400px]">
        <SheetHeader>
          <SheetTitle className="font-heading text-base">新建模型</SheetTitle>
          <SheetDescription className="text-sm">
            创建一个新的数据模型，用于定义数据结构。
          </SheetDescription>
        </SheetHeader>
        <div className="space-y-4 py-6">
          <div className="space-y-2">
            <label className="text-sm font-medium">模型标识 <span className="text-destructive">*</span></label>
            <Input
              placeholder="例如：user_profile"
              value={state.newModelName}
              onChange={(e) => state.setNewModelName(e.target.value)}
              className="text-sm"
            />
            <p className="text-xs text-muted-foreground">英文字母、数字、下划线组成，用于代码引用</p>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">模型展示名称 <span className="text-destructive">*</span></label>
            <Input
              placeholder="例如：用户档案"
              value={state.newModelTitle}
              onChange={(e) => state.setNewModelTitle(e.target.value)}
              className="text-sm"
            />
            <p className="text-xs text-muted-foreground">中文显示名称，便于理解</p>
          </div>
        </div>
        <SheetFooter className="flex gap-2 sm:justify-end">
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              state.setCreateModelOpen(false)
              state.setNewModelName('')
              state.setNewModelTitle('')
            }}
          >
            取消
          </Button>
          <Button
            size="sm"
            className="border-0 bg-[#2563eb] text-white transition-colors duration-200 hover:bg-[#1d4ed8]"
            onClick={crud.handleConfirmCreateModel}
            disabled={state.creating}
          >
            {state.creating && <Loader2 className="mr-1.5 size-3.5 animate-spin" />}
            {state.creating ? '创建中...' : '创建'}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
