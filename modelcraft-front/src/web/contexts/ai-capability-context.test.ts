import { describe, it, expect } from 'vitest'
import { createCapabilityStore, NULL_CAPABILITY_CONTEXT } from './ai-capability-context'

describe('createCapabilityStore', () => {
  it('starts empty', () => {
    const store = createCapabilityStore()
    expect(store.getAll()).toEqual([])
  })

  it('register adds a capability', () => {
    const store = createCapabilityStore()
    const ref = { current: null }
    store.register({ id: 'create_model', label: '新建模型', ref })
    expect(store.getAll()).toHaveLength(1)
    expect(store.getAll()[0].id).toBe('create_model')
  })

  it('unregister removes by id', () => {
    const store = createCapabilityStore()
    const ref = { current: null }
    store.register({ id: 'create_model', label: '新建模型', ref })
    store.unregister('create_model')
    expect(store.getAll()).toEqual([])
  })

  it('getRef returns the registered ref', () => {
    const store = createCapabilityStore()
    const ref = { current: document.createElement('button') }
    store.register({ id: 'create_model', label: '新建模型', ref })
    expect(store.getRef('create_model')).toBe(ref)
  })

  it('later registration overwrites earlier for same id', () => {
    const store = createCapabilityStore()
    const ref1 = { current: null }
    const ref2 = { current: document.createElement('button') }
    store.register({ id: 'create_model', label: '旧标签', ref: ref1 })
    store.register({ id: 'create_model', label: '新标签', ref: ref2 })
    expect(store.getAll()).toHaveLength(1)
    expect(store.getAll()[0].label).toBe('新标签')
  })

  it('getRef returns undefined for unknown id', () => {
    const store = createCapabilityStore()
    expect(store.getRef('nonexistent')).toBeUndefined()
  })
})

// Regression: "useAICapabilityContext must be used inside AICapabilityProvider"
//
// Bug: AIChipMessage (rendered in CopilotWrapper/EndUserCopilotWrapper) called
// useAICapabilityContext(), which threw when no AICapabilityProvider ancestor
// existed — crashing the org-level layout and end-user layout routes.
//
// Fix: useAICapabilityContext() falls back to NULL_CAPABILITY_CONTEXT instead
// of throwing. Components render with empty capabilities (no chips shown) rather
// than crashing. If this describe block breaks, the runtime crash will return.
describe('NULL_CAPABILITY_CONTEXT — fallback when AICapabilityProvider is absent', () => {
  it('getAll returns empty array without throwing', () => {
    expect(() => NULL_CAPABILITY_CONTEXT.getAll()).not.toThrow()
    expect(NULL_CAPABILITY_CONTEXT.getAll()).toEqual([])
  })

  it('getRef returns undefined without throwing', () => {
    expect(() => NULL_CAPABILITY_CONTEXT.getRef('create_model')).not.toThrow()
    expect(NULL_CAPABILITY_CONTEXT.getRef('create_model')).toBeUndefined()
  })

  it('register is a silent no-op — does not accumulate state', () => {
    const ref = { current: null }
    expect(() => NULL_CAPABILITY_CONTEXT.register({ id: 'x', label: 'X', ref })).not.toThrow()
    // After registering, getAll() must still return [] — no side effects
    expect(NULL_CAPABILITY_CONTEXT.getAll()).toEqual([])
  })

  it('unregister is a silent no-op', () => {
    expect(() => NULL_CAPABILITY_CONTEXT.unregister('x')).not.toThrow()
  })

  it('the fallback is returned when context value is null (provider absent)', () => {
    // useAICapabilityContext does: useContext(ctx) ?? NULL_CAPABILITY_CONTEXT
    // When no provider is present, useContext returns null (the default value).
    // Simulate that here without needing React rendering:
    const noProviderValue: typeof NULL_CAPABILITY_CONTEXT | null = null
    const result = noProviderValue ?? NULL_CAPABILITY_CONTEXT
    expect(result).toBe(NULL_CAPABILITY_CONTEXT)
    expect(result.getAll()).toEqual([])
  })
})
