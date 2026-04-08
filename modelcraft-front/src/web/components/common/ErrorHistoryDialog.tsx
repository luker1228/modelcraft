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
import { ScrollArea } from "@web/components/ui/scroll-area"
import { Separator } from "@web/components/ui/separator"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@web/components/ui/card"
import { History, Trash2, Eye, Copy } from "lucide-react"
import { useErrorStore } from '@web/stores/error'
import type { GraphQLErrorInfo, GraphQLErrorContext } from '@web/components/common/GraphQLErrorDialog'
import { toast } from "sonner"

interface ErrorHistoryDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function ErrorHistoryDialog({
  open,
  onOpenChange,
}: ErrorHistoryDialogProps) {
  const { errorHistory, clearHistory, showErrorDialog } = useErrorStore()
  const [selectedErrorId, setSelectedErrorId] = useState<string | null>(null)

  const handleViewError = (errorEntry: typeof errorHistory[0]) => {
    showErrorDialog(errorEntry.errors, errorEntry.context)
    onOpenChange(false)
  }

  const handleCopyError = (errorEntry: typeof errorHistory[0]) => {
    const errorInfo = {
      ...errorEntry,
      copiedAt: new Date().toISOString(),
    }
    
    navigator.clipboard.writeText(JSON.stringify(errorInfo, null, 2))
    toast.success("错误信息已复制到剪贴板")
  }

  const handleClearHistory = () => {
    clearHistory()
    toast.success("错误历史已清空")
  }

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })
  }

  const getErrorSummary = (errors: GraphQLErrorInfo[]) => {
    if (errors.length === 1) {
      return errors[0].message
    }
    return `${errors.length} 个错误`
  }

  const getOperationInfo = (context?: GraphQLErrorContext) => {
    if (!context) return 'Unknown'
    return `${context.operationType || 'unknown'}: ${context.operationName || 'unnamed'}`
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[80vh] max-w-4xl">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <History className="size-5" />
            错误历史记录
          </DialogTitle>
          <DialogDescription>
            查看最近发生的 GraphQL 错误记录 (最多保存 50 条)
          </DialogDescription>
        </DialogHeader>

        <ScrollArea className="max-h-[60vh]">
          <div className="space-y-3">
            {errorHistory.length === 0 ? (
              <div className="py-8 text-center text-muted-foreground">
                <History className="mx-auto mb-4 size-12 opacity-50" />
                <p>暂无错误记录</p>
              </div>
            ) : (
              errorHistory.map((errorEntry) => (
                <Card 
                  key={errorEntry.id}
                  className={`cursor-pointer transition-colors hover:bg-gray-50 ${
                    selectedErrorId === errorEntry.id ? 'ring-2 ring-blue-500' : ''
                  }`}
                  onClick={() => setSelectedErrorId(
                    selectedErrorId === errorEntry.id ? null : errorEntry.id
                  )}
                >
                  <CardHeader className="pb-3">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <CardTitle className="mb-1 text-sm font-semibold text-red-600">
                          {getErrorSummary(errorEntry.errors)}
                        </CardTitle>
                        <CardDescription className="text-xs">
                          {getOperationInfo(errorEntry.context)} • {formatTimestamp(errorEntry.timestamp)}
                        </CardDescription>
                      </div>
                      <div className="flex gap-1">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={(e) => {
                            e.stopPropagation()
                            handleCopyError(errorEntry)
                          }}
                          className="size-8 p-0"
                        >
                          <Copy className="size-3" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={(e) => {
                            e.stopPropagation()
                            handleViewError(errorEntry)
                          }}
                          className="size-8 p-0"
                        >
                          <Eye className="size-3" />
                        </Button>
                      </div>
                    </div>
                  </CardHeader>
                  
                  {selectedErrorId === errorEntry.id && (
                    <CardContent className="pt-0">
                      <Separator className="mb-3" />
                      <div className="space-y-2">
                        {errorEntry.errors.map((error, index) => (
                          <div key={index} className="text-sm">
                            <div className="mb-1 flex items-center gap-2">
                              <span className="font-semibold">错误 {index + 1}:</span>
                              {error.extensions?.code && (
                                <Badge variant="secondary" className="text-xs">
                                  {error.extensions.code}
                                </Badge>
                              )}
                            </div>
                            <p className="ml-4 text-foreground">{error.message}</p>
                            {error.path && (
                              <p className="ml-4 text-xs text-muted-foreground">
                                路径: {error.path.join(' → ')}
                              </p>
                            )}
                          </div>
                        ))}
                        
                        {errorEntry.context?.variables && (
                          <div className="mt-3">
                            <p className="mb-1 text-xs font-semibold">请求变量:</p>
                            <pre className="max-h-20 overflow-x-auto rounded bg-gray-100 p-2 text-xs">
                              {JSON.stringify(errorEntry.context.variables, null, 2)}
                            </pre>
                          </div>
                        )}
                      </div>
                    </CardContent>
                  )}
                </Card>
              ))
            )}
          </div>
        </ScrollArea>

        <DialogFooter className="flex justify-between">
          <Button
            variant="outline"
            onClick={handleClearHistory}
            disabled={errorHistory.length === 0}
            className="flex items-center gap-2"
          >
            <Trash2 className="size-4" />
            清空历史
          </Button>
          <Button onClick={() => onOpenChange(false)}>
            关闭
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
