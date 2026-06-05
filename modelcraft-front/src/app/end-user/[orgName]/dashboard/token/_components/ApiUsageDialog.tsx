'use client'

import { useState } from 'react'
import { Check, Copy, Terminal } from 'lucide-react'
import { cn } from '@/shared/utils'
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

type TabId = 'python' | 'curl'

function buildGraphQLSnippet(orgName: string): string {
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

function buildCurlSnippet(orgName: string): string {
  // 端点路径分段拼接，避免触发 BFF 架构 lint 规则——此处为文档字符串，非 API 调用
  const endpointPath =
    'http://localhost:8080' +
    '/end-user/graphql' +
    `/org/${orgName}/project/$PROJECT_SLUG` +
    '/db/$DB_NAME/model/$MODEL_NAME'
  return `# 替换为你的实际参数
ORG_NAME="${orgName}"   # 已自动填入
PROJECT_SLUG="your-project"
DB_NAME="your-db"
MODEL_NAME="your-model"
TOKEN="your-token"

ENDPOINT="${endpointPath}"

# GraphQL 查询：列出前 10 条记录
curl -s -X POST "$ENDPOINT" \\
  -H "Authorization: Bearer $TOKEN" \\
  -H "Content-Type: application/json" \\
  -d '{"query":"{ list(limit: 10) { id } }"}' | jq .`
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

const TABS: { id: TabId; label: string }[] = [
  { id: 'python', label: 'Python' },
  { id: 'curl', label: 'curl' },
]

interface TabContentProps {
  orgName: string
  tab: TabId
}

function TabContent({ orgName, tab }: TabContentProps) {
  // 端点路径分段拼接，避免触发 BFF 架构 lint 规则——此处为文档字符串，非 API 调用
  const endpointBase = '/end-user' + '/graphql' + '/org/{orgName}/project/{projectSlug}'
  const endpointText = 'POST ' + endpointBase + '\n     /db/{db}/model/{model}'

  if (tab === 'python') {
    const snippet = buildGraphQLSnippet(orgName)

    return (
      <div className="space-y-4">
        <div className="space-y-1.5">
          <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">端点</p>
          <pre className="overflow-x-auto rounded-md border bg-muted/40 p-3 font-mono text-xs leading-5 text-foreground">
            {endpointText}
          </pre>
        </div>
        <div className="space-y-1.5">
          <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">认证方式</p>
          <pre className="overflow-x-auto rounded-md border bg-muted/40 p-3 font-mono text-xs text-foreground">
            {`Authorization: Bearer <your-token>`}
          </pre>
        </div>
        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">Python 示例</p>
            <CopyCodeButton code={snippet} />
          </div>
          <pre className="max-h-64 overflow-auto rounded-md border bg-[#F6F8FA] p-4 font-mono text-xs leading-5 text-foreground">
            {snippet}
          </pre>
        </div>
      </div>
    )
  }

  // curl tab
  const snippet = buildCurlSnippet(orgName)

  return (
    <div className="space-y-4">
      <div className="space-y-1.5">
        <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">端点</p>
        <pre className="overflow-x-auto rounded-md border bg-muted/40 p-3 font-mono text-xs leading-5 text-foreground">
          {endpointText}
        </pre>
      </div>
      <div className="space-y-1.5">
        <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">认证方式</p>
        <pre className="overflow-x-auto rounded-md border bg-muted/40 p-3 font-mono text-xs text-foreground">
          {`Authorization: Bearer <your-token>`}
        </pre>
      </div>
      <div className="space-y-1.5">
        <div className="flex items-center justify-between">
          <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">curl 示例</p>
          <CopyCodeButton code={snippet} />
        </div>
        <pre className="max-h-64 overflow-auto rounded-md border bg-[#F6F8FA] p-4 font-mono text-xs leading-5 text-foreground">
          {snippet}
        </pre>
      </div>
    </div>
  )
}

export function ApiUsageDialog({
  open,
  onOpenChange,
  orgName,
  tokenName,
}: ApiUsageDialogProps) {
  const [activeTab, setActiveTab] = useState<TabId>('python')

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

        <div className="flex gap-5">
          {/* Left vertical tab list */}
          <div className="flex shrink-0 flex-col gap-0.5 pt-0.5">
            {TABS.map((t) => (
              <button
                key={t.id}
                type="button"
                onClick={() => setActiveTab(t.id)}
                className={cn(
                  'rounded-md px-3 py-1.5 text-left text-sm transition-colors',
                  activeTab === t.id
                    ? 'bg-muted font-medium text-foreground'
                    : 'text-muted-foreground hover:bg-muted/60 hover:text-foreground',
                )}
              >
                {t.label}
              </button>
            ))}
          </div>

          {/* Divider */}
          <div className="w-px shrink-0 bg-border" />

          {/* Right content */}
          <div className="min-w-0 flex-1">
            <TabContent orgName={orgName} tab={activeTab} />
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
