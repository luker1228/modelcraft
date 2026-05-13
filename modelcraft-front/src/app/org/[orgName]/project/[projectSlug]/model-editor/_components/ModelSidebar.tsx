'use client'

import {
  Table2,
  Search,
  Plus,
  Download,
  MoreVertical,
  ChevronsUpDown,
  X,
  Loader2,
  Database,
  Edit,
} from 'lucide-react'
import { cn } from '@/shared/utils'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@web/components/ui/dropdown-menu'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@web/components/ui/popover'
import type { ModelEditorState, EditorModel } from '../_hooks'
import type { ModelCRUD } from '../_hooks'
import { useOnboarding } from '@shared/onboarding/OnboardingContext'

interface DatabaseOption {
  name: string
}

interface ModelSidebarProps {
  state: ModelEditorState
  crud: ModelCRUD
  databases: DatabaseOption[]
  databasesLoading: boolean
  filteredModels: EditorModel[]
  modelsLoading: boolean
}

export function ModelSidebar({
  state,
  crud,
  databases,
  databasesLoading,
  filteredModels,
  modelsLoading,
}: ModelSidebarProps) {
  const { pendingAction, setPendingAction } = useOnboarding()

  const handleModelDetailClick = (modelId: string) => {
    state.setSelectedModelId(modelId)
  }

  const handleCreateModel = () => {
    if (!state.selectedDatabase) {
      alert('请先选择数据库')
      return
    }
    state.setCreateModelOpen(true)
  }

  // Clear pendingAction when user interacts with the spotlit sidebar elements
  const handleDatabaseSelect = (dbName: string) => {
    if (pendingAction === 'select_database') setPendingAction(null)
    state.setSelectedDatabase(dbName)
    state.setSelectedModelId(null)
    state.setDatabaseOpen(false)
  }

  const handleCreateModelClick = () => {
    if (pendingAction === 'create_model') setPendingAction(null)
    handleCreateModel()
  }

  return (
    <aside className="flex w-[260px] flex-shrink-0 flex-col border-r border-border bg-sidebar">

      {/* ── Zone 1: Database Context ── */}
      <div className="p-3">
        <Popover open={state.databaseOpen} onOpenChange={state.setDatabaseOpen}>
          <PopoverTrigger asChild>
            <Button
              variant="outline"
              size="sm"
              className={cn(
                'h-9 w-full justify-between px-3 text-sm font-medium transition-colors',
                state.selectedDatabase
                  ? 'border-primary/40 bg-primary/5 text-foreground hover:border-primary/60 hover:bg-primary/10'
                  : 'border-dashed border-muted-foreground/40 bg-background text-muted-foreground hover:border-primary/40 hover:bg-primary/5 hover:text-foreground',
                pendingAction === 'select_database' && 'ring-2 ring-amber-400 ring-offset-1 animate-pulse border-amber-400'
              )}
              disabled={databasesLoading}
            >
              <span className="flex min-w-0 items-center gap-2">
                <Database className={cn('size-3.5 shrink-0', state.selectedDatabase ? 'text-primary' : 'text-muted-foreground')} />
                {databasesLoading ? (
                  <><Loader2 className="size-3 shrink-0 animate-spin" /><span className="text-muted-foreground">加载中...</span></>
                ) : state.selectedDatabase ? (
                  <span className="truncate font-medium text-foreground">{state.selectedDatabase}</span>
                ) : (
                  <span>选择数据库</span>
                )}
              </span>
              <ChevronsUpDown className="ml-2 size-3.5 shrink-0 text-muted-foreground" />
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-[228px] border border-border p-1 shadow-lg" align="start">
            {databases.length === 0 ? (
              <div className="px-2.5 py-3 text-center text-sm text-muted-foreground">
                No databases found
              </div>
            ) : (
              databases.map((db) => (
                <button
                  key={db.name}
                  type="button"
                  className={cn(
                    'w-full text-left px-2.5 py-1.5 text-sm rounded-sm transition-colors cursor-pointer',
                    state.selectedDatabase === db.name
                      ? 'bg-accent text-foreground'
                      : 'text-muted-foreground hover:bg-accent hover:text-foreground'
                  )}
                  onClick={() => {
                    handleDatabaseSelect(db.name)
                  }}
                >
                  {db.name}
                </button>
              ))
            )}
          </PopoverContent>
        </Popover>
      </div>

      {/* ── Divider ── */}
      <div className="border-t border-border" />

      {/* ── Zone 2: Models ── */}
      <div className="flex flex-1 flex-col overflow-hidden">

        {/* Action buttons */}
        <div className="flex flex-col gap-1 px-3 py-2.5">
          <Button
            size="sm"
            variant="outline"
            className={cn(
              'h-7 w-full justify-start px-2.5 text-xs font-normal transition-colors',
              !state.selectedDatabase && 'pointer-events-none opacity-40',
              pendingAction === 'create_model' && state.selectedDatabase && 'ring-2 ring-amber-400 ring-offset-1 animate-pulse border-amber-400'
            )}
            onClick={handleCreateModelClick}
            disabled={!state.selectedDatabase}
            >
              <Plus className="mr-1 size-3.5" />
              新建模型
            </Button>
          <Button
            size="sm"
            variant="outline"
            className={cn(
              'h-7 w-full justify-start px-2.5 text-xs font-normal transition-colors',
              !state.selectedDatabase && 'pointer-events-none opacity-40'
            )}
            onClick={() => state.setImportDialogOpen(true)}
            disabled={!state.selectedDatabase}
          >
            <Download className="mr-1 size-3.5" strokeWidth={1.5} />
            导入模型
          </Button>
        </div>

        {/* Search */}
        <div className="px-3 pb-2">
          <div className="relative">
            <Search className="pointer-events-none absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
            <Input
              type="text"
              placeholder="查询模型..."
              value={state.searchQuery}
              onChange={(e) => state.setSearchQuery(e.target.value)}
              className="h-7 bg-foreground/[.026] px-8 text-xs"
            />
            {state.searchQuery && (
              <button
                type="button"
                className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground transition-colors hover:text-foreground"
                onClick={() => state.setSearchQuery('')}
              >
                <X className="size-3.5" />
              </button>
            )}
          </div>
        </div>

        {/* Model List */}
        <nav className="min-h-0 flex-1 overflow-y-auto px-2 pb-4">
          <div className="space-y-px">
            {modelsLoading && (
              <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
                <Loader2 className="mb-3 size-6 animate-spin" />
                <p className="text-sm">加载模型中...</p>
              </div>
            )}

            {!modelsLoading && filteredModels.map((model) => (
              <div
                  key={model.id}
                  role="button"
                  tabIndex={0}
                  onClick={() => handleModelDetailClick(model.id)}
                  onKeyDown={(e) => e.key === 'Enter' && handleModelDetailClick(model.id)}
                  className={cn(
                    'group flex items-center gap-2 h-7 pl-2 pr-1 rounded-md cursor-pointer transition-colors select-none border-l-[3px]',
                    state.selectedModelId === model.id
                      ? 'bg-primary/[0.08] text-primary border-l-primary'
                      : 'text-muted-foreground hover:bg-accent/60 hover:text-foreground border-l-transparent'
                  )}
                >
                <Table2 className={cn('size-[15px] shrink-0 transition-colors', state.selectedModelId === model.id ? 'text-primary' : 'text-muted-foreground group-hover:text-foreground')} />

                <span className="min-w-0 flex-1 truncate text-xs">
                  {model.name}
                </span>

                {model.createdVia === 'IMPORTED' && (
                  <span className="border-warning/30 bg-warning/10 text-warning rounded border px-1 py-0 text-[10px]">
                    托管
                  </span>
                )}

                {model.title && model.title !== model.name && (
                  <span className="max-w-[56px] shrink-0 truncate text-xs text-muted-foreground/60" title={model.title}>
                    {model.title}
                  </span>
                )}

                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <button
                      type="button"
                      className="flex size-5 shrink-0 items-center justify-center rounded opacity-0 transition-opacity hover:bg-accent hover:!opacity-100 group-hover:opacity-60"
                      onClick={(e) => e.stopPropagation()}
                    >
                      <MoreVertical className="size-3" />
                    </button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="w-40 border border-border shadow-lg">
                    <DropdownMenuItem
                      className={cn(
                        'text-xs focus:bg-accent',
                        model.createdVia === 'IMPORTED'
                          ? 'cursor-not-allowed text-muted-foreground/50 focus:text-muted-foreground/50'
                          : 'cursor-pointer text-foreground focus:text-foreground'
                      )}
                      onClick={(e) => {
                        e.stopPropagation()
                        if (model.createdVia === 'IMPORTED') {
                          return
                        }
                        crud.handleEditModel(model.id)
                      }}
                      disabled={model.createdVia === 'IMPORTED'}
                    >
                      <Edit className="mr-2 size-3.5" />
                      编辑模型
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      className="cursor-pointer text-xs focus:bg-accent focus:text-foreground"
                      onClick={async (e) => {
                        e.stopPropagation()
                        try {
                          await navigator.clipboard.writeText(model.name)
                        } catch (err) {
                          console.error('复制失败:', err)
                        }
                      }}
                    >
                      复制名称
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      className={cn(
                        'text-xs focus:bg-accent',
                        model.createdVia === 'IMPORTED'
                          ? 'cursor-not-allowed text-muted-foreground/50 focus:text-muted-foreground/50'
                          : 'cursor-pointer text-destructive focus:text-destructive'
                      )}
                      onClick={(e) => {
                        e.stopPropagation()
                        if (model.createdVia === 'IMPORTED') {
                          return
                        }
                        state.setModelToDelete(model)
                        state.setDeleteModelDialogOpen(true)
                      }}
                      disabled={model.createdVia === 'IMPORTED'}
                    >
                      删除模型
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            ))}

            {!modelsLoading && filteredModels.length === 0 && state.selectedDatabase && (
              <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
                <Table2 className="mb-3 size-10 opacity-20" />
                <p className="text-sm">暂无模型</p>
              </div>
            )}

            {!state.selectedDatabase && !databasesLoading && (
              <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
                <Database className="mb-3 size-8 opacity-20" />
                <p className="text-sm">请先选择数据库</p>
              </div>
            )}
          </div>
        </nav>
      </div>
    </aside>
  )
}
