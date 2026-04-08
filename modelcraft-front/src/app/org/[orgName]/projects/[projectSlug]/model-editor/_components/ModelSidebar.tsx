'use client'

import {
  Table2,
  Search,
  Plus,
  Download,
  MoreVertical,
  ChevronsUpDown,
  X,
  Filter,
  Loader2,
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

interface Database {
  name: string
}

interface ModelSidebarProps {
  state: ModelEditorState
  crud: ModelCRUD
  databases: Database[]
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

  return (
    <aside className="flex w-[260px] flex-shrink-0 flex-col border-r border-border bg-sidebar">
      {/* Header */}
      <header className="flex min-h-[var(--header-height,56px)] items-center border-b border-border px-6">
        <h1 className="font-heading text-lg font-semibold text-foreground">模型编辑器</h1>
      </header>

      {/* Content */}
      <div className="flex flex-1 flex-col gap-4 overflow-hidden pt-4">
        {/* Controls Section */}
        <div className="flex flex-col gap-2 px-4">
          {/* Database Selector */}
          <Popover open={state.databaseOpen} onOpenChange={state.setDatabaseOpen}>
            <PopoverTrigger asChild>
              <Button
                variant="outline"
                size="sm"
                className="border-strong hover:border-stronger h-7 w-full justify-between bg-muted px-2.5 text-xs font-normal transition-colors hover:bg-accent"
                disabled={databasesLoading}
              >
                <span className="flex items-center gap-1.5 truncate">
                  <span className="text-muted-foreground">database</span>
                  {databasesLoading ? (
                    <Loader2 className="size-3 animate-spin" />
                  ) : (
                    <span className="text-foreground">{state.selectedDatabase || 'Select...'}</span>
                  )}
                </span>
                <ChevronsUpDown className="size-3.5 shrink-0 text-muted-foreground" />
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-[228px] border border-slate-200 p-1 shadow-lg" align="start">
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
                        ? 'bg-selected text-foreground'
                        : 'text-muted-foreground hover:bg-selected hover:text-foreground'
                    )}
                    onClick={() => {
                      state.setSelectedDatabase(db.name)
                      state.setDatabaseOpen(false)
                    }}
                  >
                    <div className="flex items-center justify-between">
                      <span>{db.name}</span>
                    </div>
                  </button>
                ))
              )}
            </PopoverContent>
          </Popover>

          {/* New Model Button */}
          <Button
            size="sm"
            className="h-7 w-full justify-start border-0 bg-[#2563eb] px-2.5 text-xs font-normal text-white transition-colors duration-200 hover:bg-[#1d4ed8]"
            onClick={handleCreateModel}
          >
            <Plus className="mr-1.5 size-3.5" />
            <span>新建模型</span>
          </Button>

          {/* Import Model Button */}
          <button
            className="border-strong hover:border-stronger inline-flex h-7 w-full items-center justify-start gap-2 rounded-md border bg-muted px-2.5 text-xs font-normal shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground"
            onClick={() => state.setImportDialogOpen(true)}
            disabled={!state.selectedDatabase}
          >
            <Download className="mr-1.5 size-3.5" strokeWidth={1.5} />
            <span>导入模型</span>
          </button>
        </div>

        {/* Search & Filter */}
        <div className="flex items-center gap-2 px-4">
          <div className="relative flex-1">
            <Search className="pointer-events-none absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
            <Input
              type="text"
              placeholder="查询模型..."
              value={state.searchQuery}
              onChange={(e) => state.setSearchQuery(e.target.value)}
              className="border-control focus-visible:ring-background-control h-7 bg-foreground/[.026] px-8 text-xs focus-visible:ring-2 md:h-7"
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
          <Button
            variant="outline"
            size="icon"
            className="border-strong hover:border-stronger size-7 shrink-0 border-dashed bg-transparent transition-colors hover:bg-accent"
          >
            <Filter className="size-3.5 text-muted-foreground" />
          </Button>
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
                  'group relative flex items-center gap-3 h-7 pl-4 pr-1 rounded-sm cursor-pointer text-sm transition-colors select-none',
                  state.selectedModelId === model.id
                    ? 'bg-selected text-foreground'
                    : 'text-muted-foreground hover:bg-selected/50 hover:text-foreground'
                )}
              >
                {/* Active indicator */}
                {state.selectedModelId === model.id && (
                  <div className="absolute inset-y-0 left-0 w-0.5 bg-foreground" />
                )}

                {/* Icon */}
                <Table2 className="group-hover:text-foreground-lighter size-[15px] shrink-0 text-muted-foreground transition-colors" />

                {/* Name */}
                <span className={cn(
                  'truncate flex-1 text-sm transition-colors',
                  state.selectedModelId === model.id ? 'text-foreground' : 'text-muted-foreground group-hover:text-foreground'
                )}>
                  {model.name}
                </span>

                {/* Title tooltip */}
                {model.title && model.title !== model.name && (
                  <span className="max-w-[60px] truncate text-xs text-muted-foreground" title={model.title}>
                    {model.title}
                  </span>
                )}

                {/* More menu */}
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <button
                      type="button"
                      className={cn(
                        'opacity-0 group-hover:opacity-100 transition-opacity w-6 h-6 flex items-center justify-center hover:bg-accent rounded',
                        state.selectedModelId === model.id && 'opacity-100'
                      )}
                      onClick={(e) => e.stopPropagation()}
                    >
                      <MoreVertical className="size-3.5" />
                    </button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="w-40 border border-slate-200 shadow-lg">
                    <DropdownMenuItem
                      className="cursor-pointer text-xs focus:bg-selected focus:text-foreground"
                      onClick={(e) => {
                        e.stopPropagation()
                        crud.handleEditModel(model.id)
                      }}
                    >
                      <Edit className="mr-2 size-3.5" />
                      编辑模型
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      className="cursor-pointer text-xs focus:bg-selected focus:text-foreground"
                      onClick={async (e) => {
                        e.stopPropagation()
                        try {
                          await navigator.clipboard.writeText(model.name)
                        } catch (err) {
                          console.error('复制失败:', err)
                        }
                      }}
                    >
                      <X className="mr-2 size-3.5 opacity-0" />
                      复制名称
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      className="cursor-pointer text-xs text-destructive focus:bg-selected focus:text-destructive"
                      onClick={(e) => {
                        e.stopPropagation()
                        state.setModelToDelete(model)
                        state.setDeleteModelDialogOpen(true)
                      }}
                    >
                      <X className="mr-2 size-3.5 opacity-0" />
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
                <Table2 className="mb-3 size-10 opacity-20" />
                <p className="text-sm">请先选择数据库</p>
              </div>
            )}
          </div>
        </nav>
      </div>
    </aside>
  )
}
