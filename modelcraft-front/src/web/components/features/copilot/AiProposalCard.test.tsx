// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest'
import { describe, it, expect, vi, afterEach } from 'vitest'
import { render, screen, fireEvent, cleanup } from '@testing-library/react'
import { AiProposalCard } from './AiProposalCard'
import type { ProposalCandidate } from './types'

const actionCandidate: ProposalCandidate = {
  id: 'go-models',
  type: 'action_candidate',
  title: '数据模型管理',
  description: '进入数据模型编辑器',
  isPrimary: true,
  actions: [{ type: 'ui.navigate', args: { route: '/org/acme/project/main/model-editor' } }],
}

const clarificationCandidate: ProposalCandidate = {
  id: 'intent-config-model',
  type: 'clarification_candidate',
  title: '配置项目模型',
  description: '我想配置字段和权限',
  payload: { intent: 'configure_project_model' },
}

describe('AiProposalCard', () => {
  afterEach(cleanup)
  it('renders message and candidate titles', () => {
    render(
      <AiProposalCard
        message="找到以下结果："
        candidates={[actionCandidate]}
        onSelect={vi.fn()}
      />,
    )
    expect(screen.getByText('找到以下结果：')).toBeInTheDocument()
    expect(screen.getByText('数据模型管理')).toBeInTheDocument()
  })

  it('calls onSelect with correct candidate on click', () => {
    const onSelect = vi.fn()
    render(
      <AiProposalCard
        message="选择："
        candidates={[actionCandidate]}
        onSelect={onSelect}
      />,
    )
    fireEvent.click(screen.getByRole('button', { name: /数据模型管理/ }))
    expect(onSelect).toHaveBeenCalledWith(actionCandidate)
  })

  it('shows 推荐 badge on isPrimary candidate', () => {
    render(
      <AiProposalCard
        message="推荐："
        candidates={[actionCandidate]}
        onSelect={vi.fn()}
      />,
    )
    expect(screen.getByText('推荐')).toBeInTheDocument()
  })

  it('renders clarification_candidate', () => {
    render(
      <AiProposalCard
        message="请选择："
        candidates={[clarificationCandidate]}
        onSelect={vi.fn()}
      />,
    )
    expect(screen.getByText('配置项目模型')).toBeInTheDocument()
  })

  it('renders empty state when no candidates', () => {
    render(
      <AiProposalCard message="未找到" candidates={[]} onSelect={vi.fn()} />,
    )
    expect(screen.getByText('未找到相关页面')).toBeInTheDocument()
  })
})
