'use client'
import { createContext, useContext } from 'react'
import type { DevelopRecordWorkspaceAIRef } from '@web/components/features/model-editor/model-record-form/DevelopRecordWorkspace'

export const WorkspaceAIRefContext = createContext<React.MutableRefObject<DevelopRecordWorkspaceAIRef | null> | null>(null)

export function useWorkspaceAIRef() {
  return useContext(WorkspaceAIRefContext)
}
