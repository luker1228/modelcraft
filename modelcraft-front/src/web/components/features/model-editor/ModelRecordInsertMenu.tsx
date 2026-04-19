'use client'

import { useState } from 'react'
import { Button } from '@web/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@web/components/ui/dropdown-menu'
import { ChevronDown, Columns, Plus } from 'lucide-react'
import { InsertFieldSheet } from './InsertFieldSheet'

interface ModelRecordInsertMenuProps {
  onCreateRecord: () => void
  modelId: string
  modelName?: string
  projectSlug: string
  orgName: string
  existingFieldNames: string[]
  onInsertFieldSuccess: () => void
  canInsertField?: boolean
}

export function ModelRecordInsertMenu({
  onCreateRecord,
  modelId,
  modelName,
  projectSlug,
  orgName,
  existingFieldNames,
  onInsertFieldSuccess,
  canInsertField = true,
}: ModelRecordInsertMenuProps) {
  const [insertColumnOpen, setInsertColumnOpen] = useState(false)

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            size="sm"
            className="h-[26px] border-0 bg-primary px-2.5 text-xs font-normal text-white transition-colors duration-200 hover:bg-primary/90"
          >
            <Plus className="mr-1.5 size-3.5" />
            <span>插入</span>
            <ChevronDown className="ml-1.5 size-3 opacity-70" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" className="w-40 border border-slate-200 shadow-lg">
          <DropdownMenuItem
            onClick={onCreateRecord}
            className="cursor-pointer text-xs focus:bg-selected focus:text-foreground"
          >
            <Plus className="mr-2 size-3.5" />
            插入数据
          </DropdownMenuItem>
          {canInsertField && (
            <DropdownMenuItem
              className="cursor-pointer text-xs focus:bg-selected focus:text-foreground"
              onClick={() => setInsertColumnOpen(true)}
            >
              <Columns className="mr-2 size-3.5" />
              插入列
            </DropdownMenuItem>
          )}
        </DropdownMenuContent>
      </DropdownMenu>

      {canInsertField && (
        <InsertFieldSheet
          open={insertColumnOpen}
          onOpenChange={setInsertColumnOpen}
          modelId={modelId}
          modelName={modelName}
          projectSlug={projectSlug}
          orgName={orgName}
          existingFieldNames={existingFieldNames}
          onSuccess={onInsertFieldSuccess}
        />
      )}
    </>
  )
}
