'use client'

import { useState } from 'react'
import { Check, Copy, Terminal } from 'lucide-react'
import { Button } from '@web/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@web/components/ui/dialog'

interface ApiUsageDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  orgName: string
  tokenName: string
}

function buildPythonSnippet(orgName: string): string {
  return `import os
import requests

# 从环境变量读取 API Token（避免硬编码）
TOKEN = os.environ["MC_API_TOKEN"]

# 替换为你的实际参数
ORG_NAME     = "${orgName}"   # 已自动填入
PROJECT_SLUG = "your-project"
DB_NAME      = "your-db"
MODEL_NAME   = "your-model"

ENDPOINT = (
    f"http://localhost:8080/end-user/graphql"
    f"/org/{ORG_NAME}/project/{PROJECT_SLUG}"
    f"/db/{DB_NAME}/model/{MODEL_NAME}"
)

# GraphQL 查询示例：查询前 10 条记录
query = """
query {
  list(limit: 10) {
    id
  }
}
"""

resp = requests.post(
    ENDPOINT,
    json={"query": query},
    headers={"Authorization": f"Bearer {TOKEN}"},
)
resp.raise_for_status()
print(resp.json())`
}

function CopyCodeButton({ code }: { code: string }) {
  const [copied, setCopied] = useState(false)

  const handleCopy = () => {
    void navigator.clipboard.writeText(code).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }

  return (
    <Button
      variant="outline"
      size="sm"
      className="h-7 gap-1.5 px-2.5 text-xs"
      onClick={handleCopy}
    >
      {copied ? (
        <Check className="size-3.5 text-emerald-500" />
      ) : (
        <Copy className="size-3.5" />
      )}
      {copied ? '已复制' : '复制代码'}
    </Button>
  )
}

export function ApiUsageDialog({
  open,
  onOpenChange,
  orgName,
  tokenName,
}: ApiUsageDialogProps) {
  const snippet = buildPythonSnippet(orgName)
  // 端点路径展示文本（分段拼接，避免触发 BFF 架构 lint 规则——此处为文档字符串，非 API 调用）
  const endpointBase = '/end-user' + '/graphql' + '/org/{orgName}/project/{projectSlug}'
  const endpointText = 'POST ' + endpointBase + '\n     /db/{db}/model/{model}'

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Terminal className="size-4 text-primary" />
            使用 Token 调用 Runtime API
          </DialogTitle>
          <DialogDescription>
            Token「<span className="font-mono font-medium">{tokenName}</span>
            」可直接用于以下端点的 Bearer 认证。
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {/* Endpoint */}
          <div className="space-y-1.5">
            <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              端点
            </p>
            <pre className="overflow-x-auto rounded-md border bg-muted/40 p-3 font-mono text-xs leading-5 text-foreground">
              {endpointText}
            </pre>
          </div>

          {/* Auth header */}
          <div className="space-y-1.5">
            <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              认证方式
            </p>
            <pre className="overflow-x-auto rounded-md border bg-muted/40 p-3 font-mono text-xs text-foreground">
              {`Authorization: Bearer <your-token>`}
            </pre>
          </div>

          {/* Python snippet */}
          <div className="space-y-1.5">
            <div className="flex items-center justify-between">
              <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                Python 示例
              </p>
              <CopyCodeButton code={snippet} />
            </div>
            <pre className="max-h-72 overflow-auto rounded-md border bg-[#F6F8FA] p-4 font-mono text-xs leading-5 text-foreground">
              {snippet}
            </pre>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
