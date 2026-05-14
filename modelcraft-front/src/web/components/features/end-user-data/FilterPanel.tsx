import React, { useRef } from 'react'
import type { FieldDefinition } from '@api-client/cms/public'
import { WhereJsonEditor, type WhereJsonEditorRef } from './WhereJsonEditor'
import { FieldSchemaPanel } from './FieldSchemaPanel'
import { isValidJson, formatJson } from './filter-utils'

export interface FilterPanelProps {
  /** Field definitions from the current model's jsonSchema (runtimeFields). */
  fields: FieldDefinition[]
  /** Draft JSON string — changes on every keystroke, does NOT trigger a query. */
  whereJsonDraft: string
  /** Called on every keystroke in the editor. */
  onWhereJsonDraftChange: (json: string) => void
  /**
   * Called when the user clicks "应用筛选".
   * The parent is responsible for committing whereJsonDraft → whereJsonCommitted.
   *
   * AI note: to programmatically apply a filter, set whereJsonDraft to a valid
   * where JSON and then call onApply. The parent will commit and trigger the query.
   */
  onApply: () => void
  /**
   * Called when the user clicks "清空". Bypasses the draft/apply flow entirely
   * so the parent can atomically clear both draft and committed state.
   * (Cannot use onApply here because React batches state updates — the draft
   * cleared by setWhereJsonDraft('') would not be visible to onApply yet.)
   */
  onClear: () => void
}

export function FilterPanel({
  fields,
  whereJsonDraft,
  onWhereJsonDraftChange,
  onApply,
  onClear,
}: FilterPanelProps) {
  const editorRef = useRef<WhereJsonEditorRef>(null)

  const valid = isValidJson(whereJsonDraft)

  function handleFormat() {
    onWhereJsonDraftChange(formatJson(whereJsonDraft))
  }

  function handleClear() {
    onClear() // Parent atomically clears draft + committed state
  }

  function handleFieldClick(snippet: string) {
    editorRef.current?.insertAtCursor(snippet)
  }

  return (
    <div className="flex gap-3 border-b border-border bg-muted/30 px-4 py-3">
      <WhereJsonEditor
        ref={editorRef}
        value={whereJsonDraft}
        onChange={onWhereJsonDraftChange}
        onFormat={handleFormat}
        onClear={handleClear}
        onApply={onApply}
        isValid={valid}
      />
      <FieldSchemaPanel fields={fields} onFieldClick={handleFieldClick} />
    </div>
  )
}
