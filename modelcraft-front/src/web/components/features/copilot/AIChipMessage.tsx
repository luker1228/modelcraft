'use client'

import { memo } from 'react'
import { AssistantMessage } from '@copilotkit/react-ui'
import type { AssistantMessageProps } from '@copilotkit/react-ui'
import { useAICapabilityContext } from '@web/contexts/ai-capability-context'
import { highlightElement } from '@web/lib/highlight-element'

/** A parsed segment of an AI message. */
export type MessageSegment =
  | { type: 'text'; content: string }
  | { type: 'action'; id: string }

/**
 * Parse a message string into text and ACTION marker segments.
 * Pure function — exported for testing.
 */
export function parseActionMarkers(text: string): MessageSegment[] {
  const ACTION_REGEX = /\[ACTION:([^\]]+)\]/g
  const segments: MessageSegment[] = []
  let lastIndex = 0
  let match: RegExpExecArray | null

  while ((match = ACTION_REGEX.exec(text)) !== null) {
    if (match.index > lastIndex) {
      segments.push({ type: 'text', content: text.slice(lastIndex, match.index) })
    }
    segments.push({ type: 'action', id: match[1] })
    lastIndex = match.index + match[0].length
  }

  if (lastIndex < text.length) {
    segments.push({ type: 'text', content: text.slice(lastIndex) })
  }

  return segments
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

function getMessageContent(message: unknown): string {
  if (!isRecord(message)) {
    return ''
  }
  const content = message.content
  return typeof content === 'string' ? content : ''
}

/**
 * Drop-in replacement for CopilotKit's AssistantMessage.
 * Renders [ACTION:id] markers as clickable amber chip buttons.
 * Unknown action IDs (not registered in AICapabilityContext) render as disabled chips.
 *
 * Usage in CopilotProvider:
 *   <CopilotSidebar AssistantMessage={AIChipMessage} ... />
 */
export const AIChipMessage = memo(function AIChipMessage(props: AssistantMessageProps) {
  const { getRef, getAll } = useAICapabilityContext()
  const content = getMessageContent(props.message)

  // Only process messages that contain ACTION markers
  if (!content.includes('[ACTION:')) {
    return <AssistantMessage {...props} />
  }

  const segments = parseActionMarkers(content)
  const capabilityMap = new Map(getAll().map((c) => [c.id, c]))

  // Reconstruct the text-only content for the default renderer (removes ACTION markers)
  const textOnly = segments
    .filter((s): s is { type: 'text'; content: string } => s.type === 'text')
    .map((s) => s.content)
    .join('')

  const handleChipClick = (actionId: string) => {
    const ref = getRef(actionId)
    if (ref) {
      highlightElement(ref)
    }
  }

  return (
    <div>
      {/* Render text segments directly (ACTION markers removed). */}
      {textOnly.trim() && (
        <div className="whitespace-pre-wrap px-3 py-2 text-sm leading-6">{textOnly}</div>
      )}
      {/* Render chip buttons below the text */}
      <div className="mt-2 flex flex-wrap gap-2 px-3 pb-2">
        {segments
          .filter((s): s is { type: 'action'; id: string } => s.type === 'action')
          .map((seg, i) => {
            const known = capabilityMap.has(seg.id)
            const label = capabilityMap.get(seg.id)?.label ?? seg.id
            return (
              <button
                key={`${seg.id}-${i}`}
                type="button"
                disabled={!known}
                onClick={() => handleChipClick(seg.id)}
                title={known ? `高亮 ${label}` : '该操作当前不可用'}
                className={
                  known
                    ? 'inline-flex cursor-pointer items-center gap-1.5 rounded-full border border-amber-300 bg-amber-50 px-3 py-1 text-xs font-medium text-amber-900 transition-colors hover:bg-amber-100'
                    : 'pointer-events-none inline-flex cursor-not-allowed items-center gap-1.5 rounded-full border border-muted bg-muted/50 px-3 py-1 text-xs font-medium text-muted-foreground'
                }
              >
                {known && <span aria-hidden>✨</span>}
                {label}
              </button>
            )
          })}
      </div>
    </div>
  )
})
