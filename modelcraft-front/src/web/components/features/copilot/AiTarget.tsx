'use client'

import { useRef, useEffect, type ReactNode } from 'react'
import { useAICapabilityContext } from '@web/contexts/ai-capability-context'

type AiTargetType = 'field' | 'button' | 'section' | 'tableRow' | 'tab' | 'menu'

interface AiTargetProps {
  /** Stable unique identifier. Agent uses this as targetId in ui.highlight. */
  id: string
  /** Human-readable label for this region (shown in AI context). */
  label: string
  /** Optional hint for AI about what this region does. */
  description?: string
  /** Semantic type for AI's understanding of the element. */
  type?: AiTargetType
  children: ReactNode
  className?: string
}

/**
 * Declarative wrapper that registers a UI region as a highlight target.
 *
 * Usage:
 *   <AiTarget id="create-model-btn" label="新建模型按钮" type="button">
 *     <Button>新建模型</Button>
 *   </AiTarget>
 *
 * This adds data-ai-target="create-model-btn" to the wrapper div and
 * registers/unregisters with AICapabilityContext automatically.
 */
export function AiTarget({
  id,
  label,
  description,
  type,
  children,
  className,
}: AiTargetProps) {
  const ref = useRef<HTMLDivElement>(null)
  const { register, unregister } = useAICapabilityContext()

  useEffect(() => {
    register({ id, label, ref, description, type })
    return () => unregister(id)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id, label, description, type])

  return (
    <div ref={ref} data-ai-target={id} className={className}>
      {children}
    </div>
  )
}
