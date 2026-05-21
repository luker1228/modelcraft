'use client'

import { createContext, useContext, useState, useCallback, type RefObject } from 'react'

export type AICapability = {
  id: string
  label: string
  ref: RefObject<HTMLElement>
  description?: string
}

interface AICapabilityStore {
  register: (capability: AICapability) => void
  unregister: (id: string) => void
  getAll: () => AICapability[]
  getRef: (id: string) => RefObject<HTMLElement> | undefined
}

/** Pure factory — used both in the React context and in tests. */
export function createCapabilityStore(): AICapabilityStore {
  const map = new Map<string, AICapability>()
  return {
    register: (cap) => map.set(cap.id, cap),
    unregister: (id) => map.delete(id),
    getAll: () => Array.from(map.values()),
    getRef: (id) => map.get(id)?.ref,
  }
}

const AICapabilityContext = createContext<{
  register: (capability: AICapability) => void
  unregister: (id: string) => void
  getAll: () => AICapability[]
  getRef: (id: string) => RefObject<HTMLElement> | undefined
} | null>(null)

export function AICapabilityProvider({ children }: { children: React.ReactNode }) {
  // version counter drives re-renders so useCopilotReadable consumers pick up changes
  const [, setVersion] = useState(0)
  const [store] = useState(() => createCapabilityStore())

  const register = useCallback((cap: AICapability) => {
    store.register(cap)
    setVersion((v) => v + 1)
  }, [store])

  const unregister = useCallback((id: string) => {
    store.unregister(id)
    setVersion((v) => v + 1)
  }, [store])

  return (
    <AICapabilityContext.Provider value={{
      register,
      unregister,
      getAll: store.getAll,
      getRef: store.getRef,
    }}>
      {children}
    </AICapabilityContext.Provider>
  )
}

export function useAICapabilityContext() {
  const ctx = useContext(AICapabilityContext)
  if (!ctx) throw new Error('useAICapabilityContext must be used inside AICapabilityProvider')
  return ctx
}
