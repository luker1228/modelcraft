'use client'

import { Loader2 } from 'lucide-react'
import { Button } from '@web/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@web/components/ui/dialog'
import type { ModelEditorState } from '../_hooks'
import type { ModelCRUD } from '../_hooks'

interface DeleteModelDialogProps {
  state: ModelEditorState
  crud: ModelCRUD
}

export function DeleteModelDialog({ state, crud }: DeleteModelDialogProps) {
  return (
    <Dialog open={state.deleteModelDialogOpen} onOpenChange={state.setDeleteModelDialogOpen}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>删除模型</DialogTitle>
          <DialogDescription>
            确定要删除模型 <span className="font-mono font-semibold text-foreground">{state.modelToDelete?.name}</span> 吗？此操作无法撤销。
          </DialogDescription>
        </DialogHeader>
        <DialogFooter className="gap-2">
          <Button
            variant="outline"
            onClick={() => state.setDeleteModelDialogOpen(false)}
            disabled={state.deletingModel}
          >
            取消
          </Button>
          <Button
            variant="destructive"
            onClick={crud.handleDeleteModel}
            disabled={state.deletingModel}
          >
            {state.deletingModel ? (
              <>
                <Loader2 className="mr-2 size-4 animate-spin" />
                删除中...
              </>
            ) : (
              '删除'
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
