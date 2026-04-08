import { useState } from "react"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@web/components/ui/dialog"
import { Button } from "@web/components/ui/button"
import { Badge } from "@web/components/ui/badge"
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@web/components/ui/collapsible"
import { ScrollArea } from "@web/components/ui/scroll-area"
import { Separator } from "@web/components/ui/separator"
import { AlertTriangle, ChevronDown, ChevronRight, Copy, Bug } from "lucide-react"
import { toast } from "sonner"

export interface GraphQLErrorInfo {
  message: string
  extensions?: {
    code?: string
    exception?: {
      stacktrace?: string[]
    }
  }
  locations?: Array<{
    line: number
    column: number
  }>
  path?: (string | number)[]
}

export interface GraphQLErrorContext {
  operationName?: string
  operationType?: 'query' | 'mutation' | 'subscription'
  variables?: Record<string, unknown>
  query?: string
  networkError?: Error | unknown
  timestamp: string
  userAgent?: string
  url?: string
}

export interface GraphQLErrorDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  errors: GraphQLErrorInfo[]
  context?: GraphQLErrorContext
}

export function GraphQLErrorDialog({
  open,
  onOpenChange,
  errors,
  context,
}: GraphQLErrorDialogProps) {
  const [showDetails, setShowDetails] = useState(false)
  const [showContext, setShowContext] = useState(false)

  const copyErrorInfo = () => {
    const text = JSON.stringify(
      { errors, context, timestamp: new Date().toISOString() },
      null,
      2,
    )

    // navigator.clipboard only works on HTTPS / localhost
    if (navigator.clipboard && window.isSecureContext) {
      navigator.clipboard.writeText(text).then(
        () => toast.success("错误信息已复制到剪贴板"),
        () => fallbackCopy(text),
      )
    } else {
      fallbackCopy(text)
    }
  }

  const fallbackCopy = (text: string) => {
    const textarea = document.createElement('textarea')
    textarea.value = text
    textarea.style.cssText = 'position:fixed;top:0;left:0;opacity:0'
    document.body.appendChild(textarea)
    textarea.focus()
    textarea.select()
    const ok = document.execCommand('copy')
    document.body.removeChild(textarea)
    if (ok) {
      toast.success("错误信息已复制到剪贴板")
    } else {
      toast.error("复制失败，请手动复制")
    }
  }

  const getErrorTypeColor = (code?: string) => {
    switch (code) {
      case 'GRAPHQL_VALIDATION_FAILED':
        return 'destructive'
      case 'BAD_USER_INPUT':
        return 'secondary'
      case 'UNAUTHENTICATED':
        return 'outline'
      case 'FORBIDDEN':
        return 'outline'
      default:
        return 'destructive'
    }
  }

  const formatVariables = (variables?: Record<string, unknown>) => {
    if (!variables) return 'N/A'
    return JSON.stringify(variables, null, 2)
  }

  const formatQuery = (query?: string) => {
    if (!query) return 'N/A'
    // 简单的GraphQL格式化
    return query
      .replace(/\s+/g, ' ')
      .replace(/{\s*/g, '{\n  ')
      .replace(/\s*}/g, '\n}')
      .trim()
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[80vh] max-w-4xl">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <AlertTriangle className="size-5 text-red-500" />
            GraphQL 错误
          </DialogTitle>
          <DialogDescription>
            操作执行时发生了错误，请查看详细信息
          </DialogDescription>
        </DialogHeader>

        <ScrollArea className="max-h-[60vh]">
          <div className="space-y-4">
            {/* 错误列表 */}
            <div className="space-y-3">
              {errors.map((error, index) => (
                <div key={index} className="rounded-lg border p-4">
                  <div className="mb-2 flex items-start justify-between gap-2">
                    <div className="flex-1">
                      <p className="mb-1 font-semibold text-red-600">
                        {error.message}
                      </p>
                      {error.extensions?.code && (
                        <Badge variant={getErrorTypeColor(error.extensions.code)}>
                          {error.extensions.code}
                        </Badge>
                      )}
                    </div>
                  </div>

                  {/* 错误位置信息 */}
                  {error.locations && error.locations.length > 0 && (
                    <div className="mt-2 text-sm text-muted-foreground">
                      <strong>位置:</strong> 行 {error.locations[0].line}, 列 {error.locations[0].column}
                    </div>
                  )}

                  {/* 错误路径 */}
                  {error.path && error.path.length > 0 && (
                    <div className="mt-1 text-sm text-muted-foreground">
                      <strong>路径:</strong> {error.path.join(' → ')}
                    </div>
                  )}

                  {/* 堆栈跟踪 */}
                  {error.extensions?.exception?.stacktrace && (
                    <Collapsible>
                      <CollapsibleTrigger className="mt-2 flex items-center gap-1 text-sm text-blue-600 hover:text-blue-800">
                        <ChevronRight className="size-4" />
                        查看堆栈跟踪
                      </CollapsibleTrigger>
                      <CollapsibleContent className="mt-2">
                        <pre className="overflow-x-auto rounded bg-gray-100 p-2 text-xs">
                          {error.extensions.exception.stacktrace.join('\n')}
                        </pre>
                      </CollapsibleContent>
                    </Collapsible>
                  )}
                </div>
              ))}
            </div>

            <Separator />

            {/* 操作上下文信息 */}
            {context && (
              <Collapsible open={showContext} onOpenChange={setShowContext}>
                <CollapsibleTrigger className="flex w-full items-center gap-2 rounded p-2 text-left hover:bg-gray-50">
                  {showContext ? (
                    <ChevronDown className="size-4" />
                  ) : (
                    <ChevronRight className="size-4" />
                  )}
                  <Bug className="size-4" />
                  <span className="font-medium">操作上下文信息</span>
                </CollapsibleTrigger>
                <CollapsibleContent className="mt-2 space-y-3">
                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <strong>操作名称:</strong> {context.operationName || 'N/A'}
                    </div>
                    <div>
                      <strong>操作类型:</strong> {context.operationType || 'N/A'}
                    </div>
                    <div>
                      <strong>时间戳:</strong> {new Date(context.timestamp).toLocaleString()}
                    </div>
                    <div>
                      <strong>URL:</strong> {context.url || 'N/A'}
                    </div>
                  </div>

                  {/* 变量信息 */}
                  {context.variables && Object.keys(context.variables).length > 0 && (
                    <div>
                      <strong className="mb-1 block">请求变量:</strong>
                      <pre className="max-h-32 overflow-x-auto rounded bg-gray-100 p-2 text-xs">
                        {formatVariables(context.variables)}
                      </pre>
                    </div>
                  )}

                  {/* GraphQL查询 */}
                  {context.query && (
                    <div>
                      <strong className="mb-1 block">GraphQL 查询:</strong>
                      <pre className="max-h-40 overflow-x-auto rounded bg-gray-100 p-2 text-xs">
                        {formatQuery(context.query)}
                      </pre>
                    </div>
                  )}

                  {/* 网络错误 */}
                  {context.networkError && (
                    <div>
                      <strong className="mb-1 block">网络错误:</strong>
                      <pre className="overflow-x-auto rounded bg-red-50 p-2 text-xs">
                        {JSON.stringify(context.networkError, null, 2)}
                      </pre>
                    </div>
                  )}
                </CollapsibleContent>
              </Collapsible>
            )}
          </div>
        </ScrollArea>

        <DialogFooter className="flex justify-between">
          <Button
            variant="outline"
            onClick={copyErrorInfo}
            className="flex items-center gap-2"
          >
            <Copy className="size-4" />
            复制错误信息
          </Button>
          <div className="flex gap-2">
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              关闭
            </Button>
            <Button 
              onClick={() => window.location.reload()}
              variant="default"
            >
              刷新页面
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}