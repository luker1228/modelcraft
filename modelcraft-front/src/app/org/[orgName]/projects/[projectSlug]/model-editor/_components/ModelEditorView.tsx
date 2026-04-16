'use client'

import { Suspense, lazy, useState, useCallback } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { Table2, Loader2, AlertTriangle, ExternalLink } from 'lucide-react'
import { Button } from '@web/components/ui/button'
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
import { CreateModelDialog } from './CreateModelDialog'
import { DeleteModelDialog } from './DeleteModelDialog'
import { FieldEditSheet } from './FieldEditSheet'

const DynamicModelTable = lazy(() => import('@web/components/features/model-editor/DynamicModelTable'))

export function ModelEditorView() {
  const params = useParams()
  const router = useRouter()
  const orgName = params?.orgName as string
  const projectSlug = params?.projectSlug as string

  const state = useModelEditorState()
  const [schemaRefreshToken, setSchemaRefreshToken] = useState(0)
  const handleFieldAdded = useCallback(() => {
    setSchemaRefreshToken((prev) => prev + 1)
  }, [])

  const crud = useModelCRUD({ orgName, projectSlug, state })
  const fieldOps = useFieldOperations({ orgName, projectSlug, state })
  const fkOps = useForeignKeys({ projectSlug, state })

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
              <DialogTitle className="font-heading text-base">数据库连接失败</DialogTitle>
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
        models={crud.models}
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
      />

      {/* Right Content Area */}
      <main className="flex min-w-0 flex-1 flex-col bg-sidebar">
        {state.selectedModelId ? (
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
            <DynamicModelTable
              key={`${state.selectedModelId}-${schemaRefreshToken}`}
              modelId={state.selectedModelId}
              projectSlug={projectSlug}
              orgName={orgName}
              refreshToken={schemaRefreshToken}
            />
          </Suspense>
        ) : (
          <div className="flex flex-1 items-center justify-center">
            <div className="px-6 text-center text-muted-foreground">
              <div className="mx-auto mb-5 flex size-20 items-center justify-center rounded-2xl bg-muted/30">
                <Table2 className="size-10 opacity-30" />
              </div>
              <h2 className="mb-2 font-heading text-base font-semibold text-foreground">选择一个模型</h2>
              <p className="mx-auto max-w-[280px] text-sm leading-relaxed">
                从左侧选择一个模型开始编辑，或点击 &ldquo;新建模型&rdquo; 创建新模型
              </p>
            </div>
          </div>
        )}
      </main>

      {/* Delete Model Confirmation Dialog */}
      <DeleteModelDialog state={state} crud={crud} />
    </div>
  )
}
