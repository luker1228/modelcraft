// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { ModelApiDocsDialog } from './ModelApiDocsDialog'
import {
  buildFindManyCurlSnippet,
  buildModelApiAiPrompt,
  buildModelRuntimeEndpoint,
} from './model-api-docs'

const originalClipboard = navigator.clipboard

afterEach(() => {
  cleanup()
  Object.defineProperty(navigator, 'clipboard', {
    configurable: true,
    value: originalClipboard,
  })
  vi.restoreAllMocks()
})

const context = {
  orgName: 'acme',
  projectSlug: 'alpha-project',
  databaseName: 'users_db',
  modelName: 'User',
}

function mockClipboard() {
  const writeText = vi.fn().mockResolvedValue(undefined)

  Object.defineProperty(navigator, 'clipboard', {
    configurable: true,
    value: {
      writeText,
    },
  })

  return { writeText }
}

describe('ModelApiDocsDialog', () => {
  it('renders the five documentation sections for a complete context', () => {
    render(
      <ModelApiDocsDialog
        open
        onOpenChange={vi.fn()}
        context={context}
      />,
    )

    expect(screen.getByRole('heading', { name: 'Server URL' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'URL 含义' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Token 怎么填' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'curl 示例' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: '如何让 AI 帮你继续写' })).toBeInTheDocument()
  })

  it('renders the model-specific endpoint and findMany example', () => {
    const endpoint = buildModelRuntimeEndpoint(context)
    const curlSnippet = buildFindManyCurlSnippet(context)
    const curlLines = curlSnippet.split('\n')

    render(
      <ModelApiDocsDialog
        open
        onOpenChange={vi.fn()}
        context={context}
      />,
    )

    expect(screen.getByText(endpoint)).toBeInTheDocument()
    const urlMeaningSection = screen.getByRole('heading', { name: 'URL 含义' }).closest('section')
    const curlSection = screen.getByRole('heading', { name: 'curl 示例' }).closest('section')

    expect(urlMeaningSection).not.toBeNull()
    expect(urlMeaningSection).toHaveTextContent('org 表示 organization')
    expect(urlMeaningSection).toHaveTextContent('project 表示 project')
    expect(urlMeaningSection).toHaveTextContent('db 表示 database')
    expect(urlMeaningSection).toHaveTextContent('model 表示 model')
    expect(curlSection).not.toBeNull()
    expect(curlSection).toHaveTextContent(curlLines[0])
    expect(curlSection).toHaveTextContent(curlLines[2])
    expect(curlSection).toHaveTextContent('findMany(take: 5, skip: 0)')
  })

  it('copies the curl example when clicking the exact copy button', () => {
    const { writeText } = mockClipboard()

    render(
      <ModelApiDocsDialog
        open
        onOpenChange={vi.fn()}
        context={context}
      />,
    )

    fireEvent.click(screen.getByRole('button', { name: '复制 curl' }))

    expect(writeText).toHaveBeenCalledWith(buildFindManyCurlSnippet(context))
  })

  it('copies the full endpoint when clicking the exact endpoint copy button', () => {
    const { writeText } = mockClipboard()

    render(
      <ModelApiDocsDialog
        open
        onOpenChange={vi.fn()}
        context={context}
      />,
    )

    fireEvent.click(screen.getByRole('button', { name: '复制 endpoint' }))

    expect(writeText).toHaveBeenCalledWith(buildModelRuntimeEndpoint(context))
  })

  it('copies the header example when clicking the header copy button', () => {
    const { writeText } = mockClipboard()

    render(
      <ModelApiDocsDialog
        open
        onOpenChange={vi.fn()}
        context={context}
      />,
    )

    fireEvent.click(screen.getByRole('button', { name: '复制 Header' }))

    expect(writeText).toHaveBeenCalledWith('Authorization: Bearer <API_TOKEN>')
  })

  it('copies the AI prompt when clicking the prompt copy button', () => {
    const { writeText } = mockClipboard()

    render(
      <ModelApiDocsDialog
        open
        onOpenChange={vi.fn()}
        context={context}
      />,
    )

    fireEvent.click(screen.getByRole('button', { name: '复制 Prompt' }))

    expect(writeText).toHaveBeenCalledWith(buildModelApiAiPrompt(context))
  })

  it('renders the fallback message when context is null', () => {
    render(
      <ModelApiDocsDialog
        open
        onOpenChange={vi.fn()}
        context={null}
      />,
    )

    expect(
      screen.getByText('当前模型的运行时上下文不可用，暂时无法生成 API 文档。'),
    ).toBeInTheDocument()
  })
})
