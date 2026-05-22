'use client'

import { useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { useCopilotChat } from '@copilotkit/react-core'
import { TextMessage, MessageRole } from '@copilotkit/runtime-client-gql'
import { useAICapabilityContext } from '@web/contexts/ai-capability-context'
import { highlightHTMLElement } from '@web/lib/highlight-element'
import type { ProposalCandidate, AiAction } from '@web/components/features/copilot/types'

/**
 * Poll until targetId appears in the registry, or timeout expires.
 * Used after navigation to wait for the new page's AiTarget components to mount.
 */
function waitForTarget(
  getRef: (id: string) => React.RefObject<HTMLElement> | undefined,
  targetId: string,
  timeoutMs = 3000,
): Promise<void> {
  return new Promise((resolve) => {
    const deadline = Date.now() + timeoutMs
    const poll = () => {
      if (getRef(targetId)?.current || Date.now() >= deadline) {
        resolve()
      } else {
        requestAnimationFrame(poll)
      }
    }
    requestAnimationFrame(poll)
  })
}

export function useNavigationProposal() {
  const router = useRouter()
  const { getRef } = useAICapabilityContext()
  const { appendMessage } = useCopilotChat()

  const executeActions = useCallback(
    async (actions: AiAction[]) => {
      for (let i = 0; i < actions.length; i += 1) {
        const action = actions[i]

        if (action.type === 'ui.navigate') {
          router.push(action.args.route)

          const nextAction = actions[i + 1]
          if (nextAction && (nextAction.type === 'ui.highlight' || nextAction.type === 'ui.guide') && nextAction.args.targetId) {
            await waitForTarget(getRef, nextAction.args.targetId)
          }
          continue
        }

        if (action.type === 'ui.highlight') {
          const ref = getRef(action.args.targetId)
          const el = ref?.current
          if (!el) {
            console.warn(`[AI] targetId "${action.args.targetId}" not in registry`)
            continue
          }
          highlightHTMLElement(el, {
            message: action.args.message,
            durationMs: action.args.durationMs ?? 5000,
            scrollIntoView: action.args.scrollIntoView ?? true,
          })
          continue
        }

        if (action.type === 'ui.guide') {
          if (action.args.route) {
            router.push(action.args.route)
          }

          if (!action.args.targetId) {
            continue
          }

          const ref = getRef(action.args.targetId)
          const el = ref?.current
          if (!el) {
            console.warn(`[AI] targetId "${action.args.targetId}" not in registry`)
            continue
          }

          highlightHTMLElement(el, {
            message: action.args.message,
            durationMs: action.args.durationMs ?? 5000,
            scrollIntoView: action.args.scrollIntoView ?? true,
          })
        }
      }
    },
    [router, getRef],
  )

  const sendClarificationToAgent = useCallback(
    (candidate: Extract<ProposalCandidate, { type: 'clarification_candidate' }>) => {
      appendMessage(
        new TextMessage({
          role: MessageRole.User,
          content: `我选择了：${candidate.title}\nclarification_payload: ${JSON.stringify(candidate.payload ?? {})}`,
        }),
      )
    },
    [appendMessage],
  )

  const handleCandidateClick = useCallback(
    async (candidate: ProposalCandidate) => {
      if (candidate.type === 'action_candidate') {
        await executeActions(candidate.actions)
        return
      }

      if (candidate.type === 'clarification_candidate') {
        sendClarificationToAgent(candidate)
      }
    },
    [executeActions, sendClarificationToAgent],
  )

  return { handleCandidateClick, executeActions, sendClarificationToAgent }
}
