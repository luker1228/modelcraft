'use client'

import { Button } from '@web/components/ui/button'
import { Plus } from 'lucide-react'

interface ModelRecordInsertMenuProps {
  onCreateRecord: () => void
  canCreateRecord?: boolean
}

export function ModelRecordInsertMenu({
  onCreateRecord,
  canCreateRecord = true,
}: ModelRecordInsertMenuProps) {
  return (
    <Button
      size="sm"
      className="h-[26px] border-0 bg-primary px-2.5 text-xs font-normal text-white transition-colors duration-200 hover:bg-primary/90"
      onClick={onCreateRecord}
      disabled={!canCreateRecord}
    >
      <Plus className="mr-1.5 size-3.5" />
      <span>插入数据</span>
    </Button>
  )
}
