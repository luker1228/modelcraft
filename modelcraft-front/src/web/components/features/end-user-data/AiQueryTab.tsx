'use client'

import { useState } from 'react'
import { useCopilotReadable, useCopilotChat } from '@copilotkit/react-core'
import { Button } from '@web/components/ui/button'
import type { FieldDefinition } from '@api-client/cms/public'

interface AiQueryTabProps {
  fields: FieldDefinition[]
  /** Called after AI applies a filter via set_filter action */
  onFilterApplied?: () => void
}

const QUICK_PROMPTS = [
  '名字包含"张"',
  '今天创建的记录',
  '状态为激活',
]

export function AiQueryTab({ fields, onFilterApplied: _onFilterApplied }: AiQueryTabProps) {
  const [input, setInput] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [lastPrompt, setLastPrompt] = useState('')
  const [errorMsg, setErrorMsg] = useState('')

  const fieldSchemaText = fields
    .filter((f) => !f.name.startsWith('_'))
    .map((f) => `${f.name}(${f.storageHint ?? f.schemaType ?? 'STRING'})`)
    .join(', ')

  useCopilotReadable({
    description: '当前模型的字段列表（name:type 格式），供 nl2filter 生成 where 条件使用',
    value: fieldSchemaText,
  })

  const { append } = useCopilotChat() as {
    append: (msg: { role: string; content: string }) => Promise<void>
  }

  async function handleGenerate(prompt: string) {
    if (!prompt.trim() || isLoading) return
    setLastPrompt(prompt)
    setErrorMsg('')
    setIsLoading(true)
    try {
      await append({
        role: 'user',
        content: `请用 set_filter action 为以下条件生成筛选并应用：${prompt}`,
      })
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err)
      setErrorMsg(msg.includes('not found') || msg.includes('unreachable')
        ? 'AI 服务未启动，请先运行 modelcraft-agent'
        : `请求失败：${msg}`)
    } finally {
      setIsLoading(false)
      setInput('')
    }
  }

  return (
    <div className="flex flex-col">
      <div className="flex flex-col gap-3 p-3">
        <p className="text-xs text-muted-foreground">
          用自然语言描述筛选条件，AI 自动生成并应用：
        </p>

        <div className="flex gap-2">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') void handleGenerate(input)
            }}
            placeholder="例：名字包含张且年龄大于18的"
            className="h-8 flex-1 rounded-md border border-input bg-background px-3 text-xs focus:outline-none focus:ring-1 focus:ring-ring"
            disabled={isLoading}
          />
          <Button
            size="sm"
            className="h-8 whitespace-nowrap px-3 text-xs"
            onClick={() => void handleGenerate(input)}
            disabled={!input.trim() || isLoading}
          >
            {isLoading ? '生成中…' : '生成筛选'}
          </Button>
        </div>

        {isLoading && (
          <p className="text-xs text-muted-foreground">
            ✨ 正在生成「{lastPrompt}」的筛选条件…
          </p>
        )}

        {errorMsg && (
          <p className="text-xs text-destructive">{errorMsg}</p>
        )}

        <div className="flex flex-wrap items-center gap-1.5">
          <span className="text-[10px] text-muted-foreground">快捷：</span>
          {QUICK_PROMPTS.map((p) => (
            <button
              key={p}
              type="button"
              onClick={() => void handleGenerate(p)}
              disabled={isLoading}
              className="rounded-full border border-border bg-muted/50 px-2.5 py-0.5 text-[10px] text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50"
            >
              {p}
            </button>
          ))}
        </div>
      </div>

      <div className="flex items-center justify-between border-t border-border bg-muted/40 px-3 py-2">
        <span className="text-[10px] text-muted-foreground">由 modelcraft-agent 驱动</span>
      </div>
    </div>
  )
}
