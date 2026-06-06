import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@web/components/ui/dialog'
import { Button } from '@web/components/ui/button'
import {
  API_DOC_SERVER_URL,
  buildFindManyCurlSnippet,
  buildModelApiAiPrompt,
  buildModelRuntimeEndpoint,
  type ModelApiDocContext,
} from './model-api-docs'

export interface ModelApiDocsDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  context: ModelApiDocContext | null
}

const HEADER_EXAMPLE = 'Authorization: Bearer <API_TOKEN>'

async function copyText(text: string) {
  if (!navigator.clipboard?.writeText) {
    return
  }

  try {
    await navigator.clipboard.writeText(text)
  } catch {
    // Ignore clipboard write failures to avoid unhandled promise rejections.
  }
}

function Section({
  title,
  children,
}: {
  title: string
  children: React.ReactNode
}) {
  return (
    <section className="space-y-3 rounded-lg border p-4">
      <h3 className="text-sm font-semibold text-foreground">{title}</h3>
      {children}
    </section>
  )
}

function CodeBlock({
  value,
  className = '',
}: {
  value: string
  className?: string
}) {
  return (
    <pre
      className={`overflow-x-auto rounded-md bg-muted p-3 text-sm text-foreground ${className}`}
    >
      <code>{value}</code>
    </pre>
  )
}

export function ModelApiDocsDialog({
  open,
  onOpenChange,
  context,
}: ModelApiDocsDialogProps) {
  const endpoint = context ? buildModelRuntimeEndpoint(context) : null
  const curlSnippet = context ? buildFindManyCurlSnippet(context) : null
  const aiPrompt = context ? buildModelApiAiPrompt(context) : null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] max-w-4xl overflow-y-auto">
        <DialogHeader>
          <DialogTitle>API 文档</DialogTitle>
          <DialogDescription>
            先确认这个模型的运行时访问方式，再直接复制示例去调试、联调或继续让 AI 生成下一段请求。
          </DialogDescription>
        </DialogHeader>

        {!context ? (
          <p className="rounded-lg border border-dashed p-4 text-sm text-muted-foreground">
            当前模型的运行时上下文不可用，暂时无法生成 API 文档。
          </p>
        ) : (
          <div className="space-y-4">
            <Section title="Server URL">
              <p className="text-sm text-muted-foreground">
                这是服务地址本身。通常把它配置成环境变量，后面的模型路径再按需拼接。
              </p>
              <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <CodeBlock value={API_DOC_SERVER_URL} className="flex-1" />
                <Button variant="outline" onClick={() => copyText(API_DOC_SERVER_URL)}>
                  复制 Server URL
                </Button>
              </div>
            </Section>

            <Section title="URL 含义">
              <p className="text-sm text-muted-foreground">
                这个完整地址已经固定到当前模型。你发到这个 endpoint 的 GraphQL 请求，只会命中当前模型的运行时接口。
              </p>
              <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <CodeBlock value={endpoint} className="flex-1" />
                <Button variant="outline" onClick={() => copyText(endpoint)}>
                  复制 endpoint
                </Button>
              </div>
              <div className="space-y-2 text-sm text-muted-foreground">
                <p>
                  <span className="font-medium text-foreground">org</span>
                  {' '}
                  表示
                  {' '}
                  <span className="font-medium text-foreground">organization</span>
                  ，也就是当前所属组织。
                </p>
                <p>
                  <span className="font-medium text-foreground">project</span>
                  {' '}
                  表示
                  {' '}
                  <span className="font-medium text-foreground">project</span>
                  ，也就是当前项目。
                </p>
                <p>
                  <span className="font-medium text-foreground">db</span>
                  {' '}
                  表示
                  {' '}
                  <span className="font-medium text-foreground">database</span>
                  ，也就是当前数据库。
                </p>
                <p>
                  <span className="font-medium text-foreground">model</span>
                  {' '}
                  表示
                  {' '}
                  <span className="font-medium text-foreground">model</span>
                  ，也就是当前模型。
                </p>
              </div>
            </Section>

            <Section title="Token 怎么填">
              <p className="text-sm text-muted-foreground">
                把你拿到的 API Token 放进 Bearer 后面，不要保留尖括号占位符。
              </p>
              <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <CodeBlock value={HEADER_EXAMPLE} className="flex-1" />
                <Button variant="outline" onClick={() => copyText(HEADER_EXAMPLE)}>
                  复制 Header
                </Button>
              </div>
            </Section>

            <Section title="curl 示例">
              <p className="text-sm text-muted-foreground">
                下面是一个可直接运行的 findMany 示例，先用它验证链路通不通，再替换查询字段。
              </p>
              <div className="flex flex-col gap-3">
                <CodeBlock value={curlSnippet} />
                <div className="flex justify-end">
                  <Button variant="outline" onClick={() => copyText(curlSnippet)}>
                    复制 curl
                  </Button>
                </div>
              </div>
            </Section>

            <Section title="如何让 AI 帮你继续写">
              <p className="text-sm text-muted-foreground">
                把下面这段提示词发给 AI，再补上你的业务目标，它就能继续帮你生成 GraphQL 和 curl。
              </p>
              <div className="flex flex-col gap-3">
                <CodeBlock value={aiPrompt} className="whitespace-pre-wrap" />
                <div className="flex justify-end">
                  <Button variant="outline" onClick={() => copyText(aiPrompt)}>
                    复制 Prompt
                  </Button>
                </div>
              </div>
            </Section>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
