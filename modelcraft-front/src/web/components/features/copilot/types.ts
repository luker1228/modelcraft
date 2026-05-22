// modelcraft-front/src/web/components/features/copilot/types.ts

export type AiNavigateArgs = {
  route: string
  params?: Record<string, unknown>
  reason?: string
}

export type AiHighlightArgs = {
  targetId: string
  targetType?: 'field' | 'button' | 'section' | 'tableRow' | 'tab' | 'menu'
  label?: string
  message?: string
  durationMs?: number
  scrollIntoView?: boolean
}

export type AiGuideArgs = {
  route?: string
  targetId?: string
  message?: string
  durationMs?: number
  scrollIntoView?: boolean
}

export type AiAction =
  | { type: 'ui.navigate'; args: AiNavigateArgs }
  | { type: 'ui.highlight'; args: AiHighlightArgs }
  | { type: 'ui.guide'; args: AiGuideArgs }

export type ActionCandidate = {
  id: string
  type: 'action_candidate'
  title: string
  description?: string
  category?: 'page' | 'model' | 'table' | 'field' | 'setting' | 'action'
  confidence?: number
  isPrimary?: boolean
  actions: AiAction[]
}

export type ClarificationCandidate = {
  id: string
  type: 'clarification_candidate'
  title: string
  description?: string
  payload: {
    intent?: string
    entities?: Record<string, unknown>
    userMeaning?: string
  }
}

export type ProposalCandidate = ActionCandidate | ClarificationCandidate

export type AgentUiResponse = {
  kind: 'proposal'
  proposalId: string
  proposalType: 'navigation' | 'highlight' | 'clarification' | 'mixed'
  message: string
  query: string
  candidates: ProposalCandidate[]
}
