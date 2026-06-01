'use client'

import type { ProposalCandidate } from './types'

interface AiProposalCardProps {
  message: string
  candidates: ProposalCandidate[]
  onSelect: (candidate: ProposalCandidate) => void
}

/**
 * Renders AI navigation proposals as clickable candidate cards.
 * Both action_candidate (execute) and clarification_candidate (chat)
 * are rendered identically — the caller handles the behavioral difference.
 */
export function AiProposalCard({ message, candidates, onSelect }: AiProposalCardProps) {
  if (candidates.length === 0) {
    return (
      <div className="mt-2 rounded-lg border border-border bg-background p-3 text-sm text-muted-foreground">
        未找到相关页面
      </div>
    )
  }

  return (
    <div className="mt-2 flex flex-col gap-2">
      {message && (
        <p className="px-1 text-xs text-muted-foreground">{message}</p>
      )}
      {candidates.map((candidate) => (
        <button
          key={candidate.id}
          type="button"
          onClick={() => onSelect(candidate)}
          className="group w-full rounded-lg border border-border bg-background px-3 py-2.5 text-left transition-colors hover:border-amber-400 hover:bg-amber-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-amber-400"
        >
          <div className="flex items-start justify-between gap-2">
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-1.5">
                <span className="truncate text-sm font-medium text-foreground">
                  {candidate.title}
                </span>
                {candidate.type === 'action_candidate' && candidate.isPrimary && (
                  <span className="shrink-0 rounded-full bg-amber-100 px-1.5 py-0.5 text-[10px] font-medium text-amber-700">
                    推荐
                  </span>
                )}
                {candidate.type === 'action_candidate' && (
                  <span className="shrink-0 rounded-full bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground">
                    {candidate.category ?? 'page'}
                  </span>
                )}
              </div>
              {candidate.description && (
                <p className="mt-0.5 line-clamp-2 text-xs text-muted-foreground">
                  {candidate.description}
                </p>
              )}
            </div>
            <span className="mt-0.5 shrink-0 text-muted-foreground transition-colors group-hover:text-amber-600">
              {candidate.type === 'clarification_candidate' ? '→' : '↗'}
            </span>
          </div>
        </button>
      ))}
    </div>
  )
}
