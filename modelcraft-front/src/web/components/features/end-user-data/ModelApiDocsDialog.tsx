import { Button } from '@web/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@web/components/ui/sheet'
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

function FieldChip({
  label,
  value,
}: {
  label: string
  value: string
}) {
  return (
    <div className="rounded-lg border border-border/80 bg-background px-3 py-2.5">
      <p className="text-[11px] font-medium uppercase tracking-[0.08em] text-muted-foreground">
        {label}
      </p>
      <p className="mt-1 font-mono text-xs text-foreground">
        {value}
      </p>
    </div>
  )
}

function Section({
  eyebrow,
  title,
  description,
  action,
  children,
}: {
  eyebrow: string
  title: string
  description: string
  action?: React.ReactNode
  children: React.ReactNode
}) {
  return (
    <section className="space-y-4 rounded-2xl border border-border/80 bg-background p-5 shadow-sm">
      <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
        <div className="space-y-1.5">
          <p className="text-[11px] font-medium uppercase tracking-[0.08em] text-muted-foreground">
            {eyebrow}
          </p>
          <div className="space-y-1">
            <h3 className="text-base font-semibold text-foreground">{title}</h3>
            <p className="max-w-2xl text-sm leading-6 text-muted-foreground">
              {description}
            </p>
          </div>
        </div>
        {action ? <div className="shrink-0">{action}</div> : null}
      </div>
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
      className={`overflow-x-auto rounded-xl border border-slate-800/80 bg-[#1A1F36] p-4 text-sm leading-6 text-slate-100 shadow-sm ${className}`}
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
  const runtimePathSegments = context
    ? [
        { label: 'org', value: context.orgName, meaning: 'organization，也就是当前所属组织。' },
        { label: 'project', value: context.projectSlug, meaning: 'project，也就是当前项目。' },
        { label: 'db', value: context.databaseName, meaning: 'database，也就是当前数据库。' },
        { label: 'model', value: context.modelName, meaning: 'model，也就是当前模型。' },
      ]
    : []

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        side="right"
        className="w-[64vw] min-w-[720px] overflow-y-auto border-l border-border/80 bg-[#F6F8FA] p-0 sm:max-w-4xl lg:max-w-[67vw]"
      >
        <SheetHeader className="sticky top-0 z-10 border-b border-border/80 bg-background/95 px-6 py-5 backdrop-blur">
          <div className="space-y-3 pr-12">
            <div className="flex flex-wrap items-center gap-2">
              <span className="rounded-md bg-primary/[0.08] px-2 py-1 text-[11px] font-medium uppercase tracking-[0.08em] text-primary">
                Runtime API
              </span>
              {context ? (
                <span className="rounded-md border border-border/80 bg-muted/50 px-2 py-1 font-mono text-[11px] text-muted-foreground">
                  {context.modelName}
                </span>
              ) : null}
            </div>
            <div className="space-y-1">
              <SheetTitle>API 文档</SheetTitle>
              <SheetDescription>
                这不是一份通用 GraphQL 教程，而是当前模型的调用参考。先看清地址和认证，再复制最小示例去验证链路。
              </SheetDescription>
            </div>
          </div>
        </SheetHeader>

        {!context ? (
          <div className="px-6 py-6">
            <p className="rounded-2xl border border-dashed border-border/80 bg-background p-5 text-sm text-muted-foreground shadow-sm">
              当前模型的运行时上下文不可用，暂时无法生成 API 文档。
            </p>
          </div>
        ) : (
          <div className="space-y-6 px-6 py-6">
            <section className="rounded-2xl border border-border/80 bg-background p-5 shadow-sm">
              <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
                <FieldChip label="org" value={context.orgName} />
                <FieldChip label="project" value={context.projectSlug} />
                <FieldChip label="db" value={context.databaseName} />
                <FieldChip label="model" value={context.modelName} />
              </div>
            </section>

            <Section
              eyebrow="Connection"
              title="Server URL"
              description="这是服务地址本身。通常把它配置成环境变量，后面的模型路径再按需拼接。"
              action={
                <Button variant="outline" onClick={() => copyText(API_DOC_SERVER_URL)}>
                  复制 Server URL
                </Button>
              }
            >
              <CodeBlock value={API_DOC_SERVER_URL} />
            </Section>

            <Section
              eyebrow="Endpoint"
              title="URL 含义"
              description="这个完整地址已经固定到当前模型。你发到这个 endpoint 的 GraphQL 请求，只会命中当前模型的运行时接口。"
              action={
                <Button variant="outline" onClick={() => copyText(endpoint)}>
                  复制 endpoint
                </Button>
              }
            >
              <CodeBlock value={endpoint} />
              <div className="grid gap-3 xl:grid-cols-2">
                {runtimePathSegments.map((segment) => (
                  <div
                    key={segment.label}
                    className="rounded-xl border border-border/70 bg-muted/40 px-4 py-3"
                  >
                    <p className="font-mono text-xs font-medium text-foreground">
                      {segment.label}
                    </p>
                    <p className="mt-1 text-sm text-muted-foreground">
                      <span className="font-medium text-foreground">{segment.label}</span>
                      {' '}
                      表示
                      {' '}
                      <span className="font-medium text-foreground">{segment.meaning.split('，')[0]}</span>
                      ，{segment.meaning.split('，')[1]}
                    </p>
                  </div>
                ))}
              </div>
            </Section>

            <Section
              eyebrow="Auth"
              title="Token 怎么填"
              description="把你拿到的 API Token 放进 Bearer 后面，不要保留尖括号占位符。"
              action={
                <Button variant="outline" onClick={() => copyText(HEADER_EXAMPLE)}>
                  复制 Header
                </Button>
              }
            >
              <CodeBlock value={HEADER_EXAMPLE} />
            </Section>

            <Section
              eyebrow="Verify"
              title="curl 示例"
              description="下面是一个可直接运行的 findMany 示例，先用它验证链路通不通，再替换查询字段。"
              action={
                <Button variant="outline" onClick={() => copyText(curlSnippet)}>
                    复制 curl
                  </Button>
              }
            >
              <CodeBlock value={curlSnippet} />
            </Section>

            <Section
              eyebrow="Next Step"
              title="如何让 AI 帮你继续写"
              description="把下面这段提示词发给 AI，再补上你的业务目标，它就能继续帮你生成 GraphQL 和 curl。"
              action={
                <Button variant="outline" onClick={() => copyText(aiPrompt)}>
                    复制 Prompt
                  </Button>
              }
            >
              <CodeBlock value={aiPrompt} className="whitespace-pre-wrap" />
            </Section>
          </div>
        )}
      </SheetContent>
    </Sheet>
  )
}
