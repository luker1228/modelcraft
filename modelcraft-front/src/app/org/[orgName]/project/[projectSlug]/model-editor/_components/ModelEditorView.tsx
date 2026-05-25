'use client'

import { Suspense, lazy, useState, useCallback, useEffect } from 'react'
import { useParams, useRouter, useSearchParams } from 'next/navigation'
import { Loader2, AlertTriangle, ExternalLink } from 'lucide-react'
import { Button } from '@web/components/ui/button'
import { toast } from 'sonner'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@web/components/ui/dialog'
import { ImportModelDialog } from '@web/components/features/model-editor/ImportModelDialog'
import { useModelEditorState, useModelCRUD, useFieldOperations, useForeignKeys } from '../_hooks'
import { ModelSidebar } from './ModelSidebar'
import { ModelDetailPanel } from './ModelDetailPanel'
import { ModelSchemaPanel } from './ModelSchemaPanel'
import { CreateModelDialog } from './CreateModelDialog'
import { DeleteModelDialog } from './DeleteModelDialog'
import { FieldEditSheet } from './FieldEditSheet'
import {
  DataWorkspacePanel,
  type DataWorkspaceTab,
} from '@web/components/features/model-editor/DataWorkspacePanel'
  // Onboarding: pendingAction is consumed by ModelSidebar directly via useOnboarding

const DevelopRecordWorkspace = lazy(() => import('@web/components/features/model-editor/model-record-form/DevelopRecordWorkspace'))
const MAX_MODEL_TABS = 8

