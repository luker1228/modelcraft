import { describe, it, expect, vi } from 'vitest'

// Mock CopilotKit's react-ui — it transitively imports lit-html which calls
// document.createTreeWalker at module-init time, unavailable in the node test env.
// The component-level tests are out of scope; this file only tests parseActionMarkers.
vi.mock('@copilotkit/react-ui', () => ({ AssistantMessage: () => null }))

import { parseActionMarkers, type MessageSegment } from './AIChipMessage'

describe('parseActionMarkers', () => {
  it('returns plain text segment when no markers', () => {
    const result = parseActionMarkers('Hello world')
    expect(result).toEqual([{ type: 'text', content: 'Hello world' }])
  })

  it('extracts a single ACTION marker', () => {
    const result = parseActionMarkers('Click [ACTION:create_model] to start')
    expect(result).toEqual([
      { type: 'text', content: 'Click ' },
      { type: 'action', id: 'create_model' },
      { type: 'text', content: ' to start' },
    ])
  })

  it('extracts multiple ACTION markers', () => {
    const result = parseActionMarkers('[ACTION:create_model] or [ACTION:connect_db]')
    expect(result).toEqual([
      { type: 'action', id: 'create_model' },
      { type: 'text', content: ' or ' },
      { type: 'action', id: 'connect_db' },
    ])
  })

  it('ignores empty text segments', () => {
    const result = parseActionMarkers('[ACTION:create_model]')
    expect(result).toEqual([{ type: 'action', id: 'create_model' }])
  })

  it('handles ACTION at end of string', () => {
    const result = parseActionMarkers('请点击 [ACTION:create_model]')
    expect(result).toEqual([
      { type: 'text', content: '请点击 ' },
      { type: 'action', id: 'create_model' },
    ])
  })

  it('returns empty array for empty input', () => {
    expect(parseActionMarkers('')).toEqual([])
  })
})
