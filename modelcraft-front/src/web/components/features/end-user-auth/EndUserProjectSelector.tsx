'use client'

// src/web/components/features/end-user-auth/EndUserProjectSelector.tsx
// 终端用户选择 Project 卡片组件（EndUser v2）

import { AlertCircle, Loader2, Database, CheckCircle2 } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@web/components/ui/card'
import { Button } from '@web/components/ui/button'
import { Alert, AlertDescription } from '@web/components/ui/alert'
import { useEndUserProjectSelector } from '@web/hooks/end-user-auth-v2/useEndUserProjectSelector'
import { cn } from '@shared/utils'

interface EndUserProjectSelectorProps {
  orgName: string
}

/**
 * 终端用户选择 Project 卡片（EndUser v2）。
 * 当用户可访问多个项目时展示此页，选择确认后获取正式 access token。
 */
export function EndUserProjectSelector({ orgName }: EndUserProjectSelectorProps) {
  const { projects, selectedSlug, isLoading, error, selectProject, confirmSelection } =
    useEndUserProjectSelector(orgName)

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/40 p-4">
      <Card className="w-full max-w-md border bg-background shadow-sm">
        <CardHeader className="space-y-4 px-8 pt-8">
          <div className="flex items-center justify-center gap-3">
            <div className="flex size-10 items-center justify-center rounded-lg bg-primary">
              <Database className="size-5 text-primary-foreground" strokeWidth={1.5} />
            </div>
            <span className="text-xl font-semibold text-foreground">{orgName}</span>
          </div>
          <div className="text-center">
            <CardTitle className="text-2xl">选择项目</CardTitle>
            <CardDescription className="mt-2">
              您有访问权限的项目如下，请选择要进入的项目
            </CardDescription>
          </div>
        </CardHeader>

        <CardContent className="px-8 pb-8">
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertCircle className="size-4" />
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {projects.length === 0 && !error ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="size-6 animate-spin text-muted-foreground" />
            </div>
          ) : (
            <div className="flex flex-col gap-2">
              {projects.map((project) => (
                <button
                  key={project.slug}
                  type="button"
                  className={cn(
                    'flex w-full items-center justify-between rounded-md border px-4 py-3 text-left transition-colors',
                    selectedSlug === project.slug
                      ? 'border-primary bg-primary/5 text-foreground'
                      : 'border-border bg-background text-foreground hover:bg-muted/50'
                  )}
                  onClick={() => selectProject(project.slug)}
                >
                  <div className="flex flex-col gap-0.5">
                    <span className="text-sm font-medium">{project.title}</span>
                    <span className="text-xs text-muted-foreground">{project.slug}</span>
                  </div>
                  {selectedSlug === project.slug && (
                    <CheckCircle2 className="size-4 text-primary" />
                  )}
                </button>
              ))}

              <Button
                className="mt-2 w-full"
                disabled={!selectedSlug || isLoading || projects.length === 0}
                onClick={confirmSelection}
              >
                {isLoading && <Loader2 className="mr-2 size-4 animate-spin" />}
                进入项目
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