export function ModelEditorView() {
  const params = useParams()
  const router = useRouter()
  const orgName = params?.orgName as string
  const projectSlug = params?.projectSlug as string

  const state = useModelEditorState()
  const searchParams = useSearchParams()
  const viewMode = (searchParams.get('view') === 'data' ? 'data' : 'schema') as 'schema' | 'data'
  const [schemaRefreshToken, setSchemaRefreshToken] = useState(0)
  const [openedTabs, setOpenedTabs] = useState<DataWorkspaceTab[]>([])
  const handleFieldAdded = useCallback(() => {
    setSchemaRefreshToken((prev) => prev + 1)
  }, [])

  const crud = useModelCRUD({ orgName, projectSlug, state })
  const fieldOps = useFieldOperations({ orgName, projectSlug, state })
  const fkOps = useForeignKeys({ projectSlug, state })

  // Onboarding: pendingAction is consumed by ModelSidebar directly via useOnboarding

  useEffect(() => {
    if (!state.selectedModelId) return
    const selectedModel = crud.models.find((model) => model.id === state.selectedModelId)
    if (!selectedModel) return

    setOpenedTabs((prev) => {
      if (prev.some((tab) => tab.id === selectedModel.id)) return prev
      if (prev.length >= MAX_MODEL_TABS) {
        toast.warning(`最多可同时打开 ${MAX_MODEL_TABS} 个模型标签`)
        return prev
      }
      return [
        ...prev,
        {
          id: selectedModel.id,
          name: selectedModel.name,
          title: selectedModel.title,
        },
      ]
    })
  }, [state.selectedModelId, crud.models])

  useEffect(() => {
    setOpenedTabs([])
  }, [state.selectedDatabase])

  // Reset schema-view overlays when switching to data view
  useEffect(() => {
    if (viewMode === 'data') {
      state.setEditModelOpen(false)
      state.setInsertFieldOpen(false)
      state.setEditFieldOpen(false)
      state.setFkFormOpen(false)
    }
  }, [viewMode]) // eslint-disable-line react-hooks/exhaustive-deps

  // Auto-load model detail when in schema view and selected model changes
  useEffect(() => {
    if (viewMode === 'schema' && state.selectedModelId) {
      void crud.loadModelDetailForSchemaView(state.selectedModelId)
    }
  }, [viewMode, state.selectedModelId]) // eslint-disable-line react-hooks/exhaustive-deps

  const handleCloseTab = (tabId: string) => {
    setOpenedTabs((prev) => {
      const next = prev.filter((tab) => tab.id !== tabId)
      if (state.selectedModelId === tabId) {
        state.setSelectedModelId(next[next.length - 1]?.id ?? null)
      }
      return next
    })
  }

  return (
    <div className="relative flex size-full">
      {/* Cluster connection failure dialog */}
      <Dialog open={state.connectionFailed} onOpenChange={() => {}}>
        <DialogContent
          className="sm:max-w-md"
          onInteractOutside={(e) => e.preventDefault()}
          onEscapeKeyDown={(e) => e.preventDefault()}
        >
          <DialogHeader>
            <div className="mb-1 flex items-center gap-3">
              <div className="flex size-10 flex-shrink-0 items-center justify-center rounded-full bg-destructive/10">
                <AlertTriangle className="size-5 text-destructive" />
              </div>
              <DialogTitle className="text-base">数据库连接失败</DialogTitle>
            </div>
            <DialogDescription className="pl-[52px] text-sm leading-relaxed">
              {state.connectionError}
              <br />
              <span className="text-muted-foreground">请前往集群配置页检查数据库连接信息。</span>
            </DialogDescription>
          </DialogHeader>
          <DialogFooter className="flex gap-2 sm:justify-end">
            <Button variant="outline" size="sm" onClick={() => router.back()}>
              返回
            </Button>
            <Button size="sm" onClick={crud.handleGoToCluster}>
              <ExternalLink className="mr-1.5 size-3.5" />
              前往集群配置
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Create Model Sheet */}
      <CreateModelDialog state={state} crud={crud} />

      {/* Edit Model Drawer */}
      <ModelDetailPanel
        state={state}
        crud={crud}
        fieldOps={fieldOps}
        fkOps={fkOps}
        orgName={orgName}
        projectSlug={projectSlug}
        onFieldAdded={handleFieldAdded}
      />

      {/* Edit Field Sheet */}
      <FieldEditSheet state={state} fieldOps={fieldOps} orgName={orgName} projectSlug={projectSlug} />

      {/* Import Model Dialog */}
      <ImportModelDialog
        open={state.importDialogOpen}
        onOpenChange={state.setImportDialogOpen}
        projectSlug={projectSlug}
        databaseName={state.selectedDatabase}
        onSuccess={() => crud.refetchModels()}
      />

      {/* Connection checking overlay */}
      {state.connectionChecking && (
        <div className="absolute inset-0 z-10 flex items-center justify-center bg-background/50">
          <div className="flex flex-col items-center gap-3 text-muted-foreground">
            <Loader2 className="size-6 animate-spin" />
            <span className="text-sm">正在检查数据库连接...</span>
          </div>
        </div>
      )}

      {/* Left Sidebar - Model List */}
      <ModelSidebar
        state={state}
        crud={crud}
        databases={crud.databases}
        databasesLoading={crud.databasesLoading}
        filteredModels={crud.filteredModels}
        modelsLoading={crud.modelsLoading}
        viewMode={viewMode}
      />

      {/* Right Content Area */}
      <main className="flex min-w-0 flex-1 flex-col bg-background p-4">
        <section className="flex h-full min-h-0 flex-col gap-3">
          {viewMode === 'schema' ? (
            <ModelSchemaPanel
              state={state}
              crud={crud}
              fieldOps={fieldOps}
              fkOps={fkOps}
              orgName={orgName}
              projectSlug={projectSlug}
              onFieldAdded={handleFieldAdded}
            />
          ) : (
            <DataWorkspacePanel
              tabs={openedTabs}
              activeTabId={state.selectedModelId ?? ''}
              onTabChange={(tabId) => state.setSelectedModelId(tabId)}
              onTabClose={handleCloseTab}
              emptyText="从左侧选择模型以打开数据表"
              className="h-full min-h-0"
              renderContent={(activeTab) => (
                <Suspense
                  fallback={
                    <div className="flex flex-1 items-center justify-center">
                      <div className="flex flex-col items-center gap-3 text-muted-foreground">
                        <Loader2 className="size-6 animate-spin" />
                        <span className="text-sm">加载中...</span>
                      </div>
                    </div>
                  }
                >
                  <DevelopRecordWorkspace
                    key={`${activeTab.id}-${schemaRefreshToken}`}
                    modelId={activeTab.id}
                    projectSlug={projectSlug}
                    orgName={orgName}
                    refreshToken={schemaRefreshToken}
                  />
                </Suspense>
              )}
            />
          )}
        </section>
      </main>

      {/* Delete Model Confirmation Dialog */}
      <DeleteModelDialog state={state} crud={crud} />
    </div>
  )
}
