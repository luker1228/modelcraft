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

describe('NULL_CAPABILITY_CONTEXT', () => {
  it('getAll returns empty array (no crash outside provider)', () => {
    expect(NULL_CAPABILITY_CONTEXT.getAll()).toEqual([])
  })

  it('getRef returns undefined (no crash outside provider)', () => {
    expect(NULL_CAPABILITY_CONTEXT.getRef('create_model')).toBeUndefined()
  })

  it('register and unregister are silent no-ops', () => {
    const ref = { current: null }
    expect(() => NULL_CAPABILITY_CONTEXT.register({ id: 'x', label: 'X', ref })).not.toThrow()
    expect(() => NULL_CAPABILITY_CONTEXT.unregister('x')).not.toThrow()
    expect(NULL_CAPABILITY_CONTEXT.getAll()).toEqual([])
  })
})

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
