'use client'

import { useEffect, type RefObject } from 'react'
import { useAICapabilityContext } from '@web/contexts/ai-capability-context'

/**
 * Register a UI element's capability with the AI assistant.
 * Automatically unregisters when the component unmounts.
 *
 * @param id       Unique action identifier, e.g. "create_model"
 * @param label    Display label for AI suggestions, e.g. "新建模型"
 * @param ref      Ref to the DOM element that will be highlighted on click
 * @param description  Optional hint for the AI about what this action does
 */
export function useRegisterAICapability(
  id: string,
  label: string,
  ref: RefObject<HTMLElement>,
  description?: string,
) {
  const { register, unregister } = useAICapabilityContext()

  useEffect(() => {
    register({ id, label, ref, description })
    return () => unregister(id)
  // eslint-disable-next-line react-hooks/exhaustive-deps
  // NOTE: callers MUST pass a stable ref (e.g. from useRef). Passing an unstable
  // ref identity will silently miss re-registration.
  }, [id, label, description])
  // Note: `ref` identity is stable (useRef), no need to add it to deps.
  // `register`/`unregister` are stable callbacks from useCallback.
}
